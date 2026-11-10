package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// The codes of the editor's repository checks (sim/siml/SPEC.md §7.4), which fail a file under ui/specs.
var repositoryCodes = []string{"unregistered-name", "gear-item", "gear-slot", "gear-suffix", "gear-enchant"}

// The root of the repository the file at path sits in, relative to the working folder where it can be, so the files
// read under it print relative too; the working folder when no repository holds the file.
func repositoryRootOf(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "."
	}
	root, ok := siml.RepositoryRoot(abs)
	if !ok {
		return "."
	}
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, root); err == nil {
			return rel
		}
	}
	return root
}

// The class folder of a path under ui/specs: hunter for ui/specs/hunter/dps/apls/default.apl.siml, and "" for a file
// in no folder of ui/specs.
func classFolder(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	_, rest, _ := strings.Cut(filepath.ToSlash(abs), "/ui/specs/")
	folder, _, inFolder := strings.Cut(rest, "/")
	if !inFolder {
		return ""
	}
	return folder
}

// What is wrong with the file at path under ui/specs, which reads as f and was analyzed as doc, each
// "path:line:col: what" or "path: what": the editor's repository checks and error lints against catalog, an item.<name>
// outside the slot rows that the database does not hold, and a spec outside its class's folder.
func repositoryProblems(doc *siml.Doc, catalog *siml.DatabaseCatalog, path string, f *siml.File) []string {
	var problems []string
	at := func(p siml.Position) string { return fmt.Sprintf("%s:%d:%d", path, p.Line, p.Col) }
	if info, ok := siml.SpecOf(f.Spec); ok {
		if class := siml.ClassName(info.Class); info.Class != proto.Class_ClassUnknown && classFolder(path) != class {
			problems = append(problems, fmt.Sprintf("%s: spec %s is a %s spec, and the file is not under ui/specs/%s", path, info.Field.Name(), class, class))
		}
	}
	type place struct {
		code  string
		start siml.Position
	}
	parser := map[place]bool{}
	for _, p := range doc.Diagnostics() {
		parser[place{p.Code, p.Start}] = true
	}
	gearItems := map[siml.Position]bool{} // Where the gear check already found an item the database lacks.
	for _, d := range doc.EditorDiagnostics(catalog) {
		if parser[place{d.Code, d.Start}] {
			continue
		}
		if slices.Contains(repositoryCodes, d.Code) || d.Code != "raw-ref" && d.Severity == siml.Error {
			problems = append(problems, fmt.Sprintf("%s: %s", at(d.Start), d.Message))
		}
		if d.Code == "gear-item" {
			gearItems[d.Start] = true
		}
	}
	for _, r := range f.Names {
		if r.Namespace != siml.NamespaceItem || gearItems[r.Start] {
			continue
		}
		if _, ok := catalog.ItemName(r.ID); !ok {
			problems = append(problems, fmt.Sprintf("%s: the database holds no item %d, which %s names", at(r.Start), r.ID, r.Name))
		}
	}
	return problems
}
