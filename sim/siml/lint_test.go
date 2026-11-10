package siml

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

// Lint, over each check on its own; lint_db_test.go runs it over the committed rotations.

// The findings on the rotation the JSON text is.
func lintText(t *testing.T, text string) []Finding {
	t.Helper()
	r := &proto.APLRotation{}
	if err := protojson.Unmarshal([]byte(text), r); err != nil {
		t.Fatalf("the rotation is no APLRotation: %v\n%s", err, text)
	}
	return Lint(r)
}

// `"castSpell": …` for the spell id, to go in an action.
func castOf(id int) string { return fmt.Sprintf(`"castSpell":{"spellId":{"spellId":%d}}`, id) }

func constOfText(val string) string { return fmt.Sprintf(`{"const":{"val":%q}}`, val) }

func findingsOf(findings []Finding, code string) []Finding {
	return slices.DeleteFunc(slices.Clone(findings), func(f Finding) bool { return f.Code != code })
}

// Each finding as the given parts: "path" (joined with /), "severity", "code" or "message", joined with a space.
func findingLines(findings []Finding, parts ...string) []string {
	out := []string{}
	for _, f := range findings {
		var line []string
		for _, part := range parts {
			switch part {
			case "path":
				line = append(line, pathString(f.Path))
			case "severity":
				line = append(line, f.Severity)
			case "code":
				line = append(line, f.Code)
			case "message":
				line = append(line, f.Message)
			}
		}
		out = append(out, strings.Join(line, " "))
	}
	return out
}

func wantLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if want == nil {
		want = []string{}
	}
	if !slices.Equal(got, want) {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestLintEmptyAction(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {}},
			{"action": {"groupReference": {"groupName": "Stance ON/OFF"}}},
			{"action": {"condition": `+constOfText("true")+`}}
		],
		"groups": [{"name": "Stance ON/OFF", "actions": [{"action": {}}]}]
	}`)
	// An empty action is an error, except in a toggle group.
	wantLines(t, findingLines(findingsOf(findings, "empty-action"), "path", "severity"), "priorityList/0/action error", "priorityList/2/action error")
}

func TestLintEmptyValue(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"condition": {}, `+castOf(1)+`}},
			{"action": {"condition": {"or": {"vals": [{"gcdIsReady": {}}, {}]}}, `+castOf(2)+`}},
			{"action": {"multidot": {"spellId": {"spellId": 3}, "maxOverlap": {}}}},
			{"action": {"channelSpell": {"spellId": {"spellId": 4}, "interruptIf": {}}}},
			{"action": {"move": {"rangeFromTarget": {}}}},
			{"action": {"moveDuration": {}}},
			{"action": {"groupReference": {"groupName": "G", "variables": [{"name": "p", "value": {}}]}}},
			{"action": {"wait": {"duration": {}}}}
		],
		"groups": [{"name": "G", "actions": [{"action": {"condition": {"variablePlaceholder": {"name": "p"}}, `+castOf(5)+`}}]}]
	}`)
	// As the sim reads it in each field.
	wantLines(t, findingLines(findingsOf(findings, "empty-value"), "path", "severity", "message"),
		"priorityList/0/action/condition warning No value set; the sim reads it as no condition, so the row always runs.",
		"priorityList/1/action/condition/or/vals/1 info No value set; the sim leaves it out of the list.",
		"priorityList/2/action/multidot/maxOverlap info No value set; the sim reads as 0s.",
		"priorityList/3/action/channelSpell/interruptIf info No value set; the sim casts the channel with no interrupt condition.",
		"priorityList/4/action/move/rangeFromTarget error No value set; the sim fails to run.",
		"priorityList/6/action/groupReference/variables/0/value error No value set; the sim fails to run.",
		"priorityList/7/action/wait/duration warning No value set; the sim drops it and what holds it.",
		"priorityList/5/action/moveDuration error No duration set; the sim fails to run.",
	)
}

