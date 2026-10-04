package database

// Runs the rep item query on a fixture shaped like the game files, holding only the columns it reads.
// The factions are made up: each one's race mask names the races of the case it tests.

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestRepItemsReadFactionStandingAndSide(t *testing.T) {
	cases := []struct {
		name          string
		races         dbc.Race
		minReputation int
		side          proto.Faction
		level         proto.RepLevel
	}{
		{"alliance races only", dbc.Human | dbc.Dwarf, 7, proto.Faction_Alliance, proto.RepLevel_RepLevelExalted},
		{"horde races only", dbc.Orc | dbc.Troll, 5, proto.Faction_Horde, proto.RepLevel_RepLevelHonored},
		{"races of both sides", dbc.Human | dbc.Orc, 6, proto.Faction_Unknown, proto.RepLevel_RepLevelRevered},
		{"every race", -1, 3, proto.Faction_Unknown, proto.RepLevel_RepLevelNeutral},
		{"no race", 0, 4, proto.Faction_Unknown, proto.RepLevel_RepLevelFriendly},
		{"high order skyborne only", dbc.HighOrderSkyborne, 5, proto.Faction_Alliance, proto.RepLevel_RepLevelHonored},
		{"windshaper skyborne only", dbc.WindshaperSkyborne, 5, proto.Faction_Horde, proto.RepLevel_RepLevelHonored},
	}

	var factions, items []string
	for i, c := range cases {
		factionID, itemID := i+1, 100+i
		mask := int64(c.races)
		factions = append(factions, fmt.Sprintf("(%d, '%s', %d, %d)", factionID, c.name, int32(mask), int32(mask>>32)))
		items = append(items, fmt.Sprintf("(%d, %d, %d)", itemID, factionID, c.minReputation))
	}
	// Neither an item without a faction nor one whose faction has no row is a rep item.
	items = append(items, "(900, 0, 0)", "(901, 999, 5)")
	fixture := fmt.Sprintf(`
CREATE TABLE Faction (ID INTEGER PRIMARY KEY, Name_lang TEXT, ReputationRaceMasks0_0 INTEGER, ReputationRaceMasks0_1 INTEGER);
CREATE TABLE ItemSparse (ID INTEGER PRIMARY KEY, MinFactionID INTEGER, MinReputation INTEGER);
INSERT INTO Faction (ID, Name_lang, ReputationRaceMasks0_0, ReputationRaceMasks0_1) VALUES %s;
INSERT INTO ItemSparse (ID, MinFactionID, MinReputation) VALUES %s;`, strings.Join(factions, ", "), strings.Join(items, ", "))

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "factions.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(fixture); err != nil {
		t.Fatalf("building the fixture: %v", err)
	}
	sources, names, err := LoadRepItems(&DBHelper{db: db})
	if err != nil {
		t.Fatal(err)
	}

	if len(sources) != len(cases) {
		t.Errorf("got rep sources for %d items, want %d", len(sources), len(cases))
	}
	for i, c := range cases {
		factionID, itemID := int32(i+1), 100+i
		got := sources[itemID]
		if len(got) != 1 {
			t.Errorf("%s: got %d rep sources, want 1", c.name, len(got))
			continue
		}
		if got[0].RepFactionId != factionID || got[0].RepLevel != c.level || got[0].FactionId != c.side {
			t.Errorf("%s: got faction %d %v side %v, want faction %d %v side %v", c.name,
				got[0].RepFactionId, got[0].RepLevel, got[0].FactionId, factionID, c.level, c.side)
		}
		if names[factionID] != c.name {
			t.Errorf("%s: faction name %q", c.name, names[factionID])
		}
	}
}
