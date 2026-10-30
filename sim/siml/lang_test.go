package siml

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

// The language services on a table of four names, which needs neither the committed table nor the item database: a
// hunter spell its builds register, one no build registers, and two shared ones, the first registered by the warrior's
// builds alone.

func smallNames() *Names {
	return &Names{
		Classes: map[proto.Class]SpellNames{proto.Class_ClassHunter: {NamespaceSpell: {
			"steady": {SpellID: 900001, RegisteredBy: []string{"hunter"}},
			"kept":   {SpellID: 900002},
		}}},
		Shared: SpellNames{NamespaceSpell: {
			"warriors_only": {SpellID: 900003, RegisteredBy: []string{"dps_warrior"}},
			"everyones":     {SpellID: 900004, RegisteredBy: []string{"dps_warrior", "hunter"}},
		}},
	}
}

// The repository root, and the paths the cases read their text as.
var (
	langRoot     = func() string { root, _ := filepath.Abs(filepath.Join("..", "..")); return root }()
	hunterPath   = filepath.Join(langRoot, "x.apl.siml")
	setupPath    = filepath.Join(langRoot, "x.siml")
	gearPath     = filepath.Join(langRoot, "x.gear.siml")
	specsPath    = filepath.Join(langRoot, "ui", "specs", "hunter", "dps", "apls", "x.apl.siml")
	langOpen     = Opener(func(path string) ([]byte, error) { return readDisk(path) })
	langPreface  = "siml 1\nspec hunter\n\nvar ${Melee weave} = true\n"
	plainPreface = "siml 1\nspec hunter\n\n"
)

// The text with its `|` taken out, and the position the `|` marked.
func marked(text string) (string, Position) {
	at := strings.IndexByte(text, '|')
	text = text[:at] + text[at+1:]
	return text, newLineMap(text).position(at)
}

// Where needle starts in text, plus k.
func positionOf(text, needle string, k int) Position {
	return newLineMap(text).position(strings.Index(text, needle) + k)
}

func completionsIn(t *testing.T, names *Names, c *DatabaseCatalog, path, source string) []Completion {
	t.Helper()
	text, at := marked(source)
	return Analyze(path, []byte(text), names, langOpen).Complete(at, c)
}

func labelsOf(items []Completion) []string {
	out := make([]string, len(items))
	for i, x := range items {
		out[i] = x.Label
	}
	return out
}

// [line (0-based, as an editor counts), code, severity (1 error, 2 warning, 3 information)] of each diagnostic.
type diagRow struct {
	line     int
	code     string
	severity int
}

func diagRows(diags Diagnostics) []diagRow {
	out := make([]diagRow, len(diags))
	for i, x := range diags {
		out[i] = diagRow{x.Start.Line - 1, x.Code, int(x.Severity) + 1}
	}
	return out
}

func TestRegisteredNames(t *testing.T) {
	names := smallNames()
	var got []string
	for _, n := range names.Registered(proto.Class_ClassHunter, "hunter", NamespaceSpell) {
		got = append(got, n.Name+map[bool]string{true: " shared", false: ""}[n.Shared])
	}
	if want := []string{"steady", "everyones shared"}; !slices.Equal(got, want) {
		t.Errorf("hunter: %v, want %v", got, want)
	}
	got = nil
	for _, n := range names.Registered(proto.Class_ClassUnknown, "", NamespaceSpell) {
		got = append(got, n.Name)
	}
	if want := []string{"everyones", "warriors_only"}; !slices.Equal(got, want) {
		t.Errorf("no spec: %v, want %v", got, want)
	}
}

func TestCompleteRegisteredNamesOnly(t *testing.T) {
	names := smallNames()
	hunter := labelsOf(completionsIn(t, names, nil, hunterPath, langPreface+"priority:\n  cast |\n"))
	if !slices.Contains(hunter, "steady") || !slices.Contains(hunter, "everyones") {
		t.Errorf("hunter offers %v", hunter)
	}
	if slices.Contains(hunter, "kept") || slices.Contains(hunter, "warriors_only") {
		t.Errorf("hunter offers a name its builds do not register: %v", hunter)
	}
	noSpec := labelsOf(completionsIn(t, names, nil, filepath.Join(langRoot, "x.cooldowns.siml"), "siml 1\n\ncooldowns(id=|\n"))
	if !slices.Contains(noSpec, "warriors_only") || !slices.Contains(noSpec, "everyones") || slices.Contains(noSpec, "steady") {
		t.Errorf("no spec offers %v", noSpec)
	}
}