func TestLintGroups(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"groupReference": {"groupName": "Missing"}}},
			{"hide": true, "action": {"groupReference": {"groupName": "Off ON/OFF"}}},
			{"action": {"condition": {"actionGroupUsed": {"name": "Off ON/OFF"}}, `+castOf(1)+`}},
			{"action": {"condition": {"actionGroupUsed": {"name": "Unreferenced"}}, `+castOf(2)+`}},
			{"action": {"condition": {"actionGroupUsed": {"name": "Nowhere"}}, `+castOf(3)+`}}
		],
		"groups": [
			{"name": "Off ON/OFF", "actions": [{"action": {}}]},
			{"name": "Unreferenced", "actions": [{"action": {`+castOf(4)+`}}]},
			{"name": "Lonely", "actions": [{"action": {`+castOf(5)+`}}]}
		]
	}`)
	wantLines(t, findingLines(findingsOf(findings, "undefined-group"), "path"), "priorityList/0/action/groupReference/groupName")
	// A hidden reference is a toggle switched off, not a missing one.
	wantLines(t, findingLines(findingsOf(findings, "group-never-used"), "path"),
		"priorityList/3/action/condition/actionGroupUsed/name",
		"priorityList/4/action/condition/actionGroupUsed/name",
	)
	wantLines(t, findingLines(findingsOf(findings, "unused-group"), "path"), "groups/2/name")
}

func TestLintSequences(t *testing.T) {
	findings := lintText(t, `{
		"prepullActions": [{"action": {"sequence": {"name": "Pre", "actions": [{`+castOf(1)+`}]}}, "doAtValue": `+constOfText("-1s")+`}],
		"priorityList": [
			{"action": {"sequence": {"name": "Live", "actions": [{`+castOf(1)+`}]}}},
			{"hide": true, "action": {"sequence": {"name": "Hidden", "actions": [{`+castOf(1)+`}]}}},
			{"action": {"condition": {"sequenceIsComplete": {"sequenceName": "Live"}}, "resetSequence": {"sequenceName": "Live"}}},
			{"action": {"condition": {"sequenceIsComplete": {"sequenceName": "Hidden"}}, "resetSequence": {"sequenceName": "Pre"}}},
			{"action": {"resetSequence": {"sequenceName": "Nope"}}}
		]
	}`)
	// Sequences are found among the live priority list and group rows only.
	wantLines(t, findingLines(findingsOf(findings, "undefined-sequence"), "path"),
		"priorityList/3/action/condition/sequenceIsComplete/sequenceName",
		"priorityList/3/action/resetSequence/sequenceName",
		"priorityList/4/action/resetSequence/sequenceName",
	)
}

func TestLintVariables(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"condition": {"variableRef": {"name": "Used"}}, `+castOf(1)+`}},
			{"action": {"condition": {"variableRef": {"name": "Missing"}}, `+castOf(2)+`}},
			{"action": {"groupReference": {"groupName": "Filled", "variables": [{"name": "x", "value": `+constOfText("1s")+`}]}}},
			{"action": {"groupReference": {"groupName": "Filled"}}},
			{"action": {"condition": {"variablePlaceholder": {"name": "stray"}}, `+castOf(3)+`}}
		],
		"groups": [{
			"name": "Filled",
			"actions": [{"action": {
				"condition": {"cmp": {"op": "OpLt", "lhs": {"variablePlaceholder": {"name": "x"}}, "rhs": {"variableRef": {"name": "x"}}}},
				`+castOf(4)+`
			}}]
		}],
		"valueVariables": [
			{"name": "Used", "value": `+constOfText("true")+`},
			{"name": "Used", "value": `+constOfText("false")+`},
			{"name": "Idle", "value": `+constOfText("true")+`}
		]
	}`)
	undefined := findingsOf(findings, "undefined-variable")
	wantLines(t, findingLines(undefined, "path"), "priorityList/1/action/condition/variableRef/name", "groups/0/actions/0/action/condition/cmp/rhs/variableRef/name")
	if len(undefined) == 2 && !strings.HasPrefix(undefined[1].Message, `Only the group's variables or a reference to it set "x"; `) || !strings.HasSuffix(undefined[1].Message, " so the sim drops this.") {
		t.Errorf("the group variable's message: %s", undefined[1].Message)
	}
	wantLines(t, findingLines(findingsOf(findings, "duplicate-variable"), "path"), "valueVariables/1/name")
	wantLines(t, findingLines(findingsOf(findings, "unused-variable"), "path"), "valueVariables/2/name")
	wantLines(t, findingLines(findingsOf(findings, "placeholder-unset"), "path"), "priorityList/3/action/groupReference/groupName")
	wantLines(t, findingLines(findingsOf(findings, "placeholder-outside"), "path"), "priorityList/4/action/condition/variablePlaceholder/name")
}

