package siml

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Static checks on a rotation: what the sim would skip, warn about or never reach, found without running it. Every
// finding points at the most specific JSON node it is about.
//
// | Code                    | Severity | Catches                                                                                           |
// | ----------------------- | -------- | ------------------------------------------------------------------------------------------------- |
// | `empty-action`          | error    | An action with no kind set (`{}`); the sim skips the row. Toggle groups are exempt (below).       |
// | `empty-value`           | warning  | A value with no kind set; the sim drops it, so a condition made of it is no condition. An empty   |
// |                         |          | item of an `and`, `or`, `max` or `min`, an empty overlap or offset (0s) and an empty channel      |
// |                         |          | interrupt (a plain cast) are `info`. A move's range or duration, empty or missing, and a          |
// |                         |          | reference's value for a placeholder are errors: the sim fails to run.                             |
// | `undefined-group`       | error    | A group reference to a group no `groups` entry names, or with no name.                            |
// | `group-never-used`      | warning  | `actionGroupUsed` of a group that no row references, hidden or not: it is always false.           |
// | `undefined-sequence`    | warning  | A sequence name no live sequence in the priority list or a group carries (prepull ones don't).    |
// | `undefined-variable`    | error    | A `variableRef` to a name no value variable has. A group's variables and those a reference sets   |
// |                         |          | only replace a value variable of the same name, so a name only they set is dropped too.           |
// | `placeholder-unset`     | error    | A group reference that leaves one of the group's `variablePlaceholder`s unfilled.                 |
// | `placeholder-outside`   | warning  | A `variablePlaceholder` outside a group, which nothing replaces.                                  |
// | `duplicate-variable`    | warning  | A value variable whose name an earlier one already takes; references read the first.              |
// | `unused-variable`       | info     | A value variable nothing references.                                                              |
// | `unused-group`          | info     | A group nothing references, by group reference or `actionGroupUsed`.                              |
// | `prepull-no-time`       | warning  | A prepull row with no Do At time; the sim skips it.                                               |
// | `prepull-positive-time` | warning  | A prepull Do At time after the pull; the sim ignores the row.                                     |
// | `prepull-not-time`      | warning  | A Do At value read from the fight; the sim reads it once, before the fight, and can fail on it.   |
// |                         |          | Constant times and spell and aura durations, which need no fight, are fine.                       |
// | `prepull-order`         | info     | A prepull row listed after one with a later time; prepull rows run in time order, not list order. |
// | `never-runs`            | warning  | A condition that is always false (the sim drops the row). A literal `false` is `info`: it keeps   |
// |                         |          | the spell out of automatic cooldown use, a deliberate way to turn a cooldown off.                 |
// | `always-true`           | info     | A condition, or a comparison, `not`, `and` or `or` inside one, that is always true or false.      |
// | `shadowed-row`          | warning  | A row whose action an earlier row in the same list takes with no condition or the same one: the   |
// |                         |          | later row is only reached when that action cannot run, so it never runs.                          |
//
// Hidden rows: the sim parses none of them, so every check skips them except the references. A hidden row's references
// are still checked and still count as uses, since unhiding the row brings them back (a hidden group reference is how
// a toggle group is switched off). A sequence that only a hidden row carries is reported, as the sim does not find it.
//
// Toggle groups: a group whose rows all have empty actions is a switch ("Stance Dance ON/OFF"): hiding or unhiding the
// row that references it turns actionGroupUsed of it off or on. Its empty actions are not reported; unused-group still
// reports one nothing references.
//
// Not followed: value variables. A condition that reads a variable holding a constant is a setting the reader flips,
// not a mistake, so constant checks stop at a variableRef.
//
// The checks read the rotation as its protojson: JSON names, fields left out where protojson leaves them out.

// A finding of Lint, at the JSON path within the rotation of what it is about.
type Finding struct {
	Path     []any  `json:"path"`     // JSON names and list indices.
	Severity string `json:"severity"` // error, warning or info.
	Code     string `json:"code"`
	Message  string `json:"message"`
}

// A JSON object: its keys and values in the order protojson writes them.
type lintObject []lintField

type lintField struct {
	key   string
	value any
}

func (o lintObject) get(key string) (any, bool) {
	for _, f := range o {
		if f.key == key {
			return f.value, true
		}
	}
	return nil, false
}

func (o lintObject) value(key string) any {
	v, _ := o.get(key)
	return v
}

