package siml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The panel's edits of when blocks and defs, and the structure it reads them from.

var spliceBlocks = textLines(
	"siml 1",
	"spec hunter",
	"",
	"var $a = true",
	"var $b = true",
	"def fits(x) = spell(x).cast_time < 2s or not $a",
	"",
	"priority:",
	"  cast hunters_mark",
	"  when $a:",
	"    # cast aimed_shot",
	"    when $b:",
	"      cast multi_shot if fits(multi_shot)",
	"  cast arcane_shot",
	"",
	`group "Weave":`,
	"  move 5",
)

func TestSpliceBlockEdits(t *testing.T) {
	for _, c := range []struct {
		name, text, edits string
		want              string
	}{{
		name:  "wrap in a block",
		text:  spliceBlocks,
		edits: `[{"op": "wrap", "path": ["priorityList"], "from": 2, "to": 2, "value": {"variableRef": {"name": "b"}}}]`,
		want: textLines(
			"  when $a:",
			"    # cast aimed_shot",
			"    when $b:",
			"      when $b:",
			"        cast multi_shot if fits(multi_shot)",
			"  cast arcane_shot",
		),
	}, {
		name:  "wrap a group's row",
		text:  spliceBlocks,
		edits: `[{"op": "wrap", "path": ["groups", 0, "actions"], "from": 0, "to": 0, "value": {"variableRef": {"name": "a"}}}]`,
		want: textLines(
			`group "Weave":`,
			"  when $a:",
			"    move 5",
		),
	}, {
		name:  "wrap rows that share no conjunct",
		text:  spliceBlocks,
		edits: `[{"op": "wrap", "path": ["priorityList"], "from": 3, "to": 3, "value": {"const": {"val": "true"}}}]`,
		want: textLines(
			"  when true:",
			"    cast arcane_shot",
		),
	}, {
		name:  "a block's new condition keeps a call in a row's action",
		text:  strings.NewReplacer("def fits", "def near = 5\ndef fits", "    # cast aimed_shot", "    # cast aimed_shot\n    move near").Replace(spliceBlocks),
		edits: `[{"op": "set", "path": ["whens", 0, "condition"], "value": {"variableRef": {"name": "b"}}}]`,
		want: textLines(
			"  when $b:",
			"    # cast aimed_shot",
			"    move near",
		),
	}, {
		name:  "a block's new condition keeps a comment inside its old one",
		text:  strings.Replace(spliceBlocks, "  when $a:", "  when ($a // the opener\n    or $b):", 1),
		edits: `[{"op": "set", "path": ["whens", 0, "condition"], "value": {"variableRef": {"name": "b"}}}]`,
		want: textLines(
			"  // the opener",
			"  when $b:",
		),
	}, {
		name:  "show a block again",
		text:  strings.Replace(spliceBlocks, "  when $a:", "  # when $a:", 1),
		edits: `[{"op": "set", "path": ["whens", 0, "hide"]}]`,
		want: textLines(
			"  when $a:",
			"    # cast aimed_shot",
			"    when $b:",
			"      cast multi_shot if fits(multi_shot)",
		),
	}, {
		name:  "move a row into a block",
		text:  spliceBlocks,
		edits: `[{"op": "move", "path": ["priorityList"], "from": 3, "to": 1}, {"op": "set", "path": ["priorityList", 1, "action", "condition"], "value": {"variableRef": {"name": "a"}}}]`,
		want: textLines(
			"  when $a:",
			"    cast arcane_shot",
			"    # cast aimed_shot",
		),
	}} {
		t.Run(c.name, func(t *testing.T) {
			if got := splicedText(t, c.text, c.edits); !strings.Contains(got, c.want) {
				t.Errorf("the spliced text holds no\n%s\ngot:\n%s", c.want, got)
			}
		})
	}
}

func TestSpliceBlockRefusals(t *testing.T) {
	for _, c := range []struct{ name, edits, want string }{
		{"a block that holds blocks", `[{"op": "set", "path": ["whens", 0]}]`, "take those apart first"},
		{"rows of different blocks", `[{"op": "wrap", "path": ["priorityList"], "from": 0, "to": 1, "value": {"variableRef": {"name": "a"}}}]`, "different when blocks"},
		{"prepull rows", `[{"op": "wrap", "path": ["prepullActions"], "from": 0, "to": 0, "value": {"variableRef": {"name": "a"}}}]`, "priority list or a group"},
		{"a block with no condition", `[{"op": "set", "path": ["whens", 1, "condition"]}]`, "needs a condition"},
		{"a block that does not exist", `[{"op": "set", "path": ["whens", 7, "hide"], "value": true}]`, "does not exist"},
		{"a def name taken", `[{"op": "extract", "path": ["priorityList", 3, "action"], "name": "fits"}]`, "written already"},
		{"a def name that is a spell", `[{"op": "defInsert", "name": "multi_shot", "text": "true"}]`, "a spell"},
		{"no def call there", `[{"op": "inline", "path": ["priorityList", 0, "action"]}]`, "no def call"},
		{"a def that is not there", `[{"op": "defDelete", "name": "nothing"}]`, "no def is named nothing"},
		{"a def edit with others", `[{"op": "defSet", "name": "fits", "params": ["x"], "text": "true"}, {"op": "set", "path": ["priorityList", 0, "hide"], "value": true}]`, "goes alone"},
		{"a def value with an error", `[{"op": "defSet", "name": "fits", "params": ["x"], "text": "spell(x).nonsense"}]`, "the def has an error"},
		{"a def its calls no longer fit", `[{"op": "defSet", "name": "fits", "params": ["x", "y"], "text": "spell(x).cast_time < y"}]`, "fits takes 2 arguments"},
		{"a call that does not read", `[{"op": "set", "path": ["priorityList", 3, "action", "condition"], "call": "fits(nonsense)"}]`, "has an error"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := spliceRefused(t, splicePath, spliceBlocks, c.edits, nil); !strings.Contains(got, c.want) {
				t.Errorf("got %q, want an error holding %q", got, c.want)
			}
		})
	}
	text := strings.Replace(spliceBlocks, "def fits(x)", "def early = fits(aimed_shot)\ndef fits(x)", 1)
	if got := spliceRefused(t, splicePath, text, `[{"op": "defDelete", "name": "fits"}]`, nil); !strings.Contains(got, "the def early calls fits") {
		t.Errorf("deleting a def another calls: got %q", got)
	}
}

