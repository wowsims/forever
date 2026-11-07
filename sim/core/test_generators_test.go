package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// An AllItems test weapon replaces the main hand whatever fills the hands, a two-hander empties the
// off hand with it, a shield and an off-hand item replace the off hand, and an item of another kind
// still takes the slot EquipItem gives it.
func TestEquipTestItemTakesTheWornSlot(t *testing.T) {
	weapon := func(id int32, hand proto.HandType) Item {
		return Item{ID: id, Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeSword, HandType: hand}
	}
	dualWielder := func() Equipment {
		var equipment Equipment
		*equipment.MainHand() = weapon(1, proto.HandType_HandTypeOneHand)
		*equipment.OffHand() = weapon(2, proto.HandType_HandTypeOneHand)
		*equipment.Finger1() = Item{ID: 3, Type: proto.ItemType_ItemTypeFinger}
		return equipment
	}

	for name, tc := range map[string]struct {
		item     Item
		mainHand int32
		offHand  int32
	}{
		"a one-hander replaces the main hand":               {weapon(10, proto.HandType_HandTypeOneHand), 10, 2},
		"a main-hand weapon replaces the main hand":         {weapon(11, proto.HandType_HandTypeMainHand), 11, 2},
		"a two-hander takes both hands":                     {weapon(12, proto.HandType_HandTypeTwoHand), 12, 0},
		"an off-hand weapon replaces the off hand":          {weapon(13, proto.HandType_HandTypeOffHand), 1, 13},
		"a shield replaces the off hand":                    {Item{ID: 14, Type: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield, HandType: proto.HandType_HandTypeOffHand}, 1, 14},
		"a ring leaves the hands and takes the free finger": {Item{ID: 15, Type: proto.ItemType_ItemTypeFinger}, 1, 2},
	} {
		equipment := dualWielder()
		equipTestItem(&equipment, tc.item)
		if got := equipment.MainHand().ID; got != tc.mainHand {
			t.Errorf("%s: main hand holds %d, want %d", name, got, tc.mainHand)
		}
		if got := equipment.OffHand().ID; got != tc.offHand {
			t.Errorf("%s: off hand holds %d, want %d", name, got, tc.offHand)
		}
		if tc.item.Type == proto.ItemType_ItemTypeFinger && equipment.Finger2().ID != tc.item.ID {
			t.Errorf("%s: the second finger holds %d, want %d", name, equipment.Finger2().ID, tc.item.ID)
		}
	}
}