// The message at key; empty when there is none.
func (o lintObject) object(key string) lintObject {
	v, _ := o.value(key).(lintObject)
	return v
}

// The messages of the list at key.
func (o lintObject) list(key string) []lintObject {
	items, _ := o.value(key).([]any)
	objects := make([]lintObject, len(items))
	for i, item := range items {
		objects[i], _ = item.(lintObject)
	}
	return objects
}

// The kind an action or value is: its one key other than an action's condition and the editor's uuid.
func (o lintObject) kind() string {
	for _, f := range o {
		if f.key != "condition" && f.key != "uuid" {
			return f.key
		}
	}
	return ""
}

// m as protojson writes it: JSON names, fields in declaration order, enums by name.
func lintJSON(m protoreflect.Message) lintObject {
	var o lintObject
	fields := m.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		v, ok := fieldValue(m, fd)
		if !ok {
			continue
		}
		var x any
		switch {
		case fd.IsList():
			items := make([]any, v.List().Len())
			for k := range items {
				items[k] = lintScalar(fd, v.List().Get(k))
			}
			x = items
		case fd.IsMap():
			var entries lintObject
			v.Map().Range(func(k protoreflect.MapKey, e protoreflect.Value) bool {
				entries = append(entries, lintField{k.String(), lintScalar(fd.MapValue(), e)})
				return true
			})
			slices.SortFunc(entries, func(a, b lintField) int { return compareObjectKeys(a.key, b.key) })
			x = entries
		default:
			x = lintScalar(fd, v)
		}
		o = append(o, lintField{fd.JSONName(), x})
	}
	return o
}

func lintScalar(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return lintJSON(v.Message())
	case protoreflect.EnumKind:
		if e := fd.Enum().Values().ByNumber(v.Enum()); e != nil {
			return string(e.Name())
		}
	}
	return v.Interface()
}

// v written out with uuids left out: the same for two values the sim reads the same, as a uuid only tags a value for
// the editor.
func canonicalJSON(v any) string {
	var b strings.Builder
	var write func(v any)
	write = func(v any) {
		switch x := v.(type) {
		case lintObject:
			b.WriteByte('{')
			for _, f := range x {
				if f.key != "uuid" {
					b.WriteString(jsonString(f.key) + ":")
					write(f.value)
					b.WriteByte(',')
				}
			}
			b.WriteByte('}')
		case []any:
			b.WriteByte('[')
			for _, item := range x {
				write(item)
				b.WriteByte(',')
			}
			b.WriteByte(']')
		case string:
			b.WriteString(jsonString(x))
		default:
			fmt.Fprint(&b, x)
		}
	}
	write(v)
	return b.String()
}

// Where a row sits: whether it is hidden, and the list it belongs to.
type lintScope struct {
	hidden  bool
	prepull bool
	group   int // The groups index a group row belongs to, -1 for none.
}

type actionSite struct {
	path   []any
	action lintObject
	scope  lintScope
	nested bool // Inside a sequence or a schedule rather than a row's own action.
}

type valueSite struct {
	path      []any
	value     lintObject
	scope     lintScope
	condition bool // The condition of an action, as opposed to an operand or an action's setting.
}

type nameSite struct {
	path  []any // The name string itself.
	name  string
	scope lintScope
}

type groupRefSite struct {
	nameSite
	variables map[string][]any // The path of the value it sets, by name.
}

type missingSite struct {
	path  []any
	what  string
	scope lintScope
}

type collected struct {
	actions         []actionSite
	values          []valueSite
	groupReferences []groupRefSite
	groupUsed       []nameSite
	sequences       []nameSite
	sequenceRefs    []nameSite
	variableRefs    []nameSite
	placeholders    []nameSite
	missing         []missingSite // An action body without the value it needs.
}

