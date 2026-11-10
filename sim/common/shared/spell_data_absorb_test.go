package shared

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Mark of Resolution 17759's on-use spell 21956, A_SCHOOL_ABSORB with Misc 1 (Physical) on the caster.
// The test trinkets carrying it run on a cooldown and in a category of the tests' own.
const (
	resolutionSpell    int32 = 21956
	resolutionCategory int32 = 2554
)

func resolutionOnUse() *proto.ItemEffect {
	return testOnUse(resolutionSpell, 300000, resolutionCategory, 10000)
}

// 21956 states a spread around its amount. The tests that count the shield down read the amount alone.
func withoutSpread(t *testing.T, spellID int32) {
	t.Helper()
	editRow(t, spellID, func(s *spelldata.Spell) { s.Effects[0].Variance = 0 })
}

// The amount the row's absorb effect shields for, once its spread is gone, after checking the effect
// absorbs Physical and not Fire, which the tests below hit it with.
func physicalAbsorb(t *testing.T, spellID int32) float64 {
	t.Helper()
	withoutSpread(t, spellID)
	effect := spelldata.MustFind(spellID).AbsorbEffect()
	schools := core.SpellSchool(effect.Misc)
	if !schools.Matches(core.SpellSchoolPhysical) || schools.Matches(core.SpellSchoolFire) || effect.Average() <= 0 {
		t.Fatalf("%d absorbs %v of %v, want an amount of Physical alone", spellID, effect.Average(), schools)
	}
	return effect.Average()
}

// What is left of a hit of the school once the wearer's absorbs have taken their share.
func takenAfterAbsorbs(sim *core.Simulation, caster *testCaster, school core.SpellSchool, damage float64) float64 {
	enemy := sim.Encounter.ActiveTargetUnits[0]
	result := &core.SpellResult{Target: &caster.Unit, Damage: damage}
	(&core.Spell{SpellSchool: school, Unit: enemy}).ApplyPostOutcomeDamageModifiers(sim, result, false)
	return result.Damage
}

func absorbAura(t *testing.T, caster *testCaster, row int32, sourceID int32) *core.Aura {
	t.Helper()
	label := fmt.Sprintf("%s %d", spelldata.MustFind(row).Name, sourceID)
	aura := caster.GetAura(label)
	if aura == nil {
		t.Fatalf("no absorb aura %q is registered", label)
	}
	return aura
}

// The shield takes the row's amount off the physical damage the wearer takes: a hit of twice that
// lands for the other half, and the shield it used up is gone. A fire hit before it is not absorbed
// and leaves the shield whole.
func TestOnUseAbsorbTakesPhysicalDamageOffItsShield(t *testing.T) {
	const itemID int32 = 991201
	amount := physicalAbsorb(t, resolutionSpell)
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: resolutionOnUse()})
	spell := onUseSpell(t, caster, itemID)
	shield := absorbAura(t, caster, resolutionSpell, itemID)

	if got := caster.GetInitialMajorCooldown(spell.ActionID); !got.Type.Matches(core.CooldownTypeSurvival) || !spell.Flags.Matches(core.SpellFlagHelpful) {
		t.Errorf("the absorb is a major cooldown of type %v, helpful %v; want a survival one cast on the wearer",
			got.Type, spell.Flags.Matches(core.SpellFlagHelpful))
	}

	spell.Cast(sim, &caster.Unit)
	if !shield.IsActive() {
		t.Fatalf("using the item put up no shield")
	}

	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, 2*amount); got != 2*amount {
		t.Errorf("a fire hit of %v landed for %v, want all of it", 2*amount, got)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 2*amount); got != amount {
		t.Errorf("a physical hit of %v landed for %v, want %v past the row's %v shield", 2*amount, got, amount, amount)
	}
	if shield.IsActive() {
		t.Errorf("the shield is still up after absorbing its %v", amount)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100); got != 100 {
		t.Errorf("a physical hit after the shield broke landed for %v, want all 100", got)
	}
}

// Unused, the shield drops at the end of the row's duration and absorbs nothing after.
func TestOnUseAbsorbExpiresWithItsRow(t *testing.T) {
	const itemID int32 = 991202
	physicalAbsorb(t, resolutionSpell)
	duration := spelldata.MustFind(resolutionSpell).Duration()
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: resolutionOnUse()})
	shield := absorbAura(t, caster, resolutionSpell, itemID)

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	start := sim.CurrentTime

	stepPast(t, sim, start+duration-time.Millisecond)
	if !shield.IsActive() {
		t.Fatalf("the shield dropped before the row's %v", duration)
	}
	stepPast(t, sim, start+duration+time.Millisecond)
	if shield.IsActive() {
		t.Errorf("the shield is still up after the row's %v", duration)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100); got != 100 {
		t.Errorf("a physical hit after the shield expired landed for %v, want all 100", got)
	}
}

