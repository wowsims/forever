package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// A set item counts toward the registered set with its set id. An item whose set id no registered set has
// matches by name only where the item carries no set id, so a reissue under the same name takes no bonus.
func TestSetBonusMatchesByNameOnlyWithoutSetID(t *testing.T) {
	const setName = "Test Raiment"
	saved := sets
	sets = []*ItemSet{{ID: 990900, Name: setName, Bonuses: map[int32]ApplySetBonus{2: func(Agent, *Aura) {}}}}
	t.Cleanup(func() { sets = saved })

	for _, tc := range []struct {
		name  string
		setID int32
		want  bool
	}{
		{"registered set id", 990900, true},
		{"no set id", 0, true},
		{"set id no registered set has", 990901, false},
	} {
		var equipment Equipment
		equipment[proto.ItemSlot_ItemSlotHead] = Item{SetName: setName, SetID: tc.setID}
		equipment[proto.ItemSlot_ItemSlotChest] = Item{SetName: setName, SetID: tc.setID}
		if got := equipment.getSetBonuses().ContainsBonus(setName, 2); got != tc.want {
			t.Errorf("%s: 2-piece bonus active = %v, want %v", tc.name, got, tc.want)
		}
	}
}
