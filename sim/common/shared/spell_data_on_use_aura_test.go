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

const (
	// Eye of Diminution 23001's on-use: 28862, A_MOD_THREAT on the wearer.
	eyeOfDiminution int32 = 28862

	// Fetish of Chitinous Spikes 21488's on-use: 26168, A_DAMAGE_SHIELD of Nature on the wearer.
	chitinousSpikes int32 = 26168

	// Essence of the Pure Flame 18815's equip spell: 23266, A_DAMAGE_SHIELD of Fire on the wearer.
	fieryAura int32 = 23266

	// Cold Basilisk Eye 5079's on-use: 1139, A_MOD_DECREASE_SPEED and A_MOD_MELEE_RANGED_HASTE_2 on the
	// enemy.
	coldEye int32 = 1139

	// Weakness Analyzer 272438's on-use: 1291101, A_MOD_CRIT_PCT on the wearer, with a charge spent by a
	// non-periodic crit.
	analyzingWeaknesses int32 = 1291101

	// Adaptive Combat Assistant 272437's on-use: 1291097, A_SCHOOL_ABSORB of every school, which deals
	// 1291099's Nature to the enemies in melee range when it breaks early.
	damageSigmoid  int32 = 1291097
	sigmoidRevenge int32 = 1291099
)

// A threat reduction is worth nothing the sim scores, so the item is never used on its own: its major
// cooldown files under no type and never asks to be used. Used by hand, it takes the row's share off
// the wearer's threat for the row's duration.
func TestOnUseThreatReductionIsNeverUsedOnItsOwn(t *testing.T) {
	row := spelldata.MustFind(eyeOfDiminution)
	needAmount(t, "28862's threat", row.FirstAura(dbcenums.A_MOD_THREAT).Mean())
	threat := 1 + row.FirstAura(dbcenums.A_MOD_THREAT).Mean()/100
	const itemID int32 = 992701
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(eyeOfDiminution, 120000, 0, 0))
	NewSpellDataAuraOnUse(itemID)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	eye := tester.GetSpell(core.ActionID{ItemID: itemID})
	mcd := tester.GetMajorCooldown(core.ActionID{ItemID: itemID})
	if eye == nil || mcd == nil {
		t.Fatalf("the eye registered no on-use major cooldown")
	}
	if mcd.Type != core.CooldownTypeUnknown || !mcd.DisableAutocast {
		t.Errorf("the eye's major cooldown is of type %v and is autocast: %v; want neither", mcd.Type, !mcd.DisableAutocast)
	}

	if !eye.Cast(sim, sim.Encounter.ActiveTargetUnits[0]) {
		t.Fatalf("the eye could not be used")
	}
	start := sim.CurrentTime
	if got := tester.PseudoStats.ThreatMultiplier; !core.WithinToleranceFloat64(threat, got, 1e-6) {
		t.Errorf("threat multiplier while the eye is up %v, want %v", got, threat)
	}
	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	if got := tester.PseudoStats.ThreatMultiplier; !core.WithinToleranceFloat64(1, got, 1e-6) {
		t.Errorf("threat multiplier after its %v %v, want 1", row.Duration(), got)
	}
}

// The wearer's buff sorts the on-use: a block, dodge or parry chance or a damage shield makes it a
// survival cooldown, spirit alone a mana cooldown, and a crit chance a DPS cooldown. Each row stands in
// for the aura it carries on the wearer.
func TestOnUseAuraCooldownTypeFollowsTheWearersBuff(t *testing.T) {
	for i, tc := range []struct {
		row  int32
		aura dbcenums.EffectAuraType
		want core.CooldownType
	}{
		{1291105, dbcenums.A_MOD_BLOCK_PERCENT, core.CooldownTypeSurvival},
		{1293820, dbcenums.A_MOD_DODGE_PERCENT, core.CooldownTypeSurvival},
		{19263, dbcenums.A_MOD_PARRY_PERCENT, core.CooldownTypeSurvival},
		{chitinousSpikes, dbcenums.A_DAMAGE_SHIELD, core.CooldownTypeSurvival},
		{1291103, dbcenums.A_MOD_STAT, core.CooldownTypeMana},
		{analyzingWeaknesses, dbcenums.A_MOD_CRIT_PCT, core.CooldownTypeDPS},
	} {
		row := spelldata.MustFind(tc.row)
		if !slices.ContainsFunc(spelldata.EffectsOn(row, spelldata.AuraOnWearer), func(e int32) bool { return row.EffectN(int(e)).Aura == tc.aura }) {
			t.Fatalf("%d puts no %v on the wearer", tc.row, tc.aura)
		}

		itemID := 992711 + int32(i)
		withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(tc.row, 120000, 0, 0))
		NewSpellDataAuraOnUse(itemID)

		sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
		if got := auraTesterAt(sim, 0).GetInitialMajorCooldown(core.ActionID{ItemID: itemID}).Type; got != tc.want {
			t.Errorf("%d (%v): the on-use is of type %v, want %v", tc.row, tc.aura, got, tc.want)
		}
	}
}

