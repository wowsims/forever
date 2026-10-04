package shared

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Mark of Resolution 17759's on-use: ItemEffect casts 21956, A_SCHOOL_ABSORB 780 with Misc 1
// (Physical) on the caster for 10 s, on a 5 min cooldown in category 2554 for 10 s.
const (
	resolutionSpell    int32 = 21956
	resolutionCategory int32 = 2554
)

func resolutionOnUse() *proto.ItemEffect {
	return testOnUse(resolutionSpell, 300000, resolutionCategory, 10000)
}

// 21956 states a 0.6 spread around its 780. The tests that count the shield down read the 780 alone.
func withoutSpread(t *testing.T, spellID int32) {
	t.Helper()
	editRow(t, spellID, func(s *spelldata.Spell) { s.Effects[0].Variance = 0 })
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

// The shield takes 780 off the physical damage the wearer takes: a hit of 1000 lands for 220, and the
// shield it used up is gone. A fire hit before it is not absorbed and leaves the shield whole.
func TestOnUseAbsorbTakesPhysicalDamageOffItsShield(t *testing.T) {
	const itemID int32 = 991201
	withoutSpread(t, resolutionSpell)
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

	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, 1000); got != 1000 {
		t.Errorf("a fire hit of 1000 landed for %v, want all of it", got)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 1000); got != 220 {
		t.Errorf("a physical hit of 1000 landed for %v, want 220 past the 780 shield", got)
	}
	if shield.IsActive() {
		t.Errorf("the shield is still up after absorbing its 780")
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100); got != 100 {
		t.Errorf("a physical hit after the shield broke landed for %v, want all 100", got)
	}
}

// Unused, the shield drops at the end of its 10 s and absorbs nothing after.
func TestOnUseAbsorbExpiresWithItsRow(t *testing.T) {
	const itemID int32 = 991202
	withoutSpread(t, resolutionSpell)
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: resolutionOnUse()})
	shield := absorbAura(t, caster, resolutionSpell, itemID)

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	start := sim.CurrentTime

	stepPast(t, sim, start+10*time.Second-time.Millisecond)
	if !shield.IsActive() {
		t.Fatalf("the shield dropped before its 10 s")
	}
	stepPast(t, sim, start+10*time.Second+time.Millisecond)
	if shield.IsActive() {
		t.Errorf("the shield is still up after its 10 s")
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100); got != 100 {
		t.Errorf("a physical hit after the shield expired landed for %v, want all 100", got)
	}
}

// The item's own 5 min cooldown, and its category's 10 s that a second item in category 2554 waits out.
func TestOnUseAbsorbRunsOnTheItemsCooldownAndCategory(t *testing.T) {
	const markID, otherID int32 = 991203, 991204
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{
		markID:  resolutionOnUse(),
		otherID: testOnUse(resolutionSpell, 60000, resolutionCategory, 10000),
	})
	mark, other := onUseSpell(t, caster, markID), onUseSpell(t, caster, otherID)

	mark.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	if offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could be used while the mark's category cooldown runs", resolutionCategory)
	}

	stepPast(t, sim, start+10*time.Second)
	if !offCooldown(sim, other) {
		t.Errorf("an item sharing category %d could not be used once the 10 s ran out", resolutionCategory)
	}
	if got := mark.CD.ReadyAt(); got != start+5*time.Minute {
		t.Errorf("the mark is ready again at %v, want 5 min after its use at %v", got, start)
	}
}

// With the row's spread the shield rolls within 780 ± 234, 0.6/2 of its base points either side.
func TestOnUseAbsorbRollsTheRowsSpread(t *testing.T) {
	const itemID int32 = 991205
	effect := spelldata.MustFind(resolutionSpell).AbsorbEffect()
	sim, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, map[int32]*proto.ItemEffect{itemID: resolutionOnUse()})
	low, high := effect.Min(), effect.Max()
	if low == high {
		t.Fatalf("%d states no spread", resolutionSpell)
	}

	onUseSpell(t, caster, itemID).Cast(sim, &caster.Unit)
	if absorbed := 2000 - takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 2000); absorbed < low || absorbed > high {
		t.Errorf("the shield absorbed %v, want %v to %v", absorbed, low, high)
	}
}

// Enchant Chest - Absorption 8220: its equip aura 1249072 rolls 25% on a melee hit taken, at most once
// every 5 s, and casts 1249073, A_SCHOOL_ABSORB 50 physical until used up. A copy of the trigger
// rolling 100% stands in, so every hit taken outside the lockout procs.
func TestAbsorbProcShieldsTheWearerOnAMeleeHitTaken(t *testing.T) {
	const chestID, enchantID int32 = 991206, 991207
	const trigger, absorb int32 = 1249072, 1249073
	editRow(t, trigger, func(s *spelldata.Spell) { s.ProcChance = 100 })

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
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolFire, 100); got != 100 {
		t.Errorf("a fire hit landed for %v, want all 100", got)
	}
	if got := takenAfterAbsorbs(sim, caster, core.SpellSchoolPhysical, 100); got != 50 {
		t.Errorf("a physical hit of 100 landed for %v, want 50 past the shield", got)
	}
	procTime := sim.CurrentTime

	struck()
	if shield.IsActive() {
		t.Errorf("a hit inside the 5 s lockout put the shield back up")
	}

	stepPast(t, sim, procTime+5*time.Second)
	struck()
	if !shield.IsActive() {
		t.Errorf("a hit after the lockout put up no shield")
	}
}

// Uther's Strength 11302's equip aura 8397 states 2% on effect 1 and 4 in its ProcChance column. The
// item's proc rolls the column; a class spell stating both, Hack and Slash 13960 (5% on its effect, 1
// in the column), is not an item proc and keeps the effect's.
func TestItemProcRollsTheColumnOverAnEffectsChance(t *testing.T) {
	_, caster := newOnUseSim(t, NewSpellDataAbsorbOnUse, nil)
	uther := spelldata.MustFind(8397)
	if got := uther.StatedChance(); got != 0.02 {
		t.Fatalf("8397 states %v on its effect, want 0.02", got)
	}

	config := spellDataProcListener(&caster.Character, SpellDataProc{Name: "Uther's Strength", ItemID: 11302, TriggerSpellID: 8397},
		effectSource{id: 11302}, uther)
	if config.ProcChance != 0.04 || config.DPM != nil {
		t.Errorf("the proc rolls %v (manager %v), want the column's 0.04", config.ProcChance, config.DPM != nil)
	}

	hackAndSlash := spelldata.MustFind(13960)
	if got := spelldata.ProcTrigger(&caster.Character, hackAndSlash, nil).ProcChance; got != 0.05 {
		t.Errorf("a class spell's trigger rolls %v, want its effect's 0.05 over its column's 1", got)
	}
}
