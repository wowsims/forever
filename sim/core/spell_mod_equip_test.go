package core

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

func checkReaches(t *testing.T, when string, mod *SpellMod, spells map[string]*Spell, want ...string) {
	t.Helper()
	for name, spell := range spells {
		if got := slices.Contains(mod.AffectedSpells, spell); got != slices.Contains(want, name) {
			t.Errorf("%s: mod reaches the %s = %v, want %v", when, name, got, !got)
		}
	}
}

func autos(character *Character) map[string]*Spell {
	return map[string]*Spell{
		"main-hand auto": character.AutoAttacks.MHAuto(),
		"off-hand auto":  character.AutoAttacks.OHAuto(),
		"ranged auto":    character.AutoAttacks.RangedAuto(),
	}
}

func TestSubclassModFollowsTheWeaponBehindTheAttack(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	if !character.AutoAttacks.IsDualWielding {
		t.Fatal("sword and dagger should dual wield")
	}

	swords := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusCrit_Percent,
		FloatValue:    5,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword),
	})
	checkReaches(t, "sword and dagger", swords, autos(character), "main-hand auto")

	shooters := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    3,
		EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING,
	})
	checkReaches(t, "bow, gun or crossbow", shooters, autos(character), "ranged auto")

	offHand := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    2,
		EquipSubclass: 41105,
		EquipInvType:  dbcenums.INVTYPE_BIT_WEAPONOFFHAND,
	})
	checkReaches(t, "one-hander in the off hand", offHand, autos(character), "off-hand auto")

	offHandByAttr := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusHit_Percent,
		FloatValue:    2,
		EquipSubclass: 41105,
		EquipHand:     ProcMaskMeleeOH,
	})
	checkReaches(t, "off-hand bit", offHandByAttr, autos(character), "off-hand auto")
}

func TestShieldIsNotAWeapon(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestShieldID, 0), nil, nil)

	offHand := EquipRequirement{InvTypeMask: dbcenums.INVTYPE_BIT_SHIELD | dbcenums.INVTYPE_BIT_WEAPONOFFHAND | dbcenums.INVTYPE_BIT_HOLDABLE}
	anyWeapon := EquipRequirement{SubclassMask: -1}
	equipment := character.playerEquipment()
	for _, procMask := range []ProcMask{ProcMaskMeleeOHAuto, ProcMaskMeleeOHSpecial} {
		if offHand.matchesEquipment(equipment, procMask) || anyWeapon.matchesEquipment(equipment, procMask) {
			t.Errorf("a shield in the off hand matched a weapon requirement for proc mask %v", procMask)
		}
	}
	if !anyWeapon.matchesEquipment(equipment, ProcMaskMeleeMHSpecial) {
		t.Error("the sword in the main hand should match any weapon")
	}
	if anyWeapon.matchesEquipment(equipment, ProcMaskSpellDamage) {
		t.Error("an attack with no weapon bit should never match")
	}
}

func TestEquipRequirementNeverMatchesOtherUnits(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	anyWeapon := EquipRequirement{SubclassMask: -1}
	if !anyWeapon.matchesEquipment(character.playerEquipment(), ProcMaskMeleeMHAuto) {
		t.Fatal("the player's own sword should match")
	}

	pet := &Unit{Type: PetUnit, Env: character.Env}
	target := sim.Encounter.ActiveTargetUnits[0]
	for name, unit := range map[string]*Unit{"pet": pet, "target": target} {
		if anyWeapon.matchesEquipment(unit.playerEquipment(), ProcMaskMeleeMHAuto) {
			t.Errorf("a %s matched a weapon requirement", name)
		}
		mod := &SpellMod{Equip: anyWeapon}
		if shouldApply(&Spell{Unit: unit, ProcMask: ProcMaskMeleeMHAuto}, mod) {
			t.Errorf("a weapon mod applied to a %s's attack", name)
		}
	}
}

