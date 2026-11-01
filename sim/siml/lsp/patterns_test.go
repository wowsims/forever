package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// The patterns Insert Pattern offers, each inserted the way siml/insertPattern inserts it: its edits made on the
// rotation's text with Doc.Splice, and the text read back.

const patternPath = "pattern.apl.siml"

func tableNames(t *testing.T) *siml.Names {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, tableFile))
	if err != nil {
		t.Fatal(err)
	}
	names, err := siml.ParseNames(data)
	if err != nil {
		t.Fatal(err)
	}
	return names
}

// The ActionID a hunter file's name in ns reads as, from the committed name table.
func tableSpell(t *testing.T, ns siml.Namespace, name string) *proto.ActionID {
	t.Helper()
	id, ok := tableNames(t).Spell(proto.Class_ClassHunter, ns, name)
	if !ok {
		t.Fatalf("the name table has no %s %s for a hunter", ns, name)
	}
	return id
}

// The JSON as protojson compares it: an empty list is the default and left out, and an ActionID's rank is the name's
// own once it prints by one (SPEC.md §12.4), so ranks are not compared.
func normalized(v any) any {
	switch v := v.(type) {
	case []any:
		out := make([]any, len(v))
		for i, x := range v {
			out[i] = normalized(x)
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for k, x := range v {
			if list, ok := x.([]any); k == "rank" || ok && len(list) == 0 {
				continue
			}
			out[k] = normalized(x)
		}
		return out
	}
	return v
}

func jsonOf(t *testing.T, v any) any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func rotationProto(t *testing.T, apl any) *proto.APLRotation {
	t.Helper()
	raw, err := json.Marshal(apl)
	if err != nil {
		t.Fatal(err)
	}
	var r proto.APLRotation
	if err := protojson.Unmarshal(raw, &r); err != nil {
		t.Fatalf("%v: %s", err, raw)
	}
	return &r
}

func printed(t *testing.T, names *siml.Names, apl any) string {
	t.Helper()
	settings := &proto.IndividualSimSettings{Player: &proto.Player{Rotation: rotationProto(t, apl)}}
	return string(siml.Print(siml.Kind{Section: "apl"}, proto.Spec_SpecHunter, settings, nil, names))
}

// The rotation the text reads as, or the test fails with its diagnostics.
func readBack(t *testing.T, names *siml.Names, text, what string) any {
	t.Helper()
	d := siml.Analyze(patternPath, []byte(text), names, nil)
	if diags := d.Diagnostics(); len(diags) > 0 {
		t.Fatalf("%s: %v\n%s", what, diags, text)
	}
	var out any
	if err := json.Unmarshal(d.RotationJSON(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

// Whether apl prints as SIML that reads back to it: a field, kind or enum value the protos do not have, or a value of
// two kinds, does not.
func valid(t *testing.T, names *siml.Names, apl any) bool {
	t.Helper()
	read := readBack(t, names, printed(t, names, apl), "printed")
	return reflect.DeepEqual(normalized(read), normalized(jsonOf(t, apl)))
}

// What a reader would fill in: spell and aura ids, a group or variable name, the defaults for the rest.
func sampleValues(p pattern) map[string]any {
	values := map[string]any{}
	for i, ph := range p.Placeholders {
		if ph.Default != nil {
			continue
		}
		switch ph.Kind {
		case "spell", "aura", "target_aura":
			values[ph.Name] = float64(1000 + i)
		case "group":
			values[ph.Name] = "Base ON/OFF"
		case "variable":
			values[ph.Name] = "Use it"
		}
	}
	return values
}

func baseRotation() map[string]any {
	var out map[string]any
	_ = json.Unmarshal([]byte(`{
		"type": "TypeAPL",
		"prepullActions": [
			{"action": {"castSpell": {"spellId": {"spellId": 1}}}, "doAtValue": {"const": {"val": "-5s"}}},
			{"action": {"castSpell": {"spellId": {"spellId": 2}}}, "doAtValue": {"const": {"val": "-1s"}}}
		],
		"priorityList": [
			{"action": {"castSpell": {"spellId": {"spellId": 3}}}},
			{"action": {"condition": {"gcdIsReady": {}}, "castSpell": {"spellId": {"spellId": 4}}}},
			{"action": {"groupReference": {"groupName": "Rows"}}},
			{"action": {"groupReference": {"groupName": "Base ON/OFF"}}}
		],
		"groups": [
			{"name": "Rows", "actions": [{"action": {"castSpell": {"spellId": {"spellId": 5}}}}]},
			{"name": "Base ON/OFF", "actions": [{"action": {}}]}
		]
	}`), &out)
	return out
}

// The rotation after the pattern goes in.
func inserted(t *testing.T, names *siml.Names, apl any, p pattern, values map[string]any, place placement) map[string]any {
	t.Helper()
	text := printed(t, names, apl)
	d := siml.Analyze(patternPath, []byte(text), names, nil)
	var rotation any
	if err := json.Unmarshal(d.RotationJSON(), &rotation); err != nil {
		t.Fatalf("%s: %v\n%s", p.ID, err, text)
	}
	edits, err := patternEdits(rotation, p, values, place)
	if err != nil {
		t.Fatal(err)
	}
	out, err := d.Splice(edits)
	if err != nil {
		t.Fatalf("%s: %v", p.ID, err)
	}
	read, _ := readBack(t, names, applied(text, toTextEdits(out)), p.ID).(map[string]any)
	return read
}

//go:fix inline
func index(i int) *int { return new(i) }

func TestPatternsAreWellFormed(t *testing.T) {
	ids := map[string]bool{}
	for _, p := range patterns {
		if ids[p.ID] {
			t.Errorf("%s twice", p.ID)
		}
		ids[p.ID] = true
		names := map[string]bool{}
		for _, ph := range p.Placeholders {
			if names[ph.Name] {
				t.Errorf("%s.%s twice", p.ID, ph.Name)
			}
			names[ph.Name] = true
			if ph.Kind == "choice" && !slices.Contains(ph.Options, fmt.Sprint(ph.Default)) {
				t.Errorf("%s.%s: the default %v is no option", p.ID, ph.Name, ph.Default)
			}
		}
	}
}

func TestEveryPatternInsertsAValidRotation(t *testing.T) {
	names := tableNames(t)
	for _, p := range patterns {
		values := withDefaults(p, sampleValues(p))
		for _, ph := range p.Placeholders {
			if _, ok := values[ph.Name]; !ph.Optional && !ok {
				t.Errorf("%s: no value for %s", p.ID, ph.Name)
			}
		}
		places := []placement{{index: new(1)}, {list: []any{"groups", 0, "actions"}}}
		if p.Target == "condition" {
			places = []placement{{condition: []any{"priorityList", 0, "action"}}, {condition: []any{"priorityList", 1, "action"}}}
		}
		for _, place := range places {
			apl := inserted(t, names, baseRotation(), p, values, place)
			if !valid(t, names, apl) {
				t.Errorf("%s does not print and read back as itself", p.ID)
			}
			// The sample spells are not registered, and the base's rows do nothing in particular; nothing else may show.
			for _, f := range siml.Lint(rotationProto(t, apl)) {
				if f.Severity != "info" {
					t.Errorf("%s: %+v", p.ID, f)
				}
			}
		}
	}
}

func TestPatternIntoARotationWithNoLists(t *testing.T) {
	names := tableNames(t)
	for _, p := range patterns {
		place := placement{index: new(0)}
		if p.Target == "condition" {
			place = placement{}
		}
		apl := inserted(t, names, map[string]any{"type": "TypeAPL"}, p, sampleValues(p), place)
		if !valid(t, names, apl) {
			t.Errorf("%s does not print and read back as itself", p.ID)
		}
	}
}

func TestPrepullRowsKeepTimeOrder(t *testing.T) {
	names := tableNames(t)
	sequence, _ := patternByID("prepull-sequence")
	apl := inserted(t, names, baseRotation(), sequence, map[string]any{"first": 10.0, "second": 11.0, "third": 12.0, "start": "-3s", "gap": "1.5s"}, placement{})
	var got []string
	for _, r := range apl["prepullActions"].([]any) {
		got = append(got, fmt.Sprint(at(r, []any{"action", "castSpell", "spellId", "spellId"}), " ", at(r, []any{"doAtValue", "const", "val"})))
	}
	if want := []string{"1 -5s", "10 -3s", "11 -1.5s", "2 -1s", "12 0s"}; !reflect.DeepEqual(got, want) {
		t.Errorf("prepull = %v, want %v", got, want)
	}
	for _, f := range siml.Lint(rotationProto(t, apl)) {
		if strings.HasPrefix(f.Code, "prepull") {
			t.Errorf("%+v", f)
		}
	}
}

func TestPatternTimesReadAsAConstReadsThem(t *testing.T) {
	names := tableNames(t)
	sequence, _ := patternByID("prepull-sequence")
	for _, c := range []struct {
		start, gap string
		want       []string
	}{
		{"-1m", "30s", []string{"-60s", "-30s", "0s"}},
		{"-1h", "1e1", []string{"-3600s", "-3590s", "-3580s"}},
		{" -.5s ", "+1500ms", []string{"-0.5s", "1s", "2.5s"}},
		{"-2m30s", "1m", []string{"-150s", "-90s", "-30s"}},
	} {
		apl := inserted(t, names, map[string]any{"type": "TypeAPL"}, sequence, map[string]any{"first": 10.0, "second": 11.0, "third": 12.0, "start": c.start, "gap": c.gap}, placement{})
		var got []string
		for _, r := range apl["prepullActions"].([]any) {
			got = append(got, fmt.Sprint(at(r, []any{"doAtValue", "const", "val"})))
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("start %q, gap %q: prepull at %v, want %v", c.start, c.gap, got, c.want)
		}
	}
	for _, bad := range []string{"soon", "NaN", "1x"} {
		if _, err := patternEdits(nil, sequence, map[string]any{"first": 10.0, "second": 11.0, "start": bad}, placement{}); err == nil {
			t.Errorf("a start of %q inserts", bad)
		}
	}
}

func TestConditionJoinsTheExistingOne(t *testing.T) {
	names := tableNames(t)
	outside, _ := patternByID("outside-execute")
	execute := jsonOf(t, obj{"not": obj{"val": obj{"isExecutePhase": obj{"threshold": "E20"}}}})
	condition := func(apl map[string]any, row int) any {
		return at(inserted(t, names, apl, outside, nil, placement{condition: []any{"priorityList", row, "action"}}), []any{"priorityList", row, "action", "condition"})
	}
	if got := condition(baseRotation(), 0); !reflect.DeepEqual(got, execute) {
		t.Errorf("onto no condition: %v", got)
	}
	joined := jsonOf(t, obj{"and": obj{"vals": []any{obj{"gcdIsReady": obj{}}, execute}}})
	if got := condition(baseRotation(), 1); !reflect.DeepEqual(got, joined) {
		t.Errorf("onto a condition: %v", got)
	}
	base := baseRotation()
	base["priorityList"].([]any)[1].(map[string]any)["action"].(map[string]any)["condition"] = obj{"and": obj{"vals": []any{obj{"gcdIsReady": obj{}}}}}
	if got := condition(base, 1); !reflect.DeepEqual(got, joined) {
		t.Errorf("onto an and: %v", got)
	}
}

func TestPatternsBuildTheRepositoryShapes(t *testing.T) {
	build := func(id string, values map[string]any) fragment {
		p, _ := patternByID(id)
		return p.build(withDefaults(p, values))
	}
	check := func(what string, got, want any) {
		t.Helper()
		if g, w := jsonOf(t, got), jsonOf(t, want); !reflect.DeepEqual(g, w) {
			t.Errorf("%s = %v, want %v", what, g, w)
		}
	}
	// The arms Rend row and the prot debuff refresh.
	check("refresh-dot", at(build("refresh-dot", map[string]any{"spell": 11574.0}).rows[0], []any{"action", "condition"}),
		obj{"cmp": obj{"op": "OpLe", "lhs": obj{"dotRemainingTime": obj{"spellId": obj{"spellId": 11574}}}, "rhs": obj{"const": obj{"val": "1.5s"}}}})
	check("keep-debuff", build("keep-debuff", map[string]any{"spell": obj{"spellId": 11581.0, "rank": 6.0}, "remaining": "6s"}).rows[0], obj{"action": obj{
		"condition": obj{"cmp": obj{
			"op":  "OpLe",
			"lhs": obj{"auraRemainingTime": obj{"sourceUnit": obj{"type": "CurrentTarget"}, "auraId": obj{"spellId": 11581, "rank": 6}}},
			"rhs": obj{"const": obj{"val": "6s"}},
		}},
		"castSpell": obj{"spellId": obj{"spellId": 11581, "rank": 6}},
	}})
	// The retribution Exorcism mana floor and the warrior toggles.
	check("resource-above", build("resource-above", nil).condition, obj{"cmp": obj{"op": "OpGe", "lhs": obj{"currentManaPercent": obj{}}, "rhs": obj{"const": obj{"val": "10%"}}}})
	check("resource-above rage", at(build("resource-above", map[string]any{"resource": "currentRage", "amount": 40.0}).condition, []any{"cmp", "rhs"}), obj{"const": obj{"val": "40"}})
	toggle := build("toggle-group", map[string]any{"group": "Stance Dance ON/OFF"})
	check("toggle-group", obj{"groups": toggle.groups, "rows": toggle.rows}, obj{
		"groups": []any{obj{"name": "Stance Dance ON/OFF", "actions": []any{obj{"action": obj{}}}}},
		"rows":   []any{obj{"action": obj{"groupReference": obj{"groupName": "Stance Dance ON/OFF"}}}},
	})
	check("toggle-is-on", build("toggle-is-on", map[string]any{"group": "Stance Dance ON/OFF"}).condition, obj{"actionGroupUsed": obj{"name": "Stance Dance ON/OFF"}})
	check("prepull-cast", build("prepull-cast", map[string]any{"spell": 2687.0, "at": -3.0}).prepullRows,
		[]any{obj{"action": obj{"castSpell": obj{"spellId": obj{"spellId": 2687}}}, "doAtValue": obj{"const": obj{"val": "-3s"}}}})
}

func TestPlacement(t *testing.T) {
	cases := []struct {
		path   []any
		target string
		want   placement
	}{
		{[]any{"priorityList", 3, "action", "castSpell"}, "row", placement{list: []any{"priorityList"}, index: new(4)}},
		{[]any{"groups", 1, "actions", 0}, "row", placement{list: []any{"groups", 1, "actions"}, index: new(1)}},
		{[]any{"groups", 0, "actions", -1}, "row", placement{list: []any{"groups", 0, "actions"}, index: new(0)}}, // The start of an empty group.
		{[]any{"priorityList", 3}, "condition", placement{condition: []any{"priorityList", 3, "action"}}},
		{[]any{"groups", 0, "actions", -1}, "condition", placement{}}, // An empty group has no row.
		{[]any{"prepullActions", 0}, "condition", placement{}},        // The sim does not read a prepull row's condition.
		{[]any{}, "row", placement{}},
	}
	for _, c := range cases {
		if got := placementAt(c.path, c.target); !reflect.DeepEqual(got, c.want) {
			t.Errorf("placementAt(%v, %s) = %+v, want %+v", c.path, c.target, got, c.want)
		}
	}
}

func TestInsertPatternRequest(t *testing.T) {
	c := started(t)
	_, uri, text := fixture(t, "default.apl.siml")
	c.open(uri, text)
	var list []pattern
	c.request("siml/patterns", nil, &list)
	if len(list) != len(patterns) || list[0].ID != "refresh-dot" {
		t.Fatalf("patterns = %d, first %+v", len(list), list[0])
	}
	var result struct{ Edits []TextEdit }
	c.request("siml/insertPattern", map[string]any{"uri": uri, "version": 1, "id": "potion-at-pull", "values": map[string]any{}, "path": []any{"priorityList", 0}}, &result)
	if len(result.Edits) != 1 || !strings.Contains(result.Edits[0].NewText, "potion") || result.Edits[0].Range.Start.Line != 24 {
		t.Errorf("edits = %+v", result.Edits)
	}
	r := c.send("siml/insertPattern", map[string]any{"uri": uri, "version": 1, "id": "outside-execute", "values": map[string]any{}, "path": []any{"prepullActions", 0}})
	if r.Error == nil || r.Error.Message != "Outside the execute phase is a condition: put the cursor on a priority or group row to add it there." {
		t.Errorf("a condition outside a row = %+v", r.Error)
	}
	r = c.send("siml/insertPattern", map[string]any{"uri": uri, "version": 7, "id": "potion-at-pull", "values": map[string]any{}, "path": []any{}})
	if r.Error == nil || !strings.Contains(r.Error.Message, "changed while the pattern was being filled in") {
		t.Errorf("a stale version = %+v", r.Error)
	}
}

func TestPrepullSecondsRefusesNaNAndInfinity(t *testing.T) {
	for _, val := range []string{"NaN", "Inf", "-inf"} {
		row := map[string]any{"doAtValue": map[string]any{"const": map[string]any{"val": val}}}
		if _, ok := prepullSeconds(row); ok {
			t.Errorf("prepullSeconds(%q) reads as a time", val)
		}
	}
	row := map[string]any{"doAtValue": map[string]any{"const": map[string]any{"val": "-1.5s"}}}
	if s, ok := prepullSeconds(row); !ok || s != -1.5 {
		t.Errorf("prepullSeconds(-1.5s) = %v, %v", s, ok)
	}
}
