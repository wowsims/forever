package database

import (
	"testing"

	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/tooltip"
)

// Every spell the item database names keeps its game files text on the site: its description and its buff text
// compile. Reads the DBC inputs gen_db writes, so it skips where they are absent, as in CI. The update workflows
// run it apart from the other generator tests, so a failure still opens their pull request, which names the spells.
func TestItemSpellTextsCompile(t *testing.T) {
	withDBCInputs(t)
	instance := dbc.GetDBC()

	named := itemDatabaseSpells(instance, assetsdb.Load())
	if len(named) == 0 {
		t.Fatal("the item database names no spell, so this proves nothing")
	}
	_, failed := spellTooltips(instance, tooltip.DBCTooltipDataProvider{DBC: instance}, nil, named, nil, nil)
	for _, f := range failed {
		t.Errorf("no tooltip text for %s", f)
	}
	if len(failed) > 0 {
		t.Log("each form named above needs reading in tools/tooltip before the next gen_db -gen=tooltips")
	}
	t.Logf("%d spells the item database names, %d of whose texts do not compile", len(named), len(failed))
}
