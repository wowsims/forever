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

	familyMobTypes := map[string]string{}
	for _, entry := range data.CreatureFamilyMobTypes {
		familyMobTypes[entry.CreatureFamilyName] = entry.MobTypeName
	}
	for family, mobType := range map[proto.CreatureFamily]proto.MobType{
		proto.CreatureFamily_CreatureFamilyWolf:    proto.MobType_MobTypeBeast,
		proto.CreatureFamily_CreatureFamilyAquatic: proto.MobType_MobTypeUnknown,
	} {
		if got := familyMobTypes[family.String()]; got != mobType.String() {
			t.Errorf("CREATURE_FAMILY_MOB_TYPES gives %s %q, want %s", family, got, mobType)
		}
	}
}
