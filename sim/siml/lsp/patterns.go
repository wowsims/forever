package lsp

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// Common APL patterns as JSON fragments of an APLRotation, each with the values a reader fills in, for Insert
// Pattern. The shapes are the ones the repository's APL files use: the toggle groups of the warrior files, the
// `Use ...` switch variables of the hunter and retribution files, DoT and debuff refreshes, execute rows and prepull
// casts.

// A value a pattern takes. Kind is what it takes: a spell, an aura on the player (aura) or one on the target
// (target_aura), each an ActionID of the spec's names; or a duration, number, group, variable or choice.
type placeholder struct {
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Kind     string   `json:"kind"`
	Default  any      `json:"default,omitempty"`  // Left out when the reader must give one, or when Optional lets it be left out.
	Optional bool     `json:"optional,omitempty"` // Escape leaves it out.
	Existing bool     `json:"existing,omitempty"` // A group or variable the rotation has, rather than a new one's name.
	Options  []string `json:"options,omitempty"`  // What a choice takes.
}

// What a pattern builds, by the part of the rotation each piece belongs to.
type fragment struct {
	rows           []any // APLListItem, for the row list the pattern is inserted into.
	prepullRows    []any // APLPrepullAction.
	groups         []any // APLGroup, appended to groups.
	valueVariables []any // APLValueVariable, appended to valueVariables.
	condition      any   // APLValue.
}

