package siml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Splice: panel edits applied to the text of an x.apl.siml, over the corpus in testdata/splice and over each rule on
// its own.

const splicePath = "x.apl.siml"

func aplEdits(t *testing.T, text string) []AplEdit {
	t.Helper()
	var edits []AplEdit
	if err := json.Unmarshal([]byte(text), &edits); err != nil {
		t.Fatalf("the edits are no JSON: %v\n%s", err, text)
	}
	return edits
}

// The text with edits applied.
func applied(d *Doc, edits []TextEdit) string {
	text := d.Text()
	for _, e := range slices.Backward(edits) {
		text = text[:d.Offset(e.Range.Start)] + e.NewText + text[d.Offset(e.Range.End):]
	}
	return text
}

func splicedText(t *testing.T, text, edits string) string {
	t.Helper()
	d := Analyze(splicePath, []byte(text), corpusNames(t), nil)
	out, err := d.Splice(aplEdits(t, edits))
	if err != nil {
		t.Fatal(err)
	}
	return applied(d, out)
}

// A splice's edits as the 0-based lines they replace and the text they write.
type lineEdit struct {
	from, to int
	text     string
}

func spliceChanges(t *testing.T, text, edits string) []lineEdit {
	t.Helper()
	d := Analyze(splicePath, []byte(text), corpusNames(t), nil)
	out, err := d.Splice(aplEdits(t, edits))
	if err != nil {
		t.Fatal(err)
	}
	changes := []lineEdit{}
	for _, e := range out {
		changes = append(changes, lineEdit{e.Range.Start.Line - 1, e.Range.End.Line - 1, e.NewText})
	}
	return changes
}

func spliceRefused(t *testing.T, path, text, edits string, open Opener) string {
	t.Helper()
	d := Analyze(path, []byte(text), corpusNames(t), open)
	_, err := d.Splice(aplEdits(t, edits))
	if err == nil {
		return ""
	}
	return err.Error()
}

func textLines(lines ...string) string { return strings.Join(lines, "\n") + "\n" }

func TestSpliceCases(t *testing.T) {
	names := corpusNames(t)
	cases := corpusCases(t, "splice", ".apl.siml")
	if len(cases) <= 5 {
		t.Fatalf("testdata/splice holds %d cases", len(cases))
	}
	for _, file := range cases {
		t.Run(caseName(file), func(t *testing.T) {
			text, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			edits, err := os.ReadFile(strings.TrimSuffix(file, ".apl.siml") + ".edits.json")
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(file + ".want")
			if err != nil {
				t.Fatal(err)
			}
			// A case in a folder of its own is read at its path, with the files it uses.
			path, open := splicePath, Opener(nil)
			if filepath.Base(filepath.Dir(file)) != "splice" {
				path, open = file, caseOpener(file)
			}
			d := Analyze(path, text, names, open)
			out, err := d.Splice(aplEdits(t, string(edits)))
			if err != nil {
				t.Fatal(err)
			}
			if got := applied(d, out); got != string(want) {
				t.Errorf("the spliced text differs from %s.want\ngot:\n%s", filepath.Base(file), got)
			}
			// A file in canonical form stays in it.
			for _, src := range [][]byte{text, want} {
				if formatted, err := Format(file, src, names, open); err != nil || string(formatted) != string(src) {
					t.Errorf("format changes the text (%v):\n%s", err, formatted)
				}
			}
		})
	}
}

var spliceRows = textLines(
	"siml 1",
	"spec hunter",
	"",
	"priority:",
	"  cast hunters_mark if not target.aura(hunters_mark).is_active",
	"  // Bite whenever it is up.",
	"  /// Mongoose Bite needs a dodge first.",
	"  cast mongoose_bite",
	"  cast multi_shot",
	"",
	`group "Weave":`,
	"  move 5 if auto_time_to_next(main_hand_auto) < 1s",
	"  move 11",
)