// What the wearer's damage shield, the spell under its row's id with tag 2, has dealt to the attacker,
// once the handlers the hits queued have run.
func shieldDamage(t *testing.T, sim *core.Simulation, wearer *auraTester, row int32, attacker *auraTester) float64 {
	t.Helper()
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	shield := wearer.GetSpell(core.ActionID{SpellID: row, Tag: 2})
	if shield == nil {
		t.Fatalf("the wearer has no damage shield spell for %d", row)
	}
	return shield.SpellMetrics[attacker.UnitIndex].TotalDamage
}

// A wearer certain to land its spells, and an attacker beside it.
func newShieldSim(items map[proto.ItemSlot]int32) (*core.Simulation, *auraTester, *auraTester) {
	sim := newAuraSim(shaman("Wearer", items), shaman("Attacker", nil))
	wearer, attacker := auraTesterAt(sim, 0), auraTesterAt(sim, 1)
	wearer.AddStatsDynamic(sim, stats.Stats{stats.SpellHitPercent: 100})
	return sim, wearer, attacker
}

// Worn, Essence of the Pure Flame deals its row's Fire back to each melee attacker, and nothing back
// for a spell.
func TestEquipDamageShieldDealsItsAmountBackToAMeleeAttacker(t *testing.T) {
	const itemID int32 = 992702
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket)
	NewSpellDataEquipAura(SpellDataProc{TriggerSpellID: fieryAura},
		[]ItemVariant{{ItemID: itemID, ItemName: "Test Essence of the Pure Flame"}})
	each := spelldata.MustFind(fieryAura).FirstAura(dbcenums.A_DAMAGE_SHIELD).Average()
	needAmount(t, "23266's damage", each)

	sim, wearer, attacker := newShieldSim(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID})

	dealt(sim, attacker.bolt, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, fieryAura, attacker); got != 0 {
		t.Errorf("an arcane bolt on the wearer drew %v back, want nothing", got)
	}
	dealt(sim, attacker.strike, &wearer.Unit)
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, fieryAura, attacker); !core.WithinToleranceFloat64(2*each, got, 1e-6) {
		t.Errorf("two physical strikes on the wearer drew %v back, want %v each", got, each)
	}
}

// Used, Fetish of Chitinous Spikes deals its row's Nature back to each melee attacker for the row's
// duration.
func TestOnUseDamageShieldLastsTheRowsDuration(t *testing.T) {
	const itemID int32 = 992703
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(chitinousSpikes, 180000, 0, 0))
	NewSpellDataAuraOnUse(itemID)
	row := spelldata.MustFind(chitinousSpikes)
	each := row.FirstAura(dbcenums.A_DAMAGE_SHIELD).Average()
	needAmount(t, "26168's damage", each)

	sim, wearer, attacker := newShieldSim(map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID})
	fetish := wearer.GetSpell(core.ActionID{ItemID: itemID})
	if fetish == nil {
		t.Fatalf("the fetish registered no on-use")
	}

	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); got != 0 {
		t.Errorf("a strike before the fetish was used drew %v back, want nothing", got)
	}

	fetish.Cast(sim, sim.Encounter.ActiveTargetUnits[0])
	start := sim.CurrentTime
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); !core.WithinToleranceFloat64(each, got, 1e-6) {
		t.Errorf("a strike while the spikes are up drew %v back, want %v", got, each)
	}

	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	dealt(sim, attacker.strike, &wearer.Unit)
	if got := shieldDamage(t, sim, wearer, chitinousSpikes, attacker); !core.WithinToleranceFloat64(each, got, 1e-6) {
		t.Errorf("a strike after the %v left %v drawn back in all, want the %v from before", row.Duration(), got, each)
	}
}

