package siml

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// What the language services read a Doc with: its syntax tree, its spec and class, whether it lies under ui/specs,
// and the files its use lines name.

// The syntax tree, nil when the schema does not read.
func (d *Doc) tree() *syntaxTree {
	if d.typed == nil {
		return nil
	}
	return d.typed.tree
}

// The spec the spec line names, nil when there is none or it names no spec.
func (d *Doc) specInfo() *SpecInfo {
	if d.typed == nil {
		return nil
	}
	return d.typed.spec
}

// The spec line's field name, such as hunter; "" when the file names no spec.
func (d *Doc) specName() string {
	if s := d.specInfo(); s != nil {
		return string(s.Field.Name())
	}
	return ""
}

// The class the file's names are read with.
func (d *Doc) class() proto.Class {
	if s := d.specInfo(); s != nil {
		return s.Class
	}
	return proto.Class_ClassUnknown
}

// The repository root above a file: the first folder up that holds sim/siml/names.json.
func RepositoryRoot(file string) (string, bool) {
	for dir := filepath.Dir(file); ; dir = filepath.Dir(dir) {
		if info, err := os.Stat(filepath.Join(dir, "sim", "siml", "names.json")); err == nil && info.Mode().IsRegular() {
			return dir, true
		}
		if filepath.Dir(dir) == dir {
			return "", false
		}
	}
}

// Whether the file lies under ui/specs of the repository it sits in, where warnings are errors and raw refs are
// reported (SPEC.md §7.3, §13).
func UnderSpecs(path string) bool {
	if path == "" {
		return false
	}
	file, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	root, ok := RepositoryRoot(file)
	if !ok {
		return false
	}
	rel, err := filepath.Rel(root, file)
	parts := strings.Split(filepath.ToSlash(rel), "/")
	return err == nil && len(parts) > 2 && parts[0] == "ui" && parts[1] == "specs"
}

// Another file's text, read as Analyze reads this one, with the same names and opener.
func (d *Doc) analyze(path, text string) *Doc { return Analyze(path, []byte(text), d.names, d.open) }

// The text of the file at path: through the doc's opener, which reads an open editor's text, else nothing.
func (d *Doc) readUsed(path string) (string, bool) {
	if d.open == nil {
		return "", false
	}
	text, err := d.open(path)
	if err != nil {
		return "", false
	}
	return string(text), true
}

// The start of the line holding offset.
func lineStartOf(text string, offset int) int { return strings.LastIndexByte(text[:offset], '\n') + 1 }

func (s span) holds(offset int) bool { return s.start <= offset && offset <= s.end }

// Whether two paths name one file, whichever separator each is written with, without case on Windows.
func SamePath(a, b string) bool {
	a, b = filepath.Clean(strings.ReplaceAll(a, `\`, "/")), filepath.Clean(strings.ReplaceAll(b, `\`, "/"))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// The section each statement belongs to; a header belongs to its own section.
func sectionsByStatement(tree *syntaxTree) map[stmt]*section {
	out := map[stmt]*section{}
	for _, s := range tree.sections {
		if s.header != nil {
			out[s.header] = s
		}
		for _, b := range s.body {
			out[b] = s
		}
	}
	return out
}

// A section's text: from its header, or its first statement, to the end of its last statement.
func sectionSpan(s *section) span {
	var parts []span
	if s.header != nil {
		parts = append(parts, s.header.span)
	}
	for _, b := range s.body {
		parts = append(parts, b.base().span)
	}
	if len(parts) == 0 {
		return span{}
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out.start, out.end = min(out.start, p.start), max(out.end, p.end)
	}
	return out
}

// The group a section is, as a scope.
func scopeOf(s *section) scope {
	if s == nil || s.group == nil {
		return scope{}
	}
	return scope{group: s.group.value, inGroup: true}
}

// The spell-like entry name means in ns for a file of class: the class's own, else the shared one.
func (t *Names) spellEntry(class proto.Class, ns Namespace, name string) (SpellName, bool) {
	if e, ok := t.Classes[class][ns][name]; ok {
		return e, true
	}
	e, ok := t.Shared[ns][name]
	return e, ok
}

// A name of a spell-like namespace and its entry.
type NamedEntry struct {
	Name   string
	Entry  SpellName
	Shared bool // A shared entry: a spell or aura that gear, consumables or encounters add (SPEC.md §7.1).
}

// The names of a spell-like namespace a file of class reads (the class's own, then the shared ones it does not hide),
// each in byte order, that a build of spec registers; with no spec, those any build registers.
func (t *Names) Registered(class proto.Class, spec string, ns Namespace) []NamedEntry {
	registered := func(e SpellName) bool {
		if spec == "" {
			return len(e.RegisteredBy) > 0
		}
		return slices.Contains(e.RegisteredBy, spec)
	}
	var out []NamedEntry
	own := t.Classes[class][ns]
	for _, name := range sortedKeys(own) {
		if registered(own[name]) {
			out = append(out, NamedEntry{Name: name, Entry: own[name]})
		}
	}
	shared := t.Shared[ns]
	for _, name := range sortedKeys(shared) {
		if _, hidden := own[name]; !hidden && registered(shared[name]) {
			out = append(out, NamedEntry{Name: name, Entry: shared[name], Shared: true})
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
