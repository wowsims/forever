package siml

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"math/big"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Typing (SPEC.md §6, §7.2, §8.3): the meaning of each syntax node, from the schema and the name table, as the proto
// message or field value it writes. Typing reads every node it can: an error in one argument does not stop its
// siblings, and a node whose kind, property or field is unknown is not read further.

// The group whose statements are being typed; the zero scope is outside every group.
type scope struct {
	group   string
	inGroup bool
}

// A variable declaration: a value variable, or a group's variable.
type variableDecl struct {
	span
	name  string
	scope scope
}

// `$x` or `${X}`, whose name the checks fill in once every declaration is known (SPEC.md §7.5).
type variableUse struct {
	node *varNode
	name func(string) // Writes the declared name it refers to.
	site *expansion   // The def call it is read in, if any.
}

// A group reference, with the placeholders its with list sets.
type groupUse struct {
	span
	name  string
	scope scope
	with  []*placeholderNode
}

type nameUse struct {
	span
	name string
	site *expansion // The def call it is read in, if any.
}

type typer struct {
	ix           *schemaIndex
	names        *Names
	class        proto.Class // The class the spec line sets, whose names the file reads (SPEC.md §7.1).
	spec         string      // The spec line's Player spec field name; "" without one.
	diags        []diag
	refs         []nameRef // The names the file writes, as they are typed.
	usedRefs     []nameRef // The names used files' defs write, each at the call in this file that reads it.
	rawRefs      []span    // The raw tokens that hold an id a name could write.
	kind         kindScope // The kind whose own arguments are being typed.
	variableUses []*variableUse
	groupUses    []groupUse
	sequenceDefs map[string]bool // The sequences the rows of the file and of the files it uses hold, as sequencesIn gives them.
	sequenceUses []nameUse
	placeholders map[scope]map[string]bool // The placeholders each group's statements hold.
	scope        scope
	at           []any      // The JSON path, from IndividualSimSettings, of what is being typed: JSON names and list indices.
	spans        []pathSpan // The nodes that read, each at the JSON path it writes, in the order they are read.
	lines        *lineMap
	defs         defTyping
	site         *expansion       // The def call being read, if any.
	reported     map[[2]diag]bool // Each diagnostic as written, and the span it moves to: a node read more than once, as a when condition is for each row, reports once.
}

// A node's text and the JSON path of what it writes.
type pathSpan struct {
	span
	path []any
}

// Types what follows at the JSON path keys lead to from the current one; gives what restores the current one.
func (t *typer) in(keys ...any) func() {
	n := len(t.at)
	t.at = append(t.at, keys...)
	return func() { t.at = t.at[:n] }
}

// Records that the text at writes the current JSON path.
func (t *typer) mark(at span) {
	if t.lines != nil && at.end > len(t.lines.src) {
		return // In a used file's def.
	}
	t.spans = append(t.spans, pathSpan{span: at, path: append([]any{}, t.at...)})
}

var (
	compareOperators = map[string]proto.APLValueCompare_ComparisonOperator{
		"==": proto.APLValueCompare_OpEq, "!=": proto.APLValueCompare_OpNe, ">=": proto.APLValueCompare_OpGe,
		">": proto.APLValueCompare_OpGt, "<=": proto.APLValueCompare_OpLe, "<": proto.APLValueCompare_OpLt,
	}
	mathOperators = map[string]proto.APLValueMath_MathOperator{
		"+": proto.APLValueMath_OpAdd, "-": proto.APLValueMath_OpSub, "*": proto.APLValueMath_OpMul, "/": proto.APLValueMath_OpDiv,
	}
	listItemMessage = (&proto.APLListItem{}).ProtoReflect().Descriptor()
	prepullMessage  = (&proto.APLPrepullAction{}).ProtoReflect().Descriptor()
)

func newTyper(ix *schemaIndex, names *Names) *typer {
	return &typer{
		ix:           ix,
		names:        names,
		placeholders: map[scope]map[string]bool{},
		reported:     map[[2]diag]bool{},
	}
}

func (t *typer) report(code, message string, at span) {
	t.add(diag{span: at, severity: Error, code: code, message: message})
}

func (t *typer) warn(code, message string, at span) {
	t.add(diag{span: at, severity: Warning, code: code, message: message})
}

// Adds d, at the def call it is in when it lies in a def's value (SPEC.md §6.10), unless it is reported there already.
// The calls it came through are no part of that, so a def that a call reaches through other defs by several paths
// reports it once, with the first path.
func (t *typer) add(d diag) {
	moved := t.site.relocate(d, t.lines)
	if key := [2]diag{d, {span: moved.span}}; !t.reported[key] {
		t.reported[key] = true
		t.diags = append(t.diags, moved)
	}
}

// "; did you mean a, b?" for the candidates within edit distance 2 of word, closest first, at most three.
func didYouMean(word string, candidates []string) string {
	type scored struct {
		name     string
		distance int
	}
	var close []scored
	seen := map[string]bool{}
	for _, c := range candidates {
		if seen[c] {
			continue
		}
		seen[c] = true
		if d := editDistance(word, c); d <= 2 {
			close = append(close, scored{c, d})
		}
	}
	slices.SortFunc(close, func(a, b scored) int {
		if a.distance != b.distance {
			return a.distance - b.distance
		}
		return strings.Compare(a.name, b.name)
	})
	if len(close) == 0 {
		return ""
	}
	names := make([]string, 0, 3)
	for _, c := range close[:min(3, len(close))] {
		names = append(names, c.name)
	}
	return "; did you mean " + strings.Join(names, ", ") + "?"
}

// The edit distance of a and b when it is at most 2, else 3. A path of cost 2 or less stays within 2 of the diagonal,
// so only those cells are worked out, and a row with none at 2 or less ends the work.
func editDistance(a, b string) int {
	const far = 3
	if len(a)-len(b) > 2 || len(b)-len(a) > 2 {
		return far
	}
	prev, row := make([]int, len(b)+1), make([]int, len(b)+1)
	for j := range prev {
		prev[j] = min(j, far)
	}
	for i := 1; i <= len(a); i++ {
		row[0] = min(i, far)
		best := row[0]
		for j := 1; j <= len(b); j++ {
			if j < i-2 || j > i+2 {
				row[j] = far
				continue
			}
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			row[j] = min(prev[j]+1, row[j-1]+1, prev[j-1]+cost, far)
			best = min(best, row[j])
		}
		if best == far {
			return far
		}
		prev, row = row, prev
	}
	return prev[len(b)]
}

func article(word string) string {
	if word != "" && strings.ContainsRune("aeiou", rune(word[0])) {
		return "an " + word
	}
	return "a " + word
}

// A message's last name segment: APLValue for proto.APLValue.
func typeName[N ~string](full N) string {
	s := string(full)
	return s[strings.LastIndexByte(s, '.')+1:]
}

// --- Values (SPEC.md §6.5)

func constValue(text string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: text}}}
}

func (t *typer) value(n node) *proto.APLValue {
	if p, ok := n.(*parenNode); ok {
		return t.value(p.expr)
	}
	v := t.valueOf(n)
	if v != nil {
		t.mark(n.pos())
	}
	return v
}

