package database

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// The subset of CBOR the baked data holds, with strings as byte strings as it writes them.
func encodeCBOR(value any) []byte {
	head := func(major byte, arg int) []byte {
		switch {
		case arg < 24:
			return []byte{major<<5 | byte(arg)}
		case arg < 1<<8:
			return []byte{major<<5 | 24, byte(arg)}
		default:
			return []byte{major<<5 | 25, byte(arg >> 8), byte(arg)}
		}
	}
	switch value := value.(type) {
	case nil:
		return []byte{0xf6}
	case int:
		if value < 0 {
			return head(1, -1-value)
		}
		return head(0, value)
	case string:
		return append(head(2, len(value)), value...)
	case []any:
		out := head(4, len(value))
		for _, element := range value {
			out = append(out, encodeCBOR(element)...)
		}
		return out
	case map[any]any:
		out := head(5, len(value))
		for key, element := range value {
			out = append(append(out, encodeCBOR(key)...), encodeCBOR(element)...)
		}
		return out
	}
	panic(fmt.Sprintf("cannot encode %T", value))
}

func questieValue(value any) string {
	return base64.StdEncoding.EncodeToString(encodeCBOR(value))
}

func questieIDs(ids ...any) string {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	writer.Write(encodeCBOR(ids))
	writer.Close()
	return base64.StdEncoding.EncodeToString(compressed.Bytes())
}

// Splits a value into a "~N~" header and its N parts.
func questieChunks(key, value string, parts int) string {
	lines := []string{fmt.Sprintf("## X-%s: ~%d~", key, parts)}
	size := (len(value) + parts - 1) / parts
	for i := range parts {
		lines = append(lines, fmt.Sprintf("## X-%s-%d: %s", key, i+1, value[min(i*size, len(value)):min((i+1)*size, len(value))]))
	}
	return strings.Join(lines, "\n")
}

const questieTestItemMeta = `local meta = {
  entity = 'Item',
  keys = {
    -- Localized item name.
    ['name'] = 1,
    -- Quest IDs that reward this item.
    ['questRewards'] = 2,
    ['vendors'] = 3, -- NPC IDs for vendors that sell this item.
    ['npcDrops'] = 4, -- NPC IDs that can drop this item.
  },
  types = { [1]='string', [2]='table', [3]='table', [4]='table' },
}`

const questieTestQuestMeta = `local meta = {
  keys = {
    ['name'] = 1,
    ['zoneOrSort'] = 2, -- int, <0: QuestSort.dbc ID
  },
}`

const questieTestNpcMeta = `local meta = {
  keys = {
    ['name'] = 1,
    ['spawns'] = 2, -- table {[zoneID(int)] = {coordPair(floatVector2D),...},...}
    ['zoneID'] = 3,
  },
}`

const questieTestQuestRewards = "## X-Item-100-2: "
const questieTestQuestRow = "## X-Quest-8561-S: "

// Item 100 has a quest with no row, item 101's vendors are chunked and sold by a vendor stored as a list row and one
// with a null zone, item 102 has no table fields, and item 103 has a vendors field its presence bits leave out. Item
// 104 is dropped by the two NPCs that have rows and one that has none, and item 105 by more NPCs than are listed. A
// vendor sells items 106 and 107, whose drop lists name that vendor, the second with more NPCs than are listed.
func questieTestTOC() string {
	return strings.Join([]string{
		"## Interface: 16001",
		"## X-Flavor: Forever",
		questieChunks("Item-IDS", questieIDs(100, 101, 102, 103, 104, 105, 106, 107), 3),
		"## X-Item-106-S: " + questieValue(map[any]any{1: "Sold Sword", "p": 1<<2 | 1<<3}),
		"## X-Item-106-3: " + questieValue([]any{5}),
		"## X-Item-106-4: " + questieValue([]any{5}),
		"## X-Item-107-S: " + questieValue(map[any]any{1: "Formula: Sold Recipe", "p": 1<<2 | 1<<3}),
		"## X-Item-107-3: " + questieValue([]any{5}),
		"## X-Item-107-4: " + questieValue([]any{5, 7, 21, 22, 23, 24}),
		"## X-Item-104-S: " + questieValue(map[any]any{1: "Formula: Enchant Weapon - Crusader", "p": 1 << 3}),
		"## X-Item-104-4: " + questieValue([]any{5, 7, 9999}),
		"## X-Item-105-S: " + questieValue(map[any]any{1: "Plain Sword", "p": 1 << 3}),
		"## X-Item-105-4: " + questieValue([]any{5, 7, 21, 22, 23, 24}),
		"## X-Npc-7-S: " + questieValue([]any{"Scarlet Spellbinder", nil, 28}),
		"## X-Item-100-S: " + questieValue(map[any]any{1: "Conqueror's Crown", "p": 1 << 1}),
		questieTestQuestRewards + questieValue([]any{8561, 9999}),
		"## X-Item-101-S: " + questieValue(map[any]any{1: "Plain Ring", "p": 1 << 2}),
		questieChunks("Item-101-3", questieValue([]any{5, 6}), 2),
		"## X-Item-102-S: " + questieValue([]any{"Unsold, Unrewarded"}),
		"## X-Item-103-S: " + questieValue(map[any]any{1: "Stale Vendors"}),
		"## X-Item-103-3: " + questieValue([]any{5}),
		"## X-Quest-IDS: " + questieIDs(8561),
		questieTestQuestRow + questieValue(map[any]any{1: "Conqueror's Crown", 2: -15}),
		"## X-Npc-IDS: " + questieIDs(5, 6),
		"## X-Npc-5-S: " + questieValue([]any{`Grizzle "Rings, Bands" O'Hare`, nil, 1519}),
		"## X-Npc-6-S: " + questieValue(map[any]any{1: "Nowhere Vendor", 3: nil}),
	}, "\r\n")
}

