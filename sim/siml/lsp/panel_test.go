package lsp

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/siml"
)

// The rotation as the panel reads an .apl.siml through siml/rotation, and its edits through siml/splice.

type rotationAnswer struct {
	Version   int            `json:"version"`
	Rotation  map[string]any `json:"rotation"`
	ReadOnly  string         `json:"readOnly"`
	ClassName string         `json:"className"`
	Findings  []panelFinding `json:"findings"`
	Spans     []panelSpan    `json:"spans"`
	Used      []string       `json:"used"`
	Spec      string         `json:"spec"`
	Names     string         `json:"names"`
	Database  string         `json:"database"`
	Structure *struct {
		Whens      []map[string]any `json:"whens"`
		Calls      []map[string]any `json:"calls"`
		Defs       []map[string]any `json:"defs"`
		WhenRanges []Range          `json:"whenRanges"`
		DefRanges  []Range          `json:"defRanges"`
	} `json:"structure"`
	Registered struct {
		Status     string `json:"status"`
		Spec       string `json:"spec"`
		Registered map[string][]struct {
			Name string         `json:"name"`
			ID   map[string]any `json:"id"`
		} `json:"registered"`
	} `json:"-"` // Read through siml/registered.
}

func readRotation(t *testing.T, c *client, uri, text string) rotationAnswer {
	t.Helper()
	c.open(uri, text)
	var answer rotationAnswer
	c.request("siml/rotation", map[string]any{"uri": uri}, &answer)
	c.request("siml/registered", map[string]any{"uri": uri}, &answer.Registered)
	return answer
}

func spliceError(c *client, uri string, version int, edits ...map[string]any) string {
	r := c.send("siml/splice", map[string]any{"uri": uri, "version": version, "edits": edits})
	if r.Error == nil {
		return ""
	}
	return r.Error.Message
}

func key(path []any) string {
	parts := make([]string, len(path))
	for i, p := range path {
		parts[i] = fmt.Sprint(p)
	}
	return strings.Join(parts, "/")
}

// The rotation path the text at a position writes, through siml/pathAt.
func pathAt(c *client, uri string, at Position) []any {
	c.t.Helper()
	var path []any
	c.request("siml/pathAt", map[string]any{"uri": uri, "position": at}, &path)
	return path
}

func positionOf(text string, offset int) Position {
	before := text[:offset]
	line := strings.Count(before, "\n")
	return Position{line, offset - strings.LastIndex(before, "\n") - 1}
}

func TestRotationReadsWithItsRows(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	r := readRotation(t, c, uri, text)
	if r.ReadOnly != "" {
		t.Errorf("readOnly = %q", r.ReadOnly)
	}
	if rows, _ := r.Rotation["priorityList"].([]any); len(rows) != 13 {
		t.Errorf("%d priority rows, want 13", len(rows))
	}
	if r.ClassName != "hunter" || r.Spec != "hunter" || r.Registered.Status != "ready" || r.Registered.Spec != "hunter" {
		t.Errorf("class %q, registered %q %q", r.ClassName, r.Registered.Status, r.Registered.Spec)
	}
	serpent := strings.Index(text, "cast serpent_sting")
	if path := pathAt(c, uri, positionOf(text, serpent+2)); len(path) < 2 || key(path[:2]) != "priorityList/5" {
		t.Errorf("the path at serpent_sting = %v", path)
	}
	var row, spell *panelSpan
	for i, s := range r.Spans {
		switch key(s.Path) {
		case "priorityList/5":
			row = &r.Spans[i]
		case "priorityList/5/action/castSpell/spellId":
			spell = &r.Spans[i]
		}
	}
	if row == nil || row.Range.Start.Line != positionOf(text, serpent).Line {
		t.Errorf("the row's span = %+v, want line %d", row, positionOf(text, serpent).Line)
	}
	if spell == nil || spell.Range.Start != positionOf(text, serpent+len("cast ")) {
		t.Errorf("the spell's span = %+v", spell)
	}
}

func TestPathAtIsTheShortestSpan(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	c.open(uri, text)
	serpent := strings.Index(text, "cast serpent_sting")
	if path := pathAt(c, uri, positionOf(text, serpent+len("cast s"))); key(path) != "priorityList/5/action/castSpell/spellId" {
		t.Errorf("the path in the spell's name = %v, want the spell id's", path)
	}
	if path := pathAt(c, uri, positionOf(text, strings.Index(text, "spec hunter"))); path == nil || len(path) != 0 {
		t.Errorf("the path on the spec line = %#v, want []", path)
	}
	broken := "siml 1\nspec hunter\n\ncast\n"
	c.change(uri, 2, broken)
	if path := pathAt(c, uri, positionOf(broken, strings.Index(broken, "cast"))); path == nil || len(path) != 0 {
		t.Errorf("the path in a file that does not read = %#v, want []", path)
	}
	if r := c.send("siml/pathAt", map[string]any{"uri": uri + ".closed", "position": pos(0, 0)}); r.Error == nil {
		t.Error("siml/pathAt answered for a file that is not open")
	}
}

