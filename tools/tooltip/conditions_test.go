package tooltip

import (
	"testing"

	"github.com/wowsims/forever/tools/database/dbc"
)

// A condition on a spell or an aura no talent rank answers reads the player's expression the conditions hold for
// it, and is false for one they do not hold; without conditions, the provider decides it as the item tooltips do.
func TestSpellAndAuraConditionsReadThePlayer(t *testing.T) {
	const text = "$?s7[Known][Unknown]. $?a8[In][Out]. $?p9[Up][Down]."
	provider := fixtureProvider(fixtureRow{spell: dbc.Spell{ID: 101}})
	if got := segmentJSON(t, compileWith(t, provider, 101, text)); got != `["Known. In. Up."]` {
		t.Errorf("without conditions, %q compiles to %s, want every condition true", text, got)
	}

	provider.Conditions = &Conditions{
		Known: map[int64]Expr{7: PlayerOfClass(5, 11)},
		Auras: map[int64]Expr{8: PlayerInForm(18)},
	}
	want := `[{"if":{"a":[{"a":[{"v":"CLASS"},5],"o":"="},{"a":[{"v":"CLASS"},11],"o":"="}],"o":"or"},"n":["Unknown"],"y":["Known"]},". ",` +
		`{"if":{"a":[{"v":"FORM"},18],"o":"="},"n":["Out"],"y":["In"]},". Down."]`
	if got := segmentJSON(t, compileWith(t, provider, 101, text)); got != want {
		t.Errorf("with conditions, %q compiles to %s, want %s", text, got, want)
	}
}