func TestSplice(t *testing.T) {
	check := func(t *testing.T, got, want []lineEdit) {
		t.Helper()
		if !slices.Equal(got, want) {
			t.Errorf("got %q\nwant %q", got, want)
		}
	}

	t.Run("moves a row with its notes and the comment above it, and changes no other line", func(t *testing.T) {
		block := "  // Bite whenever it is up.\n  /// Mongoose Bite needs a dodge first.\n  cast mongoose_bite\n"
		check(t, spliceChanges(t, spliceRows, `[{"op":"move","path":["priorityList"],"from":1,"to":2}]`), []lineEdit{{5, 8, ""}, {9, 9, block}})
	})

	t.Run("prints only the row an edit inside a group changes", func(t *testing.T) {
		edits := `[{"op":"set","path":["groups",0,"actions",1,"action","move","rangeFromTarget","const","val"],"value":"8"}]`
		check(t, spliceChanges(t, spliceRows, edits), []lineEdit{{12, 13, "  move 8\n"}})
	})

	t.Run("writes a row with a new spell by its name, and changes no other line", func(t *testing.T) {
		edits := `[{"op":"insert","path":["priorityList"],"index":3,"value":{"action":{"castSpell":{"spellId":{"spellId":3045}}}}}]`
		check(t, spliceChanges(t, spliceRows, edits), []lineEdit{{9, 9, "  cast rapid_fire\n"}})
	})

	t.Run("names a spell whatever its rank, and writes an id the table does not name as raw JSON", func(t *testing.T) {
		ranked := `[{"op":"set","path":["priorityList",2],"value":{"action":{"castSpell":{"spellId":{"spellId":2643,"rank":3}}}}}]`
		check(t, spliceChanges(t, spliceRows, ranked), []lineEdit{})
		tagged := `[{"op":"set","path":["priorityList",2],"value":{"action":{"castSpell":{"spellId":{"spellId":2643,"tag":7}}}}}]`
		check(t, spliceChanges(t, spliceRows, tagged), []lineEdit{{8, 9, "  cast `{\"spellId\":2643,\"tag\":7}`\n"}})
	})

	t.Run("reads each edit against the rotation the edits before it leave", func(t *testing.T) {
		edits := `[
			{"op":"insert","path":["groups",0,"actions"],"index":0,"value":{"action":{"castSpell":{"spellId":{"spellId":2643}}}}},
			{"op":"set","path":["groups",0,"actions",0,"action","condition"],"value":{"const":{"val":"true"}}},
			{"op":"move","path":["groups",0,"actions"],"from":0,"to":2}
		]`
		check(t, spliceChanges(t, spliceRows, edits), []lineEdit{{13, 13, "  cast multi_shot if true\n"}})
	})

	t.Run("refuses a path through an index the rotation does not hold, or to a place it does not have", func(t *testing.T) {
		for edit, want := range map[string]string{
			`{"op":"set","path":["priorityList",3,"hide"],"value":true}`:                                 "priorityList[3] does not exist",
			`{"op":"set","path":["groups",1,"name"],"value":"G"}`:                                        "groups[1] does not exist",
			`{"op":"set","path":["groups",0,"actions",2]}`:                                               "groups[0].actions[2] does not exist",
			`{"op":"set","path":["priorityList",0,"action","castSpell","target","type"],"value":"Self"}`: "priorityList[0].action.castSpell.target.type does not exist",
			`{"op":"set","path":["priorityList",0,"notes"]}`:                                             "priorityList[0].notes does not exist",
			`{"op":"insert","path":["priorityList"],"index":4,"value":{"action":{}}}`:                    "priorityList holds 3, so nothing goes in at 4",
			`{"op":"move","path":["groups",0,"actions"],"from":2,"to":0}`:                                "groups[0].actions[2] does not exist",
			`{"op":"move","path":["priorityList",0,"action"],"from":0,"to":1}`:                           "priorityList[0].action is no list",
			`{"op":"set","path":["type"],"value":"TypeAuto"}`:                                            "type is not in a row, var line or group, which are what the panel edits",
		} {
			if got := spliceRefused(t, splicePath, spliceRows, "["+edit+"]", nil); got != want {
				t.Errorf("%s: got %q, want %q", edit, got, want)
			}
		}
	})

	t.Run("refuses a file of another kind, a text with errors, and a rotation put together from other files", func(t *testing.T) {
		open := func(string) ([]byte, error) {
			return []byte(textLines("siml 1", "spec hunter", "", "priority:", "  cast arcane_shot")), nil
		}
		for _, c := range []struct{ path, text, want string }{
			{"x.siml", spliceRows, "the panel edits the rotation of an x.apl.siml file"},
			{splicePath, strings.Replace(spliceRows, "cast multi_shot", "cast multi", 1), "the text has errors; fix them to edit the rotation"},
			{splicePath, textLines("siml 1", "spec hunter", `use "y.apl.siml"`), "the rotation is put together with y.apl.siml; edit it as text"},
		} {
			if got := spliceRefused(t, c.path, c.text, "[]", open); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		}
	})

	t.Run("edits a rotation whose used file brings only variables and defs, reading its calls and references back", func(t *testing.T) {
		used := textLines("siml 1", "spec hunter", "", "var $range = 12", "def far = self.unit_distance > $range")
		open := func(string) ([]byte, error) { return []byte(used), nil }
		text := textLines("siml 1", "spec hunter", `use "y.apl.siml"`, "", "var $a = 1", "", "priority:", "  cast multi_shot if far", "  cast arcane_shot")
		d := Analyze(splicePath, []byte(text), corpusNames(t), open)
		out, err := d.Splice(aplEdits(t, `[{"op":"set","path":["valueVariables",0,"value"],"value":{"variableRef":{"name":"range"}}}]`))
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != 1 || out[0].NewText != "var $a = $range\n" {
			t.Errorf("edits %+v", out)
		}
		// The panel's JSON lists the used file's variables after the file's own: a new one goes after the file's own, and
		// those of the used file are edited there.
		out, err = d.Splice(aplEdits(t, `[{"op":"insert","path":["valueVariables"],"index":2,"value":{"name":"b","value":{"const":{"val":"2"}}}}]`))
		if err != nil || len(out) != 1 || out[0].Range.Start.Line != 6 || out[0].NewText != "var $b = 2\n" {
			t.Errorf("a variable put after the used file's: %+v, %v", out, err)
		}
		for _, edits := range []string{
			`[{"op":"set","path":["valueVariables",1,"value"],"value":{"const":{"val":"3"}}}]`,
			`[{"op":"move","path":["valueVariables"],"from":1,"to":0}]`,
		} {
			if _, err := d.Splice(aplEdits(t, edits)); err == nil || err.Error() != "range is declared in y.apl.siml; edit it there" {
				t.Errorf("%s: %v", edits, err)
			}
		}
		// A file whose every variable a used file brings holds none of its own.
		d = Analyze(splicePath, []byte(textLines("siml 1", "spec hunter", `use "y.apl.siml"`)), corpusNames(t), open)
		out, err = d.Splice(aplEdits(t, `[{"op":"insert","path":["priorityList"],"index":0,"value":{"action":{"castSpell":{"spellId":{"spellId":2643}}}}}]`))
		if err != nil || len(out) != 1 || out[0].NewText != "\npriority:\n  cast multi_shot\n" {
			t.Errorf("a row in a file of a use line alone: %+v, %v", out, err)
		}
	})

	t.Run("refuses a def rename that a file using this one would miss", func(t *testing.T) {
		common := filepath.Join("testdata", "splice", "use", "common.apl.siml")
		text, err := os.ReadFile(common)
		if err != nil {
			t.Fatal(err)
		}
		for edits, want := range map[string]string{
			`[{"op":"defRename","name":"at_range","text":"far_off"}]`: "the def at_range is called in main.apl.siml; rename it in the text with F2, which renames those calls too",
			`[{"op":"defRename","name":"spare","text":"kept"}]`:       "",
		} {
			if got := spliceRefused(t, common, string(text), edits, os.ReadFile); got != want {
				t.Errorf("%s: got %q, want %q", edits, got, want)
			}
		}
	})

	t.Run("refuses an edit that changes what a use line's with list reads", func(t *testing.T) {
		main := filepath.Join("testdata", "splice", "use", "main.apl.siml")
		text, err := os.ReadFile(main)
		if err != nil {
			t.Fatal(err)
		}
		want := "a use line's with list reads a variable the edits rename or remove; change the with list first"
		for _, edits := range []string{`[{"op":"set","path":["valueVariables",0]}]`, `[{"op":"set","path":["valueVariables",0,"name"],"value":"Weave time"}]`} {
			if got := spliceRefused(t, main, string(text), edits, os.ReadFile); got != want {
				t.Errorf("%s: got %q", edits, got)
			}
		}
	})

	t.Run("refuses an edit that leaves the rotation with an error", func(t *testing.T) {
		text := textLines("siml 1", "spec hunter", "", "var $a = 1", "var $b = $a")
		got := spliceRefused(t, splicePath, text, `[{"op":"set","path":["valueVariables",0,"value"],"value":{"variableRef":{"name":"b"}}}]`, nil)
		if want := "the edited rotation has an error: a refers to itself through other variables"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("prints again a statement whose meaning an edit changes", func(t *testing.T) {
		text := textLines("siml 1", "spec hunter", "", "var ${Use Multi-Shot} = true", "", "priority:", "  cast multi_shot if $use_multi_shot")
		check(t, spliceChanges(t, text, `[{"op":"set","path":["valueVariables",0]}]`), []lineEdit{{3, 5, ""}, {6, 7, "  cast multi_shot if ${Use Multi-Shot}\n"}})
	})

	t.Run("moves a group with the changes made inside it", func(t *testing.T) {
		text := textLines("siml 1", "spec hunter", "", `group "A":`, "  // Shoot.", "  cast multi_shot", "", `group "B":`, "  none")
		edits := `[
			{"op":"set","path":["groups",0,"name"],"value":"C"},
			{"op":"set","path":["groups",0,"actions",0,"hide"],"value":true},
			{"op":"move","path":["groups"],"from":0,"to":1}
		]`
		want := textLines("siml 1", "spec hunter", "", `group "B":`, "  none", "", `group "C":`, "  // Shoot.", "  # cast multi_shot")
		if got := splicedText(t, text, edits); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("puts a row written on its header line on a line of its own when its list changes", func(t *testing.T) {
		text := textLines("siml 1", "spec hunter", "", "priority: cast multi_shot // first", "  none")
		want := textLines("siml 1", "spec hunter", "", "priority:", "  none", "  cast multi_shot // first")
		if got := splicedText(t, text, `[{"op":"move","path":["priorityList"],"from":1,"to":0}]`); got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
		check(t, spliceChanges(t, text, `[{"op":"set","path":["priorityList",1,"hide"],"value":true}]`), []lineEdit{{4, 5, "  # none\n"}})
	})

	t.Run("replaces a list too long to compare line by line in one edit", func(t *testing.T) {
		rows := make([]string, 1100)
		for k := range rows {
			rows[k] = fmt.Sprintf("  cast multi_shot if current_time > %ds", k)
		}
		text := textLines(append([]string{"siml 1", "spec hunter", "", "priority:"}, rows...)...)
		var rotation struct{ PriorityList []json.RawMessage }
		if err := json.Unmarshal(Analyze(splicePath, []byte(text), corpusNames(t), nil).RotationJSON(), &rotation); err != nil {
			t.Fatal(err)
		}
		slices.Reverse(rotation.PriorityList)
		value, _ := json.Marshal(rotation.PriorityList)
		slices.Reverse(rows)
		check(t, spliceChanges(t, text, `[{"op":"set","path":["priorityList"],"value":`+string(value)+`}]`), []lineEdit{{4, 1104, textLines(rows...)}})
	})

	t.Run("writes new lines with the line breaks of the text, and adds none at its end", func(t *testing.T) {
		move := `[{"op":"move","path":["priorityList"],"from":1,"to":2}]`
		crlf := strings.ReplaceAll(spliceRows, "\n", "\r\n")
		if got, want := splicedText(t, crlf, move), strings.ReplaceAll(splicedText(t, spliceRows, move), "\n", "\r\n"); got != want {
			t.Errorf("got:\n%q\nwant:\n%q", got, want)
		}
		open := strings.TrimRight(textLines("siml 1", "spec hunter", "", "priority:", "  none"), "\n")
		if got, want := splicedText(t, open, `[{"op":"insert","path":["priorityList"],"index":1,"value":{"action":{}}}]`), open+"\n  none"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
