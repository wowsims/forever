package printer

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The class of each game files spell family, which is SpellClassOptions.SpellClassSet.
var familyClasses = map[int32]string{
	3: "mage", 4: "warrior", 5: "warlock", 6: "priest", 7: "druid",
	8: "rogue", 9: "hunter", 10: "paladin", 11: "shaman",
}

// Writes every class spell in the store to <dir>/<class>.txt as its text card, sorted by name and
// id and split by a line of =======, so two dumps diff spell by spell. The ladder calls are left
// out: they say where the sim reads a row, not what the game files changed.
func runDump(dir string) error {
	byClass := map[string][]*spelldata.Spell{}
	for _, s := range spelldata.All() {
		if class, ok := familyClasses[s.ClassFlags.Family]; ok {
			byClass[class] = append(byClass[class], s)
		}
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for class, rows := range byClass {
		slices.SortFunc(rows, func(a, b *spelldata.Spell) int {
			return cmp.Or(strings.Compare(a.Name, b.Name), cmp.Compare(a.ID, b.ID))
		})

		var b strings.Builder
		for i, s := range rows {
			if i > 0 {
				b.WriteString("=======\n")
			}
			c := newCard(s, s.Rank, 0)
			c.Ladder = nil
			c.writeText(&b)
		}
		if err := os.WriteFile(filepath.Join(dir, class+".txt"), []byte(b.String()), 0644); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "spelldata: %d %s spells\n", len(rows), class)
	}
	return nil
}
