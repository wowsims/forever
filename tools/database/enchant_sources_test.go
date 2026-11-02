package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func soldBySource(npcID int32) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: npcID, NpcName: "Vendor", ZoneId: 1519}}}
}

func dropSource(npcID, zoneID int32) *proto.UIItemSource {
	return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: npcID, ZoneId: zoneID}}}
}

func itemWithSources(id int32, sources ...*proto.UIItemSource) *proto.UIItem {
	return &proto.UIItem{Id: id, Sources: sources}
}

func TestAddEnchantSources(t *testing.T) {
	crafted := &proto.CraftedSource{Profession: proto.Profession_Leatherworking, SpellId: 2152}
	rep := &proto.RepSource{RepFactionId: 59, RepLevel: proto.RepLevel_RepLevelHonored}
	atlasDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 1, ZoneId: 2017, Difficulty: proto.DungeonDifficulty_DifficultyDungeon}}}
	atlasRep := &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: &proto.RepSource{RepFactionId: 59}}}

	inputs := EnchantSourceInputs{
		Crafted: map[int][]*proto.CraftedSource{10: {crafted}},
		Rep:     map[int][]*proto.RepSource{50: {rep}},
		Atlas:   NewWowDatabase(),
		Questie: NewWowDatabase(),
	}
	inputs.Atlas.Items[40] = itemWithSources(40, atlasDrop)
	inputs.Atlas.Items[50] = itemWithSources(50, atlasRep)
	inputs.Questie.Items[20] = itemWithSources(20, soldBySource(7))
	inputs.Questie.Items[30] = itemWithSources(30, soldBySource(8))
	inputs.Questie.Items[31] = itemWithSources(31, soldBySource(8))
	inputs.Questie.Items[32] = itemWithSources(32, dropSource(9, 28))
	inputs.Questie.Items[40] = itemWithSources(40, dropSource(1, 1585))
	inputs.Atlas.Items[41] = itemWithSources(41, atlasDrop)
	inputs.Questie.Items[41] = itemWithSources(41, soldBySource(8))

	repSource := &proto.UIItemSource{Source: &proto.UIItemSource_Rep{Rep: rep}}
	craftedSource := &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: crafted}}
	for _, test := range []struct {
		name    string
		itemID  int32
		recipes []int
		want    []*proto.UIItemSource
	}{
		{name: "its own item", itemID: 10, want: []*proto.UIItemSource{craftedSource}},
		{name: "its own item is the recipe", itemID: 20, recipes: []int{20}, want: []*proto.UIItemSource{soldBySource(7)}},
		{name: "a scroll whose recipe is another item", itemID: 99, recipes: []int{20}, want: []*proto.UIItemSource{soldBySource(7)}},
		{name: "a source two recipes share comes once", itemID: 99, recipes: []int{30, 31}, want: []*proto.UIItemSource{soldBySource(8)}},
		{name: "a recipe no vendor sells has its drop", itemID: 99, recipes: []int{32}, want: []*proto.UIItemSource{dropSource(9, 28)}},
		{name: "own item before recipe", itemID: 10, recipes: []int{20}, want: []*proto.UIItemSource{craftedSource, soldBySource(7)}},
		{name: "AtlasLoot's drop stands in for QuestieDB's", itemID: 40, want: []*proto.UIItemSource{atlasDrop}},
		{name: "a sold item with an AtlasLoot drop keeps both", itemID: 41, want: []*proto.UIItemSource{atlasDrop, soldBySource(8)}},
		{name: "the game files' rep stands in for AtlasLoot's", itemID: 50, want: []*proto.UIItemSource{repSource}},
		{name: "no source anywhere", itemID: 99},
		{name: "no item", itemID: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			enchant := &proto.UIEnchant{EffectId: 1, ItemId: test.itemID, Name: test.name}
			key := EnchantToDBKey(enchant)
			AddEnchantSources(map[EnchantDBKey]*proto.UIEnchant{key: enchant}, map[EnchantDBKey][]int{key: test.recipes}, inputs)

			if len(enchant.Sources) != len(test.want) {
				t.Fatalf("sources = %v, want %v", enchant.Sources, test.want)
			}
			for i := range test.want {
				if !googleProto.Equal(enchant.Sources[i], test.want[i]) {
					t.Errorf("source %d = %v, want %v", i, enchant.Sources[i], test.want[i])
				}
			}
		})
	}
}

func TestAddSourceZonesAndNpcs(t *testing.T) {
	db := NewWowDatabase()
	db.Zones[28] = &proto.UIZone{Id: 28, Name: "Held Zone"}
	db.Npcs[1] = &proto.UINPC{Id: 1, Name: "Held NPC", ZoneId: 28}
	questie := NewWowDatabase()
	questie.Npcs[1] = &proto.UINPC{Id: 1, Name: "QuestieDB's NPC", ZoneId: 139}
	questie.Npcs[2] = &proto.UINPC{Id: 2, Name: "Scarlet Archmage", ZoneId: 139}
	questie.Npcs[3] = &proto.UINPC{Id: 3, Name: "No Source Names Me", ZoneId: 139}

	worldDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{OtherName: WorldDropName}}}
	enchant := &proto.UIEnchant{EffectId: 1, Sources: []*proto.UIItemSource{dropSource(1, 28), dropSource(2, 139), worldDrop, soldBySource(3)}}
	questie.Npcs[4] = &proto.UINPC{Id: 4, Name: "On A Leftover Item", ZoneId: 28}
	from := NewWowDatabase()
	from.Enchants[EnchantToDBKey(enchant)] = enchant
	from.Items[10] = &proto.UIItem{Id: 10, Sources: []*proto.UIItemSource{dropSource(4, 28)}}
	AddSourceZonesAndNpcs(db, from, map[int32]string{28: "AreaTable", 139: "Eastern Plaguelands"}, questie)

	if db.Zones[28].Name != "Held Zone" || db.Npcs[1].Name != "Held NPC" {
		t.Errorf("a row db holds moved: zone %v, npc %v", db.Zones[28], db.Npcs[1])
	}
	if db.Zones[139].GetName() != "Eastern Plaguelands" || db.Npcs[2].GetName() != "Scarlet Archmage" {
		t.Errorf("the missing rows did not land: zone %v, npc %v", db.Zones[139], db.Npcs[2])
	}
	if len(db.Zones) != 2 || len(db.Npcs) != 3 {
		t.Errorf("zones %v and npcs %v hold more than the drops name", db.Zones, db.Npcs)
	}
}