func TestSpliceDefLines(t *testing.T) {
	got := splicedText(t, spliceBlocks, `[{"op": "defSet", "name": "fits", "params": ["x"], "text": "spell(x).cast_time  <  3s"}]`)
	if !strings.Contains(got, "def fits(x) = spell(x).cast_time < 3s\n") {
		t.Errorf("defSet wrote:\n%s", got)
	}
	// The panel shows a def's value without the comments in it and sends it back as it shows it; the comments go above
	// the def.
	commented := strings.Replace(spliceBlocks, "< 2s or not $a", "< 2s // short casts\n    or not $a", 1)
	d := Analyze(splicePath, []byte(commented), corpusNames(t), nil)
	value := d.Structure().Defs[0].Value
	if value != "spell(x).cast_time < 2s or not $a" {
		t.Errorf("the def's value reads %q", value)
	}
	out, err := d.Splice([]AplEdit{{Op: "defSet", Name: "fits", Params: []string{"x"}, Text: value}})
	if err != nil {
		t.Fatal(err)
	}
	if got := applied(d, out); !strings.Contains(got, "\n// short casts\ndef fits(x) = spell(x).cast_time < 2s or not $a\n") {
		t.Errorf("defSet of a def with a comment in it wrote:\n%s", got)
	}
	got = splicedText(t, spliceBlocks, `[{"op": "defRename", "name": "fits", "text": "fits_weave"}]`)
	if !strings.Contains(got, "def fits_weave(x) =") || !strings.Contains(got, "if fits_weave(multi_shot)") {
		t.Errorf("defRename wrote:\n%s", got)
	}
	// A file with no var or def line takes the new def after its top-level lines, a blank line between.
	bare := textLines("siml 1", "spec hunter", "", "priority:", "  cast hunters_mark if current_time > 1s")
	got = splicedText(t, bare, `[{"op": "extract", "path": ["priorityList", 0, "action", "condition"], "name": "late"}]`)
	if want := textLines("siml 1", "spec hunter", "", "def late = current_time > 1s", "", "priority:", "  cast hunters_mark if late"); got != want {
		t.Errorf("extract wrote:\n%s\nwant:\n%s", got, want)
	}
}

func TestStructure(t *testing.T) {
	d := Analyze(splicePath, []byte(spliceBlocks), corpusNames(t), nil)
	s := d.Structure()
	if s == nil {
		t.Fatal("no structure")
	}
	got, _ := json.Marshal(s)
	for _, want := range []string{
		`"whens":[{"id":0,"list":["priorityList"],"parent":-1,"first":1,"last":2,"condition":{"variableRef":{"name":"a"}},"conjuncts":[{"variableRef":{"name":"a"}}]}`,
		`{"id":1,"list":["priorityList"],"parent":0,"first":2,"last":2,"condition":{"variableRef":{"name":"b"}},"conjuncts":[{"variableRef":{"name":"b"}}]}]`,
		`"rows":[{"path":["priorityList",1],"inherited":1,"own":false},{"path":["priorityList",2],"inherited":2,"own":true}]`,
		`"calls":[{"path":["priorityList",2,"action","condition","and","vals",2],"name":"fits","args":["multi_shot"]}]`,
		`"defs":[{"name":"fits","params":["x"],"value":"spell(x).cast_time `,
		`or not $a","used":true}]`,
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("the structure holds no %s\ngot: %s", want, got)
		}
	}
	if broken := Analyze(splicePath, []byte(spliceBlocks+"  cast nonsense\n"), corpusNames(t), nil); broken.Structure() != nil {
		t.Error("a text with an error has a structure")
	}
	// The calls of a file with a use line, of its own defs and of the used file's.
	main := filepath.Join("testdata", "splice", "use", "main.apl.siml")
	text, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	got, _ = json.Marshal(Analyze(main, text, corpusNames(t), os.ReadFile).Structure().Calls)
	if want := `[{"path":["priorityList",0,"action","condition"],"name":"far","args":[]},{"path":["priorityList",1,"action","condition"],"name":"at_range","args":[]}]`; string(got) != want {
		t.Errorf("the calls of a file with a use line: %s", got)
	}
}