var (
	// The action kinds whose value the sim reads with no check, and fails without: the move's range and duration.
	requiredValues = map[string]struct{ field, what string }{
		"move":         {"rangeFromTarget", "range"},
		"moveDuration": {"duration", "duration"},
	}
	// How the sim reads an empty value in a field it has a fallback for, by kind.field.
	emptyFallback = map[string]string{
		"multidot.maxOverlap":          "reads as 0s",
		"strictMultidot.maxOverlap":    "reads as 0s",
		"multishield.maxOverlap":       "reads as 0s",
		"auraShouldRefresh.maxOverlap": "reads as 0s",
		"multipleCdUsages.offset":      "reads as 0s",
		"channelSpell.interruptIf":     "casts the channel with no interrupt condition",
	}
	// The kinds whose state lasts between evaluations: two of them are two separate states, never a copy.
	statefulKinds = map[string]bool{"sequence": true, "strictSequence": true, "groupReference": true, "schedule": true}
	// The value kinds whose duration needs no fight: the sim reads them before the pull with no harm.
	staticDurations = map[string]bool{"spellCastTime": true, "spellGcdHastedDuration": true, "inputDelay": true, "channelClipDelay": true, "dotBaseDuration": true, "auraDuration": true}
)

// A name as the sim reads it: a missing one is "".
func nameOf(v any) string {
	s, _ := v.(string)
	return s
}

func at(path []any, keys ...any) []any { return slices.Concat(path, keys) }

// The JSON names apl.proto gives the lint's walk: the value fields of action and value bodies, and the value kinds
// that name a sequence.
type lintFields struct {
	actionValues, valueValues, sequenceValues map[string]bool
}

var lintFieldsOnce lazy[*lintFields]

func lintFieldsOf() *lintFields {
	f, err := lintFieldsOnce.get(func() (*lintFields, error) {
		ix, err := loadIndex()
		if err != nil {
			return nil, err
		}
		f := &lintFields{
			actionValues:   valueFields(actionMessage.Oneofs().ByName("action")),
			valueValues:    valueFields(valueMessage.Oneofs().ByName("value")),
			sequenceValues: map[string]bool{},
		}
		for _, k := range ix.valueKindList {
			if k.Accessor != nil && k.Accessor.Head == "sequence" {
				f.sequenceValues[k.JSON] = true
			}
		}
		return f, nil
	})
	if err != nil {
		panic(err)
	}
	return f
}

// The JSON names of the single APLValue fields of oneof's kinds.
func valueFields(oneof protoreflect.OneofDescriptor) map[string]bool {
	out := map[string]bool{}
	kinds := oneof.Fields()
	for i := range kinds.Len() {
		fields := kinds.Get(i).Message().Fields()
		for j := range fields.Len() {
			if fd := fields.Get(j); !fd.IsList() && fd.Message() != nil && fd.Message().FullName() == valueMessage.FullName() {
				out[fd.JSONName()] = true
			}
		}
	}
	return out
}

