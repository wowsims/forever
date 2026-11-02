package database

import (
	"slices"

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

// One item's sources in the order an item in the database lists them. AtlasLoot's rep source gives way to the game
// files' and QuestieDB's drops to AtlasLoot's, which name the instance's NPC with its difficulty.
func (inputs EnchantSourceInputs) itemSources(itemID int) []*proto.UIItemSource {
	var sources []*proto.UIItemSource
	for _, crafted := range inputs.Crafted[itemID] {
		sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: crafted}})
	}
	reps := inputs.Rep[itemID]
	for _, rep := range reps {
		sources = append(sources, &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: rep}})
	}
	hasAtlasDrop := false
	if atlasItem := inputs.Atlas.Items[int32(itemID)]; atlasItem != nil {
		for _, source := range atlasItem.Sources {
			if source.GetRep() != nil && len(reps) > 0 {
				continue
			}
			hasAtlasDrop = hasAtlasDrop || source.GetDrop() != nil
			sources = append(sources, source)
		}
	}
	if questieItem := inputs.Questie.Items[int32(itemID)]; questieItem != nil {
		for _, source := range questieItem.Sources {
			if source.GetDrop() != nil && hasAtlasDrop {
				continue
			}
			sources = append(sources, source)
		}
	}
	return sources
}

// The zones and NPCs the enchants' drop sources name, each only where db lacks it, so a row db holds stays as it is.
// A zone takes its AreaTable name and an NPC its QuestieDB row; AtlasLoot's are already in db whole.
func AddEnchantSourceZonesAndNpcs(db *WowDatabase, enchants map[EnchantDBKey]*proto.UIEnchant, areaNames map[int32]string, questie *WowDatabase) {
	for _, enchant := range enchants {
		for _, source := range enchant.Sources {
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
}
