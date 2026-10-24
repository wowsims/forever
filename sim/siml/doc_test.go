package siml

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

// Doc: what an editor reads of a file beyond Parse.

// A JSON path as its keys and indices joined with '/'.
func pathString(path []any) string {
	parts := make([]string, len(path))
	for i, k := range path {
		parts[i] = fmt.Sprint(k)
	}
	return strings.Join(parts, "/")
}

// The text the span of a JSON path covers, or "" when no span has the path.
func spanText(d *Doc, text, path string) string {
	for _, s := range d.Spans() {
		if pathString(s.Path) == path {
			return text[d.Offset(s.Range.Start):d.Offset(s.Range.End)]
		}
	}
	return ""
}

// A Doc's refs as "name path namespace", the path "-" for a ref outside what the file loads.
func refLines(d *Doc) []string {
	var lines []string
	for _, r := range d.Refs() {
		path := "-"
		if r.Path != nil {
			path = pathString(r.Path)
		}
		lines = append(lines, fmt.Sprintf("%s %s %s", r.Name, path, r.Namespace))
	}
	return lines
}

var docText = textLines(
	"siml 1",
	"spec hunter",
	"",
	"var $burst = true",
	"",
	"priority:",
	"  cast multi_shot if $burst",
	"  cast rapid_fire",
)

func TestDocSpansAndPaths(t *testing.T) {
	d := Analyze(splicePath, []byte(docText), corpusNames(t), nil)
	if _, err := d.File(); err != nil {
		t.Fatal(err)
	}
	spans := d.Spans()
	has := func(path string, r Range) bool {
		return slices.ContainsFunc(spans, func(s PathSpan) bool { return pathString(s.Path) == path && s.Range == r })
	}
	for _, c := range []struct {
		path string
		r    Range
	}{
		{"valueVariables/0", Range{Position{4, 1}, Position{4, 18}}},
		{"priorityList/0", Range{Position{7, 3}, Position{7, 28}}},
		{"priorityList/0/action/castSpell/spellId", Range{Position{7, 8}, Position{7, 18}}},
		{"priorityList/0/action/condition", Range{Position{7, 22}, Position{7, 28}}},
		{"priorityList/1", Range{Position{8, 3}, Position{8, 18}}},
	} {
		if !has(c.path, c.r) {
			t.Errorf("no span %s at %v", c.path, c.r)
		}
	}
	if got := pathString(d.PathAt(Position{7, 25})); got != "priorityList/0/action/condition" {
		t.Errorf("the path at 7:25 is %s", got)
	}
	if r, ok := d.SpanFor([]any{"priorityList", 1, "action", "castSpell", "spellId", "spellId"}); !ok || r != (Range{Position{8, 8}, Position{8, 18}}) {
		t.Errorf("the span for the spell id of row 2 is %v, %t", r, ok)
	}
	if got := string(d.RotationJSON()); !strings.HasPrefix(got, `{"type":"TypeAPL","priorityList":[`) {
		t.Errorf("the rotation's JSON is %s", got)
	}
}

// The rotation panel finds a row of a when block by the row's own line, and a conjunct the block gives it by the when
// line (SPEC.md §6.9).
func TestDocSpansInWhenBlock(t *testing.T) {
	text := textLines("siml 1", "spec hunter", "", "var $burst = true", "", "priority:", "  when $burst:", "    cast multi_shot if current_time > 1s", "    cast rapid_fire")
	d := Analyze(splicePath, []byte(text), corpusNames(t), nil)
	if _, err := d.File(); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]Range{
		"priorityList/0": {Position{8, 5}, Position{8, 41}},
		"priorityList/0/action/condition/and/vals/0": {Position{7, 8}, Position{7, 14}},
		"priorityList/1": {Position{9, 5}, Position{9, 20}},
	} {
		if !slices.ContainsFunc(d.Spans(), func(s PathSpan) bool { return pathString(s.Path) == path && s.Range == want }) {
			t.Errorf("no span %s at %v", path, want)
		}
	}
	if got := pathString(d.PathAt(Position{9, 10})); got != "priorityList/1/action/castSpell/spellId" {
		t.Errorf("the path at 9:10 is %s", got)
	}
}

func TestDocWithErrors(t *testing.T) {
	d := Analyze(splicePath, []byte(strings.Replace(docText, "rapid_fire", "rapid_fir", 1)), corpusNames(t), nil)
	if _, err := d.File(); err == nil {
		t.Fatal("the file reads")
	}
	if d.RotationJSON() != nil {
		t.Error("a text with an error gives the rotation's JSON")
	}
	if diags := d.Diagnostics(); len(diags) != 1 || diags[0].Code != "unknown-name" {
		t.Errorf("the diagnostics are %v", diags)
	}
	refs := d.Refs()
	if len(refs) != 2 || !refs[0].Known || refs[1].Known || refs[1].Name != "rapid_fir" || pathString(refs[1].Path) != "priorityList/1/action/castSpell/spellId" {
		t.Errorf("the refs are %+v", refs)
	}
	f, _ := d.File()
	if len(f.Names) != 1 {
		t.Errorf("the file's names are %+v", f.Names)
	}
}