func TestParseQuestieData(t *testing.T) {
	db, err := parseQuestieData(questieTestTOC(), questieTestItemMeta, questieTestQuestMeta, questieTestNpcMeta)
	if err != nil {
		t.Fatal(err)
	}

	want := map[int32][]*proto.UIItemSource{
		100: {{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: 8561, Name: "Conqueror's Crown"}}}},
		101: {
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 5, NpcName: `Grizzle "Rings, Bands" O'Hare`, ZoneId: 1519}}},
			{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 6, NpcName: "Nowhere Vendor"}}},
		},
		104: {
			{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 5, ZoneId: 1519}}},
			{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 7, ZoneId: 28}}},
		},
		105: {{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{OtherName: WorldDropName}}}},
		106: {{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 5, NpcName: `Grizzle "Rings, Bands" O'Hare`, ZoneId: 1519}}}},
		107: {{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 5, NpcName: `Grizzle "Rings, Bands" O'Hare`, ZoneId: 1519}}}},
	}
	if len(db.Items) != len(want) {
		t.Errorf("got %d items, want %d: an item with no quest, vendor or drop is left out, and so is a table field without its presence bit", len(db.Items), len(want))
	}
	if len(db.Npcs) != 2 || db.Npcs[5].GetName() != `Grizzle "Rings, Bands" O'Hare` || db.Npcs[7].GetZoneId() != 28 {
		t.Errorf("Npcs = %v, want the two NPCs a listed drop names", db.Npcs)
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

func TestParseQuestieDataRejectsMalformedData(t *testing.T) {
	toc := questieTestTOC()
	vendorsPart := "## X-Item-101-3-2: "
	for name, test := range map[string]struct{ toc, itemMeta string }{
		"missing chunk":      {strings.Replace(toc, vendorsPart, "## X-Item-101-3-9: ", 1), questieTestItemMeta},
		"bad base64":         {toc + "\n" + vendorsPart + "!", questieTestItemMeta},
		"truncated CBOR":     {toc + "\n" + questieTestQuestRewards + questieValue([]any{8561, 9999})[:4], questieTestItemMeta},
		"id is not a number": {toc + "\n" + questieTestQuestRewards + questieValue([]any{"8561"}), questieTestItemMeta},
		"name is a number":   {toc + "\n" + questieTestQuestRow + questieValue([]any{8561}), questieTestItemMeta},
		"missing key":        {toc, strings.Replace(questieTestItemMeta, "['vendors'] = 3,", "", 1)},
		"no IDS":             {strings.Replace(toc, "## X-Npc-IDS: ", "## X-Npc-Ids: ", 1), questieTestItemMeta},
		"bad field index":    {toc, strings.Replace(questieTestItemMeta, "['name'] = 1,", "['name'] = 64,", 1)},
	} {
		if _, err := parseQuestieData(test.toc, test.itemMeta, questieTestQuestMeta, questieTestNpcMeta); err == nil {
			t.Errorf("%s: parsed without an error", name)
		}
	}
}

func TestPruneQuestieSources(t *testing.T) {
	drop := func(npcID, zoneID int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: npcID, ZoneId: zoneID}}}
	}
	soldBy := func(zoneID int32) *proto.UIItemSource {
		return &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 1, ZoneId: zoneID}}}
	}
	quest := &proto.UIItemSource{Source: &proto.UIItemSource_Quest{Quest: &proto.QuestSource{Id: 1, Name: "Bounty"}}}
	worldDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{OtherName: WorldDropName}}}
	vendor, openWorld, inDungeon, inRaid := soldBy(28), drop(1, 28), drop(2, 2057), drop(3, 2717)

	questie := NewWowDatabase()
	questie.Items[1] = &proto.UIItem{Id: 1, Sources: []*proto.UIItemSource{quest, vendor, soldBy(1585), openWorld, inDungeon, inRaid, drop(4, 1585), drop(5, 0), worldDrop}}
	PruneQuestieSources(questie, map[int32]bool{2717: true}, map[int32]bool{2057: true}, map[int32]bool{28: true, 2057: true, 2717: true})

	want := []*proto.UIItemSource{quest, vendor, openWorld, inDungeon, inRaid, worldDrop}
	got := questie.Items[1].Sources
	if len(got) != len(want) {
		t.Fatalf("sources = %v, want the quest, the vendor and NPCs in known zones, and the world drop", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source %d = %v, want %v", i, got[i], want[i])
		}
	}
	for source, difficulty := range map[*proto.UIItemSource]proto.DungeonDifficulty{
		openWorld: proto.DungeonDifficulty_DifficultyUnknown,
		inDungeon: proto.DungeonDifficulty_DifficultyDungeon,
		inRaid:    proto.DungeonDifficulty_DifficultyRaid,
		worldDrop: proto.DungeonDifficulty_DifficultyUnknown,
	} {
		if source.GetDrop().Difficulty != difficulty {
			t.Errorf("%v has difficulty %v, want %v", source, source.GetDrop().Difficulty, difficulty)
		}
	}
}

