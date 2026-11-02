package database

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

// Runs the budget query on a fixture holding only the columns it reads, with made-up budgets.
const randPropPointsFixture = `
CREATE TABLE RandPropPoints (ID INTEGER PRIMARY KEY, DamageReplaceStat INTEGER,
	Epic_0 INTEGER, Epic_1 INTEGER, Epic_2 INTEGER, Epic_3 INTEGER, Epic_4 INTEGER,
	Superior_0 INTEGER, Superior_1 INTEGER, Superior_2 INTEGER, Superior_3 INTEGER, Superior_4 INTEGER,
	Good_0 INTEGER, Good_1 INTEGER, Good_2 INTEGER, Good_3 INTEGER, Good_4 INTEGER);
INSERT INTO RandPropPoints VALUES (60, 0, 50, 40, 30, 20, 10, 45, 35, 25, 15, 5, 44, 34, 24, 14, 4);
`

func TestLegendaryItemsSpendTheEpicBudget(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "budget.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec(randPropPointsFixture); err != nil {
		t.Fatalf("building the fixture: %v", err)
	}

	inputsDir := t.TempDir()
	if _, err := LoadAndWriteRandomPropAllocations(&DBHelper{db: db}, inputsDir); err != nil {
		t.Fatal(err)
	}

	data, err := dbc.ReadGzipFile(filepath.Join(inputsDir, "dbc", "rand_prop_points.json"))
	if err != nil {
		t.Fatal(err)
	}
	var written dbc.RandomPropAllocationsByIlvl
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatal(err)
	}
	budget := written[60]
	if got, want := budget[proto.ItemQuality_ItemQualityLegendary], budget[proto.ItemQuality_ItemQualityEpic]; got != want || got == [5]int32{} {
		t.Errorf("legendary budget = %v, want the epic budget %v", got, want)
	}
}
