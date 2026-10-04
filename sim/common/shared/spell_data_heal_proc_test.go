package shared

import (
	"math"
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Recovery 8721's rows: the equip aura 1248761 and the 5% E_HEAL_PCT heal 1248759 it casts. The aura
// carries the hint its grant 1248760 states, "when you are Parried or Dodged".
const (
	recoveryTrigger int32 = 1248761
	recoveryHeal    int32 = 1248759
)

// A caster wearing a test weapon enchanted with a heal proc on the given rows, at half health so a
// heal shows on the health bar, and with the spell crit given.
func newHealProcSim(t *testing.T, weaponID, enchantID int32, healSpellID int32, spellCrit float64) (*core.Simulation, *testCaster, *core.Spell) {
	t.Helper()
	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{testOneHander(weaponID)},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Heal Enchant",
			Type: proto.ItemType_ItemTypeWeapon}},
	})
	registerSpellDataHealProc(SpellDataProc{Name: "Test Heal Enchant", EnchantID: enchantID,
		TriggerSpellID: recoveryTrigger, BuffSpellID: healSpellID})

	sim := newTestCasterSim(testHands(&proto.ItemSpec{Id: weaponID, Enchant: enchantID}, &proto.ItemSpec{}), nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	caster.AddStatsDynamic(sim, stats.Stats{stats.SpellCritPercent: spellCrit - caster.GetStat(stats.SpellCritPercent)})
	caster.RemoveHealth(sim, caster.MaxHealth()/2)

	heal := caster.GetSpell(core.ActionID{SpellID: healSpellID})
	if heal == nil {
		t.Fatalf("the heal %d the proc casts was not registered", healSpellID)
	}
	return sim, caster, heal
}

// What the caster heals for when the target meets its main-hand swing with the outcome.
func swingHealed(sim *core.Simulation, caster *testCaster, outcome core.HitOutcome, damage float64) float64 {
	before := caster.CurrentHealth()
	caster.OnSpellHitDealt(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto},
		&core.SpellResult{Target: sim.Encounter.ActiveTargetUnits[0], Outcome: outcome, Damage: damage})
	return caster.CurrentHealth() - before
}

func stepPast(t *testing.T, sim *core.Simulation, until time.Duration) {
	t.Helper()
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) {}}))
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// Recovery heals the wearer for 5% of its maximum health when the target dodges or parries the
// wearer's melee attack. A landed swing does not proc it, nor does the wearer dodging an attack it
// takes, and a second avoided swing inside the 10 s ProcCategoryRecovery heals nothing.
func TestRecoveryHealsWhenTheWearersAttackIsDodgedOrParried(t *testing.T) {
	sim, caster, heal := newHealProcSim(t, 990971, 990972, recoveryHeal, 0)
	want := 0.05 * caster.MaxHealth()

	if got := swingHealed(sim, caster, core.OutcomeHit, 100); got != 0 {
		t.Errorf("a landed swing healed %v, want nothing", got)
	}

	before := caster.CurrentHealth()
	enemy := sim.Encounter.ActiveTargetUnits[0]
	caster.OnSpellHitTaken(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, Unit: enemy},
		&core.SpellResult{Target: &caster.Unit, Outcome: core.OutcomeDodge})
	if got := caster.CurrentHealth() - before; got != 0 {
		t.Errorf("the wearer dodging an attack it takes healed %v, want nothing", got)
	}

	if got := swingHealed(sim, caster, core.OutcomeDodge, 0); math.Abs(got-want) > 1e-6 {
		t.Fatalf("a dodged swing healed %v, want 5%% of %v, %v", got, caster.MaxHealth(), want)
	}
	procTime := sim.CurrentTime

	if got := swingHealed(sim, caster, core.OutcomeParry, 0); got != 0 {
		t.Errorf("a parried swing inside the 10 s lockout healed %v, want nothing", got)
	}

	stepPast(t, sim, procTime+10*time.Second)
	if got := swingHealed(sim, caster, core.OutcomeParry, 0); math.Abs(got-want) > 1e-6 {
		t.Errorf("a parried swing after the lockout healed %v, want %v", got, want)
	}

	if got := heal.SpellMetrics[caster.UnitIndex].TotalHealing; math.Abs(got-2*want) > 1e-6 {
		t.Errorf("healing metrics = %v, want the two heals' %v", got, 2*want)
	}
}

