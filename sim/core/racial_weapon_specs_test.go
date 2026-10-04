package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const (
	racialTestMaceID int32 = 991511 + iota
	racialTestTwoHandAxeID
	racialTestTwoHandMaceID
)

func addRacialTestWeapons() {
	weapon := func(id int32, weaponType proto.WeaponType, handType proto.HandType) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Weapon",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     weaponType,
			HandType:       handType,
			WeaponSpeed:    2.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
		}
	}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		weapon(racialTestMaceID, proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeOneHand),
		weapon(racialTestTwoHandAxeID, proto.WeaponType_WeaponTypeAxe, proto.HandType_HandTypeTwoHand),
		weapon(racialTestTwoHandMaceID, proto.WeaponType_WeaponTypeMace, proto.HandType_HandTypeTwoHand),
	}})
}

func TestRacialWeaponSpecializationFollowsItemSwaps(t *testing.T) {
	daggers := WeaponTestSlots(WeaponTestDaggerID, WeaponTestDaggerID, 0)
	_, withoutSword := NewWeaponTestSim(t, proto.Race_RaceHuman, daggers, daggers, nil)
	base := withoutSword.GetStat(stats.PhysicalCritPercent)

	for _, tc := range []struct {
		name                          string
		mainHand, offHand             int32
		swapMainHand, swapOffHand     int32
		critBeforeSwap, critAfterSwap float64
	}{
		{"sword swapped out of the main hand", WeaponTestSwordID, WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestDaggerID, 2, 0},
		{"sword swapped into the off hand", WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestSwordID, 0, 2},
		{"sword in the off hand only, swapped out", WeaponTestDaggerID, WeaponTestSwordID, WeaponTestDaggerID, WeaponTestDaggerID, 2, 0},
	} {
		sim, character := NewWeaponTestSim(t, proto.Race_RaceHuman,
			WeaponTestSlots(tc.mainHand, tc.offHand, 0), WeaponTestSlots(tc.swapMainHand, tc.swapOffHand, 0), nil)
		if got := character.GetStat(stats.PhysicalCritPercent) - base; math.Abs(got-tc.critBeforeSwap) > 1e-9 {
			t.Errorf("%s: +%v%% crit before the swap, want +%v%%", tc.name, got, tc.critBeforeSwap)
		}

		character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
		if got := character.GetStat(stats.PhysicalCritPercent) - base; math.Abs(got-tc.critAfterSwap) > 1e-9 {
			t.Errorf("%s: +%v%% crit after the swap, want +%v%%", tc.name, got, tc.critAfterSwap)
		}
	}
}

func TestRacialWeaponSpecializationsRaiseAllCrit(t *testing.T) {
	addRacialTestWeapons()

	crit := func(race proto.Race, mainHand int32) (physical float64, spell float64) {
		_, character := NewWeaponTestSim(t, race, WeaponTestSlots(mainHand, 0, 0), nil, nil)
		return character.GetStat(stats.PhysicalCritPercent), character.GetStat(stats.SpellCritPercent)
	}

	for _, tc := range []struct {
		name            string
		race            proto.Race
		weapon, without int32
		crit            float64
	}{
		{"human with a sword", proto.Race_RaceHuman, WeaponTestSwordID, WeaponTestDaggerID, 2},
		{"human with a mace", proto.Race_RaceHuman, racialTestMaceID, WeaponTestDaggerID, 0},
		{"orc with a two-handed axe", proto.Race_RaceOrc, racialTestTwoHandAxeID, racialTestTwoHandMaceID, 1},
		{"dwarf with a mace", proto.Race_RaceDwarf, racialTestMaceID, WeaponTestDaggerID, 1},
		{"dwarf with a two-handed mace", proto.Race_RaceDwarf, racialTestTwoHandMaceID, racialTestTwoHandAxeID, 1},
	} {
		physical, spell := crit(tc.race, tc.weapon)
		basePhysical, baseSpell := crit(tc.race, tc.without)
		if got := physical - basePhysical; math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("%s: physical crit +%v%%, want +%v%%", tc.name, got, tc.crit)
		}
		if got := spell - baseSpell; math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("%s: spell crit +%v%%, want +%v%%", tc.name, got, tc.crit)
		}
	}
}