// The item's own cooldown, and its category's that a second item in the category waits out, as the
// on-use effects state them.
func TestOnUseAbsorbRunsOnTheItemsCooldownAndCategory(t *testing.T) {
	const markID, otherID int32 = 991203, 991204
	markEffect := resolutionOnUse()
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{
		markID:  markEffect,
		otherID: testOnUse(resolutionSpell, 60000, resolutionCategory, markEffect.GetOnUse().CategoryCooldownMs),
	})
	mark, other := onUseSpell(t, caster, markID), onUseSpell(t, caster, otherID)
	cooldown := core.DurationFromMillis(markEffect.GetOnUse().CooldownMs)
	categoryCooldown := core.DurationFromMillis(markEffect.GetOnUse().CategoryCooldownMs)

	mark.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	if offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could be used while the mark's category cooldown runs", resolutionCategory)
	}

	stepPast(t, sim, start+categoryCooldown)
	if !offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could not be used once its %v ran out", resolutionCategory, categoryCooldown)
	}
	if got := mark.CD.ReadyAt(); got != start+cooldown {
		t.Errorf("the mark is ready again at %v, want %v after its use at %v", got, cooldown, start)
	}
}

// With the row's spread the shield rolls within its effect's ends.
func TestOnUseAbsorbRollsTheRowsSpread(t *testing.T) {
	const itemID int32 = 991205
	effect := spelldata.MustFind(resolutionSpell).AbsorbEffect()
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: resolutionOnUse()})
	low, high := effect.Min(), effect.Max()
	if low == high {
		t.Fatalf("%d states no spread", resolutionSpell)
	}

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	if absorbed := 2*high - takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 2*high); absorbed < low || absorbed > high {
		t.Errorf("the shield absorbed %v, want %v to %v", absorbed, low, high)
	}
}

// Blazing Emblem 2802's on-use casts 13744, A_SCHOOL_ABSORB of Fire beside A_MOD_RESISTANCE of Fire for
// the row's duration; its effect entry carries the resistance as a stat. The resistance is up with the
// shield and stays while a hit eats part of it; the hit that uses the shield up removes it, resistance
// and all.
func TestOnUseAbsorbCarriesTheStatsItsEffectStates(t *testing.T) {
	const itemID, emblemSpell int32 = 991208, 13744
	if spelldata.Find(emblemSpell) == spelldata.Nil {
		t.Skip("spell 13744 is not in the live game files; this runs again once they hold it")
	}
	withoutSpread(t, emblemSpell)
	row := spelldata.MustFind(emblemSpell)
	amount, resistance := row.AbsorbEffect().Average(), row.FirstAura(dbcenums.A_MOD_RESISTANCE).Average()
	if amount <= 0 || resistance <= 0 {
		t.Fatalf("%d states %v absorbed and %v resistance, want both", emblemSpell, amount, resistance)
	}
	emblem := testOnUse(emblemSpell, 600000, 0, 0)
	emblem.EffectDurationMs = row.DurationMs
	emblem.ScalingOptions = map[int32]*proto.ScalingItemEffectProperties{
		0: {Stats: map[int32]float64{int32(proto.Stat_StatFireResistance): resistance}},
	}
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: emblem})
	shield := absorbAura(t, caster, emblemSpell, itemID)
	fireResistance := func() float64 { return caster.GetStat(stats.FireResistance) }
	base := fireResistance()

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	if got := fireResistance() - base; !shield.IsActive() || got != resistance {
		t.Fatalf("the shield is up %v with %v Fire resistance, want up with the entry's %v", shield.IsActive(), got, resistance)
	}

	first := amount * 3 / 4
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, first); got != 0 {
		t.Errorf("a fire hit of %v landed for %v, want none of it", first, got)
	}
	if got := fireResistance() - base; got != resistance {
		t.Errorf("with %v of the shield left the wearer has %v Fire resistance, want %v", amount-first, got, resistance)
	}

	second := amount - first + 100
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, second); math.Abs(got-100) > 1e-9 {
		t.Errorf("a fire hit of %v landed for %v, want 100 past the %v left", second, got, amount-first)
	}
	if got := fireResistance() - base; shield.IsActive() || got != 0 {
		t.Errorf("the used-up shield is up %v with %v Fire resistance, want gone with none", shield.IsActive(), got)
	}
}