// A proc's heal crits at the caster's spell crit unless its row carries ATTR_EX_2_CANT_CRIT. 1248759
// is not flagged, so a flagged copy stands in for the rows that are.
func TestProcHealCritsUnlessTheRowRulesItOut(t *testing.T) {
	sim, caster, heal := newHealProcSim(t, 990973, 990974, recoveryHeal, 100)
	want := 0.05 * caster.MaxHealth() * heal.CritDamageMultiplier(nil)
	if got := swingHealed(sim, caster, core.OutcomeDodge, 0); math.Abs(got-want) > 1e-6 {
		t.Errorf("at 100%% spell crit the heal was %v, want the critical %v", got, want)
	}

	editRow(t, recoveryHeal, func(s *spelldata.Spell) { s.Attr[dbcenums.ATTR_INDEX_EX_2] |= dbcenums.ATTR_EX_2_CANT_CRIT })

	sim, caster, _ = newHealProcSim(t, 990975, 990976, recoveryHeal, 100)
	want = 0.05 * caster.MaxHealth()
	if got := swingHealed(sim, caster, core.OutcomeDodge, 0); math.Abs(got-want) > 1e-6 {
		t.Errorf("a heal whose row cannot crit healed %v at 100%% spell crit, want %v", got, want)
	}
}

// Darkmoon Card: Heroism's 23682 states its heal as E_HEAL 150 with a 0.4 spread, which the proc
// rolls rather than reading as a share of maximum health.
func TestProcHealRollsAStatedAmount(t *testing.T) {
	const fixedHeal int32 = 23682
	effect := spelldata.MustFind(fixedHeal).ProcHealEffect()
	if effect.Type != dbcenums.E_HEAL {
		t.Fatalf("%d heals through %v, want E_HEAL", fixedHeal, effect.Type)
	}

	sim, caster, _ := newHealProcSim(t, 990977, 990978, fixedHeal, 0)
	low, high := effect.Min(), effect.Max()
	if got := swingHealed(sim, caster, core.OutcomeDodge, 0); got < low || got > high {
		t.Errorf("the heal was %v, want %v to %v", got, low, high)
	}
}

// Julie's Blessing 8348, Julie's Dagger 6660's heal: A_PERIODIC_HEAL 13 on the caster every 2 s for 12 s.
const julieHot int32 = 8348

// Edits the store's row of a spell for the test, restoring it afterwards. The effects are copied so
// an edit to one does not reach the original.
func editRow(t *testing.T, id int32, edit func(*spelldata.Spell)) {
	t.Helper()
	row := spelldata.MustFind(id)
	original := *row
	row.Effects = slices.Clone(row.Effects)
	edit(row)
	t.Cleanup(func() { *row = original })
}

// A trigger whose row states no rate, rolled at 100% for the test.
func everyHit(t *testing.T, id int32) {
	editRow(t, id, func(s *spelldata.Spell) {
		s.ProcChance, s.ProcChanceSource, s.ProcChanceEffect = 100, spelldata.ProcChanceColumn, 0
	})
}

// What the caster heals for between now and the given time after start, stepping just past it so a
// tick landing on it is counted.
func healedUntil(t *testing.T, sim *core.Simulation, caster *testCaster, start, after time.Duration) float64 {
	t.Helper()
	before := caster.CurrentHealth()
	stepPast(t, sim, start+after+time.Millisecond)
	return caster.CurrentHealth() - before
}

// The HoT lands nothing when it is applied, then heals the wearer 13 every 2 s until its 12 s run out:
// six ticks, 78 in all, measured as the spell's periodic healing.
func TestProcHotTicksItsAmountEveryPeriodOnTheWearer(t *testing.T) {
	sim, caster, hot := newHealProcSim(t, 990979, 990980, julieHot, 0)

	if got := swingHealed(sim, caster, core.OutcomeDodge, 0); got != 0 {
		t.Errorf("applying the HoT healed %v, want nothing until the first tick", got)
	}
	start := sim.CurrentTime
	if !hot.SelfHot().IsActive() {
		t.Fatalf("the proc did not apply the HoT to the wearer")
	}

	for tick := 1; tick <= 6; tick++ {
		if got := healedUntil(t, sim, caster, start, time.Duration(tick)*2*time.Second); got != 13 {
			t.Errorf("tick %d healed %v, want 13", tick, got)
		}
	}
	if hot.SelfHot().IsActive() {
		t.Errorf("the HoT is still up after its 12 s")
	}
	if got := healedUntil(t, sim, caster, start, 14*time.Second); got != 0 {
		t.Errorf("the HoT healed %v after it ran out, want nothing", got)
	}

	metrics := hot.SpellMetrics[caster.UnitIndex]
	if metrics.Ticks != 6 || metrics.CritTicks != 0 || metrics.TotalHealing != 78 {
		t.Errorf("metrics = %d ticks, %d crit ticks, %v healing; want 6, 0, 78",
			metrics.Ticks, metrics.CritTicks, metrics.TotalHealing)
	}
}

