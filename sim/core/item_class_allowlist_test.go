package core

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// An item effect applies only to a class the item's class allowlist names, whether the item is
// worn or waiting in the item swap set; an item without an allowlist applies to every class.
func TestItemEffectAppliesOnlyToAllowedClasses(t *testing.T) {
	const paladinOnlyID, anyClassID int32 = 990801, 990802

	trinket := func(id int32, classes ...proto.Class) *proto.SimItem {
		return &proto.SimItem{Id: id, Name: "Test Trinket", Type: proto.ItemType_ItemTypeTrinket, ClassAllowlist: classes,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}
	}
	addToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		trinket(paladinOnlyID, proto.Class_ClassPaladin),
		trinket(anyClassID),
	}})

	var applied []int32
	for _, id := range []int32{paladinOnlyID, anyClassID} {
		NewItemEffect(id, func(Agent) { applied = append(applied, id) })
	}

	trinkets := func(first, second int32) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotTrinket1] = &proto.ItemSpec{Id: first}
		items[proto.ItemSlot_ItemSlotTrinket2] = &proto.ItemSpec{Id: second}
		return items
	}

	for _, tc := range []struct {
		name     string
		class    proto.Class
		equipped []*proto.ItemSpec
		swapped  []*proto.ItemSpec
		want     []int32
	}{
		{"paladin wearing both", proto.Class_ClassPaladin, trinkets(paladinOnlyID, anyClassID), nil, []int32{paladinOnlyID, anyClassID}},
		{"warrior wearing both", proto.Class_ClassWarrior, trinkets(paladinOnlyID, anyClassID), nil, []int32{anyClassID}},
		{"paladin swapping to both", proto.Class_ClassPaladin, trinkets(0, 0), trinkets(paladinOnlyID, anyClassID), []int32{paladinOnlyID, anyClassID}},
		{"warrior swapping to both", proto.Class_ClassWarrior, trinkets(0, 0), trinkets(paladinOnlyID, anyClassID), []int32{anyClassID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			applied = nil
			player := &proto.Player{
				Name:      "Class Tester",
				Class:     tc.class,
				Race:      proto.Race_RaceHuman,
				Spec:      &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}},
				Equipment: &proto.EquipmentSpec{Items: tc.equipped},
			}
			if tc.class == proto.Class_ClassPaladin {
				player.Spec = &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{}}
			}
			if tc.swapped != nil {
				player.EnableItemSwap = true
				player.ItemSwap = &proto.ItemSwap{Items: tc.swapped}
			}
			agent := &FakeAgent{Character: NewCharacter(&Party{}, 0, player)}
			agent.Character.ItemSwap.initialize(&agent.Character)
			agent.Character.applyItemEffects(agent)

			slices.Sort(applied)
			if !slices.Equal(applied, tc.want) {
				t.Errorf("applied effects = %v, want %v", applied, tc.want)
			}
		})
	}
}

// The test generator's item filter drops an item its class may not use.
func TestItemFilterDropsItemsTheClassMayNotUse(t *testing.T) {
	paladinOnly := Item{ID: 990811, Type: proto.ItemType_ItemTypeTrinket, ClassAllowlist: []proto.Class{proto.Class_ClassPaladin}}
	anyClass := Item{ID: 990812, Type: proto.ItemType_ItemTypeTrinket}

	for _, tc := range []struct {
		class proto.Class
		item  Item
		want  bool
	}{
		{proto.Class_ClassPaladin, paladinOnly, true},
		{proto.Class_ClassWarrior, paladinOnly, false},
		{proto.Class_ClassUnknown, paladinOnly, true},
		{proto.Class_ClassWarrior, anyClass, true},
	} {
		filter := ItemFilter{Class: tc.class}
		if got := filter.Matches(tc.item, true); got != tc.want {
			t.Errorf("%s filter matches item %d = %v, want %v", tc.class, tc.item.ID, got, tc.want)
		}
	}
}