func TestDocOffsets(t *testing.T) {
	text := "siml 1\nspec hunter\n// 😀 é x\n"
	d := Analyze(splicePath, []byte(text), corpusNames(t), nil)
	x := strings.Index(text, "x")
	if got := d.Position(x); got != (Position{3, 9}) {
		t.Errorf("x is at %v", got)
	}
	if got := d.Offset(Position{3, 9}); got != x {
		t.Errorf("3:9 is offset %d, not %d", got, x)
	}
	if got := d.Offset(Position{2, 99}); got != strings.Index(text, "\n//") {
		t.Errorf("a column past its line's end is offset %d", got)
	}
}

var refsText = textLines(
	"siml 1",
	"spec hunter",
	"",
	"var $v = 1s",
	"",
	"prepull:",
	"  -1s cast hunters_mark",
	"",
	"priority:",
	"  cast hunters_mark if target.aura(hunters_mark).is_active and pet1.aura(frenzy).is_active",
	"  cast item.goblin_sapper_charge if boss_spell(x).is_casting or $v > 0s",
)

// Every row, var line and node has the span of its text, and every name the namespace it is read in.
func TestDocRotationSpansAndRefs(t *testing.T) {
	d := Analyze(splicePath, []byte(refsText), corpusNames(t), nil)
	if _, err := d.File(); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		"valueVariables/0":                           "var $v = 1s",
		"valueVariables/0/value":                     "1s",
		"prepullActions/0":                           "-1s cast hunters_mark",
		"prepullActions/0/doAtValue":                 "-1s",
		"priorityList/0/action":                      "cast hunters_mark if target.aura(hunters_mark).is_active and pet1.aura(frenzy).is_active",
		"priorityList/0/action/condition/and/vals/1": "pet1.aura(frenzy).is_active",
		"priorityList/0/action/condition/and/vals/0/auraIsActive/sourceUnit": "target",
		"priorityList/1/action/castSpell/spellId":                            "item.goblin_sapper_charge",
	} {
		if got := spanText(d, refsText, path); got != want {
			t.Errorf("the span of %s covers %q, want %q", path, got, want)
		}
	}
	want := []string{
		"hunters_mark prepullActions/0/action/castSpell/spellId spell",
		"hunters_mark priorityList/0/action/castSpell/spellId spell",
		"hunters_mark priorityList/0/action/condition/and/vals/0/auraIsActive/auraId target_aura",
		"frenzy priorityList/0/action/condition/and/vals/1/auraIsActive/auraId pet_aura",
		"goblin_sapper_charge priorityList/1/action/castSpell/spellId item",
		"x priorityList/1/action/condition/or/vals/0/bossSpellIsCasting/spellId target_spell",
	}
	if got := refLines(d); !slices.Equal(got, want) {
		t.Errorf("the refs are\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// A file with an error keeps the spans and refs of what reads; a name in a section the file kind does not hold has no
// path, as it lies outside the rotation.
func TestDocRefsWithErrors(t *testing.T) {
	text := strings.Replace(refsText, "item.goblin_sapper_charge", "item.goblin_sapper_chrge", 1)
	d := Analyze(splicePath, []byte(text), corpusNames(t), nil)
	if diags := d.Diagnostics(); len(diags) != 1 || diags[0].Message != "no item is named goblin_sapper_chrge; did you mean goblin_sapper_charge?" {
		t.Errorf("the diagnostics are %v", diags)
	}
	if got, want := spanText(d, text, "priorityList/1"), "cast item.goblin_sapper_chrge if boss_spell(x).is_casting or $v > 0s"; got != want {
		t.Errorf("the second row's span covers %q, want %q", got, want)
	}
	if !slices.Contains(refLines(d), "goblin_sapper_chrge priorityList/1/action/castSpell/spellId item") {
		t.Errorf("the refs are %v", refLines(d))
	}

	foreign := Analyze(splicePath, []byte("siml 1\nspec hunter\n\ncooldowns:\n  cooldowns(id=hunters_mark)\n"), corpusNames(t), nil)
	if diags := foreign.Diagnostics(); len(diags) != 1 || diags[0].Code != "section-not-allowed" {
		t.Errorf("the diagnostics are %v", diags)
	}
	if got := refLines(foreign); !slices.Equal(got, []string{"hunters_mark - spell"}) {
		t.Errorf("the refs are %v", got)
	}
}

// A gear file's slot rows have the spans of their text, and its names the item, enchant and suffix namespaces.
func TestDocGearSpansAndRefs(t *testing.T) {
	text := textLines("siml 1", "head       helm + agility", "waist      belt of_the_tiger")
	d := Analyze("p1.gear.siml", []byte(text), corpusNames(t), nil)
	if _, err := d.File(); err != nil {
		t.Fatal(err)
	}
	if got := spanText(d, text, "items/0"); got != "head       helm + agility" {
		t.Errorf("the head's span covers %q", got)
	}
	if got := spanText(d, text, "items/7/randomSuffix"); got != "of_the_tiger" {
		t.Errorf("the waist's suffix span covers %q", got)
	}
	want := []string{"helm items/0/id item", "agility items/0/enchant enchant", "belt items/7/id item", "of_the_tiger items/7/randomSuffix suffix"}
	if got := refLines(d); !slices.Equal(got, want) {
		t.Errorf("the refs are %v, want %v", got, want)
	}
}