// The value n writes at the JSON path keys lead to.
func (t *typer) valueIn(n node, keys ...any) *proto.APLValue {
	defer t.in(keys...)()
	return t.value(n)
}

// Records n, whose declared name the checks write through name.
func (t *typer) useVariable(n *varNode, name func(string)) {
	use := &variableUse{node: n, name: name, site: t.site}
	t.variableUses = append(t.variableUses, use)
}

func (t *typer) valueOf(n node) *proto.APLValue {
	switch n := n.(type) {
	case *literalNode:
		return constValue(n.text)
	case *stringNode:
		return constValue(n.value)
	case *boolNode:
		return constValue(strconv.FormatBool(n.value))
	case *noneNode:
		return &proto.APLValue{}
	case *varNode:
		ref := &proto.APLValueVariableRef{}
		t.useVariable(n, func(name string) { ref.Name = name })
		return &proto.APLValue{Value: &proto.APLValue_VariableRef{VariableRef: ref}}
	case *placeholderNode:
		names := t.placeholders[t.scope]
		if names == nil {
			names = map[string]bool{}
			t.placeholders[t.scope] = names
		}
		names[n.name] = true
		return &proto.APLValue{Value: &proto.APLValue_VariablePlaceholder{VariablePlaceholder: &proto.APLValueVariablePlaceholder{Name: n.name}}}
	case *rawNode:
		if m := t.raw(valueMessage, n); m != nil {
			return m.Interface().(*proto.APLValue)
		}
		return nil
	case *notNode:
		if val := t.valueIn(n.operand, "not", "val"); val != nil {
			return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: val}}}
		}
		return nil
	case *logicNode:
		vals, ok := t.values(n.op, n.operands)
		switch {
		case !ok:
			return nil
		case n.op == "and":
			return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: vals}}}
		}
		return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: vals}}}
	case *compareNode:
		lhs, rhs := t.valueIn(n.left, "cmp", "lhs"), t.valueIn(n.right, "cmp", "rhs")
		if lhs == nil || rhs == nil {
			return nil
		}
		return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: compareOperators[n.op], Lhs: lhs, Rhs: rhs}}}
	case *mathNode:
		lhs, rhs := t.valueIn(n.left, "math", "lhs"), t.valueIn(n.right, "math", "rhs")
		if lhs == nil || rhs == nil {
			return nil
		}
		return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: mathOperators[n.op], Lhs: lhs, Rhs: rhs}}}
	case *callNode:
		return t.valueCall(n)
	case *memberNode:
		return t.unitPrefixed(n)
	case *accessorNode:
		return t.accessor(n)
	case *listNode:
		t.report("wrong-type", "a value goes here, not a list", n.span)
		return nil
	}
	t.report("wrong-type", "a value goes here, not an action", n.pos())
	return nil
}

// Each of nodes as a value of the vals of op, and or or; false when any fails, after typing them all.
func (t *typer) values(op string, nodes []node) ([]*proto.APLValue, bool) {
	out := make([]*proto.APLValue, len(nodes))
	ok := true
	for i, n := range nodes {
		out[i] = t.valueIn(n, op, "vals", i)
		ok = ok && out[i] != nil
	}
	return out, ok
}

// The APLValue holding body as the value kind k.
func kindValue(k *aplKind, body protoreflect.Message) *proto.APLValue {
	v := &proto.APLValue{}
	v.ProtoReflect().Set(k.field, protoreflect.ValueOfMessage(body))
	return v
}

// The APLAction holding body as the action kind k.
func kindAction(k *aplKind, body protoreflect.Message) *proto.APLAction {
	a := &proto.APLAction{}
	a.ProtoReflect().Set(k.field, protoreflect.ValueOfMessage(body))
	return a
}

// A kind's message from its arguments; a kind written with no arguments has none of its fields set.
func (t *typer) kindBody(k *aplKind, args *argList) protoreflect.Message {
	body := newMessage(k.message.desc)
	if args == nil {
		return body
	}
	defer t.in(k.field.JSONName())()
	defer t.enter(k, nil, args)()
	if !t.kindArgs(k, args, body) {
		return nil
	}
	return body
}

// `kind` or `kind(args)`, or a call of a def, which reads as the def's value.
func (t *typer) valueCall(c *callNode) *proto.APLValue {
	if d := t.defs.lookup(c.name.text); d != nil {
		return t.expand(c, d)
	}
	k := t.ix.valueKinds[c.name.text]
	if k == nil {
		t.report("unknown-kind", t.unknownKind(c.name.text, true), c.name.span)
		return nil
	}
	if body := t.kindBody(k, c.args); body != nil {
		return kindValue(k, body)
	}
	return nil
}

func (t *typer) unknownKind(name string, value bool) string {
	noun, kinds, others := "action", t.ix.actionKinds, t.ix.valueKinds
	if value {
		noun, kinds, others = "value", t.ix.valueKinds, t.ix.actionKinds
	}
	if !value && t.defs.lookup(name) != nil {
		return fmt.Sprintf("%s is a def, which reads as a value, and an action goes here", name)
	}
	if others[name] != nil {
		other := "a value"
		if value {
			other = "an action"
		}
		return fmt.Sprintf("%s is %s, and %s goes here", name, other, article(noun))
	}
	if held := t.names.Holding(t.class, name); len(held) > 0 {
		return fmt.Sprintf("%s is %s, and %s goes here", name, namespaceThings[held[0]], article(noun))
	}
	names := make([]string, 0, len(kinds))
	for n := range kinds {
		names = append(names, n)
	}
	return fmt.Sprintf("no %s is named %s%s", noun, name, didYouMean(name, names))
}

// `u.kind`: a kind whose only field is a unit (SPEC.md §6.8).
func (t *typer) unitPrefixed(m *memberNode) *proto.APLValue {
	k := t.ix.valueKinds[m.name.text]
	if k == nil || k.UnitField == "" {
		var prefixed []string
		for _, vk := range t.ix.valueKindList {
			if vk.UnitField != "" {
				prefixed = append(prefixed, vk.Kind)
			}
		}
		message := "no value is named " + m.name.text + didYouMean(m.name.text, prefixed)
		if k != nil {
			message = fmt.Sprintf("%s takes no unit; the values a unit goes in front of are %s", k.Kind, strings.Join(prefixed, ", "))
		}
		t.report("unknown-kind", message, m.name.span)
		return nil
	}
	unitField := k.message.byProto[k.UnitField]
	pop := t.in(k.field.JSONName(), unitField.JSON)
	unit := t.unitName(m.object)
	pop()
	if unit == nil {
		return nil
	}
	body := newMessage(k.message.desc)
	body.Set(unitField.desc, protoreflect.ValueOfMessage(unit.ProtoReflect()))
	return kindValue(k, body)
}