func collect(apl lintObject) *collected {
	fields := lintFieldsOf()
	out := &collected{}
	var visitValue func(value any, path []any, scope lintScope, condition bool)
	visitValue = func(value any, path []any, scope lintScope, condition bool) {
		o, ok := value.(lintObject)
		if !ok {
			return
		}
		out.values = append(out.values, valueSite{path, o, scope, condition})
		kind := o.kind()
		body, here := o.object(kind), at(path, kind)
		// A reference with no name set reads as one to the name "", which nothing has.
		name := nameSite{at(here, "name"), nameOf(body.value("name")), scope}
		switch {
		case kind == "variableRef":
			out.variableRefs = append(out.variableRefs, name)
		case kind == "actionGroupUsed":
			out.groupUsed = append(out.groupUsed, name)
		case kind == "variablePlaceholder" && name.name != "":
			out.placeholders = append(out.placeholders, name)
		case fields.sequenceValues[kind]:
			out.sequenceRefs = append(out.sequenceRefs, nameSite{at(here, "sequenceName"), nameOf(body.value("sequenceName")), scope})
		}
		for _, f := range body {
			if fields.valueValues[f.key] {
				visitValue(f.value, at(here, f.key), scope, false)
			}
		}
		for i, item := range body.list("vals") {
			visitValue(item, at(here, "vals", i), scope, false)
		}
	}

	var visitAction func(action lintObject, path []any, scope lintScope, nested bool)
	visitAction = func(action lintObject, path []any, scope lintScope, nested bool) {
		out.actions = append(out.actions, actionSite{path, action, scope, nested})
		visitValue(action.value("condition"), at(path, "condition"), scope, true)
		kind := action.kind()
		body, here := action.object(kind), at(path, kind)
		switch kind {
		case "groupReference":
			variables := map[string][]any{}
			for i, v := range body.list("variables") {
				path := at(here, "variables", i, "value")
				if name, ok := v.value("name").(string); ok {
					variables[name] = path
				}
				visitValue(v.value("value"), path, scope, false)
			}
			out.groupReferences = append(out.groupReferences, groupRefSite{nameSite{at(here, "groupName"), nameOf(body.value("groupName")), scope}, variables})
		case "sequence":
			if name, ok := body.value("name").(string); ok {
				out.sequences = append(out.sequences, nameSite{at(here, "name"), name, scope})
			}
		case "resetSequence":
			out.sequenceRefs = append(out.sequenceRefs, nameSite{at(here, "sequenceName"), nameOf(body.value("sequenceName")), scope})
		}
		if r, ok := requiredValues[kind]; ok {
			if _, set := body.get(r.field); !set {
				out.missing = append(out.missing, missingSite{here, r.what, scope})
			}
		}
		for _, f := range body {
			switch {
			case fields.actionValues[f.key]:
				visitValue(f.value, at(here, f.key), scope, false)
			case f.key == "innerAction":
				visitAction(body.object(f.key), at(here, f.key), scope, true)
			}
		}
		for i, item := range body.list("actions") {
			visitAction(item, at(here, "actions", i), scope, true)
		}
	}

	for i, row := range apl.list("prepullActions") {
		scope := lintScope{hidden: row.value("hide") == true, prepull: true, group: -1}
		visitAction(row.object("action"), []any{"prepullActions", i, "action"}, scope, false)
		visitValue(row.value("doAtValue"), []any{"prepullActions", i, "doAtValue"}, scope, false)
	}
	for i, row := range apl.list("priorityList") {
		visitAction(row.object("action"), []any{"priorityList", i, "action"}, lintScope{hidden: row.value("hide") == true, group: -1}, false)
	}
	for g, group := range apl.list("groups") {
		for v, variable := range group.list("variables") {
			visitValue(variable.value("value"), []any{"groups", g, "variables", v, "value"}, lintScope{group: g}, false)
		}
		for i, row := range group.list("actions") {
			visitAction(row.object("action"), []any{"groups", g, "actions", i, "action"}, lintScope{hidden: row.value("hide") == true, group: g}, false)
		}
	}
	for i, variable := range apl.list("valueVariables") {
		visitValue(variable.value("value"), []any{"valueVariables", i, "value"}, lintScope{group: -1}, false)
	}
	return out
}

// --- Constants

// A constant's number as the sim's newValueConst parses it: a duration in seconds, an int32, a percent as a fraction,
// or a float. False when the sim reads it as a string.
func constNumber(val string) (float64, bool) {
	if d, err := time.ParseDuration(val); err == nil {
		return d.Seconds(), true
	}
	if n, err := strconv.Atoi(val); err == nil {
		return float64(int32(n)), true
	}
	if number, ok := strings.CutSuffix(val, "%"); ok && len(val) > 1 {
		if n, err := strconv.ParseFloat(number, 64); err == nil {
			return n / 100, true
		}
	}
	n, err := strconv.ParseFloat(val, 64)
	return n, err == nil
}

// ConstSeconds is a const value's number as the sim reads it, a duration in seconds; false when the sim reads it as a
// string.
func ConstSeconds(val string) (float64, bool) { return constNumber(val) }

// A const value's text, a missing val being ""; false when value is no const.
func constOf(value lintObject) (string, bool) {
	return nameOf(value.object("const").value("val")), value.kind() == "const"
}

// What a constant reads as for a condition: true or false in any case, otherwise true unless empty (a const of "0"
// read as a condition is true).
func constBool(val string) bool {
	lower := strings.ToLower(val)
	if lower == "false" {
		return false
	}
	return lower == "true" || val != ""
}

var compareConstants = map[string]func(a, b float64) bool{
	"OpEq": func(a, b float64) bool { return a == b },
	"OpNe": func(a, b float64) bool { return a != b },
	"OpLt": func(a, b float64) bool { return a < b },
	"OpLe": func(a, b float64) bool { return a <= b },
	"OpGt": func(a, b float64) bool { return a > b },
	"OpGe": func(a, b float64) bool { return a >= b },
}

// Whether a value present in the file is one the sim builds: an empty value, or an operator whose operands are all
// empty, comes out as nothing.
func absent(value lintObject) bool {
	body := value.object(value.kind())
	switch value.kind() {
	case "":
		return true
	case "and", "or":
		return !slices.ContainsFunc(body.list("vals"), func(v lintObject) bool { return !absent(v) })
	case "not":
		return absent(body.object("val"))
	case "cmp":
		return absent(body.object("lhs")) || absent(body.object("rhs"))
	}
	return false
}

