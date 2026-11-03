// Starts tools/spelldata for an editor, from the repository root, through tools/lsplaunch.
package main

import (
	"os"
	"slices"

	"github.com/wowsims/forever/tools/lsplaunch"
)

func main() {
	args := os.Args[1:]
	os.Exit(lsplaunch.Run(lsplaunch.Server{
		Name:    "wowsims-spelldata",
		Package: "./tools/spelldata",
		Args:    args,
		LSP:     slices.Contains(args, "-lsp"),
	}))
}