type pattern struct {
	ID           string        `json:"id"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	Target       string        `json:"target"` // Where its main piece goes: row, condition, group (with the row that references it) or prepull.
	Placeholders []placeholder `json:"placeholders"`
	build        func(v map[string]any) fragment
}

type obj = map[string]any

// An action id from a value: a spell id number, or an action id object as the file writes one.
func actionIDOf(value any) any {
	switch v := value.(type) {
	case float64:
		return obj{"spellId": v}
	case int:
		return obj{"spellId": v}
	case map[string]any:
		return v
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(text(value)), 64)
	if err != nil || math.IsNaN(n) {
		n = 0
	}
	return obj{"spellId": n}
}

// A value as text, as JavaScript's String writes it.
func text(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return number(v)
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	}
	return fmt.Sprint(value)
}

func number(v float64) string {
	if v == 0 {
		return "0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// A time as const takes it: a string as written ("1.5s", "-3s") without the spaces around it, a number as seconds.
func duration(value any) string {
	switch v := value.(type) {
	case float64:
		return number(v) + "s"
	case int:
		return strconv.Itoa(v) + "s"
	}
	return strings.TrimSpace(text(value))
}

func constant(val string) obj { return obj{"const": obj{"val": val}} }

func castOf(spell any) obj { return obj{"castSpell": obj{"spellId": actionIDOf(spell)}} }

func cmp(op string, lhs, rhs any) obj { return obj{"cmp": obj{"op": op, "lhs": lhs, "rhs": rhs}} }

func row(action obj, condition any) obj {
	if condition == nil {
		return obj{"action": action}
	}
	merged := obj{"condition": condition}
	maps.Copy(merged, action)
	return obj{"action": merged}
}

// The resources a threshold reads, and whether the amount is a percent.
var resources = []struct {
	key     string
	percent bool
}{
	{"currentManaPercent", true},
	{"currentHealthPercent", true},
	{"currentRage", false},
	{"currentEnergy", false},
	{"currentComboPoints", false},
}

func resourceAtLeast(v map[string]any) obj {
	resource, amount := text(v["resource"]), text(v["amount"])
	for _, r := range resources {
		if r.key == resource && r.percent {
			amount += "%"
		}
	}
	return cmp("OpGe", obj{resource: obj{}}, constant(amount))
}

func spellPlaceholder(name, label string) placeholder {
	return placeholder{Name: name, Label: label, Kind: "spell"}
}

var (
	// The thresholds is_execute_phase takes, in the enum's order, Unknown left out; the lowest is the default.
	thresholdPlaceholder = func() placeholder {
		p := placeholder{Name: "threshold", Label: "Execute below", Kind: "choice"}
		values := proto.APLValueIsExecutePhase_Unknown.Descriptor().Values()
		for i := range values.Len() {
			if v := values.Get(i); v.Number() != 0 {
				p.Options = append(p.Options, string(v.Name()))
			}
		}
		p.Default = p.Options[0]
		return p
	}()
	resourcePlaceholder = func() placeholder {
		p := placeholder{Name: "resource", Label: "Resource", Kind: "choice", Default: "currentManaPercent"}
		for _, r := range resources {
			p.Options = append(p.Options, r.key)
		}
		return p
	}()
)

// The value of a placeholder, else of a second one.
func or(v map[string]any, name, fallback string) any {
	if value, ok := v[name]; ok && value != nil {
		return value
	}
	return v[fallback]
}

var patterns = []pattern{
	{
		ID:           "refresh-dot",
		Label:        "Refresh a DoT",
		Description:  "Casts a DoT when it is off the target or its remaining time is at or below a threshold.",
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "DoT"), {Name: "remaining", Label: "Refresh at", Kind: "duration", Default: "1.5s"}},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), cmp("OpLe", obj{"dotRemainingTime": obj{"spellId": actionIDOf(v["spell"])}}, constant(duration(v["remaining"]))))}}
		},
	},
	{
		ID:          "keep-debuff",
		Label:       "Keep a debuff on the target",
		Description: "Casts a spell when its debuff on the current target is missing or about to run out.",
		Target:      "row",
		Placeholders: []placeholder{
			spellPlaceholder("spell", "Spell"),
			{Name: "aura", Label: "Debuff (default: the spell)", Kind: "target_aura", Optional: true},
			{Name: "remaining", Label: "Refresh at", Kind: "duration", Default: "3s"},
		},
		build: func(v map[string]any) fragment {
			remaining := obj{"auraRemainingTime": obj{"sourceUnit": obj{"type": "CurrentTarget"}, "auraId": actionIDOf(or(v, "aura", "spell"))}}
			return fragment{rows: []any{row(castOf(v["spell"]), cmp("OpLe", remaining, constant(duration(v["remaining"]))))}}
		},
	},
	{
		ID:           "keep-buff",
		Label:        "Keep a buff up",
		Description:  "Casts a spell whenever its buff is not on you.",
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "Spell"), {Name: "aura", Label: "Buff (default: the spell)", Kind: "aura", Optional: true}},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), obj{"not": obj{"val": obj{"auraIsActive": obj{"auraId": actionIDOf(or(v, "aura", "spell"))}}}})}}
		},
	},
	{
		ID:           "cooldown-during-aura",
		Label:        "Use a cooldown while another is active",
		Description:  "Casts a spell only while an aura, such as another cooldown, is on you.",
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "Cooldown"), {Name: "aura", Label: "While active", Kind: "aura"}},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), obj{"auraIsActive": obj{"auraId": actionIDOf(v["aura"])}})}}
		},
	},
	{
		ID:          "line-up-cooldowns",
		Label:       "Line up two cooldowns",
		Description: "Casts a cooldown while another one is active, or when the other is too far from ready to wait for.",
		Target:      "row",
		Placeholders: []placeholder{
			spellPlaceholder("spell", "Cooldown"),
			spellPlaceholder("other", "Line up with"),
			{Name: "otherAura", Label: "Its aura (default: that spell)", Kind: "aura", Optional: true},
			{Name: "maxWait", Label: "Wait at most", Kind: "duration", Default: "20s"},
		},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), obj{"or": obj{"vals": []any{
				obj{"auraIsActive": obj{"auraId": actionIDOf(or(v, "otherAura", "other"))}},
				cmp("OpGt", obj{"spellTimeToReady": obj{"spellId": actionIDOf(v["other"])}}, constant(duration(v["maxWait"]))),
			}}})}}
		},
	},
	{
		ID:           "execute-cast",
		Label:        "Cast in the execute phase",
		Description:  "Casts a spell only once the target is below the execute threshold.",
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "Spell"), thresholdPlaceholder},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), obj{"isExecutePhase": obj{"threshold": v["threshold"]}})}}
		},
	},
	{
		ID:           "execute-group",
		Label:        "Execute-phase block",
		Description:  "A group of rows that runs only in the execute phase, and the row that references it.",
		Target:       "group",
		Placeholders: []placeholder{{Name: "group", Label: "Group name", Kind: "group", Default: "Execute"}, thresholdPlaceholder},
		build: func(v map[string]any) fragment {
			group := text(v["group"])
			return fragment{
				groups: []any{obj{"name": group, "actions": []any{}}},
				rows:   []any{row(obj{"groupReference": obj{"groupName": group}}, obj{"isExecutePhase": obj{"threshold": v["threshold"]}})},
			}
		},
	},
	{
		ID:           "outside-execute",
		Label:        "Outside the execute phase",
		Description:  "True while the target is above the execute threshold.",
		Target:       "condition",
		Placeholders: []placeholder{thresholdPlaceholder},
		build: func(v map[string]any) fragment {
			return fragment{condition: obj{"not": obj{"val": obj{"isExecutePhase": obj{"threshold": v["threshold"]}}}}}
		},
	},
	{
		ID:           "resource-above",
		Label:        "Resource at or above",
		Description:  "True while a resource is at or above an amount; percent resources take a percent.",
		Target:       "condition",
		Placeholders: []placeholder{resourcePlaceholder, {Name: "amount", Label: "At least", Kind: "number", Default: float64(10)}},
		build:        func(v map[string]any) fragment { return fragment{condition: resourceAtLeast(v)} },
	},
	{
		ID:           "cast-above-resource",
		Label:        "Cast above a resource amount",
		Description:  "Casts a spell only while a resource is at or above an amount, keeping some back for higher rows.",
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "Spell"), resourcePlaceholder, {Name: "amount", Label: "At least", Kind: "number", Default: float64(10)}},
		build: func(v map[string]any) fragment {
			return fragment{rows: []any{row(castOf(v["spell"]), resourceAtLeast(v))}}
		},
	},
	{
		ID:           "toggle-group",
		Label:        "On/off toggle",
		Description:  `An empty group and the row that references it: hiding the row turns the toggle off. Conditions read it with "Toggle is on" (actionGroupUsed).`,
		Target:       "group",
		Placeholders: []placeholder{{Name: "group", Label: "Toggle name", Kind: "group", Default: "Toggle ON/OFF"}},
		build: func(v map[string]any) fragment {
			group := text(v["group"])
			return fragment{
				groups: []any{obj{"name": group, "actions": []any{obj{"action": obj{}}}}},
				rows:   []any{row(obj{"groupReference": obj{"groupName": group}}, nil)},
			}
		},
	},
	{
		ID:           "toggle-is-on",
		Label:        "Toggle is on",
		Description:  "True while the row referencing a toggle group is shown.",
		Target:       "condition",
		Placeholders: []placeholder{{Name: "group", Label: "Toggle", Kind: "group", Existing: true}},
		build: func(v map[string]any) fragment {
			return fragment{condition: obj{"actionGroupUsed": obj{"name": text(v["group"])}}}
		},
	},
	{
		ID:           "switch-variable",
		Label:        "Row behind a switch variable",
		Description:  `A "Use ..." variable set to true or false, and a row that casts the spell only while it is true.`,
		Target:       "row",
		Placeholders: []placeholder{spellPlaceholder("spell", "Spell"), {Name: "variable", Label: "Variable name", Kind: "variable", Default: "Use spell"}},
		build: func(v map[string]any) fragment {
			variable := text(v["variable"])
			return fragment{
				valueVariables: []any{obj{"name": variable, "value": constant("true")}},
				rows:           []any{row(castOf(v["spell"]), obj{"variableRef": obj{"name": variable}})},
			}
		},
	},
	{
		ID:           "potion-at-pull",
		Label:        "Potion",
		Description:  "Drinks the selected offensive potion as soon as it is ready; placed near the top, that is at the pull.",
		Target:       "row",
		Placeholders: []placeholder{},
		build: func(map[string]any) fragment {
			return fragment{rows: []any{row(obj{"castSpell": obj{"spellId": obj{"otherId": "OtherActionOffensivePotion"}}}, nil)}}
		},
	},
	{
		ID:           "prepull-cast",
		Label:        "Prepull cast",
		Description:  "Casts a spell at a time before the pull.",
		Target:       "prepull",
		Placeholders: []placeholder{spellPlaceholder("spell", "Spell"), {Name: "at", Label: "At", Kind: "duration", Default: "-1.5s"}},
		build: func(v map[string]any) fragment {
			return fragment{prepullRows: []any{obj{"action": castOf(v["spell"]), "doAtValue": constant(duration(v["at"]))}}}
		},
	},
	{
		ID:          "prepull-sequence",
		Label:       "Prepull sequence",
		Description: "Casts up to three spells before the pull, the first at a start time and each next one a gap later.",
		Target:      "prepull",
		Placeholders: []placeholder{
			spellPlaceholder("first", "First spell"),
			spellPlaceholder("second", "Second spell"),
			{Name: "third", Label: "Third spell", Kind: "spell", Optional: true},
			{Name: "start", Label: "First at", Kind: "duration", Default: "-3s"},
			{Name: "gap", Label: "Gap", Kind: "duration", Default: "1.5s"},
		},
		build: func(v map[string]any) fragment {
			start, _ := seconds(v["start"])
			gap, _ := seconds(v["gap"])
			var rows []any
			for _, spell := range []any{v["first"], v["second"], v["third"]} {
				if spell == nil || spell == "" {
					continue
				}
				at := math.Round((start+float64(len(rows))*gap)*1000) / 1000
				rows = append(rows, obj{"action": castOf(spell), "doAtValue": constant(number(at) + "s")})
			}
			return fragment{prepullRows: rows}
		},
	},
}

// Seconds from a time as a const reads one: a number, or a time string ("1.5s", "500ms", "2m30s", "-3"). False for
// text the sim reads as no number, and for NaN or an infinity, which order against nothing.
func seconds(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, !math.IsNaN(v) && !math.IsInf(v, 0)
	case int:
		return float64(v), true
	}
	n, ok := siml.ConstSeconds(strings.TrimSpace(text(value)))
	return n, ok && !math.IsNaN(n) && !math.IsInf(n, 0)
}

func patternList() []pattern { return patterns }

func patternByID(id string) (pattern, bool) {
	for _, p := range patterns {
		if p.ID == id {
			return p, true
		}
	}
	return pattern{}, false
}

// A pattern's values with the defaults filled in for those the reader left out.
func withDefaults(p pattern, values map[string]any) map[string]any {
	filled := map[string]any{}
	maps.Copy(filled, values)
	for _, ph := range p.Placeholders {
		if _, ok := filled[ph.Name]; !ok && ph.Default != nil {
			filled[ph.Name] = ph.Default
		}
	}
	return filled
}

// Where a pattern goes in the rotation: the row list and index of a row, or the action whose condition a condition
// joins.
type placement struct {
	list      []any // ['priorityList'] or ['groups', g, 'actions']; nil for the priority list.
	index     *int  // Where in that list; the end when nil.
	condition []any // The action whose condition a condition pattern joins: ['priorityList', 3, 'action'].
}

// Where a pattern goes for a rotation path: after the row the path is in (an index of -1 is the start of an empty
// list), and for a condition pattern, onto that row's action. A condition has no place outside a priority or group
// row: the sim does not read a prepull row's condition. Outside any row, a row pattern goes at the end of the
// priority list.
func placementAt(path []any, target string) placement {
	var r []any
	switch {
	case len(path) >= 4 && path[0] == "groups" && isInt(path[1]) && path[2] == "actions" && isInt(path[3]):
		r = path[:4]
	case len(path) >= 2 && (path[0] == "priorityList" || path[0] == "prepullActions") && isInt(path[1]):
		r = path[:2]
	}
	if r == nil || r[0] == "prepullActions" {
		return placement{}
	}
	index := r[len(r)-1].(int)
	if target == "condition" {
		if index < 0 {
			return placement{}
		}
		return placement{condition: append(append([]any{}, r...), "action")}
	}
	next := index + 1
	return placement{list: append([]any{}, r[:len(r)-1]...), index: &next}
}

func isInt(v any) bool {
	_, ok := v.(int)
	return ok
}

// The node at a path of decoded JSON; nil when there is none.
func at(node any, path []any) any {
	for _, key := range path {
		switch n := node.(type) {
		case map[string]any:
			k, ok := key.(string)
			if !ok {
				return nil
			}
			node = n[k]
		case []any:
			i, ok := key.(int)
			if !ok || i < 0 || i >= len(n) {
				return nil
			}
			node = n[i]
		default:
			return nil
		}
	}
	return node
}

func edit(op string, path []any, value any, index int) (siml.AplEdit, error) {
	e := siml.AplEdit{Op: op, Path: path, Index: index}
	if value != nil {
		raw, err := json.Marshal(value)
		if err != nil {
			return e, err
		}
		e.Value = raw
	}
	return e, nil
}

// The prepull time a row runs at, when it is a plain constant time.
func prepullSeconds(r any) (float64, bool) {
	val, ok := at(r, []any{"doAtValue", "const", "val"}).(string)
	if !ok {
		return 0, false
	}
	return seconds(val)
}

// The edits that put a built pattern into the rotation. Prepull rows go where their time keeps the prepull list in
// time order; a condition joins an action's existing condition with and; groups and variables are appended.
func patternEdits(apl any, p pattern, values map[string]any, place placement) ([]siml.AplEdit, error) {
	root, _ := apl.(map[string]any)
	if root == nil {
		root = map[string]any{}
	}
	filled := withDefaults(p, values)
	for _, ph := range p.Placeholders {
		if v := filled[ph.Name]; ph.Kind == "duration" && v != nil && v != "" {
			if _, ok := seconds(v); !ok {
				return nil, fmt.Errorf("%s takes a time such as 1.5s or 300ms, not %q", ph.Label, text(v))
			}
		}
	}
	f := p.build(filled)
	var edits []siml.AplEdit
	add := func(op string, path []any, value any, index int) error {
		e, err := edit(op, path, value, index)
		edits = append(edits, e)
		return err
	}
	appendTo := func(key string, items []any) error {
		if len(items) == 0 {
			return nil
		}
		list, ok := root[key].([]any)
		if !ok {
			return add("set", []any{key}, items, 0)
		}
		for i, item := range items {
			if err := add("insert", []any{key}, item, len(list)+i); err != nil {
				return err
			}
		}
		return nil
	}
	if err := appendTo("valueVariables", f.valueVariables); err != nil {
		return nil, err
	}
	if err := appendTo("groups", f.groups); err != nil {
		return nil, err
	}

	if len(f.prepullRows) > 0 {
		existing, ok := root["prepullActions"].([]any)
		if !ok {
			if err := add("set", []any{"prepullActions"}, f.prepullRows, 0); err != nil {
				return nil, err
			}
		} else {
			existing = append([]any{}, existing...)
			for _, r := range f.prepullRows {
				index := len(existing)
				if time, ok := prepullSeconds(r); ok {
					for i, other := range existing {
						if t, ok := prepullSeconds(other); ok && t > time {
							index = i
							break
						}
					}
				}
				if err := add("insert", []any{"prepullActions"}, r, index); err != nil {
					return nil, err
				}
				existing = append(existing[:index], append([]any{r}, existing[index:]...)...)
			}
		}
	}

	if len(f.rows) > 0 {
		listPath := place.list
		if listPath == nil {
			listPath = []any{"priorityList"}
		}
		list, ok := at(root, listPath).([]any)
		if !ok {
			if err := add("set", listPath, f.rows, 0); err != nil {
				return nil, err
			}
		} else {
			start := len(list)
			if place.index != nil {
				start = min(max(*place.index, 0), len(list))
			}
			for i, r := range f.rows {
				if err := add("insert", listPath, r, start+i); err != nil {
					return nil, err
				}
			}
		}
	}

	if f.condition != nil && place.condition != nil {
		conditionPath := append(append([]any{}, place.condition...), "condition")
		existing, _ := at(root, conditionPath).(map[string]any)
		meaningful := 0
		for key := range existing {
			if key != "uuid" {
				meaningful++
			}
		}
		and, _ := existing["and"].(map[string]any)
		vals, isList := and["vals"].([]any)
		var err error
		switch {
		case meaningful == 0:
			err = add("set", conditionPath, f.condition, 0)
		case isList:
			err = add("insert", append(append([]any{}, conditionPath...), "and", "vals"), f.condition, len(vals))
		default:
			err = add("set", conditionPath, obj{"and": obj{"vals": []any{existing, f.condition}}}, 0)
		}
		if err != nil {
			return nil, err
		}
	}
	return edits, nil
}
