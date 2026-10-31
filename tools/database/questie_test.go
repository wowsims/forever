package database

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

const questieTestItems = `QuestieDB.itemKeys = {
    ['name'] = 1, -- string
    ['questRewards'] = 2, -- table or nil, quest IDs
    ['vendors'] = 3, -- table or nil, NPC IDs
}

QuestieDB.itemData = [[return {
[100] = {'Conqueror\'s Crown',{8561,9999},nil},
[101] = {"Plain Ring",nil,{5,6}},
[102] = {'Unsold, Unrewarded',nil,nil},
}]]
`

const questieTestQuests = `QuestieDB.questKeys = {
    ['name'] = 1, -- string
    ['zoneOrSort'] = 2, -- int, <0: QuestSort.dbc ID
}

QuestieDB.questData = [[return {
[8561] = {"Conqueror's Crown",-15},
}]]
`

const questieTestNpcs = `QuestieDB.npcKeys = {
    ['name'] = 1, -- string
    ['spawns'] = 2, -- table {[zoneID(int)] = {coordPair(floatVector2D),...},...}
    ['zoneID'] = 3, -- guess as to where this NPC is most common
}

QuestieDB.npcData = [[return {
[5] = {'Grizzle "Rings, Bands" O\'Hare',{[1519]={{25.06,38.2},{-1,.5}}},1519},
[6] = {'Nowhere Vendor',nil,nil},
}]]
`

func TestParseQuestieData(t *testing.T) {
	db, err := parseQuestieData(questieTestItems, questieTestQuests, questieTestNpcs)
	if err != nil {
		t.Fatal(err)
	}

	want := map[int32][]*proto.UIItemSource{
		100: {{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: 8561, Name: "Conqueror's Crown"}}}},
		101: {
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 5, NpcName: `Grizzle "Rings, Bands" O'Hare`, ZoneId: 1519}}},
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 6, NpcName: "Nowhere Vendor"}}},
		},
	}
	if len(db.Items) != len(want) {
		t.Errorf("got %d items, want %d: an item with neither quests nor vendors is left out", len(db.Items), len(want))
	}
	for id, sources := range want {
		item, ok := db.Items[id]
		if !ok {
			t.Errorf("item %d is missing", id)
			continue
		}
		if len(item.Sources) != len(sources) {
			t.Errorf("item %d has %d sources, want %d: a quest with no row is left out", id, len(item.Sources), len(sources))
			continue
		}
		for i := range sources {
			if !googleProto.Equal(item.Sources[i], sources[i]) {
				t.Errorf("item %d source %d = %v, want %v", id, i, item.Sources[i], sources[i])
			}
		}
	}
}

func TestParseQuestieDataRejectsMalformedLua(t *testing.T) {
	for name, items := range map[string]string{
		"unclosed string": `QuestieDB.itemKeys = {['questRewards'] = 2, ['vendors'] = 3}
QuestieDB.itemData = [[return {[1] = {'Broken,nil,nil}}]]`,
		"missing key": `QuestieDB.itemKeys = {['questRewards'] = 2}
QuestieDB.itemData = [[return {}]]`,
		"unexpected word": `QuestieDB.itemKeys = {['questRewards'] = 2, ['vendors'] = 3}
QuestieDB.itemData = [[return {[1] = {'A',nil,true}}]]`,
		"missing separator": `QuestieDB.itemKeys = {['questRewards'] = 2, ['vendors'] = 3}
QuestieDB.itemData = [[return {[1] = {'A' nil}}]]`,
	} {
		if _, err := parseQuestieData(items, questieTestQuests, questieTestNpcs); err == nil {
			t.Errorf("%s: parsed without an error", name)
		}
	}
}