// The value's truth if it is the same whatever happens in the fight, read as a condition; false when it is not or the
// sim does not build it.
func constantTruth(value lintObject) (truth, known bool) {
	if absent(value) {
		return false, false
	}
	kind := value.kind()
	body := value.object(kind)
	switch kind {
	case "const":
		val, _ := constOf(value)
		return constBool(val), true
	case "not":
		truth, known := constantTruth(body.object("val"))
		return !truth, known
	case "and", "or":
		decisive, all := kind == "or", true
		for _, val := range slices.DeleteFunc(body.list("vals"), absent) {
			truth, known := constantTruth(val)
			if known && truth == decisive {
				return decisive, true
			}
			all = all && known && truth == !decisive
		}
		return !decisive, all
	case "cmp":
		lhs, lok := constOf(body.object("lhs"))
		rhs, rok := constOf(body.object("rhs"))
		a, aok := constNumber(lhs)
		b, bok := constNumber(rhs)
		if compare := compareConstants[nameOf(body.value("op"))]; lok && rok && aok && bok && compare != nil {
			return compare(a, b), true
		}
	}
	return false, false
}

type constantPart struct {
	path  []any
	truth bool
	kind  string
}

// The most specific operators inside a condition that read the same whatever happens, when the condition as a whole
// does not.
func constantParts(value lintObject, path []any, found []constantPart) []constantPart {
	kind := value.kind()
	body := value.object(kind)
	switch kind {
	case "", "const":
		return found
	case "cmp", "not", "and", "or":
		if truth, known := constantTruth(value); known {
			return append(found, constantPart{path, truth, kind})
		}
	}
	for _, field := range []string{"val", "lhs", "rhs"} {
		found = constantParts(body.object(field), at(path, kind, field), found)
	}
	for i, val := range body.list("vals") {
		found = constantParts(val, at(path, kind, "vals", i), found)
	}
	return found
}

func rowLabel(path []any) string {
	if path[0] == "groups" {
		return fmt.Sprintf("row %d of the group", path[3].(int)+1)
	}
	return fmt.Sprintf("row %d", path[1].(int)+1)
}

// --- Checks

type lintReport func(path []any, severity, code, message string)

