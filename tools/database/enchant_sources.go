package database

import (
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// What says where an item comes from, keyed by item id. Atlas and Questie are the pruned inputs.
type EnchantSourceInputs struct {
	Crafted map[int][]*proto.CraftedSource
	Rep     map[int][]*proto.RepSource
	Atlas   *WowDatabase
	Questie *WowDatabase
}

// Sets each enchant's sources to those of its own item and the items that teach it, by their ids in recipes, without
// repeats. An enchant neither has any source for takes none: the game files hold no trainers, so one a trainer teaches
// looks the same.
func AddEnchantSources(enchants map[EnchantDBKey]*proto.UIEnchant, recipes map[EnchantDBKey][]int, inputs EnchantSourceInputs) {
	for key, enchant := range enchants {
		itemIDs := append([]int{int(enchant.ItemId)}, recipes[key]...)
		var sources []*proto.UIItemSource
		for i, itemID := range itemIDs {
			if itemID == 0 || slices.Contains(itemIDs[:i], itemID) {
				continue
			}
			for _, source := range inputs.itemSources(itemID) {
				if !slices.ContainsFunc(sources, func(seen *proto.UIItemSource) bool { return googleProto.Equal(seen, source) }) {
					sources = append(sources, source)
				}
			}
		}
		enchant.Sources = sources
	}
}

// One item's sources in the order an item in the database lists them.
func (inputs EnchantSourceInputs) itemSources(itemID int) []*proto.UIItemSource {
	var sources []*proto.UIItemSource
	for _, crafted := range inputs.Crafted[itemID] {
		sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: crafted}})
	}
	for _, rep := range inputs.Rep[itemID] {
		sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: rep}})
	}
	sources = append(sources, AtlasLootSourcesBeside(sources, inputs.Atlas.Items[int32(itemID)].GetSources())...)
	return append(sources, questieSourcesBeside(sources, inputs.Questie.Items[int32(itemID)])...)
}

// AtlasLoot's sources for an item that already holds existing from the game files, less the ones they repeat: the
// crafting spell both state, and a rep source of the faction the game files name.
func AtlasLootSourcesBeside(existing, atlasLoot []*proto.UIItemSource) []*proto.UIItemSource {
	firstRepFaction := int32(-1)
	if i := slices.IndexFunc(existing, func(source *proto.UIItemSource) bool { return source.GetRep() != nil }); i >= 0 {
		firstRepFaction = existing[i].GetRep().RepFactionId
	}
	return core.FilterSlice(atlasLoot, func(source *proto.UIItemSource) bool {
		if rep := source.GetRep(); rep != nil && rep.RepFactionId == firstRepFaction {
			return false
		}
		return !slices.ContainsFunc(existing, func(held *proto.UIItemSource) bool { return googleProto.Equal(held, source) })
	})
}

// The zones and NPCs the drop sources of from's items and enchants name, each only where db lacks it, so a row db holds
// stays as it is. A zone takes its AreaTable name and an NPC its QuestieDB row; AtlasLoot's are already in db whole.
func AddSourceZonesAndNpcs(db, from *WowDatabase, areaNames map[int32]string, questie *WowDatabase) {
	add := func(sources []*proto.UIItemSource) {
		for _, source := range sources {
			drop := source.GetDrop()
			if drop == nil {
				continue
			}
			if name, ok := areaNames[drop.ZoneId]; ok && db.Zones[drop.ZoneId] == nil {
				db.Zones[drop.ZoneId] = &proto.UIZone{Id: drop.ZoneId, Name: name}
			}
			if npc, ok := questie.Npcs[drop.NpcId]; ok && db.Npcs[drop.NpcId] == nil {
				db.Npcs[drop.NpcId] = npc
			}
		}
	}
	for _, item := range from.Items {
		add(item.Sources)
	}
	for _, enchant := range from.Enchants {
		add(enchant.Sources)
	}
}