// `[u.]head(subject).prop[(args)]` (SPEC.md §6.7).
func (t *typer) accessor(a *accessorNode) *proto.APLValue {
	head := a.head.text
	if !slices.Contains(t.ix.heads, head) {
		message := didYouMean(head, t.ix.heads)
		if message == "" {
			message = "; the heads are " + strings.Join(t.ix.heads, ", ")
		}
		t.report("unknown-kind", head+" is no accessor head"+message, a.head.span)
		return nil
	}
	k := t.ix.accessors[head+"."+a.prop.text]
	if k == nil {
		var props []string
		for _, vk := range t.ix.valueKindList {
			if vk.Accessor != nil && vk.Accessor.Head == head {
				props = append(props, vk.Accessor.Prop)
			}
		}
		t.report("unknown-property", fmt.Sprintf("%s(…) has no property %s%s", head, a.prop.text, didYouMean(a.prop.text, props)), a.prop.span)
		return nil
	}
	acc := k.Accessor
	body := newMessage(k.message.desc)
	given := map[*field]bool{}
	ok := true
	defer t.in(k.field.JSONName())()
	defer t.enter(k, a.unit, a.args)()
	if a.unit != nil {
		unitField := k.message.byProto[acc.Unit]
		if unitField == nil {
			t.report("wrong-type", fmt.Sprintf("%s(…).%s takes no unit", head, a.prop.text), a.unit.span)
			ok = false
		} else {
			given[unitField] = true
			pop := t.in(unitField.JSON)
			unit := t.unitName(*a.unit)
			pop()
			if unit != nil {
				body.Set(unitField.desc, protoreflect.ValueOfMessage(unit.ProtoReflect()))
			} else {
				ok = false
			}
		}
	}
	subject := k.message.byProto[acc.Subject]
	given[subject] = true
	pop := t.in(subject.JSON)
	v, good := t.single(subject, a.subject)
	pop()
	if good {
		body.Set(subject.desc, v)
	} else {
		ok = false
	}
	positional := fmt.Sprintf("arguments on a property are named: %s(field=value)", a.prop.text)
	if a.args != nil && !t.namedArgs(k, a.args, body, given, positional) {
		ok = false
	}
	if !ok {
		return nil
	}
	return kindValue(k, body)
}

// The arguments of a generic or plain-argument kind (SPEC.md §8.3 rules 5 and 6): the positional field, or
// the list's items, unnamed; every other field named; a bool field also by its name alone.
func (t *typer) kindArgs(k *aplKind, args *argList, body protoreflect.Message) bool {
	fields := k.message.fields
	given := map[*field]bool{}
	var list, positional *field
	if k.ListField != "" {
		list = k.message.byProto[k.ListField]
	} else if k.Positional != "" {
		positional = k.message.byProto[k.Positional]
	}
	var items []protoreflect.Value
	positionalWritten := false
	ok := true
	for _, a := range args.items {
		if a.name != nil {
			ok = t.namedArg(k, *a.name, a.value, body, given) && ok
			continue
		}
		if f := flagField(a.value, fields); f != nil {
			ok = t.setFlag(f, a.value.pos(), body, given) && ok
			continue
		}
		target := list
		if target == nil {
			target = positional
		}
		if target == nil || positionalWritten || given[target] && len(items) == 0 {
			ok = false
			switch {
			case target != nil && given[target] && !positionalWritten:
				t.report("duplicate-field", target.Siml+" is written twice", a.span)
			case target != nil:
				t.report("positional-not-allowed", fmt.Sprintf("%s takes one positional argument, %s; name the others", k.Kind, target.Siml), a.span)
			case len(fields) > 0:
				t.report("positional-not-allowed", fmt.Sprintf("%s takes no positional argument; name its fields: %s", k.Kind, strings.Join(k.message.fieldNames(), ", ")), a.span)
			default:
				t.report("positional-not-allowed", k.Kind+" takes no arguments", a.span)
			}
			continue
		}
		given[target] = true
		pop := t.in(target.JSON)
		if list != nil {
			t.at = append(t.at, len(items))
		}
		v, good := t.single(target, a.value)
		pop()
		ok = ok && good
		if list != nil {
			items = append(items, v)
			continue
		}
		positionalWritten = true
		if good {
			body.Set(target.desc, v)
		}
	}
	if !ok {
		return false
	}
	if len(items) > 0 {
		l := body.Mutable(list.desc).List()
		for _, v := range items {
			l.Append(v)
		}
	}
	return true
}

// Named arguments, and bare bool flags; any other unnamed argument is positional-not-allowed.
func (t *typer) namedArgs(k *aplKind, args *argList, body protoreflect.Message, given map[*field]bool, positionalMessage string) bool {
	ok := true
	for _, a := range args.items {
		switch f := flagField(a.value, k.message.fields); {
		case a.name != nil:
			ok = t.namedArg(k, *a.name, a.value, body, given) && ok
		case f != nil:
			ok = t.setFlag(f, a.value.pos(), body, given) && ok
		default:
			t.report("positional-not-allowed", positionalMessage, a.span)
			ok = false
		}
	}
	return ok
}

func (t *typer) namedArg(k *aplKind, n name, value node, body protoreflect.Message, given map[*field]bool) bool {
	f := k.message.byName[n.text]
	if f == nil {
		names := k.message.fieldNames()
		message := didYouMean(n.text, names)
		if message == "" {
			message = "; its fields are " + strings.Join(names, ", ")
		}
		t.report("unknown-field", fmt.Sprintf("%s has no field %s%s", k.Kind, n.text, message), n.span)
		return false
	}
	if given[f] {
		t.report("duplicate-field", f.Siml+" is written twice", n.span)
		return false
	}
	given[f] = true
	pop := t.in(f.JSON)
	v, ok := t.term(f, value)
	pop()
	if ok {
		body.Set(f.desc, v)
	}
	return ok
}

// The bool field a bare word names, or nil.
func flagField(n node, fields []*field) *field {
	c, ok := n.(*callNode)
	if !ok || c.args != nil {
		return nil
	}
	for _, f := range fields {
		if f.Printed == "bool" && !f.Repeated && (f.Siml == c.name.text || slices.Contains(f.Aliases, c.name.text)) {
			return f
		}
	}
	return nil
}

func (t *typer) setFlag(f *field, at span, body protoreflect.Message, given map[*field]bool) bool {
	if given[f] {
		t.report("duplicate-field", f.Siml+" is written twice", at)
		return false
	}
	given[f] = true
	body.Set(f.desc, protoreflect.ValueOfBool(true))
	defer t.in(f.JSON)()
	t.mark(at)
	return true
}

// --- Terms: the value of a field, by its printed type (SPEC.md §5.2, §8.3)

// m as a field value; false when typing it failed, which leaves m a nil pointer.
func messageValue(m googleProto.Message) (protoreflect.Value, bool) {
	r := m.ProtoReflect()
	if !r.IsValid() {
		return protoreflect.Value{}, false
	}
	return protoreflect.ValueOfMessage(r), true
}

// Whether f holds a list whose items are written one by one; an indexed array is written whole, as one value.
func (f *field) isList() bool { return f.Repeated && f.Printed != "indexed" }

// A new list for the repeated field f.
func newList(f *field) protoreflect.List {
	return newMessage(f.desc.ContainingMessage()).NewField(f.desc).List()
}

