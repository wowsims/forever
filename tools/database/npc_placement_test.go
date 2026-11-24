package database

import (
	"os"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools"
	googleProto "google.golang.org/protobuf/proto"
)

// Rows as the game files hold them: Dalaran's map sits on Eastern Kingdoms (0), Shaper's Terrace on Kalimdor (1), the
// Scarlet Enclave and Karazhan Crypts are Season of Discovery maps with no Forever dungeon difficulty, Molten Core is
// a raid, and Blackwing Lair a raid given that difficulty here.
func testNpcPlacement() *NpcPlacement {
	return newNpcPlacement(
		[]instanceMapRow{
			{ID: 36, Name: "Deadmines", InstanceType: 1, CosmeticParentMapID: -1, ForeverDifficulty: true},
			{ID: 230, Name: "Blackrock Depths", InstanceType: 1, AreaTableID: 1584, CosmeticParentMapID: -1, ForeverDifficulty: true},
			{ID: 329, Name: "Stratholme", InstanceType: 1, CosmeticParentMapID: -1, ForeverDifficulty: true},
			{ID: 409, Name: "Molten Core", InstanceType: 2, AreaTableID: 2717, CosmeticParentMapID: -1},
			{ID: 469, Name: "Blackwing Lair", InstanceType: 2, AreaTableID: 2677, CosmeticParentMapID: -1, ForeverDifficulty: true},
			{ID: 2856, Name: "Scarlet Enclave", InstanceType: 2, CosmeticParentMapID: -1},
			{ID: 2875, Name: "Karazhan Crypts", InstanceType: 1, AreaTableID: 16074, CosmeticParentMapID: -1},
			{ID: 2959, Name: "City of Dalaran", InstanceType: 1, AreaTableID: 16544, CosmeticParentMapID: 0, ForeverDifficulty: true},
			{ID: 2998, Name: "Excavation Site: Wetlands", InstanceType: 1, CosmeticParentMapID: -1, ForeverDifficulty: true},
			{ID: 3001, Name: "The Shaper's Terrace", InstanceType: 1, CosmeticParentMapID: 1, ForeverDifficulty: true},
		},
		[]instanceAreaRow{
			{ID: 11, Name: "Wetlands", ContinentID: 0},
			{ID: 36, Name: "Alterac Mountains", ContinentID: 0},
			{ID: 331, Name: "Ashenvale", ContinentID: 1},
			{ID: 1581, Name: "The Deadmines", ContinentID: 36},
			{ID: 1584, Name: "Blackrock Depths", ContinentID: 0},
			{ID: 2017, Name: "Stratholme", ContinentID: 329},
			{ID: 2717, Name: "Molten Core", ContinentID: 409},
			{ID: 16074, Name: "Karazhan Crypts", ContinentID: 2875},
			{ID: 16236, Name: "Scarlet Enclave", ContinentID: 2856},
			{ID: 16544, Name: "City of Dalaran", ContinentID: 2959},
			{ID: 16560, Name: "City of Dalaran", ContinentID: 2959},
			{ID: 16619, Name: "Excavation Site - Bael Modan", ContinentID: 3001},
			{ID: 16667, Name: "The Shaper's Terrace", ContinentID: 3001},
			{ID: 16668, Name: "Terrace Steps", ContinentID: 3001, ParentID: 16667},
			{ID: 16732, Name: "Excavation Site: Wetlands", ContinentID: 2998},
		},
		[]encounterRow{
			{Name: "Shade of the Archmage", MapID: 2959},
			{Name: "Lord Roccor", MapID: 230},
			{Name: "Balnazzar", MapID: 329},
			{Name: "Balnazzar", MapID: 2856},
			{Name: "Highland Horror", MapID: 2998},
			{Name: "Titanosaur", MapID: 3001},
			{Name: "Dark Rider", MapID: 2875},
			{Name: "Ragnaros", MapID: 409},
			{Name: "Nefarian", MapID: 469},
			{Name: "Edwin VanCleef", MapID: 36},
			{Name: "Mr. Smite", MapID: 36},
			{Name: "Mr. Smite", MapID: 230},
		},
	)
}