func TestSubclassModFollowsAnItemSwap(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead,
		WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, 0),
		WeaponTestSlots(WeaponTestDaggerID, WeaponTestSwordID, 0), nil)

	swords := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusCrit_Percent,
		FloatValue:    5,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword),
	})
	swords.Activate()
	mhAuto, ohAuto := character.AutoAttacks.MHAuto(), character.AutoAttacks.OHAuto()
	baseMH, baseOH := mhAuto.BonusCritPercent-5, ohAuto.BonusCritPercent

	check := func(when string, mh, oh float64) {
		t.Helper()
		if got := mhAuto.BonusCritPercent - baseMH; got != mh {
			t.Errorf("%s: main-hand auto has %v%% from the mod, want %v%%", when, got, mh)
		}
		if got := ohAuto.BonusCritPercent - baseOH; got != oh {
			t.Errorf("%s: off-hand auto has %v%% from the mod, want %v%%", when, got, oh)
		}
	}
	check("sword in the main hand", 5, 0)

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	check("sword swapped to the off hand", 0, 5)

	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Main, false)
	check("sword swapped back", 5, 0)

	swords.Deactivate()
	character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	check("inactive mod across a swap", 0, 0)
	swords.Activate()
	check("mod turned on with the sword in the off hand", 0, 5)
}

func TestSwapThatKeepsAModsReachLeavesItAlone(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead,
		WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, 0),
		WeaponTestSlots(WeaponTestDaggerID, WeaponTestSwordID, 0), nil)

	applies, removes := 0, 0
	anyWeapon := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_Custom,
		ApplyCustom:   func(_ *SpellMod, _ *Spell) { applies++ },
		RemoveCustom:  func(_ *SpellMod, _ *Spell) { removes++ },
		EquipSubclass: -1,
	})
	anyWeapon.Activate()
	crit := character.AddDynamicMod(SpellModConfig{
		Kind:          SpellMod_BonusCrit_Percent,
		FloatValue:    0.1,
		EquipSubclass: -1,
	})
	crit.Activate()
	checkReaches(t, "sword and dagger", anyWeapon, autos(character), "main-hand auto", "off-hand auto")

	mhAuto, ohAuto := character.AutoAttacks.MHAuto(), character.AutoAttacks.OHAuto()
	mhCrit, ohCrit := math.Float64bits(mhAuto.BonusCritPercent), math.Float64bits(ohAuto.BonusCritPercent)
	wantApplies, wantRemoves := applies, removes
	for _, swap := range []proto.APLActionItemSwap_SwapSet{proto.APLActionItemSwap_Swap1, proto.APLActionItemSwap_Main} {
		character.ItemSwap.SwapItems(sim, swap, false)
		checkReaches(t, swap.String(), anyWeapon, autos(character), "main-hand auto", "off-hand auto")
		if applies != wantApplies || removes != wantRemoves {
			t.Errorf("%s: the mod was applied %d and removed %d times, want %d and %d", swap, applies, removes, wantApplies, wantRemoves)
		}
		if math.Float64bits(mhAuto.BonusCritPercent) != mhCrit || math.Float64bits(ohAuto.BonusCritPercent) != ohCrit {
			t.Errorf("%s: crit moved to %v and %v", swap, mhAuto.BonusCritPercent, ohAuto.BonusCritPercent)
		}
	}
}

func TestWeaponChangeRunsOncePerSwapThatMovesAWeapon(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead,
		WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, 0),
		WeaponTestSlots(WeaponTestDaggerID, WeaponTestSwordID, 0), nil)
	calls := 0
	character.ItemSwap.onWeaponChange = append(character.ItemSwap.onWeaponChange, func(*Simulation) { calls++ })

	for _, step := range []struct {
		when string
		swap func()
		want int
	}{
		{"swapped", func() { character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false) }, 1},
		{"reset to the main set", func() { character.ItemSwap.reset(sim) }, 2},
		{"reset again", func() { character.ItemSwap.reset(sim) }, 2},
	} {
		step.swap()
		if calls != step.want {
			t.Errorf("%s: the weapon change ran %d times, want %d", step.when, calls, step.want)
		}
	}
}