// The codes the parser reports too have the parser's severities, and a sequence only a hidden or prepull row holds is
// undefined-sequence in both.
func TestLintAgreesWithTheParser(t *testing.T) {
	text := "siml 1\nspec hunter\n\nvar $dup = 1\nvar $dup = 2\n\nprepull:\n  -1s sequence \"Pre\" [cast x]\n\n" +
		"priority:\n  # sequence \"Hidden\" [cast x]\n  group \"Nowhere\"\n  reset_sequence(\"Pre\") if $nowhere\n  reset_sequence(\"Hidden\")\n"
	_, err := Parse("x.apl.siml", []byte(text), corpusNames(t), nil)
	parser := map[string][]Severity{}
	for _, d := range diagnosticsOf(t, nil, err) {
		parser[d.Code] = append(parser[d.Code], d.Severity)
	}
	findings := lintText(t, `{
		"prepullActions": [{"action": {"sequence": {"name": "Pre", "actions": [{`+castOf(1)+`}]}}, "doAtValue": `+constOfText("-1s")+`}],
		"priorityList": [
			{"hide": true, "action": {"sequence": {"name": "Hidden", "actions": [{`+castOf(1)+`}]}}},
			{"action": {"groupReference": {"groupName": "Nowhere"}}},
			{"action": {"condition": {"variableRef": {"name": "nowhere"}}, "resetSequence": {"sequenceName": "Pre"}}},
			{"action": {"resetSequence": {"sequenceName": "Hidden"}}}
		],
		"valueVariables": [{"name": "dup", "value": `+constOfText("1")+`}, {"name": "dup", "value": `+constOfText("2")+`}]
	}`)
	for _, code := range parserChecks {
		var lint []Severity
		for _, f := range findingsOf(findings, code) {
			lint = append(lint, lintSeverity[f.Severity])
		}
		if len(lint) == 0 || !slices.Equal(lint, parser[code]) {
			t.Errorf("%s: the lint gives %v, the parser %v", code, lint, parser[code])
		}
	}
}

func TestLintUnnamedReferences(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"condition": {"variableRef": {}}, `+castOf(1)+`}},
			{"action": {"groupReference": {}}},
			{"action": {"condition": {"actionGroupUsed": {}}, "resetSequence": {}}},
			{"action": {"condition": {"sequenceIsComplete": {}}, `+castOf(2)+`}}
		]
	}`)
	// A reference with no name set refers to nothing.
	serious := slices.DeleteFunc(findings, func(f Finding) bool { return f.Severity == "info" })
	wantLines(t, findingLines(serious, "code", "path", "message"),
		"undefined-group priorityList/1/action/groupReference/groupName No group name set; the sim skips this.",
		"group-never-used priorityList/2/action/condition/actionGroupUsed/name No group name set, so this is always false.",
		"undefined-sequence priorityList/2/action/resetSequence/sequenceName No sequence name set.",
		"undefined-sequence priorityList/3/action/condition/sequenceIsComplete/sequenceName No sequence name set.",
		"undefined-variable priorityList/0/action/condition/variableRef/name No variable name set; the sim drops this.",
	)
}

func TestLintPrepullTimes(t *testing.T) {
	findings := lintText(t, `{
		"prepullActions": [
			{"action": {`+castOf(1)+`}, "doAtValue": `+constOfText("-2s")+`},
			{"action": {`+castOf(2)+`}, "doAtValue": `+constOfText("-3s")+`},
			{"action": {`+castOf(3)+`}},
			{"action": {`+castOf(4)+`}, "doAtValue": `+constOfText("1s")+`},
			{"action": {`+castOf(5)+`}, "doAtValue": {"currentTime": {}}},
			{"action": {`+castOf(6)+`}, "doAtValue": `+constOfText("soon")+`},
			{"action": {`+castOf(7)+`}, "doAtValue": {"variableRef": {"name": "Start"}}},
			{"action": {`+castOf(8)+`}, "doAtValue": `+constOfText("0s")+`},
			{"hide": true, "action": {`+castOf(9)+`}, "doAtValue": `+constOfText("5s")+`},
			{"action": {`+castOf(10)+`}, "doAtValue": {"math": {"op": "OpSub", "lhs": `+constOfText("0s")+`, "rhs": {"spellCastTime": {"spellId": {"spellId": 10}}}}}}
		],
		"valueVariables": [{"name": "Start", "value": {"math": {"op": "OpSub", "lhs": `+constOfText("-1s")+`, "rhs": `+constOfText("500ms")+`}}}]
	}`)
	// Missing, after the pull, read from the fight, not a time and out of order.
	got := findingLines(findings, "code", "path")
	wantLines(t, got,
		"prepull-order prepullActions/1/doAtValue",
		"prepull-no-time prepullActions/2",
		"prepull-positive-time prepullActions/3/doAtValue",
		"prepull-not-time prepullActions/4/doAtValue",
		"prepull-not-time prepullActions/5/doAtValue",
	)
	if len(findings) > 0 && findings[0].Message != "Runs at -3s, before row 1 above it at -2s: prepull rows run in time order, not list order." {
		t.Errorf("the prepull-order message: %s", findings[0].Message)
	}
}

func TestLintConstantConditions(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"condition": `+constOfText("false")+`, `+castOf(1)+`}},
			{"action": {"condition": {"and": {"vals": [{"gcdIsReady": {}}, {"not": {"val": `+constOfText("true")+`}}]}}, `+castOf(2)+`}},
			{"action": {"condition": {"cmp": {"op": "OpLt", "lhs": `+constOfText("1s")+`, "rhs": `+constOfText("500ms")+`}}, `+castOf(3)+`}},
			{"action": {"condition": {"or": {"vals": [{"gcdIsReady": {}}, `+constOfText("true")+`]}}, `+castOf(4)+`}},
			{"action": {"condition": {"and": {"vals": [{"gcdIsReady": {}}, {"cmp": {"op": "OpGe", "lhs": `+constOfText("10%")+`, "rhs": `+constOfText("0.1")+`}}]}}, `+castOf(5)+`}},
			{"action": {"condition": `+constOfText("0")+`, `+castOf(6)+`}},
			{"action": {"condition": {"variableRef": {"name": "Off"}}, `+castOf(7)+`}}
		],
		"valueVariables": [{"name": "Off", "value": `+constOfText("false")+`}]
	}`)
	wantLines(t, findingLines(findings, "code", "severity", "path"),
		"never-runs info priorityList/0/action/condition",
		"never-runs warning priorityList/1/action/condition",
		"never-runs warning priorityList/2/action/condition",
		"always-true info priorityList/3/action/condition",
		"always-true info priorityList/4/action/condition/and/vals/1",
		"always-true info priorityList/5/action/condition",
	)
}

