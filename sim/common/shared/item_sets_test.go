package shared

import (
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// A proc trigger row registers when its rate is stated and the spell it fires deals damage at once, fills
// the wearer's bar or puts auras on the wearer alone; any other row goes to the parse.
func TestSetBonusUnsupportedReadsProcRowsOffTheirListenerAndFiredSpell(t *testing.T) {
	for _, c := range []struct {
		row  int32
		want string // A reason the row has to carry, or "" for none at all.
	}{
		{18979, ""},   // Stormshroud Armor 2: 10% on melee hits, 18980's Nature damage
		{23863, ""},   // Stormshroud Armor 3: 4% on melee hits, 23864's energy
		{9233, ""},    // Volcanic Armor 3: 10% on melee hits, 9057's Fire damage
		{1299743, ""}, // Deathbone Guardian 3: 15% when struck, 1299746's spell damage and healing
		{9142, ""},    // Stormshroud Armor 4: +14 attack power, which the parse reads
		{26107, "names class family 7 and none of its spells"},
		{17332, "17333 puts auras on an enemy"},
		{1299737, spelldata.ReasonStatesNoRate},
	} {
		got := SetBonusUnsupported(spelldata.MustFind(c.row))
		switch {
		case c.want == "" && len(got) > 0:
			t.Errorf("%d is refused for %q, want it to register", c.row, got)
		case c.want != "" && !slices.ContainsFunc(got, func(reason string) bool { return strings.Contains(reason, c.want) }):
			t.Errorf("%d is refused for %q, want a reason naming %q", c.row, got, c.want)
		}
	}
}