// Enchant Chest - Absorption 8220: its equip aura 1249072 rolls a chance on a melee hit taken, at most
// once per its internal cooldown, and casts 1249073, A_SCHOOL_ABSORB of Physical until used up. A copy
// of the trigger rolling 100% stands in, so every hit taken outside the lockout procs.
func TestAbsorbProcShieldsTheWearerOnAMeleeHitTaken(t *testing.T) {
	const chestID, enchantID int32 = 991206, 991207
	const trigger, absorb int32 = 1249072, 1249073
	editRow(t, trigger, func(s *spelldata.Spell) { s.ProcChance = 100 })
	amount := physicalAbsorb(t, absorb)
	lockout := spelldata.MustFind(trigger).ICD()
	if lockout <= 0 {
		t.Fatalf("%d states no internal cooldown", trigger)
	}

	core.AddToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{{Id: chestID, Name: "Test Chest", Type: proto.ItemType_ItemTypeChest,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}},
		Enchants: []*proto.SimEnchant{{EffectId: enchantID, Name: "Test Absorb Enchant", Type: proto.ItemType_ItemTypeChest}},
	})
	registerSpellDataAbsorbProc(SpellDataProc{Name: "Test Absorb Enchant", EnchantID: enchantID,
		TriggerSpellID: trigger, BuffSpellID: absorb})

	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotChest+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotChest] = &proto.ItemSpec{Id: chestID, Enchant: enchantID}
	sim := newTestCasterSim(items, nil)
	caster := sim.Raid.Parties[0].Players[0].(*testCaster)
	shield := absorbAura(t, caster, absorb, enchantID)
	enemy := sim.Encounter.ActiveTargetUnits[0]

	struck := func() {
		caster.OnSpellHitTaken(sim, &core.Spell{ProcMask: core.ProcMaskMeleeMHAuto, SpellSchool: core.SpellSchoolPhysical, Unit: enemy},
			&core.SpellResult{Target: &caster.Unit, Outcome: core.OutcomeHit, Damage: 100})
	}

	struck()
	if !shield.IsActive() {
		t.Fatalf("a melee hit taken put up no shield")
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, 2*amount); got != 2*amount {
		t.Errorf("a fire hit of %v landed for %v, want all of it", 2*amount, got)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 2*amount); got != amount {
		t.Errorf("a physical hit of %v landed for %v, want %v past the row's shield", 2*amount, got, amount)
	}
	procTime := sim.CurrentTime

	struck()
	if shield.IsActive() {
		t.Errorf("a hit inside the row's %v lockout put the shield back up", lockout)
	}

	stepPast(t, sim, procTime+lockout)
	struck()
	if !shield.IsActive() {
		t.Errorf("a hit after the lockout put up no shield")
	}
}

// Twin Blades of the Hakkari's equip aura 468256 states one chance on an effect and another in its
// ProcChance column. A test item's proc on it rolls the column; a class spell stating both, Hack and
// Slash 13960, is not an item proc and keeps the effect's.
func TestItemProcRollsTheColumnOverAnEffectsChance(t *testing.T) {
	_, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, nil)
	column := func(s *spelldata.Spell) float64 { return float64(s.ProcChance) / 100 }

	const itemID int32 = 992801
	twinBlades := spelldata.MustFind(468256)
	if twinBlades.StatedChance() == column(twinBlades) || twinBlades.StatedChance() <= 0 {
		t.Fatalf("468256 states %v on its effect and %v in its column, want two chances", twinBlades.StatedChance(), column(twinBlades))
	}
	config := spellDataProcListener(&caster.Character, SpellDataProc{Name: "Twin Blades of the Hakkari", ItemID: itemID, TriggerSpellID: 468256},
		effectSource{id: itemID}, twinBlades)
	if config.ProcChance != column(twinBlades) || config.DPM != nil {
		t.Errorf("the proc rolls %v (manager %v), want the column's %v", config.ProcChance, config.DPM != nil, column(twinBlades))
	}

	hackAndSlash := spelldata.MustFind(13960)
	if hackAndSlash.StatedChance() == column(hackAndSlash) || hackAndSlash.StatedChance() <= 0 {
		t.Fatalf("13960 states %v on its effect and %v in its column, want two chances", hackAndSlash.StatedChance(), column(hackAndSlash))
	}
	if got := spelldata.ProcTrigger(&caster.Character, hackAndSlash, nil).ProcChance; got != hackAndSlash.StatedChance() {
		t.Errorf("a class spell's trigger rolls %v, want its effect's %v over its column's %v", got, hackAndSlash.StatedChance(), column(hackAndSlash))
	}
}