func TestLintShadowedRows(t *testing.T) {
	gcd := `{"gcdIsReady": {}}`
	findings := lintText(t, `{
		"priorityList": [
			{"action": {"condition": `+gcd+`, `+castOf(1)+`}},
			{"action": {"condition": {"gcdIsReady": {}, "uuid": {"value": "x"}}, `+castOf(1)+`}},
			{"action": {`+castOf(2)+`}},
			{"action": {"condition": `+gcd+`, `+castOf(2)+`}},
			{"action": {"condition": {"not": {"val": `+gcd+`}}, `+castOf(1)+`}},
			{"action": {"sequence": {"name": "A", "actions": [{`+castOf(3)+`}]}}},
			{"action": {"sequence": {"name": "A", "actions": [{`+castOf(3)+`}]}}},
			{"hide": true, "action": {`+castOf(4)+`}},
			{"action": {`+castOf(4)+`}}
		],
		"groups": [{"name": "G", "actions": [{"action": {`+castOf(5)+`}}, {"action": {`+castOf(5)+`}}]}]
	}`)
	wantLines(t, findingLines(findingsOf(findings, "shadowed-row"), "path", "message"),
		"priorityList/1 row 1 above takes this action with the same condition, so this row never runs.",
		"priorityList/3 row 3 above takes this action with no condition, so this row never runs.",
		"groups/0/actions/1 row 1 of the group above takes this action with no condition, so this row never runs.",
	)
}

func TestLintHiddenRows(t *testing.T) {
	findings := lintText(t, `{
		"priorityList": [
			{"hide": true, "action": {"condition": `+constOfText("false")+`, "groupReference": {"groupName": "G"}}},
			{"hide": true, "action": {}}
		],
		"groups": [{"name": "G", "actions": [{"action": {"condition": {"variableRef": {"name": "V"}}, `+castOf(1)+`}}]}],
		"valueVariables": [{"name": "V", "value": `+constOfText("true")+`}]
	}`)
	// Hidden rows are skipped, but their references count as uses.
	wantLines(t, findingLines(findings, "code"))
}

func TestLintNoRotation(t *testing.T) {
	if findings := Lint(nil); findings != nil {
		t.Errorf("Lint(nil) gives %v", findings)
	}
}