// The findings on r, in the order the checks find them.
func Lint(r *proto.APLRotation) []Finding {
	if r == nil {
		return nil
	}
	apl := lintJSON(r.ProtoReflect())
	var findings []Finding
	report := func(path []any, severity, code, message string) {
		findings = append(findings, Finding{Path: path, Severity: severity, Code: code, Message: message})
	}
	c := collect(apl)

	groups := apl.list("groups")
	referencedGroups, usedGroups := map[string]bool{}, map[string]bool{}
	for _, ref := range c.groupReferences {
		referencedGroups[ref.name] = true
	}
	for _, ref := range c.groupUsed {
		usedGroups[ref.name] = true
	}
	groupIndex := func(name string) int {
		return slices.IndexFunc(groups, func(g lintObject) bool { return g.value("name") == name })
	}
	// The placeholders each group's live rows read.
	groupPlaceholders := make([][]string, len(groups))
	for _, p := range c.placeholders {
		if g := p.scope.group; g >= 0 && !p.scope.hidden && !slices.Contains(groupPlaceholders[g], p.name) {
			groupPlaceholders[g] = append(groupPlaceholders[g], p.name)
		}
	}

	// Toggle groups: every row's action empty.
	toggle := func(g int) bool {
		rows := groups[g].list("actions")
		return len(rows) > 0 && !slices.ContainsFunc(rows, func(row lintObject) bool { return row.object("action").kind() != "" })
	}
	for _, site := range c.actions {
		if site.scope.hidden || site.action.kind() != "" || !site.nested && site.scope.group >= 0 && toggle(site.scope.group) {
			continue
		}
		report(site.path, "error", "empty-action", "No action set; the sim skips this.")
	}

	// The reference values that fill a placeholder: the sim fails on an empty one.
	filling := map[string]bool{}
	for _, ref := range c.groupReferences {
		if g := groupIndex(ref.name); g >= 0 {
			for name, path := range ref.variables {
				if slices.Contains(groupPlaceholders[g], name) {
					filling[fmt.Sprint(path)] = true
				}
			}
		}
	}
	for _, site := range c.values {
		if site.scope.hidden || site.value.kind() != "" {
			continue
		}
		n := len(site.path)
		last, second := site.path[n-1], site.path[n-2]
		fallback := emptyFallback[fmt.Sprintf("%v.%v", second, last)]
		required, isRequired := requiredValues[fmt.Sprint(second)]
		crashes := isRequired && required.field == last || filling[fmt.Sprint(site.path)]
		switch {
		case site.condition:
			report(site.path, "warning", "empty-value", "No value set; the sim reads it as no condition, so the row always runs.")
		case second == "vals":
			report(site.path, "info", "empty-value", "No value set; the sim leaves it out of the list.")
		case fallback != "":
			report(site.path, "info", "empty-value", "No value set; the sim "+fallback+".")
		case crashes:
			report(site.path, "error", "empty-value", "No value set; the sim fails to run.")
		default:
			report(site.path, "warning", "empty-value", "No value set; the sim drops it and what holds it.")
		}
	}
	for _, site := range c.missing {
		if !site.scope.hidden {
			report(site.path, "error", "empty-value", "No "+site.what+" set; the sim fails to run.")
		}
	}

	for _, ref := range c.groupReferences {
		switch {
		case groupIndex(ref.name) >= 0:
		case ref.name != "":
			report(ref.path, "error", "undefined-group", fmt.Sprintf("No group is named \"%s\".", ref.name))
		default:
			report(ref.path, "error", "undefined-group", "No group name set; the sim skips this.")
		}
	}
	for _, ref := range c.groupUsed {
		switch {
		case ref.scope.hidden:
		case groupIndex(ref.name) < 0 && ref.name != "":
			report(ref.path, "warning", "group-never-used", fmt.Sprintf("No group is named \"%s\", so this is always false.", ref.name))
		case groupIndex(ref.name) < 0:
			report(ref.path, "warning", "group-never-used", "No group name set, so this is always false.")
		case !referencedGroups[ref.name]:
			report(ref.path, "warning", "group-never-used", fmt.Sprintf("No row references group \"%s\", so this is always false.", ref.name))
		}
	}
	for g, group := range groups {
		if name, ok := group.value("name").(string); ok && !referencedGroups[name] && !usedGroups[name] {
			report([]any{"groups", g, "name"}, "info", "unused-group", fmt.Sprintf("Nothing references group \"%s\".", name))
		}
	}

	// Sequences: the sim finds them among the live priority list and group rows.
	liveSequences, anySequences := map[string]bool{}, map[string]bool{}
	for _, seq := range c.sequences {
		if !seq.scope.hidden && !seq.scope.prepull {
			liveSequences[seq.name] = true
		}
		anySequences[seq.name] = true
	}
	for _, ref := range c.sequenceRefs {
		switch {
		case liveSequences[ref.name]:
		case ref.name == "":
			report(ref.path, "warning", "undefined-sequence", "No sequence name set.")
		case anySequences[ref.name]:
			report(ref.path, "warning", "undefined-sequence", fmt.Sprintf("Only a hidden or prepull row carries sequence \"%s\"; the sim does not find it.", ref.name))
		default:
			report(ref.path, "warning", "undefined-sequence", fmt.Sprintf("No sequence is named \"%s\".", ref.name))
		}
	}

	// Variables.
	variables := apl.list("valueVariables")
	variableNames, referencedVariables := map[string]bool{}, map[string]bool{}
	for i, variable := range variables {
		if name, ok := variable.value("name").(string); ok {
			if variableNames[name] {
				report([]any{"valueVariables", i, "name"}, "warning", "duplicate-variable", fmt.Sprintf("An earlier variable is also named \"%s\"; references read that one.", name))
			}
			variableNames[name] = true
		}
	}
	for _, ref := range c.variableRefs {
		referencedVariables[ref.name] = true
	}
	// Whether group g's variables or a reference to it set name. The sim reads a group row's references as the value
	// variables' and lets these replace one only where a value variable of the name exists.
	groupSets := func(g int, name string) bool {
		return slices.ContainsFunc(groups[g].list("variables"), func(v lintObject) bool { return v.value("name") == name }) ||
			slices.ContainsFunc(c.groupReferences, func(ref groupRefSite) bool {
				_, set := ref.variables[name]
				return set && groups[g].value("name") == ref.name
			})
	}
	for _, ref := range c.variableRefs {
		message := fmt.Sprintf("No variable is named \"%s\".", ref.name)
		switch {
		case variableNames[ref.name]:
			continue
		case ref.name == "":
			message = "No variable name set; the sim drops this."
		case ref.scope.group >= 0 && groupSets(ref.scope.group, ref.name):
			message = fmt.Sprintf("Only the group's variables or a reference to it set \"%s\"; they replace a value variable of that name, and there is none, so the sim drops this.", ref.name)
		}
		report(ref.path, "error", "undefined-variable", message)
	}
	for i, variable := range variables {
		if name, ok := variable.value("name").(string); ok && !referencedVariables[name] {
			report([]any{"valueVariables", i, "name"}, "info", "unused-variable", fmt.Sprintf("Nothing references variable \"%s\".", name))
		}
	}

	// Placeholders: each live reference to a group must fill every placeholder the group's live rows read.
	for _, p := range c.placeholders {
		if !p.scope.hidden && p.scope.group < 0 {
			report(p.path, "warning", "placeholder-outside", fmt.Sprintf("Placeholder \"%s\" is outside a group; nothing fills it, so it reads as 0 or false.", p.name))
		}
	}
	for _, ref := range c.groupReferences {
		g := groupIndex(ref.name)
		if g < 0 {
			continue
		}
		var missing []string
		for _, name := range groupPlaceholders[g] {
			if _, set := ref.variables[name]; !set {
				missing = append(missing, fmt.Sprintf("\"%s\"", name))
			}
		}
		if len(missing) > 0 {
			report(ref.path, "error", "placeholder-unset", fmt.Sprintf("Group \"%s\" needs %s set on this reference.", ref.name, strings.Join(missing, ", ")))
		}
	}

	lintPrepull(apl, variables, report)
	lintConditions(c, report)
	lintShadowed(c, report)
	return findings
}