func TestNpcPlacementZone(t *testing.T) {
	placement := testNpcPlacement()
	for _, test := range []struct {
		why       string
		name      string
		zone      int32
		questMaps []int32
		want      int32
	}{
		{"an encounter on a map whose cosmetic parent holds the NPC's world zone", "Shade of the Archmage", 36, nil, 16544},
		{"a quest in the instance lists the NPC as a kill", "Mana Echo", 36, []int32{2959}, 16544},
		{"a zone not in AreaTable", "Lord Roccor", 1585, nil, 1584},
		{"already in the instance", "Lord Roccor", 1584, nil, 1584},
		{"a world zone on a map with no cosmetic parent", "Highland Horror", 11, nil, 11},
		{"a world zone on another continent than the cosmetic parent", "Titanosaur", 36, nil, 36},
		{"an encounter on two maps", "Balnazzar", 0, nil, 0},
		{"an encounter on two maps, one of them the NPC's", "Balnazzar", 2017, nil, 2017},
		{"no zone, a live dungeon: the top-level area named as the map", "Titanosaur", 0, nil, 16667},
		{"no zone, a dungeon with no Forever difficulty", "Dark Rider", 0, nil, 0},
		{"no zone, a raid", "Ragnaros", 0, nil, 0},
		{"no zone, a raid with the Forever dungeon difficulty", "Nefarian", 0, nil, 0},
		{"no area named as the map", "Edwin VanCleef", 0, nil, 0},
		{"an encounter on two maps, one with no zone to move to", "Mr. Smite", 0, nil, 0},
		{"another instance's zone on the cosmetic parent's continent", "Shade of the Archmage", 1584, nil, 1584},
		{"no encounter and no quest", "Dalaran Guard", 36, nil, 36},
	} {
		if got := placement.zone(test.name, test.zone, test.questMaps); got != test.want {
			t.Errorf("%s: %s in zone %d placed in %d, want %d", test.why, test.name, test.zone, got, test.want)
		}
	}
}

func TestNpcPlacementInstanceOf(t *testing.T) {
	placement := testNpcPlacement()
	for zone, want := range map[int32]int32{16544: 2959, 16560: 2959, 2017: 329, 1584: 230, 36: 0, 1585: 0} {
		if got, _ := placement.instanceOf(zone); got != want {
			t.Errorf("zone %d is in instance %d, want %d", zone, got, want)
		}
	}
}

// Mana Echo is a kill of a quest in City of Dalaran, Shade of the Archmage an encounter there with no zone, and Dalaran
// Guard neither.
func TestParseQuestieDataPlacesNpcsInTheirInstance(t *testing.T) {
	toc := strings.Join([]string{
		"## X-Item-IDS: " + questieIDs(200),
		"## X-Item-200-S: " + questieValue(map[any]any{1: "Arcane Mote", "p": 1 << 3}),
		"## X-Item-200-4: " + questieValue([]any{61, 62, 63}),
		"## X-Quest-IDS: " + questieIDs(92457, 92458),
		"## X-Quest-92457-S: " + questieValue(map[any]any{1: "Starving Arcane", 2: 16544, "p": 1 << 3}),
		"## X-Quest-92457-4: " + questieValue([]any{[]any{[]any{61, "Mana Echo slain"}}}),
		"## X-Quest-92458-S: " + questieValue(map[any]any{1: "Guard Duty", 2: 36, "p": 1 << 3}),
		"## X-Quest-92458-4: " + questieValue(map[any]any{1: []any{[]any{63}}}),
		"## X-Npc-IDS: " + questieIDs(61, 62, 63),
		"## X-Npc-61-S: " + questieValue([]any{"Mana Echo", nil, 36}),
		"## X-Npc-62-S: " + questieValue(map[any]any{1: "Shade of the Archmage", 3: nil}),
		"## X-Npc-63-S: " + questieValue([]any{"Dalaran Guard", nil, 36}),
	}, "\n")
	db, _, err := parseQuestieData(toc, questieTestItemMeta, questieTestQuestMeta, questieTestNpcMeta, testNpcPlacement())
	if err != nil {
		t.Fatal(err)
	}
	want := []*proto.UIItemSource{dropSource(61, 16544), dropSource(62, 16544), dropSource(63, 36)}
	got := db.Items[200].GetSources()
	if len(got) != len(want) {
		t.Fatalf("sources = %v, want %v", got, want)
	}
	for i := range want {
		if !googleProto.Equal(got[i], want[i]) {
			t.Errorf("source %d = %v, want %v", i, got[i], want[i])
		}
	}
	if db.Npcs[61].GetZoneId() != 16544 || db.Npcs[62].GetZoneId() != 16544 {
		t.Errorf("Npcs = %v, want Mana Echo and Shade of the Archmage in 16544", db.Npcs)
	}
}