func (t *typer) term(f *field, n node) (protoreflect.Value, bool) {
	if !f.isList() {
		return t.single(f, n)
	}
	if r, ok := n.(*rawNode); ok {
		return t.rawField(f, r, true)
	}
	l, ok := n.(*listNode)
	if !ok {
		return t.wrongType(f, "a list, such as [a, b]", n)
	}
	list := newList(f)
	ok = true
	for i, item := range l.items {
		pop := t.in(i)
		v, good := t.single(f, item)
		pop()
		if good && ok {
			list.Append(v)
		}
		ok = ok && good
	}
	if ok {
		t.mark(l.span)
	}
	return protoreflect.ValueOfList(list), ok
}

// One value of f.
func (t *typer) single(f *field, n node) (protoreflect.Value, bool) {
	if r, ok := n.(*rawNode); ok && f.Printed != "value" && f.Printed != "action" {
		return t.rawField(f, r, f.Printed == "indexed")
	}
	v, ok := t.singleOf(f, n)
	// A message marks its own text.
	if ok && f.desc.Message() == nil {
		t.mark(n.pos())
	}
	return v, ok
}

func (t *typer) singleOf(f *field, n node) (protoreflect.Value, bool) {
	switch f.Printed {
	case "value":
		return messageValue(t.value(n))
	case "action":
		return messageValue(t.actionTerm(n))
	case "action_id":
		return messageValue(t.actionID(n, t.kind.namespace(f), t.kind.casts(f)))
	case "unit":
		return messageValue(t.unit(n))
	case "bool":
		if b, ok := n.(*boolNode); ok {
			return protoreflect.ValueOfBool(b.value), true
		}
		return t.wrongType(f, "true or false", n)
	case "string":
		if s, ok := n.(*stringNode); ok {
			if t.ix.sequenceFields[f.Name] {
				t.sequenceUses = append(t.sequenceUses, nameUse{s.span, s.value, t.site})
			}
			return protoreflect.ValueOfString(s.value), true
		}
		return t.wrongType(f, "a string in quotes", n)
	case "number":
		return t.number(f, n)
	case "enum", "flag_enum":
		return t.enumValue(f.Type, n, f, false)
	case "stat":
		if l, ok := n.(*literalNode); ok && isIntegerText(l.text) {
			return t.number(f, n)
		}
		return t.enumValue(statEnum, n, f, true)
	case "ref":
		return t.ref(f, n)
	case "message":
		return t.message(f, n)
	case "indexed":
		return t.indexed(f, n)
	}
	return t.wrongType(f, "raw JSON", n)
}

func (t *typer) wrongType(f *field, what string, n node) (protoreflect.Value, bool) {
	t.report("wrong-type", fmt.Sprintf("%s takes %s", f.Siml, what), n.pos())
	return protoreflect.Value{}, false
}

func isIntegerText(s string) bool { return isInt(strings.TrimPrefix(s, "-")) }

func isDecimal(s string) bool {
	whole, frac, found := strings.Cut(strings.TrimPrefix(s, "-"), ".")
	return isInt(whole) && (!found || isInt(frac))
}

// A number with no unit, as the value of the number field f.
func (t *typer) number(f *field, n node) (protoreflect.Value, bool) {
	l, ok := n.(*literalNode)
	if !ok {
		return t.wrongType(f, "a number", n)
	}
	if !isDecimal(l.text) {
		t.report("bad-literal", f.Siml+" takes a number with no unit", l.span)
		return protoreflect.Value{}, false
	}
	switch f.desc.Kind() {
	case protoreflect.DoubleKind:
		if v, ok := parseDouble(l.text); ok {
			return protoreflect.ValueOfFloat64(v), true
		}
		return t.wrongType(f, "a number that fits double", n)
	case protoreflect.FloatKind:
		if v, ok := parseFloat(l.text); ok {
			return protoreflect.ValueOfFloat32(v), true
		}
		return t.wrongType(f, "a number that fits float", n)
	}
	if !isIntegerText(l.text) {
		return t.wrongType(f, "a whole number", n)
	}
	v, _ := new(big.Int).SetString(l.text, 10)
	lo, hi := intRange(f.desc.Kind())
	if v.Cmp(lo) < 0 || v.Cmp(hi) > 0 {
		return t.wrongType(f, "a number that fits "+f.Scalar, n)
	}
	if v.Sign() > 0 && !v.IsInt64() {
		return protoreflect.ValueOfUint64(v.Uint64()), true
	}
	return intValue(f.desc, v.Int64()), true
}

// A decimal as the double nearest it; false past the range of a double. -0 reads as 0, as a JSON number has no
// negative zero.
func parseDouble(text string) (float64, bool) {
	v, err := strconv.ParseFloat(text, 64)
	return v + 0, err == nil
}

// A decimal as a float field holds it (SPEC.md §5.2): the double nearest it, rounded to a float32, as the site holds
// it; false past the range of a float32.
func parseFloat(text string) (float32, bool) {
	v, ok := parseDouble(text)
	return float32(v), ok && !math.IsInf(float64(float32(v)), 0)
}

var (
	int32Range  = [2]*big.Int{big.NewInt(-1 << 31), big.NewInt(1<<31 - 1)}
	uint32Range = [2]*big.Int{big.NewInt(0), big.NewInt(1<<32 - 1)}
	int64Range  = [2]*big.Int{big.NewInt(-1 << 63), big.NewInt(1<<63 - 1)}
	uint64Range = [2]*big.Int{big.NewInt(0), new(big.Int).SetUint64(1<<64 - 1)}
)

func intRange(k protoreflect.Kind) (*big.Int, *big.Int) {
	r := int32Range
	switch scalarName(k) {
	case "uint32":
		r = uint32Range
	case "int64":
		r = int64Range
	case "uint64":
		r = uint64Range
	}
	return r[0], r[1]
}

// n as a value of the integer, float or enum field fd.
func intValue(fd protoreflect.FieldDescriptor, n int64) protoreflect.Value {
	switch scalarName(fd.Kind()) {
	case "uint32":
		return protoreflect.ValueOfUint32(uint32(n))
	case "int64":
		return protoreflect.ValueOfInt64(n)
	case "uint64":
		return protoreflect.ValueOfUint64(uint64(n))
	case "double":
		return protoreflect.ValueOfFloat64(float64(n))
	case "float":
		return protoreflect.ValueOfFloat32(float32(n))
	case "enum":
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(n))
	}
	return protoreflect.ValueOfInt32(int32(n))
}

// An enum value by its short or proto name; a stat field holds the number.
func (t *typer) enumValue(enumName string, n node, f *field, stat bool) (protoreflect.Value, bool) {
	c, ok := n.(*callNode)
	if !ok || c.args != nil {
		if stat {
			return t.wrongType(f, "a stat, such as attack_power", n)
		}
		return t.wrongType(f, "one of its values", n)
	}
	e := t.ix.enums[enumName]
	v := e.byName[c.name.text]
	if v == nil {
		t.report("unknown-enum", fmt.Sprintf("%s has no value %s%s", f.Siml, c.name.text, didYouMean(c.name.text, e.shortNames())), c.name.span)
		return protoreflect.Value{}, false
	}
	return intValue(f.desc, int64(v.Number)), true
}