func TestCompleteClassNamesFirst(t *testing.T) {
	var offered []Completion
	for _, x := range completionsIn(t, smallNames(), nil, hunterPath, langPreface+"priority:\n  cast |\n") {
		if x.Kind == CompletionConstant {
			offered = append(offered, x)
		}
	}
	slices.SortFunc(offered, func(a, b Completion) int { return strings.Compare(a.SortText, b.SortText) })
	var got [][2]string
	for _, x := range offered {
		got = append(got, [2]string{x.Label, x.Detail})
	}
	want := [][2]string{{"steady", "(900001)"}, {"everyones", "(900004), from gear, consumables or encounters"}}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestChecksUnregisteredName(t *testing.T) {
	text := plainPreface + "priority:\n  cast kept\n  cast steady\n"
	outside := Analyze(hunterPath, []byte(text), smallNames(), langOpen).EditorDiagnostics(nil)
	if got, want := diagRows(outside), []diagRow{{4, "unregistered-name", 2}}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := outside[0].Message, "kept is a spell no hunter build registers"; got != want {
		t.Errorf("message %q, want %q", got, want)
	}
	inside := Analyze(specsPath, []byte(text), smallNames(), langOpen).EditorDiagnostics(nil)
	if len(inside) == 0 || inside[0].Severity != Error {
		t.Errorf("under ui/specs: %v", diagRows(inside))
	}
}

func TestQuickFixRewrites(t *testing.T) {
	text := langPreface + "priority:\n  cast steady if current_time > 1s and and(gcd_is_ready) or (gcd_is_ready or current_time < 5s)\n"
	d := Analyze(hunterPath, []byte(text), smallNames(), langOpen)
	var got []string
	for _, f := range d.QuickFixes(Range{Position{1, 1}, Position{10, 1}}) {
		if f.Kind == "refactor.rewrite" {
			got = append(got, f.Edits[0].NewText)
		}
	}
	if want := []string{"gcd_is_ready", "gcd_is_ready or current_time < 5s"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestQuickFixSkipsListOperand(t *testing.T) {
	text := langPreface + "priority:\n  cast steady if and([gcd_is_ready, true]) and or([gcd_is_ready]) and and([true])\n"
	d := Analyze(hunterPath, []byte(text), smallNames(), langOpen)
	for _, f := range d.QuickFixes(Range{Position{1, 1}, Position{10, 1}}) {
		if f.Kind == "refactor.rewrite" {
			t.Errorf("rewrites a list operand: %q", f.Edits[0].NewText)
		}
	}
}

func TestIdentifierRefusal(t *testing.T) {
	for _, c := range []struct {
		kind, name string
		ok         bool
	}{
		{"variable", "Melee weave", true},
		{"variable", "a}b", false},
		{"placeholder", "a\nb", false},
		{"group", "a}b", true},
		{"sequence", "", false},
	} {
		if got := identifierRefusal(c.kind, c.name) == ""; got != c.ok {
			t.Errorf("identifierRefusal(%q, %q) is empty: %v", c.kind, c.name, got)
		}
	}
}

func readDisk(path string) ([]byte, error) { return os.ReadFile(path) }

// A def with two parameters, called in a row and in a when block (SPEC.md §6.9, §6.10).
var defsText = langPreface + textLines(
	"def fits(x, y) = spell(x).cast_time < y or not $melee_weave",
	"",
	"priority:",
	"  cast steady if fits(steady, 2s)",
	"  when $melee_weave:",
	"    when current_time > 1s:",
	"      cast steady if fits(everyones, 1s)",
)

func defsDoc() *Doc { return Analyze(hunterPath, []byte(defsText), smallNames(), langOpen) }

// Each location as line:column, 1-based.
func defLocations(locations []Location) []string {
	out := []string{}
	for _, l := range locations {
		out = append(out, fmt.Sprintf("%d:%d", l.Range.Start.Line, l.Range.Start.Col))
	}
	return out
}

func TestDefNavigation(t *testing.T) {
	d := defsDoc()
	if got, want := defLocations(d.Definition(positionOf(defsText, "fits(steady", 1))), []string{"5:5"}; !slices.Equal(got, want) {
		t.Errorf("definition of a call: %v, want %v", got, want)
	}
	if got, want := defLocations(d.References(positionOf(defsText, "fits(x", 1), true)), []string{"5:5", "8:18", "11:22"}; !slices.Equal(got, want) {
		t.Errorf("references of the def: %v, want %v", got, want)
	}
	if got, want := defLocations(d.References(positionOf(defsText, "x, y", 0), true)), []string{"5:10", "5:24"}; !slices.Equal(got, want) {
		t.Errorf("references of a parameter: %v, want %v", got, want)
	}
	if got, want := defLocations(d.Definition(positionOf(defsText, "< y", 2))), []string{"5:13"}; !slices.Equal(got, want) {
		t.Errorf("definition of a parameter: %v, want %v", got, want)
	}
	edits, err := d.Rename(positionOf(defsText, "fits(steady", 1), "fits_weave")
	if err != nil || len(edits) != 1 || len(edits[0].Edits) != 3 || edits[0].Edits[2].NewText != "fits_weave" {
		t.Errorf("rename of the def: %v, %v", edits, err)
	}
	edits, err = d.Rename(positionOf(defsText, "< y", 2), "limit")
	if err != nil || len(edits) != 1 || len(edits[0].Edits) != 2 || edits[0].Edits[0].NewText != "limit" {
		t.Errorf("rename of a parameter: %v, %v", edits, err)
	}
	for _, c := range []struct{ needle, newName, want string }{
		{"fits(x", "target", "target is a unit; a def takes a name of its own."},
		{"fits(x", "2x", "A def's name is a word: letters, digits and _, not starting with a digit."},
		{"x, y", "and", "and is a reserved word."},
		{"x, y", "y", "x.apl.siml has a parameter named y already."},
	} {
		if _, err := d.Rename(positionOf(defsText, c.needle, 0), c.newName); err == nil || err.Error() != c.want {
			t.Errorf("%s → %s: %v, want %q", c.needle, c.newName, err, c.want)
		}
	}
}

// A rename leaves the rotation as it was, its new name aside, or is refused: a name that another word of a def's
// value, a def's caller or a used file's variable already has would make that word read as the renamed symbol
// (SPEC.md §6.10).
func TestRenameKeepsRotation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.apl.siml"), []byte(plainPreface+"var ${Ranged range} = 12\ndef lib_def = current_time > 2s\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "b.apl.siml")
	text := plainPreface + textLines(
		`use "a.apl.siml"`,
		"",
		"var ${Foo} = 3",
		"def g = current_time > 1s",
		"def f(alpha) = spell(alpha).is_ready and spell(everyones).is_ready",
		"def h(x) = x > 1 and g",
		"",
		"priority:",
		"  cast steady if f(steady) and h(2) and $ranged_range > 1 and $foo > 1 and lib_def",
	)
	d := Analyze(path, []byte(text), smallNames(), langOpen)
	before := string(d.RotationJSON())
	if before == "" {
		t.Fatalf("the text holds no rotation: %v", d.Diagnostics())
	}
	for _, c := range []struct {
		needle, newName string
		refused         bool
	}{
		{"(alpha", "everyones", true},
		{"(x", "g", true},
		{"def g", "x", true},
		{"${Foo", "ranged_range", true},
		// The name is taken in a file the rename edits: this one, which uses the file that declares the symbol.
		{"$ranged_range", "Foo", true},
		{"lib_def", "g", true},
		{"def g", "lib_def", true},
		{"(alpha", "beta", false},
		{"${Foo", "Bar", false},
	} {
		at := positionOf(text, c.needle, len(c.needle)-1)
		edits, err := d.Rename(at, c.newName)
		if (err != nil) != c.refused {
			t.Errorf("%s → %s: refused %v (%v), want refused %v", c.needle, c.newName, err != nil, err, c.refused)
		}
		if err != nil || len(edits) != 1 {
			continue
		}
		want := before
		if strings.HasPrefix(c.needle, "$") {
			want = strings.ReplaceAll(want, `"Foo"`, jsonString(c.newName))
		}
		if got := string(Analyze(path, []byte(applied(d, edits[0].Edits)), smallNames(), langOpen).RotationJSON()); got != want {
			t.Errorf("%s → %s changes the rotation:\n%s\nwant\n%s", c.needle, c.newName, got, want)
		}
	}
}

func TestDefCompletion(t *testing.T) {
	find := func(items []Completion, label string) (Completion, bool) {
		i := slices.IndexFunc(items, func(x Completion) bool { return x.Label == label })
		if i < 0 {
			return Completion{}, false
		}
		return items[i], true
	}
	names := smallNames()
	row := completionsIn(t, names, nil, hunterPath, strings.Replace(defsText, "if fits(steady", "if |fits(steady", 1))
	if c, ok := find(row, "fits"); !ok || c.Detail != "def fits(x, y)" || c.InsertText != "fits(" {
		t.Errorf("a value offers the def as %+v (%v)", c, ok)
	}
	if _, ok := find(row, "x"); ok {
		t.Errorf("a row offers a def's parameter: %v", labelsOf(row))
	}
	value := completionsIn(t, names, nil, hunterPath, strings.Replace(defsText, "< y or", "< | or", 1))
	if _, ok := find(value, "y"); !ok {
		t.Errorf("the def's value does not offer its parameter: %v", labelsOf(value))
	}
	if own := completionsIn(t, names, nil, hunterPath, strings.Replace(defsText, "< y or", "< |y or", 1)); slices.Contains(labelsOf(own), "fits") {
		t.Errorf("the def's value offers the def itself, which reference-cycle refuses: %v", labelsOf(own))
	}
	ref := completionsIn(t, names, nil, hunterPath, strings.Replace(defsText, "spell(x)", "spell(|)", 1))
	if _, ok := find(ref, "x"); !ok {
		t.Errorf("a ref in the def's value does not offer the parameter: %v", labelsOf(ref))
	}
	// A def with no parameters that is a name is offered where a name goes, and one with parameters is not.
	named := strings.Replace(defsText, "\npriority:", "def opener = steady\n\npriority:", 1)
	cast := completionsIn(t, names, nil, hunterPath, strings.Replace(named, "cast steady if", "cast | if", 1))
	if c, ok := find(cast, "opener"); !ok || c.Detail != "def opener = steady" {
		t.Errorf("an action's spell offers the def as %+v (%v): %v", c, ok, labelsOf(cast))
	}
	if _, ok := find(cast, "fits"); ok {
		t.Errorf("an action's spell offers a def with parameters: %v", labelsOf(cast))
	}
	valued := strings.Replace(named, "def opener = steady", "def opener = steady\ndef late = current_time", 1)
	if _, ok := find(completionsIn(t, names, nil, hunterPath, strings.Replace(valued, "cast steady if", "cast | if", 1)), "late"); ok {
		t.Error("an action's spell offers a def whose value is no name")
	}
	if _, ok := find(completionsIn(t, names, nil, hunterPath, defsText+"  |\n"), "when"); !ok {
		t.Error("a priority row does not offer when")
	}
	if _, ok := find(completionsIn(t, names, nil, hunterPath, plainPreface+"prepull:\n  |\n"), "when"); ok {
		t.Error("a prepull row offers when")
	}
	// A variable written where a name goes offers the variables.
	if _, ok := find(completionsIn(t, names, nil, hunterPath, plainPreface+"var $weave = true\n\npriority:\n  cast $we|\n"), "$weave"); !ok {
		t.Error("cast $… does not offer the variables")
	}
	if _, ok := find(completionsIn(t, names, nil, hunterPath, plainPreface+"|\n"), "def"); !ok {
		t.Error("a top line does not offer def")
	}
	when := completionsIn(t, names, nil, hunterPath, strings.Replace(defsText, "when $melee_weave:", "when $melee_weave |", 1))
	if c, ok := find(when, ":"); !ok || c.InsertText != ":" {
		t.Errorf("a when condition does not offer its colon: %v", labelsOf(when))
	}
}

// A use line offers the files near its own that it can use: none of a kind this file cannot hold, and none that uses
// this file, which section-not-allowed and use-cycle would refuse.
func TestCompleteUsePaths(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"apls/a.apl.siml":        plainPreface + "use \"b.apl.siml\"\n",
		"apls/b.apl.siml":        plainPreface,
		"apls/lib.apl.siml":      plainPreface,
		"gear_sets/p1.gear.siml": plainPreface,
		"builds/x.siml":          plainPreface,
		"defaults.siml":          plainPreface,
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		file string
		want []string
	}{
		{"apls/b.apl.siml", []string{`"lib.apl.siml"`}},
		{"gear_sets/p1.gear.siml", []string{}},
		{"builds/x.siml", []string{`"../apls/a.apl.siml"`, `"../apls/b.apl.siml"`, `"../apls/lib.apl.siml"`, `"../defaults.siml"`, `"../gear_sets/p1.gear.siml"`}},
	} {
		path := filepath.Join(dir, filepath.FromSlash(c.file))
		if got := labelsOf(completionsIn(t, smallNames(), nil, path, plainPreface+"use \"|\n")); !slices.Equal(got, c.want) {
			t.Errorf("%s offers %q, want %q", c.file, got, c.want)
		}
	}
}

