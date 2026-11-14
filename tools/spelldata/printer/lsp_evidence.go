package printer

import (
	"fmt"
	"path/filepath"
	"regexp"
)

// An evidence id (docs/evidence) or a ruling id (docs/rulings), captured in group 1.
var idFilePattern = regexp.MustCompile(`\b((?:EV|R)-\d{4})\b`)

// The file an evidence or ruling id names, as the editor holds it, else as it stands on disk.
func (w *workspace) idFileMarkdown(id string) (string, error) {
	folder := "evidence"
	if id[0] == 'R' {
		folder = "rulings"
	}
	files, _ := filepath.Glob(filepath.Join(w.root, "docs", folder, id+"-*.md"))
	if len(files) != 1 {
		return "", fmt.Errorf("docs/%s holds %d files for %s", folder, len(files), id)
	}
	return w.text(files[0])
}