// An item, enchant or suffix ref field: a name of that namespace, written as the id.
func (t *typer) ref(f *field, n node) (protoreflect.Value, bool) {
	if value, done, ok := t.named(n); ok {
		defer done()
		if value == nil {
			return protoreflect.Value{}, false
		}
		return t.ref(f, value)
	}
	ns := Namespace(f.Ref)
	c, ok := n.(*callNode)
	if !ok || c.args != nil {
		return t.wrongType(f, "the name of "+namespaceThings[ns]+", written alone", n)
	}
	r, ok := t.lookup(c.name, ns, f.Siml, false)
	if !ok {
		return protoreflect.Value{}, false
	}
	t.addRef(r)
	return intValue(f.desc, int64(r.id)), true
}

// `field(args)` for a message field: named arguments, bare bool flags and flag-value enums, and nested
// `field(args)` for message fields; or none for the empty message.
func (t *typer) message(f *field, n node) (protoreflect.Value, bool) {
	if _, ok := n.(*noneNode); ok {
		return protoreflect.ValueOfMessage(newMessage(f.desc.Message())), true
	}
	c, ok := n.(*callNode)
	if !ok || c.args == nil || c.name.text != f.Siml && !slices.Contains(f.Aliases, c.name.text) {
		return t.wrongType(f, fmt.Sprintf("its fields in %s(…), or raw JSON", f.Siml), n)
	}
	if body := t.messageArgs(t.ix.messages[f.Type], c.args); body != nil {
		t.mark(c.span)
		return protoreflect.ValueOfMessage(body), true
	}
	return protoreflect.Value{}, false
}

// The fields of a message as arguments: named, a bool or flag-value enum by its name alone, a message field as
// field(…), and each element of a list of messages as a field(…) of its own.
func (t *typer) messageArgs(m *message, args *argList) protoreflect.Message {
	body := newMessage(m.desc)
	given := map[*field]bool{}
	elements := map[*field]bool{} // Lists of messages written an element at a time.
	ok := true
	for _, a := range args.items {
		n := a.name
		call, isCall := a.value.(*callNode)
		if n == nil && isCall {
			n = &call.name
		}
		if n == nil {
			t.report("positional-not-allowed", "the fields of a message are named: field=value", a.span)
			ok = false
			continue
		}
		f := m.byName[n.text]
		if f == nil {
			t.report("unknown-field", fmt.Sprintf("%s has no field %s%s", typeName(m.desc.FullName()), n.text, didYouMean(n.text, m.fieldNames())), n.span)
			ok = false
			continue
		}
		element := a.name == nil && isCall && call.args != nil && f.isList() && f.Printed == "message"
		if given[f] || !element && elements[f] {
			t.report("duplicate-field", f.Siml+" is written twice", n.span)
			ok = false
			continue
		}
		var v protoreflect.Value
		good := false
		switch {
		case element:
			elements[f] = true
			pop := t.in(f.JSON, body.Get(f.desc).List().Len())
			v, good = t.single(f, a.value)
			pop()
			if good {
				body.Mutable(f.desc).List().Append(v)
			}
			ok = ok && good
			continue
		case a.name != nil || call.args != nil:
			pop := t.in(f.JSON)
			v, good = t.term(f, a.value)
			pop()
		default:
			v, good = t.bareFlag(f, a.value.pos())
		}
		given[f] = true
		if good {
			body.Set(f.desc, v)
		}
		ok = ok && good
	}
	if !ok {
		return nil
	}
	return body
}

// A field written by its name alone: a bool's true, or a flag-value enum's flag value.
func (t *typer) bareFlag(f *field, at span) (protoreflect.Value, bool) {
	if f.Printed == "bool" && !f.Repeated {
		return protoreflect.ValueOfBool(true), true
	}
	if f.Printed == "flag_enum" {
		if e := t.ix.enums[f.Type]; e.flag != "" {
			return protoreflect.ValueOfEnum(protoreflect.EnumNumber(e.byName[e.flag].Number)), true
		}
	}
	t.report("wrong-type", fmt.Sprintf("%s needs a value: %[1]s <value>", f.Siml), at)
	return protoreflect.Value{}, false
}

// `field(<short name>=<number>, …)` for an array indexed by an enum, written out to the enum's length.
func (t *typer) indexed(f *field, n node) (protoreflect.Value, bool) {
	c, ok := n.(*callNode)
	if !ok || c.args == nil || c.name.text != f.Siml && !slices.Contains(f.Aliases, c.name.text) {
		return t.wrongType(f, f.Siml+"(<name>=<number>, …)", n)
	}
	e := t.ix.enums[f.Index]
	out := make([]protoreflect.Value, t.ix.enumLength(f.Index))
	for i := range out {
		out[i] = intValue(f.desc, 0)
	}
	seen := map[int32]bool{}
	ok = true
	for _, a := range c.args.items {
		if a.name == nil {
			t.report("positional-not-allowed", fmt.Sprintf("%s names each entry: %[1]s(<name>=<number>)", f.Siml), a.span)
			ok = false
			continue
		}
		v := e.byName[a.name.text]
		if v == nil {
			t.report("unknown-enum", fmt.Sprintf("%s has no entry %s%s", f.Siml, a.name.text, didYouMean(a.name.text, e.shortNames())), a.name.span)
			ok = false
			continue
		}
		if seen[v.Number] {
			t.report("duplicate-field", v.Short+" is written twice", a.name.span)
			ok = false
			continue
		}
		seen[v.Number] = true
		number, good := t.number(f, a.value)
		if good {
			out[v.Number] = number
		}
		ok = ok && good
	}
	t.mark(c.span)
	if !ok {
		return protoreflect.Value{}, false
	}
	list := newList(f)
	for _, v := range out {
		list.Append(v)
	}
	return protoreflect.ValueOfList(list), true
}

// --- Refs and units (SPEC.md §6.6, §7.2)

