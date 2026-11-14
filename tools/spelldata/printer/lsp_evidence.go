package printer

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// An evidence id (docs/evidence) and a ruling id (a skill's references/rulings.md), each captured in group 1.
var (
	evidencePattern = regexp.MustCompile(`\b(EV-\d{4})\b`)
	rulingPattern   = regexp.MustCompile(`\b(R-\d{4})\b`)
)

// The evidence file an id names, as it stands on disk.
func (w *workspace) evidenceMarkdown(id string) (string, error) {
	files, _ := filepath.Glob(filepath.Join(w.root, "docs", "evidence", id+"-*.md"))
	if len(files) != 1 {
		return "", fmt.Errorf("docs/evidence holds %d files for %s", len(files), id)
	}
	text, err := os.ReadFile(files[0])
	if err != nil {
		return "", err
	}
	return string(text), nil
}

// The rulings bullet an id heads, dedented, under the skill it sits in. A bullet runs to the next bullet,
// heading or blank line.
func (w *workspace) rulingMarkdown(id string) (string, error) {
	files, _ := filepath.Glob(filepath.Join(w.root, ".github", "skills", "*", "references", "rulings.md"))
	head := "- **" + id + "**"
	for _, file := range files {
		text, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		lines := strings.Split(string(text), "\n")
		for i, line := range lines {
			if !strings.HasPrefix(line, head) {
				continue
			}
			bullet := []string{strings.TrimPrefix(line, "- ")}
			for _, next := range lines[i+1:] {
				if next == "" || strings.HasPrefix(next, "- ") || strings.HasPrefix(next, "#") {
					break
				}
				bullet = append(bullet, strings.TrimSpace(next))
			}
			skill := filepath.Base(filepath.Dir(filepath.Dir(file)))
			return fmt.Sprintf("### %s · %s\n\n%s\n", id, skill, strings.Join(bullet, " ")), nil
		}
	}
	return "", fmt.Errorf("no references/rulings.md bullet carries %s", id)
}