func TestRotationHasItsBlocksAndDefs(t *testing.T) {
	c := started(t)
	_, uri, _ := fixture(t, "default.apl.siml")
	text := "siml 1\nspec hunter\n\ndef early = current_time < 10s\n\npriority:\n  when early:\n    cast multi_shot\n"
	r := readRotation(t, c, uri, text)
	s := r.Structure
	if s == nil || len(s.Whens) != 1 || len(s.Defs) != 1 || len(s.Calls) != 1 {
		t.Fatalf("structure = %+v", s)
	}
	if got := key(s.Calls[0]["path"].([]any)); got != "whens/0/condition" {
		t.Errorf("the call's path = %s", got)
	}
	if s.WhenRanges[0].Start.Line != 6 || s.DefRanges[0].Start.Line != 3 {
		t.Errorf("ranges: when %+v, def %+v", s.WhenRanges[0], s.DefRanges[0])
	}
}

func TestFileOutsideARepository(t *testing.T) {
	c := started(t)
	r := readRotation(t, c, "file:///nowhere/x.apl.siml", "siml 1\nspec hunter\n")
	if !strings.Contains(r.ReadOnly, "not in a WoWSims repository") {
		t.Errorf("readOnly = %q", r.ReadOnly)
	}
}

func TestTextWithAnErrorIsReadOnly(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	r := readRotation(t, c, uri, strings.Replace(text, "cast mongoose_bite", "cast mongoose_bit", 1))
	if r.Rotation != nil {
		t.Error("a text with an error gives a rotation")
	}
	if want := "The text has errors (line 27: no spell is named mongoose_bit; did you mean mongoose_bite?). Edits wait until it reads again."; r.ReadOnly != want {
		t.Errorf("readOnly = %q, want %q", r.ReadOnly, want)
	}
	if err := spliceError(c, uri, 1, map[string]any{"op": "set", "path": []any{"priorityList", 0, "hide"}, "value": true}); !strings.Contains(err, "the text has errors") {
		t.Errorf("splice = %q", err)
	}
}

func TestRotationPutTogetherWithAnotherFile(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "hawks.apl.siml")
	r := readRotation(t, c, uri, text)
	if rows, _ := r.Rotation["priorityList"].([]any); len(rows) != 13 {
		t.Errorf("%d priority rows, want 13", len(rows))
	}
	if want := "The rotation is put together with default.apl.siml; edit it as text."; r.ReadOnly != want {
		t.Errorf("readOnly = %q, want %q", r.ReadOnly, want)
	}
	if len(r.Used) != 1 || !strings.HasSuffix(r.Used[0], "default.apl.siml") {
		t.Errorf("used = %v", r.Used)
	}
	if err := spliceError(c, uri, 1, map[string]any{"op": "set", "path": []any{"priorityList", 0, "hide"}, "value": true}); !regexp.MustCompile(`put together with default\.apl\.siml`).MatchString(err) {
		t.Errorf("splice = %q", err)
	}
}

