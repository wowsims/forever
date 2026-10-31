// Starts the SIML language server, `wowsimcli siml lsp`, for an editor, from the repository root, through
// tools/lsplaunch. The server is built with the item database, as the sims it runs need it.
package main

import (
	"os"

	"github.com/wowsims/forever/tools/lsplaunch"
)

func main() {
	os.Exit(lsplaunch.Run(lsplaunch.Server{
		Name:    "wowsims-siml",
		Package: "./cmd/wowsimcli",
		Tags:    []string{"with_db"},
		Args:    []string{"siml", "lsp"},
		LSP:     true,
	}))
}
