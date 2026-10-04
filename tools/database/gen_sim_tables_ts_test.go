package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func TestSimTablesTSData(t *testing.T) {
	data := simTablesTSData()

	itemTypes := map[string]bool{}
	for _, entry := range data.ItemTypeToSlots {
		itemTypes[entry.ItemTypeName] = true
	}
	for itemType := range core.ItemTypeToSlotsMap {
		if !itemTypes[proto.ItemType_name[int32(itemType)]] {
			t.Errorf("ITEM_TYPE_TO_SLOTS is missing %s", proto.ItemType_name[int32(itemType)])
		}
	}
}