func TestFindingsAreAtTheirRows(t *testing.T) {
	c := started(t)
	path, _, _ := fixture(t, "default.apl.siml")
	r := readRotation(t, c, pathURI(strings.TrimSuffix(path, "default.apl.siml")+"findings.apl.siml"),
		"siml 1\nspec hunter\n\npriority:\n  cast hunters_mark\n  move none\n  cast hunters_mark\n")
	var got [][3]string
	for _, f := range r.Findings {
		rowKey := key(f.Path)
		if len(f.Path) >= 2 {
			rowKey = key(f.Path[:2])
		}
		got = append(got, [3]string{f.Code, f.Severity, rowKey})
	}
	if want := [][3]string{{"empty-value", "error", "priorityList/1"}, {"shadowed-row", "warning", "priorityList/2"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("findings = %v, want %v", got, want)
	}
}

func TestMovedRowIsOneLineInAndOneOut(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	before := readRotation(t, c, uri, text)
	var result struct{ Edits []TextEdit }
	c.request("siml/splice", map[string]any{"uri": uri, "version": 1, "edits": []any{map[string]any{"op": "move", "path": []any{"priorityList"}, "from": 4, "to": 1}}}, &result)
	var texts []string
	for _, e := range result.Edits {
		texts = append(texts, e.NewText)
	}
	if want := []string{"  cast strider_kick\n", ""}; !reflect.DeepEqual(texts, want) {
		t.Errorf("new texts = %q, want %q", texts, want)
	}
	c.change(uri, 2, applied(text, result.Edits))
	var after rotationAnswer
	c.request("siml/rotation", map[string]any{"uri": uri}, &after)
	if after.Version != 2 || !reflect.DeepEqual(after.Rotation["priorityList"].([]any)[1], before.Rotation["priorityList"].([]any)[4]) {
		t.Errorf("row 1 after the move = %v", after.Rotation["priorityList"].([]any)[1])
	}
	if err := spliceError(c, uri, 1); !strings.Contains(err, "version 2") {
		t.Errorf("a splice at a stale version = %q", err)
	}
}

func TestPathThatDoesNotExistIsRefused(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	c.open(uri, text)
	if err := spliceError(c, uri, 1, map[string]any{"op": "set", "path": []any{"priorityList", 13, "hide"}, "value": true}); err != "priorityList[13] does not exist" {
		t.Errorf("set past the end = %q", err)
	}
	if err := spliceError(c, uri, 1, map[string]any{"op": "set", "path": []any{"groups", 0, "actions", -1, "action", "condition"}, "value": map[string]any{}}); !strings.Contains(err, "does not exist") {
		t.Errorf("set at -1 = %q", err)
	}
	if err := spliceError(c, uri, 1, map[string]any{"op": "move", "path": []any{"priorityList"}, "from": 20, "to": 0}); !strings.Contains(err, "does not exist") {
		t.Errorf("move from past the end = %q", err)
	}
}

func TestPickersWriteNamedActionIDs(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	r := readRotation(t, c, uri, text)
	find := func(list, name string) map[string]any {
		for _, n := range r.Registered.Registered[list] {
			if n.Name == name {
				return n.ID
			}
		}
		return nil
	}
	innervate := find("spells", "innervate")
	want := tableSpell(t, siml.NamespaceSpell, "innervate")
	if tag, _ := innervate["tag"].(float64); innervate["spellId"] != float64(want.GetSpellId()) || tag != float64(want.GetTag()) {
		t.Errorf("innervate = %v, want the table's %v", innervate, want)
	}
	mark := find("targetAuras", "hunters_mark")
	items := r.Registered.Registered["items"]
	if len(items) == 0 {
		t.Fatal("no items with a use")
	}
	potion := items[0]
	for _, item := range items {
		if item.Name == "item.major_mana_potion" {
			potion = item
		}
	}
	rows := []any{
		map[string]any{"action": map[string]any{"castSpell": map[string]any{"spellId": innervate}}},
		map[string]any{"action": map[string]any{
			"condition": map[string]any{"auraIsActive": map[string]any{"sourceUnit": map[string]any{"type": "CurrentTarget"}, "auraId": mark}},
			"castSpell": map[string]any{"spellId": potion.ID},
		}},
	}
	var edits []any
	for i, row := range rows {
		edits = append(edits, map[string]any{"op": "insert", "path": []any{"priorityList"}, "index": 13 + i, "value": row})
	}
	var result struct{ Edits []TextEdit }
	c.request("siml/splice", map[string]any{"uri": uri, "version": 1, "edits": edits}, &result)
	var written strings.Builder
	for _, e := range result.Edits {
		written.WriteString(e.NewText)
	}
	if want := fmt.Sprintf("  cast innervate\n  cast %s if target.aura(hunters_mark).is_active\n", potion.Name); written.String() != want {
		t.Errorf("written = %q, want %q", written.String(), want)
	}
}

func TestDatabaseNamesAndIcons(t *testing.T) {
	c := started(t)
	var db struct {
		Spells map[string][2]string `json:"spells"`
		Items  map[string][2]string `json:"items"`
	}
	c.request("siml/database", map[string]any{"uri": smokeURI}, &db)
	mark := fmt.Sprint(tableSpell(t, siml.NamespaceSpell, "hunters_mark").GetSpellId())
	if db.Spells[mark][0] != "Hunter's Mark" {
		t.Errorf("spell %s = %v", mark, db.Spells[mark])
	}
	if len(db.Items) == 0 {
		t.Error("no items")
	}
}

// Whether a lies before b.
func before(a, b Position) bool {
	return a.Line < b.Line || a.Line == b.Line && a.Character < b.Character
}

// Whether r holds p, its ends included.
func contains(r Range, p Position) bool { return !before(p, r.Start) && !before(r.End, p) }