// The seconds a Do At value reads as before the fight, known for a constant, a variable holding one, or arithmetic on
// those, NaN for a constant that is no number; static when it reads nothing of the fight: constants, those durations,
// and variables and arithmetic on them.
func doAtSeconds(value lintObject, variables []lintObject, seen []string) (seconds float64, known, static bool) {
	kind := value.kind()
	body := value.object(kind)
	switch {
	case kind == "const":
		val, _ := constOf(value)
		if n, ok := constNumber(val); ok {
			return n, true, true
		}
		return math.NaN(), true, true
	case staticDurations[kind]:
		return 0, false, true
	case kind == "variableRef":
		name, ok := body.value("name").(string)
		k := slices.IndexFunc(variables, func(v lintObject) bool { return v.value("name") == name })
		if !ok || slices.Contains(seen, name) || k < 0 {
			return 0, false, false
		}
		return doAtSeconds(variables[k].object("value"), variables, append(slices.Clone(seen), name))
	case kind == "math":
		a, aKnown, aStatic := doAtSeconds(body.object("lhs"), variables, seen)
		b, bKnown, bStatic := doAtSeconds(body.object("rhs"), variables, seen)
		if aKnown && bKnown {
			switch body.value("op") {
			case "OpAdd":
				return a + b, true, true
			case "OpSub":
				return a - b, true, true
			case "OpMul":
				return a * b, true, true
			case "OpDiv":
				return a / b, true, true
			}
		}
		return 0, false, aStatic && bStatic
	case kind == "max" || kind == "min":
		vals := body.list("vals")
		known, static := len(vals) > 0, len(vals) > 0
		all := make([]float64, len(vals))
		for i, val := range vals {
			var valKnown, valStatic bool
			all[i], valKnown, valStatic = doAtSeconds(val, variables, seen)
			known, static = known && valKnown, static && valStatic
		}
		switch {
		case !known:
			return 0, false, static
		case slices.ContainsFunc(all, math.IsNaN): // Math.max and Math.min give NaN when any value is NaN.
			return math.NaN(), true, true
		case kind == "max":
			return slices.Max(all), true, true
		}
		return slices.Min(all), true, true
	}
	return 0, false, false
}

// Seconds rounded to milliseconds, as a time: -1.5s.
func formatSeconds(seconds float64) string {
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(seconds, 'f', 3, 64), 64)
	return jsNumber(rounded, 64) + "s"
}

