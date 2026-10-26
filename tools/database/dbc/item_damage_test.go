package dbc

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
)

func TestUnroundedDPS(t *testing.T) {
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	for name, value := range map[string]float64{
		"ItemDamageOneHand": 14.695897, "ItemDamageTwoHand": 19.808466,
		"ItemDamageOneHandCaster": 9.63904, "ItemDamageTwoHandCaster": 11.927145,
	} {
		dbcInstance.ItemDamageTable[name] = map[int]ItemDamageTable{
			22: {Quality: []float64{0, 0, 0, value, 30.25, 0, 0}},
			23: {Quality: make([]float64, 7)},
		}
	}
	cases := []struct {
		name                    string
		inventory, class, level int
		quality                 ItemQuality
		flags                   ItemStaticFlags1
		dps, low, high          float64
	}{
		{"one hand", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 3, 0, 14.695897, 23, 35},
		{"two hand", INVTYPE_2HWEAPON, ITEM_CLASS_WEAPON, 22, 3, 0, 19.808466, 31, 48},
		{"caster one hand", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 3, CASTER_WEAPON, 9.63904, 15, 23},
		{"caster two hand", INVTYPE_2HWEAPON, ITEM_CLASS_WEAPON, 22, 3, CASTER_WEAPON, 11.927145, 19, 29},
		{"heirloom", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 22, 7, 0, 30.25, 48, 73},
		{"zero", INVTYPE_WEAPON, ITEM_CLASS_WEAPON, 23, 3, 0, 0, 0, 0},
		{"nonweapon", 0, 0, 22, 0, 0, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			item := Item{Id: 1, ItemClass: c.class, InventoryType: c.inventory, ItemDelay: 2000,
				ItemLevel: c.level, OverallQuality: c.quality, Flags1: c.flags, DmgVariance: 0.4}
			generated := item.ToUIItem().ScalingOptions[0]
			field := generated.ProtoReflect().Descriptor().Fields().ByName("unrounded_dps")
			if field == nil {
				t.Fatal("scaling properties discard unrounded weapon table DPS")
			}
			if got, want := generated.UnroundedDps, float32(c.dps); got != want {
				t.Fatalf("unrounded DPS = %.9f, want %.9f", got, want)
			}
			if generated.WeaponDamageMin != c.low || generated.WeaponDamageMax != c.high {
				t.Fatalf("equipped endpoints = %v-%v, want %v-%v", generated.WeaponDamageMin, generated.WeaponDamageMax, c.low, c.high)
			}
			for _, format := range []string{"json", "binary"} {
				t.Run(format, func(t *testing.T) {
					decoded := &proto.ScalingItemProperties{}
					var data []byte
					var err error
					if format == "json" {
						data, err = protojson.Marshal(generated)
					} else {
						data, err = googleProto.Marshal(generated)
					}
					if err != nil {
						t.Fatal(err)
					}
					if format == "json" {
						err = protojson.Unmarshal(data, decoded)
					} else {
						err = googleProto.Unmarshal(data, decoded)
					}
					if err != nil {
						t.Fatal(err)
					}
					if !googleProto.Equal(generated, decoded) {
						t.Fatalf("%s round trip changes scaling properties", format)
					}
				})
			}
		})
	}
}

func TestWeaponDPSIsZeroWhenSourceTableIsMissing(t *testing.T) {
	// Given
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	item := Item{ItemClass: ITEM_CLASS_WEAPON, InventoryType: INVTYPE_WEAPON, OverallQuality: 2, ItemLevel: 18}
	// When
	row := item.ToUIItem().ScalingOptions[0]
	// Then
	if row.UnroundedDps != 0 {
		t.Fatalf("missing source row gives table DPS %v, want 0", row.UnroundedDps)
	}
}

// The level 60 rare budget's first column is 100 and its one-hand weapon column 50. A healing weapon
// adds 3.76 healing and a third of that as spell damage per point of the first column, rounded down;
// a spell power weapon 2 spell power per point; both on top of the weapon's own stats. Reduced damage
// alone adds none.
func TestCasterWeaponSpellPower(t *testing.T) {
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	dbcInstance.RandomPropertiesByIlvl[60] = RandomPropAllocationMap{proto.ItemQuality_ItemQualityRare: {100, 80, 60, 50, 40}}

	for _, tc := range []struct {
		name            string
		flags           ItemStaticFlags4
		ownStat         int
		healing, damage float64
	}{
		{"healing weapon", CASTER_HEALING_POWER, 0, 376, 125},
		{"spell power weapon with no stat of its own", CASTER_SPELL_POWER, 0, 200, 200},
		{"spell power weapon on top of its own spell power", CASTER_SPELL_POWER, ITEM_MOD_SPELL_POWER, 205, 205},
		{"reduced damage only, its own healing", REDUCED_CASTER_DPS, ITEM_MOD_SPELL_HEALING_DONE, 5, 0},
	} {
		item := Item{Id: 1, ItemClass: ITEM_CLASS_WEAPON, ItemSubClass: ITEM_SUBCLASS_WEAPON_MACE, InventoryType: INVTYPE_WEAPON,
			ItemLevel: 60, OverallQuality: RARE, Flags4: tc.flags}
		if tc.ownStat != 0 {
			item.BonusStat, item.StatAlloc = []int{tc.ownStat}, []float64{1000}
		}
		stats := item.GetStats(item.ItemLevel)
		if got, want := [2]float64{stats[proto.Stat_StatHealingPower], stats[proto.Stat_StatSpellDamage]}, [2]float64{tc.healing, tc.damage}; got != want {
			t.Errorf("%s: healing %v, spell damage %v, want %v and %v", tc.name, got[0], got[1], want[0], want[1])
		}
	}
}

