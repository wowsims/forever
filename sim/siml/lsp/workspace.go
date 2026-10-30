package lsp

import (
	"path/filepath"
	"sync"

	"github.com/wowsims/forever/sim/siml"
)

// The name table every check reads, from the repository the file sits in (`make siml-names`), read again
// whenever it changes on disk; the schema is the one the server is built with.
var tableFile = filepath.Join("sim", "siml", "names.json")

type workspaces struct {
	mu     sync.Mutex
	tables map[string]*siml.NamesFile
}

func newWorkspaces() *workspaces {
	return &workspaces{tables: map[string]*siml.NamesFile{}}
}

// The name table of the repository at root, read again when the file changed since the last read.
func (w *workspaces) names(root string) (*siml.Names, error) {
	w.mu.Lock()
	t := w.tables[root]
	if t == nil {
		t = &siml.NamesFile{Path: filepath.Join(root, tableFile)}
		w.tables[root] = t
	}
	w.mu.Unlock()
	return t.Get()
}