func lintPrepull(apl lintObject, variables []lintObject, report lintReport) {
	latest, latestRow := 0.0, -1
	for i, row := range apl.list("prepullActions") {
		if row.value("hide") == true {
			continue
		}
		path := []any{"prepullActions", i, "doAtValue"}
		if absent(row.object("doAtValue")) {
			if _, set := row.get("doAtValue"); !set {
				path = []any{"prepullActions", i}
			}
			report(path, "warning", "prepull-no-time", "No Do At time; the sim skips this prepull row.")
			continue
		}
		seconds, known, static := doAtSeconds(row.object("doAtValue"), variables, nil)
		switch {
		case !known:
			if !static {
				report(path, "warning", "prepull-not-time", `Do At is read once, before the fight, where a value of the fight is not there to read and can fail the sim: write a time such as "-1.5s".`)
			}
		case math.IsNaN(seconds):
			report(path, "warning", "prepull-not-time", "Do At is not a time; the sim reads it as 0s.")
		case seconds > 0:
			report(path, "warning", "prepull-positive-time", fmt.Sprintf("Do At %s is after the pull; the sim ignores this prepull row. Prepull times are 0s or earlier.", formatSeconds(seconds)))
		case latestRow >= 0 && seconds < latest:
			report(path, "info", "prepull-order", fmt.Sprintf("Runs at %s, before row %d above it at %s: prepull rows run in time order, not list order.", formatSeconds(seconds), latestRow+1, formatSeconds(latest)))
		default:
			latest, latestRow = seconds, i
		}
	}
}

func lintConditions(c *collected, report lintReport) {
	for _, site := range c.values {
		if !site.condition || site.scope.hidden || absent(site.value) {
			continue
		}
		truth, known := constantTruth(site.value)
		switch {
		case known && !truth && site.value.kind() == "const":
			report(site.path, "info", "never-runs", "Always false: this never runs. The sim still keeps its spell out of automatic cooldown use.")
		case known && !truth:
			report(site.path, "warning", "never-runs", "Always false: this never runs.")
		case known:
			report(site.path, "info", "always-true", "Always true: the same as no condition.")
		default:
			for _, part := range constantParts(site.value, site.path, nil) {
				what := "comparison"
				if part.kind != "cmp" {
					what = fmt.Sprintf("\"%s\"", part.kind)
				}
				report(part.path, "info", "always-true", fmt.Sprintf("This %s is always %t.", what, part.truth))
			}
		}
	}
}

// A row is shadowed when an earlier live row of the same list takes the same action, with no condition or the same
// one: the later row is only reached when that action cannot run, and then it cannot either. Prepull rows each run at
// their own time, and stateful actions keep a state per row, so neither is compared. Earlier rows are kept by their
// list, kind and action, with the first that always runs and the first for each condition, so each row is looked up
// once.
func lintShadowed(c *collected, report lintReport) {
	type earlier struct {
		always     int            // The first row that runs whenever it is reached, or -1.
		conditions map[string]int // The first row with each condition.
	}
	taken := map[string]*earlier{}
	for i, site := range c.actions {
		kind := site.action.kind()
		if site.nested || site.scope.hidden || site.scope.prepull || kind == "" || statefulKinds[kind] {
			continue
		}
		condition := site.action.object("condition")
		hasCondition := !absent(condition)
		truth, known := constantTruth(condition)
		if hasCondition && known && !truth {
			continue
		}
		key := fmt.Sprint(site.scope.group, " ", kind, " ", canonicalJSON(site.action.value(kind)))
		rows := taken[key]
		if rows == nil {
			rows = &earlier{always: -1, conditions: map[string]int{}}
			taken[key] = rows
		}
		var text string
		first := rows.always
		if hasCondition {
			text = canonicalJSON(condition)
			if same, has := rows.conditions[text]; has && (first < 0 || same < first) {
				first = same
			}
		}
		if first >= 0 {
			how := "the same condition"
			if absent(c.actions[first].action.object("condition")) {
				how = "no condition"
			}
			report(site.path[:len(site.path)-1], "warning", "shadowed-row", fmt.Sprintf("%s above takes this action with %s, so this row never runs.", rowLabel(c.actions[first].path), how))
		}
		if !hasCondition || known && truth {
			if rows.always < 0 {
				rows.always = i
			}
		} else if _, has := rows.conditions[text]; !has {
			rows.conditions[text] = i
		}
	}
}