// The line right under a when line is its first row; under an unfinished condition, the condition goes on (SPEC.md
// §6.9).
func TestCompleteWhenBody(t *testing.T) {
	for _, c := range []struct{ source, want string }{
		{"priority:\n  when current_time > 1s:\n    |\n", "cast"},
		{"priority:\n  when current_time > 1s\n      and current_time < 2s:\n    |\n", "cast"},
		{"priority:\n  when current_time > 1s and\n    |\n", "current_time"},
	} {
		if got := labelsOf(completionsIn(t, smallNames(), nil, hunterPath, plainPreface+c.source)); !slices.Contains(got, c.want) {
			t.Errorf("%q offers %q, without %s", c.source, got, c.want)
		}
	}
}

func TestDefAndWhenHover(t *testing.T) {
	d := defsDoc()
	hover := func(at Position) string {
		h, ok := d.Hover(at, nil)
		if !ok {
			t.Fatalf("no hover at %v", at)
		}
		return h.Markdown
	}
	call := hover(positionOf(defsText, "fits(steady", 1))
	for _, want := range []string{"def **fits**", "def fits(x, y) = spell(x).cast_time < y or not $melee_weave", "spell(steady).cast_time < 2s or not $melee_weave"} {
		if !strings.Contains(call, want) {
			t.Errorf("the hover of a call lacks %q:\n%s", want, call)
		}
	}
	if got := hover(positionOf(defsText, "< y", 2)); !strings.Contains(got, "parameter **y** of def fits") {
		t.Errorf("the hover of a parameter:\n%s", got)
	}
	if got := hover(positionOf(defsText, "when current_time", 1)); !strings.Contains(got, "$melee_weave and current_time > 1s") {
		t.Errorf("the hover of a nested when line:\n%s", got)
	}
}

func TestDefTokensAndOutline(t *testing.T) {
	d := defsDoc()
	var got [][2]string
	for _, tok := range d.Tokens() {
		if tok.Type == "function" || tok.Type == "parameter" {
			got = append(got, [2]string{defsText[d.Offset(tok.Range.Start):d.Offset(tok.Range.End)], tok.Type})
		}
	}
	want := [][2]string{
		{"fits", "function"}, {"x", "parameter"}, {"y", "parameter"}, {"spell", "function"}, {"x", "parameter"},
		{"y", "parameter"}, {"fits", "function"}, {"current_time", "function"}, {"fits", "function"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q\nwant %q", got, want)
	}
	symbols := d.Symbols()
	var outline []string
	var walk func(prefix string, symbols []Symbol)
	walk = func(prefix string, symbols []Symbol) {
		for _, s := range symbols {
			outline = append(outline, prefix+s.Name)
			walk(prefix+"  ", s.Children)
		}
	}
	walk("", symbols)
	if want := []string{"Melee weave", "fits", "priority", "  when $melee_weave", "    when current_time > 1s"}; !slices.Equal(outline, want) {
		t.Errorf("outline %q, want %q", outline, want)
	}
}