// An ActionID: a name of ns, item.<name>, other.<short>, none or raw JSON. In a cast position (cast), a name is of a
// spell the file's spec lets a rotation cast.
func (t *typer) actionID(n node, ns Namespace, cast bool) *proto.ActionID {
	if value, done, ok := t.named(n); ok {
		defer done()
		if value == nil {
			return nil
		}
		return t.actionID(value, ns, cast)
	}
	switch n := n.(type) {
	case *noneNode:
		t.mark(n.span)
		return &proto.ActionID{}
	case *rawNode:
		t.mark(n.span)
		if m := t.raw(actionIDMessage, n); m != nil {
			return m.Interface().(*proto.ActionID)
		}
		return nil
	case *memberNode:
		switch n.object.text {
		case "other":
			e := t.ix.enums[otherEnum]
			v := e.byName[n.name.text]
			if v == nil {
				t.report("unknown-enum", fmt.Sprintf("other.%s names no other action%s", n.name.text, didYouMean(n.name.text, e.shortNames())), n.name.span)
				return nil
			}
			t.mark(n.span)
			return &proto.ActionID{RawId: &proto.ActionID_OtherId{OtherId: proto.OtherAction(v.Number)}}
		case "item":
			r, ok := t.lookup(n.name, NamespaceItem, "item.", false)
			if !ok {
				return nil
			}
			if ns == NamespaceAura && slices.Contains(t.names.ICDOnlyItems[n.name.text], t.spec) && !t.kind.readsICD() {
				t.report("icd-trigger", fmt.Sprintf("item.%s is an internal cooldown trigger, which only icd_is_ready, internal_cooldown and trigger_icd find", n.name.text), n.name.span)
			}
			t.addRef(r)
			t.mark(n.span)
			return &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: r.id}}
		}
	case *varNode:
		t.report("spell-variable", "a variable holds no spell: ActionID has no field to name one", n.span)
		return nil
	case *callNode:
		if n.args != nil {
			break
		}
		r, ok := t.lookup(n.name, ns, "this position", cast)
		if !ok {
			return nil
		}
		e, _ := t.names.spellEntry(t.class, ns, n.name.text)
		if slices.Contains(e.ICDOnlyBy, t.spec) && !t.kind.readsICD() {
			t.report("icd-trigger", fmt.Sprintf("%s is an internal cooldown trigger, which only icd_is_ready, internal_cooldown and trigger_icd find", n.name.text), n.name.span)
		}
		if cast && e.Uncastable {
			t.warn("not-castable", t.notCastable(n.name.text), n.name.span)
		}
		t.addRef(r)
		t.mark(n.span)
		return r.actionID
	}
	t.report("wrong-type", "a name goes here: a spell, item.<name> or other.<name>", n.pos())
	return nil
}

func (t *typer) unit(n node) *proto.UnitReference {
	switch n := n.(type) {
	case *noneNode:
		t.mark(n.span)
		return &proto.UnitReference{}
	case *rawNode:
		t.mark(n.span)
		if m := t.raw(unitMessage, n); m != nil {
			return m.Interface().(*proto.UnitReference)
		}
		return nil
	case *callNode:
		if n.args == nil {
			return t.unitName(n.name)
		}
	}
	t.report("wrong-type", "a unit goes here, such as target", n.pos())
	return nil
}

func (t *typer) unitName(n name) *proto.UnitReference {
	if n.text == "none" {
		t.mark(n.span)
		return &proto.UnitReference{}
	}
	u := t.ix.unitReference(n.text)
	if u == nil {
		words := t.ix.unitWordList()
		message := didYouMean(n.text, words)
		if message == "" {
			message = "; the units are " + strings.Join(words, ", ") + ", and targetN, playerN, petN"
		}
		t.report("wrong-type", "no unit is named "+n.text+message, n.span)
	} else {
		t.mark(n.span)
	}
	return u
}

// --- Actions (SPEC.md §6.4)

func (t *typer) action(n *actionNode) *proto.APLAction {
	out := t.step(n.step)
	var condition *proto.APLValue
	if n.condition != nil {
		condition = t.valueIn(n.condition, "condition")
	}
	if out == nil || n.condition != nil && condition == nil {
		return nil
	}
	if condition != nil {
		if out.Condition != nil {
			t.report("duplicate-field", "the raw action holds its condition already", n.condition.pos())
			return nil
		}
		out.Condition = condition
	}
	t.mark(n.span)
	return out
}

// An action's step: the APLAction without its condition.
func (t *typer) step(n node) *proto.APLAction {
	switch s := n.(type) {
	case *noneNode:
		return &proto.APLAction{}
	case *rawNode:
		if m := t.raw(actionMessage, s); m != nil {
			return m.Interface().(*proto.APLAction)
		}
		return nil
	case *castStep:
		defer t.in("castSpell")()
		spell, target, ok := t.spellAndTarget(s.spell, s.target)
		if !ok {
			return nil
		}
		return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: spell, Target: target}}}
	case *channelStep:
		defer t.in("channelSpell")()
		spell, target, ok := t.spellAndTarget(s.spell, s.target)
		body := &proto.APLActionChannelSpell{SpellId: spell, Target: target, AllowRecast: s.recast}
		if s.until != nil {
			body.InterruptIf = t.valueIn(s.until, "interruptIf")
			ok = ok && body.InterruptIf != nil
		}
		if !ok {
			return nil
		}
		return &proto.APLAction{Action: &proto.APLAction_ChannelSpell{ChannelSpell: body}}
	case *waitStep:
		keys := []any{"wait", "duration"}
		if s.until {
			keys = []any{"waitUntil", "condition"}
		}
		v := t.valueIn(s.value, keys...)
		switch {
		case v == nil:
			return nil
		case s.until:
			return &proto.APLAction{Action: &proto.APLAction_WaitUntil{WaitUntil: &proto.APLActionWaitUntil{Condition: v}}}
		}
		return &proto.APLAction{Action: &proto.APLAction_Wait{Wait: &proto.APLActionWait{Duration: v}}}
	case *moveStep:
		keys := []any{"move", "rangeFromTarget"}
		if s.duration {
			keys = []any{"moveDuration", "duration"}
		}
		v := t.valueIn(s.value, keys...)
		switch {
		case v == nil:
			return nil
		case s.duration:
			return &proto.APLAction{Action: &proto.APLAction_MoveDuration{MoveDuration: &proto.APLActionMoveDuration{Duration: v}}}
		}
		return &proto.APLAction{Action: &proto.APLAction_Move{Move: &proto.APLActionMove{RangeFromTarget: v, RecheckWhileMoving: s.recheck}}}
	case *groupStep:
		return t.groupStep(s)
	case *scheduleStep:
		pop := t.in("schedule", "innerAction")
		inner := t.step(s.step)
		if inner != nil {
			t.mark(s.step.pos())
		}
		pop()
		if inner == nil {
			return nil
		}
		return &proto.APLAction{Action: &proto.APLAction_Schedule{Schedule: &proto.APLActionSchedule{Schedule: s.schedule.value, InnerAction: inner}}}
	case *sequenceStep:
		actions := make([]*proto.APLAction, len(s.actions))
		ok := true
		for i, a := range s.actions {
			pop := t.in("sequence", "actions", i)
			actions[i] = t.action(a)
			pop()
			ok = ok && actions[i] != nil
		}
		if !ok {
			return nil
		}
		return &proto.APLAction{Action: &proto.APLAction_Sequence{Sequence: &proto.APLActionSequence{Name: s.name.value, Actions: actions}}}
	case *callNode:
		return t.actionCall(s)
	}
	panic(fmt.Sprintf("siml: a step of type %T", n))
}

// A cast's or channel's spell and its `on` unit, both typed whatever either gives.
func (t *typer) spellAndTarget(spell, target node) (*proto.ActionID, *proto.UnitReference, bool) {
	pop := t.in("spellId")
	id := t.actionID(spell, NamespaceSpell, true)
	pop()
	var unit *proto.UnitReference
	if target != nil {
		pop := t.in("target")
		unit = t.unit(target)
		pop()
	}
	return id, unit, id != nil && (target == nil || unit != nil)
}