// Lord Roccor is a Blackrock Depths encounter in 1585, a zone AreaTable lacks, so Verek, in 1585 with no encounter or
// quest, follows him to 1584. Nothing places the NPC in 10074, so it keeps that zone.
func TestParseQuestieDataMovesAZoneWithItsPlacedNpcs(t *testing.T) {
	toc := strings.Join([]string{
		"## X-Item-IDS: " + questieIDs(201),
		"## X-Item-201-S: " + questieValue(map[any]any{1: "Verek's Loyal Wraps", "p": 1 << 3}),
		"## X-Item-201-4: " + questieValue([]any{64, 65, 66}),
		"## X-Quest-IDS: " + questieIDs(),
		"## X-Npc-IDS: " + questieIDs(64, 65, 66),
		"## X-Npc-64-S: " + questieValue([]any{"Lord Roccor", nil, 1585}),
		"## X-Npc-65-S: " + questieValue([]any{"Verek", nil, 1585}),
		"## X-Npc-66-S: " + questieValue([]any{"Unplaced", nil, 10074}),
	}, "\n")
	db, _, err := parseQuestieData(toc, questieTestItemMeta, questieTestQuestMeta, questieTestNpcMeta, testNpcPlacement())
	if err != nil {
		t.Fatal(err)
	}
	want := []*proto.UIItemSource{dropSource(64, 1584), dropSource(65, 1584), dropSource(66, 10074)}
	got := db.Items[201].GetSources()
	if len(got) != len(want) {
		t.Fatalf("sources = %v, want %v", got, want)
	}
	for i := range want {
		if !googleProto.Equal(got[i], want[i]) {
			t.Errorf("source %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// Lord Roccor moves from 1585 to 1584 and Ragnaros to 2717, so 1585 stands for no other zone and Verek keeps it.
// Titanosaur, with no zone, joins The Shaper's Terrace, and the other NPC with no zone stays out of it.
func TestParseQuestieDataKeepsAZoneWhoseNpcsMoveApart(t *testing.T) {
	toc := strings.Join([]string{
		"## X-Item-IDS: " + questieIDs(202),
		"## X-Item-202-S: " + questieValue(map[any]any{1: "Verek's Loyal Wraps", "p": 1 << 3}),
		"## X-Item-202-4: " + questieValue([]any{67, 68, 69, 70, 71}),
		"## X-Quest-IDS: " + questieIDs(),
		"## X-Npc-IDS: " + questieIDs(67, 68, 69, 70, 71),
		"## X-Npc-67-S: " + questieValue([]any{"Lord Roccor", nil, 1585}),
		"## X-Npc-68-S: " + questieValue([]any{"Ragnaros", nil, 1585}),
		"## X-Npc-69-S: " + questieValue([]any{"Verek", nil, 1585}),
		"## X-Npc-70-S: " + questieValue([]any{"Titanosaur"}),
		"## X-Npc-71-S: " + questieValue([]any{"Unplaced"}),
	}, "\n")
	db, _, err := parseQuestieData(toc, questieTestItemMeta, questieTestQuestMeta, questieTestNpcMeta, testNpcPlacement())
	if err != nil {
		t.Fatal(err)
	}
	want := []*proto.UIItemSource{dropSource(67, 1584), dropSource(68, 2717), dropSource(69, 1585), dropSource(70, 16667)}
	got := db.Items[202].GetSources()
	if len(got) != len(want) {
		t.Fatalf("sources = %v, want %v", got, want)
	}
	for i := range want {
		if !googleProto.Equal(got[i], want[i]) {
			t.Errorf("source %d = %v, want %v", i, got[i], want[i])
		}
	}
}

// Placing the committed NPCs again moves none of them. The zone stand-in needs QuestieDB's other NPCs, so this cannot
// see it; TestParseQuestieDataMovesAZoneWithItsPlacedNpcs covers it. Skips without tools/database/wowsims.db.
func TestCommittedQuestieNpcsKeepTheirPlacement(t *testing.T) {
	DatabasePath = "wowsims.db"
	if _, err := os.Stat(DatabasePath); err != nil {
		t.Skipf("no game files database at %s", DatabasePath)
	}
	helper, err := NewDBHelper()
	if err != nil {
		t.Fatal(err)
	}
	defer helper.Close()
	placement, err := LoadNpcPlacement(helper)
	if err != nil {
		t.Fatal(err)
	}
	questie := ReadDatabaseFromJson(tools.ReadFile("../../assets/db_inputs/questie_db.json"))
	for _, npc := range questie.Npcs {
		if got := placement.zone(npc.Name, npc.ZoneId, nil); got != npc.ZoneId {
			t.Errorf("%s (%d) in zone %d places in %d; run gen_db -gen=questie", npc.Name, npc.Id, npc.ZoneId, got)
		}
	}
}