// The tables state 33 DPS one-handed and 39 two-handed; the variance is 0.4, so the range is
// DPS x speed x 0.8 to 1.2, low floored and high rounded.
func TestWeaponDamageRange(t *testing.T) {
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	for name, dps := range map[string]float64{
		"ItemDamageOneHand": 33, "ItemDamageOneHandCaster": 33, "ItemDamageTwoHand": 39, "ItemDamageTwoHandCaster": 39,
	} {
		dbcInstance.ItemDamageTable[name] = map[int]ItemDamageTable{50: {Quality: []float64{0, 0, 0, dps, 0, 0, 0}}}
	}

	for _, tc := range []struct {
		name            string
		inventory       int
		delay           int
		caster          ItemStaticFlags1
		reduced         ItemStaticFlags4
		qualityModifier float64
		min, max        float64
	}{
		{"one hand, 33 DPS at 2.0", INVTYPE_WEAPON, 2000, 0, 0, 0, 52, 79},
		{"caster one hand without the reduced flag, the whole table", INVTYPE_WEAPON, 2000, CASTER_WEAPON, 0, 0, 52, 79},
		{"reduced caster one hand, 2/3 of the table: 22 DPS", INVTYPE_WEAPON, 2000, CASTER_WEAPON, REDUCED_CASTER_DPS, 0, 35, 53},
		{"reduced caster two hand, 29/39 of the table: 29 DPS at 3.0", INVTYPE_2HWEAPON, 3000, CASTER_WEAPON, REDUCED_CASTER_DPS, 0, 69, 104},
		{"reduced caster two hand with a QualityModifier of 5.5 DPS", INVTYPE_2HWEAPON, 3000, CASTER_WEAPON, REDUCED_CASTER_DPS, 5.5, 86, 121},
	} {
		item := Item{Id: 1, ItemClass: ITEM_CLASS_WEAPON, InventoryType: tc.inventory, ItemDelay: tc.delay, ItemLevel: 50,
			OverallQuality: 3, Flags1: tc.caster, Flags4: tc.reduced, DmgVariance: 0.4, QualityModifier: tc.qualityModifier}
		if got, want := [2]float64{item.WeaponDmgMin(0), item.WeaponDmgMax(0)}, [2]float64{tc.min, tc.max}; got != want {
			t.Errorf("%s: damage %v-%v, want %v-%v", tc.name, got[0], got[1], want[0], want[1])
		}
	}
}

// A random suffix whose bonus list reduces caster DPS leaves the weapon's own damage alone and gives
// it a second range: the share of the table cell, floored and rounded as the game shows it, with the
// item's own variance.
func TestReducedCasterWeaponDPS(t *testing.T) {
	once.Do(func() {})
	previous := dbcInstance
	dbcInstance = NewDBC()
	t.Cleanup(func() { dbcInstance = previous })
	for name, dps := range map[string]float64{"ItemDamageOneHand": 30, "ItemDamageTwoHand": 39} {
		dbcInstance.ItemDamageTable[name] = map[int]ItemDamageTable{50: {Quality: []float64{0, 0, 0, dps, 0, 0, 0}}}
	}

	for _, tc := range []struct {
		name           string
		inventory      int
		delay          int
		variance       float64
		ownMin, ownMax float64
		dps            float64
		min, max       float64
	}{
		{"two hand, 29/39 of 39 DPS at 3.0 with 0.4 variance", INVTYPE_2HWEAPON, 3000, 0.4, 93, 140, 29, 69, 104},
		{"two hand with 0.6 variance", INVTYPE_2HWEAPON, 3000, 0.6, 81, 152, 29, 60, 113},
		{"one hand, 2/3 of 30 DPS at 2.0", INVTYPE_WEAPON, 2000, 0.4, 48, 72, 20, 32, 48},
	} {
		item := Item{Id: 1, ItemClass: ITEM_CLASS_WEAPON, InventoryType: tc.inventory, ItemDelay: tc.delay, ItemLevel: 50,
			OverallQuality: 3, DmgVariance: tc.variance}
		dps, present := dbcInstance.ReducedCasterWeaponDPS(&item, 0)
		if !present || math.Abs(dps-tc.dps) > 1e-9 {
			t.Errorf("%s: reduced DPS %v (%v), want %v", tc.name, dps, present, tc.dps)
		}
		if got, want := [2]float64{item.WeaponDmgMin(0), item.WeaponDmgMax(0)}, [2]float64{tc.ownMin, tc.ownMax}; got != want {
			t.Errorf("%s: the item's own damage is %v-%v, want the whole table's %v-%v", tc.name, got[0], got[1], want[0], want[1])
		}
		if low, high := item.WeaponDamage(dps); low != tc.min || high != tc.max {
			t.Errorf("%s: reduced damage %v-%v, want %v-%v", tc.name, low, high, tc.min, tc.max)
		}
	}
}
