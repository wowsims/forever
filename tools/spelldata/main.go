// Command spelldata prints one game files spell row the way the store carries it, so a bare id in
// hand-written code can be read without grepping the generated table.
//
// Run from the repository root:
//
//	go run ./tools/spelldata 11574           # the row, humanised, with the game files' own literal
//	go run ./tools/spelldata Whirlwind       # every row with that name, then the highest rank
//	go run ./tools/spelldata Rend -all       # every match in full
//	go run ./tools/spelldata 11574 -json     # the same as JSON, for a script or an agent: see README.md
//	go run ./tools/spelldata -family warrior/Execute   # the ladder's ranks, then its highest in full
//	go run ./tools/spelldata -expr 'spellData.Execute.Rank(3)' -package warrior   # the row that reaches
//	go run ./tools/spelldata -expr 'spellData.Execute.Highest().EffectN(1).Average()' -package warrior   # the value that reads
//	go run ./tools/spelldata -config 'spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))' -package warrior   # the config the resolver builds
//	go run ./tools/spelldata -config 'core.SpellModConfig{Kind: core.SpellMod_BonusCrit_Percent, EquipSubclass: dbcenums.ITEM_SUBCLASS_BIT_WEAPON_1H_AXE | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_2H_AXE}'   # a mod's fields and the weapon it needs
//	go run ./tools/spelldata -hover sim/warrior/execute.go 12:40   # the markdown an editor hover shows at line:column, 1-based
//	go run ./tools/spelldata -lsp            # a language server on stdio answering those hovers
//	go run ./tools/spelldata -dump out/      # every class spell in the store, one <class>.txt each
package main

import (
	"fmt"
	"os"

	"github.com/wowsims/forever/tools/spelldata/printer"
)

func main() {
	if err := printer.Run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "spelldata:", err)
		os.Exit(1)
	}
}