func (t *typer) groupStep(s *groupStep) *proto.APLAction {
	use := groupUse{span: s.name.span, name: s.name.value, scope: t.scope}
	for _, b := range s.with {
		use.with = append(use.with, b.placeholder)
	}
	t.groupUses = append(t.groupUses, use)
	ref := &proto.APLActionGroupReference{GroupName: s.name.value}
	ok := true
	for i, b := range s.with {
		pop := t.in("groupReference", "variables", i)
		v := t.valueIn(b.value, "value")
		if v != nil {
			t.mark(b.span)
		}
		pop()
		ok = ok && v != nil
		ref.Variables = append(ref.Variables, &proto.APLValueVariable{Name: b.placeholder.name, Value: v})
	}
	if !ok {
		return nil
	}
	return &proto.APLAction{Action: &proto.APLAction_GroupReference{GroupReference: ref}}
}

func (t *typer) actionCall(c *callNode) *proto.APLAction {
	k := t.ix.actionKinds[c.name.text]
	if k == nil {
		t.report("unknown-kind", t.unknownKind(c.name.text, false), c.name.span)
		return nil
	}
	body := t.kindBody(k, c.args)
	if body == nil {
		return nil
	}
	if m, ok := body.Interface().(*proto.APLActionGroupReference); ok {
		t.groupUses = append(t.groupUses, groupUse{span: c.span, name: m.GroupName, scope: t.scope})
	}
	return kindAction(k, body)
}

// An action in a field: an action term, a bare or called action kind, raw JSON or none.
func (t *typer) actionTerm(n node) *proto.APLAction {
	switch n := n.(type) {
	case *actionNode:
		return t.action(n)
	case *callNode, *rawNode, *noneNode:
		out := t.step(n)
		if out != nil {
			t.mark(n.pos())
		}
		return out
	}
	t.report("wrong-type", "an action goes here", n.pos())
	return nil
}

// --- Raw JSON (SPEC.md §11): it must type-check against the node it stands for.

// The message raw JSON of type md writes; nil, and bad-raw, when it is none.
func (t *typer) raw(md protoreflect.MessageDescriptor, r *rawNode) protoreflect.Message {
	m := newMessage(md)
	if problem := t.ix.readJSON(r.json, m, protojson.UnmarshalOptions{}); problem != "" {
		t.report("bad-raw", fmt.Sprintf("the raw JSON is no %s: %s", typeName(md.FullName()), problem), r.span)
		return nil
	}
	t.noteRaw(r, m)
	return m
}

// Raw JSON as a value of f: for a repeated field, the whole list when whole, otherwise one item of it.
func (t *typer) rawField(f *field, r *rawNode, whole bool) (protoreflect.Value, bool) {
	text := r.json
	if f.Repeated && !whole {
		text = "[" + text + "]"
	}
	holder := newMessage(f.desc.ContainingMessage())
	problem := ""
	if strings.TrimSpace(r.json) == "null" {
		problem = "null is no value"
	} else {
		key, _ := json.Marshal(f.JSON)
		problem = t.ix.readJSON("{"+string(key)+":"+text+"}", holder, protojson.UnmarshalOptions{})
	}
	if problem != "" {
		t.report("bad-raw", fmt.Sprintf("the raw JSON is no value of %s: %s", f.Siml, problem), r.span)
		return protoreflect.Value{}, false
	}
	t.noteRaw(r, holder)
	t.mark(r.span)
	if f.Repeated && !whole {
		return holder.Get(f.desc).List().Get(0), true
	}
	if f.desc.IsList() || f.desc.IsMap() {
		// Get gives an empty list or map read-only, which another message cannot take.
		return holder.Mutable(f.desc), true
	}
	return holder.Get(f.desc), true
}

// Reads raw JSON into m as SPEC.md §11 reads it: as JSON.parse reads the text, then as protojson reads that value
// with the options o, with what protojson alone does not check: a field the schema skips (a uuid) is no field, and
// an enum number must be one of the enum's values. The problem, or "".
func (ix *schemaIndex) readJSON(text string, m protoreflect.Message, o protojson.UnmarshalOptions) string {
	parsed, err := parsedJSON(text, m.Descriptor())
	if err != nil {
		return err.Error()
	}
	if err := o.Unmarshal([]byte(parsed), m.Interface()); err != nil {
		return err.Error()
	}
	dropNegativeZeros(m)
	return ix.unread(m)
}

// Raw JSON of message md as JSON.parse reads it, written back as JSON.stringify writes it (SPEC.md §11): a key
// written twice keeps its last value, and a number is the double nearest it. A string that escapes a lone surrogate,
// a number past the range of a double, and a string in place of a number that is not exactly a JSON number, are
// errors.
func parsedJSON(text string, md protoreflect.MessageDescriptor) (string, error) {
	if !json.Valid([]byte(text)) {
		return "", errors.New("the text is no JSON value")
	}
	for i := 0; i < len(text); i++ {
		if text[i] != '"' {
			continue
		}
		open := i
		for i++; text[i] != '"'; i++ { // Every string of a JSON text closes.
			if text[i] == '\\' {
				i++
			}
		}
		if err := badEscape(text, open, i); err != nil {
			return "", errors.New(err.message)
		}
	}
	dec := json.NewDecoder(strings.NewReader(text))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	var b strings.Builder
	var err error
	if object, ok := v.(map[string]any); ok {
		err = writeParsedMessage(&b, object, md)
	} else {
		err = writeParsed(&b, v, nil)
	}
	return b.String(), err
}

var (
	jsonNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)
	floatWords = []string{"NaN", "Infinity", "-Infinity"} // The strings protojson writes for a float that is no number.
)

// object, the JSON of a message md, with each field as protojson finds it: by JSON name, else by name.
func writeParsedMessage(b *strings.Builder, object map[string]any, md protoreflect.MessageDescriptor) error {
	b.WriteByte('{')
	for i, k := range slices.Sorted(maps.Keys(object)) {
		if i > 0 {
			b.WriteByte(',')
		}
		writeString(b, k)
		b.WriteByte(':')
		fd := md.Fields().ByJSONName(k)
		if fd == nil {
			fd = md.Fields().ByTextName(k)
		}
		if err := writeParsed(b, object[k], fd); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

// v as JSON.stringify writes it, where v is a value of fd, an element of it when it is a list or the values of a map
// when fd is a map's value; fd is nil where no field types v.
func writeParsed(b *strings.Builder, v any, fd protoreflect.FieldDescriptor) error {
	switch v := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(v))
	case string:
		if fd != nil && isNumber(fd.Kind()) && (isInteger(fd.Kind()) || !slices.Contains(floatWords, v)) {
			if !jsonNumber.MatchString(v) {
				return fmt.Errorf("%s is no number of %s", jsonString(v), fd.Name())
			}
			if fd.Kind() == protoreflect.FloatKind {
				return writeParsed(b, json.Number(v), fd)
			}
		}
		writeString(b, v)
	case json.Number:
		if fd != nil && fd.Kind() == protoreflect.FloatKind {
			f, ok := parseFloat(string(v))
			if !ok {
				return fmt.Errorf("%s is past the range of a float", v)
			}
			b.WriteString(jsNumber(float64(f), 32))
			break
		}
		f, ok := parseDouble(string(v))
		switch {
		case ok && fd != nil && isInteger(fd.Kind()) && f == math.Trunc(f):
			// The double's own integer, as BigInt gives it: 123456789012345664, not the 123456789012345660 that
			// JSON.stringify writes for that double.
			b.WriteString(new(big.Float).SetFloat64(f).Text('f', 0))
		case ok:
			b.WriteString(jsNumber(f, 64))
		case fd == nil:
			b.WriteString("null") // Past the range of a double where no field reads it, as under a key that names no field.
		default:
			return fmt.Errorf("%s is past the range of a double", v)
		}
	case []any:
		b.WriteByte('[')
		for i, x := range v {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := writeParsed(b, x, fd); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		if fd != nil && !fd.IsMap() && fd.Message() != nil {
			return writeParsedMessage(b, v, fd.Message())
		}
		var values protoreflect.FieldDescriptor // A map's values; nil for an object no field types.
		if fd != nil && fd.IsMap() {
			values = fd.MapValue()
		}
		b.WriteByte('{')
		for i, k := range slices.Sorted(maps.Keys(v)) {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(b, k)
			b.WriteByte(':')
			if err := writeParsed(b, v[k], values); err != nil {
				return err
			}
		}
		b.WriteByte('}')
	}
	return nil
}

// Clears each -0, which protojson keeps from a string such as "-0", in a field with no presence: a JSON number has
// no negative zero, so -0 is the default value, which the JSON leaves out.
func dropNegativeZeros(m protoreflect.Message) {
	walkMessages(m, func(x protoreflect.Message) {
		x.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.Message() == nil && !fd.IsList() && !fd.IsMap() && isDefault(fd, v) {
				x.Clear(fd)
			}
			return true
		})
	})
}

