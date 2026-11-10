package lsp

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/siml"
)

// A rotation of 2000 priority rows, each with a condition.
func longRotation() string {
	var b strings.Builder
	b.WriteString("siml 1\nspec hunter\n\npriority:\n")
	spells := []string{"multi_shot", "arcane_shot", "serpent_sting", "aimed_shot", "raptor_strike"}
	for i := range 2000 {
		fmt.Fprintf(&b, "  cast %s if current_time >= %ds and not target.aura(hunters_mark).is_active\n", spells[i%len(spells)], i%300)
	}
	return b.String()
}

// A check of the 2000-row rotation as the server makes it: read, then the parser's diagnostics and the checks.
func BenchmarkDiagnose(b *testing.B) {
	s := &server{ws: newWorkspaces(), docs: map[string]*document{}, analyses: map[string]*analysis{}, timers: map[string]*time.Timer{}}
	path := filepath.Join(repository, "ui", "specs", "hunter", "dps", "apls", "long.apl.siml")
	uri := pathURI(path)
	text := longRotation()
	siml.CatalogAt(repository)
	for i := range b.N {
		s.docs[uri] = &document{uri: uri, path: path, version: i + 1, text: text}
		s.diagnosticsLocked(s.analyzeLocked(uri), siml.CatalogAt(repository))
	}
}