// Each tick reads the healing power the wearer has when it lands. 8348 states no spell power share,
// so a copy stating 1 stands in for the rows that do.
func TestProcHotTicksOnTheHealingPowerOfTheTick(t *testing.T) {
	editRow(t, julieHot, func(s *spelldata.Spell) { s.Effects[0].SPCoef = 1 })
	sim, caster, _ := newHealProcSim(t, 990981, 990982, julieHot, 0)
	caster.AddStatsDynamic(sim, stats.Stats{stats.HealingPower: -caster.GetStat(stats.HealingPower)})

	swingHealed(sim, caster, core.OutcomeDodge, 0)
	start := sim.CurrentTime
	if got := healedUntil(t, sim, caster, start, 2*time.Second); got != 13 {
		t.Errorf("the first tick at no healing power healed %v, want 13", got)
	}

	caster.AddStatsDynamic(sim, stats.Stats{stats.HealingPower: 100})
	if got := healedUntil(t, sim, caster, start, 4*time.Second); got != 113 {
		t.Errorf("the tick after gaining 100 healing power healed %v, want 113", got)
	}
}

// A tick crits only where the row states Periodic Can Crit and does not rule crits out. 8348 states
// the flag; a copy without it ticks plain at 100% spell crit.
func TestProcHotCritsOnlyWhereTheRowLetsItsTicksCrit(t *testing.T) {
	editRow(t, julieHot, func(s *spelldata.Spell) {
		s.Attr[dbcenums.ATTR_INDEX_EX_8] &^= dbcenums.ATTR_EX_8_PERIODIC_CAN_CRIT
	})
	sim, caster, hot := newHealProcSim(t, 990983, 990984, julieHot, 100)
	swingHealed(sim, caster, core.OutcomeDodge, 0)
	if got := healedUntil(t, sim, caster, sim.CurrentTime, 2*time.Second); got != 13 {
		t.Errorf("an unflagged tick at 100%% spell crit healed %v, want 13", got)
	}

	editRow(t, julieHot, func(s *spelldata.Spell) {
		s.Attr[dbcenums.ATTR_INDEX_EX_8] |= dbcenums.ATTR_EX_8_PERIODIC_CAN_CRIT
	})
	sim, caster, hot = newHealProcSim(t, 990985, 990986, julieHot, 100)
	want := 13 * hot.CritDamageMultiplier(nil)
	swingHealed(sim, caster, core.OutcomeDodge, 0)
	if got := healedUntil(t, sim, caster, sim.CurrentTime, 2*time.Second); math.Abs(got-want) > 1e-6 {
		t.Errorf("a tick that can crit healed %v at 100%% spell crit, want the critical %v", got, want)
	}
	if got := hot.SpellMetrics[caster.UnitIndex].CritTicks; got != 1 {
		t.Errorf("crit ticks = %d, want 1", got)
	}

	editRow(t, julieHot, func(s *spelldata.Spell) {
		s.Attr[dbcenums.ATTR_INDEX_EX_2] |= dbcenums.ATTR_EX_2_CANT_CRIT
	})
	sim, caster, _ = newHealProcSim(t, 990987, 990988, julieHot, 100)
	swingHealed(sim, caster, core.OutcomeDodge, 0)
	if got := healedUntil(t, sim, caster, sim.CurrentTime, 2*time.Second); got != 13 {
		t.Errorf("a tick whose row cannot crit healed %v at 100%% spell crit, want 13", got)
	}
}

// A second proc while the HoT runs starts it over: the ticks already landed stay landed and a full
// 12 s of ticks follows the second proc. 1248761's 10 s lockout is the earliest it can come.
func TestProcHotStartsOverOnASecondProc(t *testing.T) {
	sim, caster, hot := newHealProcSim(t, 990989, 990990, julieHot, 0)
	swingHealed(sim, caster, core.OutcomeDodge, 0)
	start := sim.CurrentTime

	if got := healedUntil(t, sim, caster, start, 10*time.Second); got != 5*13 {
		t.Fatalf("the first 10 s healed %v, want five ticks, %v", got, 5*13)
	}
	swingHealed(sim, caster, core.OutcomeDodge, 0)
	restart := sim.CurrentTime

	if got := healedUntil(t, sim, caster, start, 13*time.Second); got != 13 || !hot.SelfHot().IsActive() {
		t.Errorf("past the first application's 12 s the HoT healed %v and is up: %v; want 13 and up",
			got, hot.SelfHot().IsActive())
	}
	if got := healedUntil(t, sim, caster, restart, 12*time.Second); got != 5*13 {
		t.Errorf("the rest of the second application healed %v, want %v", got, 5*13)
	}
	if hot.SelfHot().IsActive() {
		t.Errorf("the HoT is still up 12 s after the second proc")
	}
	if got := hot.SpellMetrics[caster.UnitIndex].Ticks; got != 11 {
		t.Errorf("ticks = %d, want 5 before the second proc and 6 after", got)
	}
}