// Cold Basilisk Eye slows the enemy's attacks by the row's share for the row's duration; its movement
// slow changes nothing the sim runs.
func TestColdBasiliskEyeSlowsTheEnemysAttacks(t *testing.T) {
	const itemID int32 = 992704
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(coldEye, 420000, 0, 0))
	NewSpellDataAuraOnUse(itemID)
	row := spelldata.MustFind(coldEye)
	needAmount(t, "1139's attack slow", row.FirstAura(dbcenums.A_MOD_MELEE_RANGED_HASTE_2).Mean())
	slowed := 1 / (1 + math.Abs(row.FirstAura(dbcenums.A_MOD_MELEE_RANGED_HASTE_2).Mean())/100)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	eye := tester.GetSpell(core.ActionID{ItemID: itemID})
	if eye == nil {
		t.Fatalf("the eye registered no on-use")
	}

	if !eye.Cast(sim, target) {
		t.Fatalf("the eye could not be used")
	}
	start := sim.CurrentTime
	if got := target.PseudoStats.AttackSpeedMultiplier; !core.WithinToleranceFloat64(slowed, got, 1e-6) {
		t.Errorf("the enemy's attack speed multiplier while slowed %v, want %v", got, slowed)
	}
	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	if got := target.PseudoStats.AttackSpeedMultiplier; !core.WithinToleranceFloat64(1, got, 1e-6) {
		t.Errorf("the enemy's attack speed multiplier after the %v %v, want 1", row.Duration(), got)
	}
}

// Weakness Analyzer adds its row's crit until the wearer's first crit, which spends its one charge: a
// hit that does not crit leaves it up.
func TestWeaknessAnalyzerDropsOnTheFirstCrit(t *testing.T) {
	const itemID int32 = 992705
	withAuraItem(itemID, proto.ItemType_ItemTypeTrinket, testOnUse(analyzingWeaknesses, 90000, 1141, 20000))
	NewSpellDataAuraOnUse(itemID)

	sim := newAuraSim(shaman("Shaman", map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotTrinket1: itemID}))
	tester := auraTesterAt(sim, 0)
	target := sim.Encounter.ActiveTargetUnits[0]
	analyzer := tester.GetSpell(core.ActionID{ItemID: itemID})
	if analyzer == nil {
		t.Fatalf("the analyzer registered no on-use")
	}
	crit := tester.GetStat(stats.PhysicalCritPercent)
	bonus := spelldata.MustFind(analyzingWeaknesses).FirstAura(dbcenums.A_MOD_CRIT_PCT).Mean()
	needAmount(t, "1291101's crit", bonus)

	analyzer.Cast(sim, target)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit+bonus, got, 1e-6) {
		t.Fatalf("crit while the analyzer is up %v, want %v", got, crit+bonus)
	}

	tester.OnSpellHitDealt(sim, tester.strike, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit+bonus, got, 1e-6) {
		t.Errorf("crit after a hit that did not crit %v, want %v", got, crit+bonus)
	}

	tester.OnSpellHitDealt(sim, tester.bolt, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit, Damage: 150})
	stepPast(t, sim, sim.CurrentTime+core.SpellBatchWindow+time.Millisecond)
	if got := tester.GetStat(stats.PhysicalCritPercent); !core.WithinToleranceFloat64(crit, got, 1e-6) {
		t.Errorf("crit after the first crit %v, want %v", got, crit)
	}
}

// Adaptive Combat Assistant's shield deals its revenge spell's Nature to the enemy when a hit breaks it
// before its duration ends, and nothing when it runs out unbroken.
func TestAdaptiveCombatAssistantDealsDamageWhenItsShieldBreaks(t *testing.T) {
	const itemID int32 = 992706
	row := spelldata.MustFind(damageSigmoid)
	shield := row.AbsorbEffect()
	strike := spelldata.MustFind(sigmoidRevenge).DamageEffect()
	needAmount(t, "1291097's shield", shield.Min())
	needAmount(t, "1291099's damage", strike.Min())
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse,
		map[int32]*proto.ItemEffect{itemID: testOnUse(damageSigmoid, 90000, 1141, 15000)})
	assistant := onUseSpell(t, caster, itemID)
	revenge := caster.GetSpell(core.ActionID{SpellID: sigmoidRevenge})
	if revenge == nil {
		t.Fatalf("the assistant registered no %d", sigmoidRevenge)
	}
	target := sim.Encounter.ActiveTargetUnits[0]
	revengeDealt := func() float64 { return revenge.SpellMetrics[target.UnitIndex].TotalDamage }

	assistant.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, shield.Min()/2)
	stepPast(t, sim, start+row.Duration()+time.Millisecond)
	if got := revengeDealt(); got != 0 {
		t.Errorf("a shield that ran out unbroken dealt %v, want nothing", got)
	}

	assistant.CD.Reset()
	assistant.SharedCD.Reset()
	assistant.Cast(sim, &caster.Unit)
	hit := shield.Max() + 100
	if got, low, high := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, hit), hit-shield.Max(), hit-shield.Min(); got < low || got > high {
		t.Errorf("a fire hit of %v landed for %v, want %v to %v past the shield", hit, got, low, high)
	}
	stepPast(t, sim, sim.CurrentTime+time.Second)
	if got, low, high := revengeDealt(), strike.Min(), strike.Max(); got < low || got > high {
		t.Errorf("breaking the shield dealt %v, want %v to %v", got, low, high)
	}
}