// An item keeps what it has. An AtlasLoot drop keeps QuestieDB's drops out and not its vendors, and an item QuestieDB
// has no row for stays as it was, so nothing makes a world drop of the absence of data.
func TestAppendQuestieSources(t *testing.T) {
	atlasDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 1, ZoneId: 2017, Difficulty: proto.DungeonDifficulty_DifficultyDungeon}}}
	questieDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{NpcId: 2, ZoneId: 28}}}
	worldDrop := &proto.UIItemSource{Source: &proto.UIItemSource_Drop{Drop: &proto.DropSource{OtherName: WorldDropName}}}
	crafted := &proto.UIItemSource{Source: &proto.UIItemSource_Crafted{Crafted: &proto.CraftedSource{SpellId: 5}}}
	questieVendor := &proto.UIItemSource{Source: &proto.UIItemSource_SoldBy{SoldBy: &proto.SoldBySource{NpcId: 4}}}

	db, questie := NewWowDatabase(), NewWowDatabase()
	db.Items[1] = &proto.UIItem{Id: 1, Sources: []*proto.UIItemSource{atlasDrop}}
	db.Items[2] = &proto.UIItem{Id: 2, Sources: []*proto.UIItemSource{crafted}}
	db.Items[3] = &proto.UIItem{Id: 3}
	db.Items[4] = &proto.UIItem{Id: 4}
	db.Items[5] = &proto.UIItem{Id: 5, Sources: []*proto.UIItemSource{atlasDrop}}
	questie.Items[1] = &proto.UIItem{Id: 1, Sources: []*proto.UIItemSource{questieVendor}}
	questie.Items[2] = &proto.UIItem{Id: 2, Sources: []*proto.UIItemSource{questieDrop}}
	questie.Items[3] = &proto.UIItem{Id: 3, Sources: []*proto.UIItemSource{worldDrop}}
	questie.Items[5] = &proto.UIItem{Id: 5, Sources: []*proto.UIItemSource{worldDrop}}
	questie.Items[99] = &proto.UIItem{Id: 99, Sources: []*proto.UIItemSource{worldDrop}}
	AppendQuestieSources(db, questie)

	for id, want := range map[int32][]*proto.UIItemSource{
		1: {atlasDrop, questieVendor},
		2: {crafted, questieDrop},
		3: {worldDrop},
		4: nil,
		5: {atlasDrop},
	} {
		got := db.Items[id].Sources
		if len(got) != len(want) {
			t.Errorf("item %d sources = %v, want %v", id, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("item %d source %d = %v, want %v", id, i, got[i], want[i])
			}
		}
	}
	if len(db.Items) != 5 {
		t.Errorf("%d items, want the 5 the database has", len(db.Items))
	}
}