// A field of m the schema skips (SPEC.md §5.1), or "". An enum number that names no value reads, as protojson reads
// it.
func (ix *schemaIndex) unread(m protoreflect.Message) string {
	problem := ""
	walkMessages(m, func(x protoreflect.Message) {
		msg := ix.messages[string(x.Descriptor().FullName())]
		x.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
			if problem == "" && (msg == nil || msg.byProto[string(fd.Name())] == nil) {
				problem = fmt.Sprintf("%s has no field %s", typeName(x.Descriptor().FullName()), fd.JSONName())
			}
			return problem == ""
		})
	})
	return problem
}

// --- Entries (SPEC.md §5.2): a section's lines that set the fields of its message.

func isEntry(s stmt) bool {
	switch s.(type) {
	case *fieldEntryStmt, *flagListStmt, *messageEntryStmt, *rawEntryStmt:
		return true
	}
	return false
}

// A value an entry line or slot row writes: one value of a field of its section, at index in the field's list, or
// -1 for a field that holds one value.
type entryWrite struct {
	placed *placed
	index  int
	value  protoreflect.Value
}

// Types an entry line of a section whose fields are fields, writing each into the message holder gives for it, and
// gives the values the line writes. set holds the fields the section's lines have written.
func (t *typer) entry(s stmt, fields []*placed, holder func(*placed) protoreflect.Message, set map[*field]bool) []entryWrite {
	var written []entryWrite
	find := func(name string, at span, raw bool) *placed {
		var p *placed
		for _, x := range fields {
			if raw && (x.JSON == name || x.Name == name) || !raw && (x.Siml == name || slices.Contains(x.Aliases, name)) {
				p = x
				break
			}
		}
		switch {
		case p == nil && raw:
			t.report("bad-raw", "no field here is named "+name, at)
		case p == nil:
			names := make([]string, len(fields))
			for i, x := range fields {
				names[i] = x.Siml
			}
			t.report("unknown-field", "no field here is named "+name+didYouMean(name, names), at)
		case set[p.field] && !(p.isList() && p.Printed == "message"):
			// A list of messages takes one line per message, any other list one line of all its values.
			t.report("duplicate-field", p.Siml+" is written twice", at)
			return nil
		default:
			set[p.field] = true
		}
		return p
	}
	// Types what follows at the JSON path of p's next value.
	at := func(p *placed) func() {
		keys := p.keys()
		if p.isList() {
			keys = append(keys, holder(p).Get(p.desc).List().Len())
		}
		return t.in(keys...)
	}
	put := func(p *placed, v protoreflect.Value, ok bool) {
		if !ok {
			return
		}
		m := holder(p)
		index := -1
		if p.isList() {
			l := m.Mutable(p.desc).List()
			index = l.Len()
			l.Append(v)
		} else {
			m.Set(p.desc, v)
		}
		written = append(written, entryWrite{placed: p, index: index, value: v})
	}
	switch s := s.(type) {
	case *fieldEntryStmt:
		p := find(s.name.text, s.name.span, false)
		switch {
		case p == nil:
		case !p.isList() && len(s.values) > 1:
			t.report("wrong-type", p.Siml+" takes one value", s.values[1].pos())
		default:
			for _, value := range s.values {
				pop := at(p)
				v, ok := t.single(p.field, value)
				pop()
				put(p, v, ok)
			}
		}
	case *flagListStmt:
		for _, fl := range s.flags {
			p := find(fl.name.text, fl.name.span, false)
			switch {
			case p == nil:
			case fl.value != nil:
				pop := at(p)
				v, ok := t.single(p.field, fl.value)
				pop()
				put(p, v, ok)
			default:
				v, ok := t.bareFlag(p.field, fl.span)
				put(p, v, ok)
			}
		}
	case *messageEntryStmt:
		if p := find(s.name.text, s.name.span, false); p != nil {
			pop := at(p)
			v, ok := t.single(p.field, &callNode{span: s.span, name: s.name, args: s.args})
			pop()
			put(p, v, ok)
		}
	case *rawEntryStmt:
		entries, ok := objectEntries(s.value.json)
		if !ok {
			t.report("bad-raw", "a raw entry is a JSON object of the fields it sets", s.value.span)
			return written
		}
		for _, e := range entries {
			p := find(e.key, s.value.span, true)
			if p == nil {
				continue
			}
			pop := t.in(p.keys()...)
			v, ok := t.rawField(p.field, &rawNode{span: s.value.span, json: string(e.value)}, p.Repeated)
			pop()
			if !ok || !p.isList() {
				put(p, v, ok)
				continue
			}
			for i := range v.List().Len() {
				put(p, v.List().Get(i), true)
			}
		}
	}
	return written
}

type jsonEntry struct {
	key   string
	value json.RawMessage
}

// The members of a JSON object in the order their keys first appear, a repeated key holding its last value as
// JSON.parse reads it; false when text is no object.
func objectEntries(text string) ([]jsonEntry, bool) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &object); err != nil || object == nil {
		return nil, false
	}
	// The text is a JSON object, so the decoder reads it without error: '{', then each key and its value.
	dec := json.NewDecoder(strings.NewReader(text))
	_, _ = dec.Token()
	var entries []jsonEntry
	seen := map[string]bool{}
	for dec.More() {
		key, _ := dec.Token()
		var value json.RawMessage
		_ = dec.Decode(&value)
		if k := key.(string); !seen[k] {
			seen[k] = true
			entries = append(entries, jsonEntry{k, object[k]})
		}
	}
	return entries, true
}
