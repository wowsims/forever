package siml

import (
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Canonical printing (SPEC.md §12): settings as SIML text, for every file kind. Each node takes the shortest form
// that reads back to it, the generic call when no shorter one fits, and raw JSON when no form does; every printed
// statement is then checked by reading the whole text back.

const (
	printWidth   = 100
	indentWidth  = 2
	continuation = 4
	slotWidth    = 11 // A slot row's slot word, padded.
)

// Precedence, loosest first (SPEC.md §6.5).
const (
	precTop = iota
	precOr
	precAnd
	precNot
	precCmp
	precAdd
	precMul
	precOperand
)

var (
	identPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	unprintable  = regexp.MustCompile(`[\x00-\x09\x0b-\x1f\x7f]| (\n|$)`)
	logicPrec    = map[string]int{"and": precAnd, "or": precOr}
	mathPrec     = map[proto.APLValueMath_MathOperator]int{
		proto.APLValueMath_OpAdd: precAdd, proto.APLValueMath_OpSub: precAdd, proto.APLValueMath_OpMul: precMul, proto.APLValueMath_OpDiv: precMul,
	}
	// Words that never read as the name of a kind: the stop words, and true, false and none.
	notNames = func() map[string]bool {
		words := map[string]bool{"true": true, "false": true, "none": true}
		for w := range stopWords {
			words[w] = true
		}
		return words
	}()
	// Stop words that are calls when ( follows directly (SPEC.md §4 note 8).
	callWords = map[string]bool{"and": true, "or": true, "not": true}
	// Action kinds named by an action keyword, which no call can write (SPEC.md §4 note 7).
	keywordActions = map[string]bool{"wait": true, "move": true}
	// The keyword steps whose one field is a value (SPEC.md §6.4).
	valueSteps = map[string]struct{ keyword, field string }{
		"wait":          {"wait ", "duration"},
		"wait_until":    {"wait until ", "condition"},
		"move":          {"move ", "range_from_target"},
		"move_duration": {"move for ", "duration"},
	}
	compareText = invert(compareOperators)
	mathText    = invert(mathOperators)
)

func invert[K, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// Printed text in pieces; nil where no form writes a node.
type text []string

func word(s string) text { return text{s} }

func rawWord(json string) text { return word(rawText(json)) }

func cat(texts ...text) text { return slices.Concat(texts...) }

func joinTexts(texts []text, sep string) text {
	var out text
	for i, t := range texts {
		if i > 0 {
			out = append(out, sep)
		}
		out = append(out, t...)
	}
	return out
}

func wrap(t text, needed bool) text {
	if !needed {
		return t
	}
	return cat(word("("), t, word(")"))
}

type stmtRole int

const (
	roleNone stmtRole = iota
	roleRow
	roleVar
	roleEntry
	roleSlot
	roleUse
)

// A printed statement: its parts joined by spaces, or, past the width, the first part followed by each other part
// on a continuation line (SPEC.md §12.3).
type printedStmt struct {
	indent   int
	notes    []string // Its /// lines.
	parts    []text
	trailing string    // A comment after its last line.
	role     stmtRole  // A statement that is read back and checked.
	expect   *expected // Nil for a statement written as raw JSON.
}

// What a printed statement must read back as, the JSON of what it was printed from (as writtenJSON gives it), and
// the statement written instead when it does not.
type expected struct {
	json     string
	fallback func() *printedStmt
}

// A line of text, a blank line, or a statement. A header line that joins takes its section's one statement onto it
// when that is one line and the two fit in the width.
type item struct {
	blank bool
	line  string
	join  bool
	stmt  *printedStmt
}

func blankLine() *item { return &item{blank: true} }

// A field's values as a section's entry lines write them: one value, a list's values, or one message of a list.
type write struct {
	placed  *placed
	values  []protoreflect.Value // An invalid value writes the field's default.
	indices []int                // Each value's place in the field's list, or -1.
}

type printer struct {
	ix           *schemaIndex
	names        *Names
	class        proto.Class // The class of the printed file's spec line, whose names it writes (SPEC.md §7.1).
	declarations []declared  // The variables the rotation declares, which decide how a reference is written.
	kind         kindScope   // The kind whose own fields are being printed.
	// The values def calls read as, which format writes as the calls.
	calls map[*proto.APLValue]string
	uses  *uses // How format reads the printed text back: following its use lines, or nil for one file alone.
}

// The call a value is read from, as format writes it, or "".
func (p *printer) defCall(v protoreflect.Message) string {
	if m, ok := v.Interface().(*proto.APLValue); ok {
		return p.calls[m]
	}
	return ""
}

// The schema field of m named name in the protos.
func (p *printer) fieldOf(m protoreflect.Message, name string) *field {
	return p.ix.messages[string(m.Descriptor().FullName())].byProto[name]
}

// The value of m's field name, as protojson holds it; false when m has no such field or does not set it.
func (p *printer) get(m protoreflect.Message, name string) (protoreflect.Value, bool) {
	f := p.fieldOf(m, name)
	if f == nil {
		return protoreflect.Value{}, false
	}
	return fieldValue(m, f.desc)
}

// The schema fields m sets, in declaration order.
func (p *printer) setFields(m protoreflect.Message) []*field {
	if !m.IsValid() {
		return nil
	}
	var out []*field
	for _, f := range p.ix.messages[string(m.Descriptor().FullName())].fields {
		if _, ok := fieldValue(m, f.desc); ok {
			out = append(out, f)
		}
	}
	return out
}

// Whether every field m sets is one of names.
func (p *printer) only(m protoreflect.Message, names ...string) bool {
	return !slices.ContainsFunc(p.setFields(m), func(f *field) bool { return !slices.Contains(names, f.Name) })
}

// The kind m holds: the one field it sets, when that is a kind of kinds.
func (p *printer) kindOf(m protoreflect.Message, kinds map[string]*aplKind, except string) (*aplKind, protoreflect.Message) {
	set := slices.DeleteFunc(p.setFields(m), func(f *field) bool { return f.Name == except })
	if len(set) != 1 || kinds[set[0].Name] == nil {
		return nil, nil
	}
	return kinds[set[0].Name], m.Get(set[0].desc).Message()
}

// --- Values (SPEC.md §6.5, §8.3)

func (p *printer) value(v protoreflect.Message, prec int, sc scope) text {
	if !v.IsValid() {
		return rawWord("null")
	}
	if c := p.defCall(v); c != "" {
		return word(c)
	}
	if len(p.setFields(v)) == 0 {
		return word("none")
	}
	k, body := p.kindOf(v, p.ix.valueKinds, "")
	if k == nil {
		return rawWord(p.ix.jsonText(v))
	}
	if t := p.shortValue(k, body, prec, sc); t != nil {
		return t
	}
	if t := p.call(k, body, sc, false); t != nil {
		return t
	}
	return rawWord(p.ix.jsonText(v))
}

// The operator and operands of a value written as a run of two or more `and` or `or` operands.
func (p *printer) infix(v protoreflect.Message, sc scope) (string, []text, bool) {
	set := p.setFields(v)
	if len(set) != 1 || logicPrec[set[0].Name] == 0 || p.defCall(v) != "" {
		return "", nil, false
	}
	operands, ok := p.operands(set[0].Name, v.Get(set[0].desc).Message(), sc)
	return set[0].Name, operands, ok
}

// The operands of an `and` or `or` body that holds two or more.
func (p *printer) operands(op string, body protoreflect.Message, sc scope) ([]text, bool) {
	vals, ok := p.get(body, "vals")
	if !p.only(body, "vals") || !ok || vals.List().Len() < 2 {
		return nil, false
	}
	operands := make([]text, vals.List().Len())
	for i := range operands {
		operands[i] = p.value(vals.List().Get(i).Message(), logicPrec[op]+1, sc)
	}
	return operands, true
}

func (p *printer) shortValue(k *aplKind, body protoreflect.Message, prec int, sc scope) text {
	switch k.Kind {
	case "const":
		if !p.only(body, "val") {
			return nil
		}
		val := body.Get(p.fieldOf(body, "val").desc).String()
		if isLiteral(val) || val == "true" || val == "false" {
			return word(val)
		}
		return word(jsonString(val))
	case "variable_ref":
		if !p.only(body, "name") {
			return nil
		}
		return p.variableRef(body.Get(p.fieldOf(body, "name").desc).String())
	case "variable_placeholder":
		if !p.only(body, "name") {
			return nil
		}
		return placeholderText(body.Get(p.fieldOf(body, "name").desc).String())
	case "and", "or":
		operands, ok := p.operands(k.Kind, body, sc)
		if !ok {
			return nil
		}
		return wrap(joinTexts(operands, " "+k.Kind+" "), prec > logicPrec[k.Kind])
	case "not":
		val, ok := p.get(body, "val")
		if !ok || !p.only(body, "val") {
			return nil
		}
		return wrap(cat(word("not "), p.value(val.Message(), precNot, sc)), prec > precNot)
	case "cmp", "math":
		op := body.Get(p.fieldOf(body, "op").desc).Enum()
		lhs, okL := p.get(body, "lhs")
		rhs, okR := p.get(body, "rhs")
		var opText string
		var at, left, right int
		if k.Kind == "cmp" {
			opText = compareText[proto.APLValueCompare_ComparisonOperator(op)]
			at, left, right = precCmp, precAdd, precAdd
		} else {
			mathOp := proto.APLValueMath_MathOperator(op)
			opText = mathText[mathOp]
			at, left, right = mathPrec[mathOp], mathPrec[mathOp], mathPrec[mathOp]+1
		}
		if opText == "" || !okL || !okR || !p.only(body, "op", "lhs", "rhs") {
			return nil
		}
		return wrap(cat(p.value(lhs.Message(), left, sc), word(" "+opText+" "), p.value(rhs.Message(), right, sc)), prec > at)
	}
	switch k.Form {
	case "accessor":
		return p.accessor(k, body, sc)
	case "unit_prefix":
		if unit, ok := p.get(body, k.UnitField); ok && p.only(body, k.UnitField) {
			if w := p.unitWord(unit.Message()); w != "" {
				return word(w + "." + k.Kind)
			}
		}
	}
	return nil
}

// `$x` when SPEC.md §7.5 reads it back as name, else `${name}`.
func (p *printer) variableRef(name string) text {
	declared := declaredIn(p.declarations, scope{})
	readsBack := func(written string) bool {
		r := resolveVariable(written, false, declared)
		if r.found {
			return r.name == name
		}
		return len(r.ambiguous) == 0 && written == name
	}
	if identPattern.MatchString(name) && readsBack(name) {
		return word("$" + name)
	}
	if folded := Identifier(name); identPattern.MatchString(folded) && readsBack(folded) {
		return word("$" + folded)
	}
	return word(braced("$", name))
}

// `%x`, or `%{name}` when the name is no ident.
func placeholderText(name string) text {
	if identPattern.MatchString(name) {
		return word("%" + name)
	}
	return word(braced("%", name))
}

var bracedEscapes = strings.NewReplacer(`\`, `\\`, "}", `\}`, "\n", `\n`)

// `${name}` or `%{name}`, with `\`, `}` and line feeds escaped (SPEC.md §2).
func braced(sigil, name string) string { return sigil + "{" + bracedEscapes.Replace(name) + "}" }

// `[u.]head(subject).prop[(f=v, …)]`, when the subject is set and the unit, if set, is a unit word.
func (p *printer) accessor(k *aplKind, body protoreflect.Message, sc scope) text {
	defer p.enter(k, body)()
	acc := k.Accessor
	subject := k.message.byProto[acc.Subject]
	unitField := k.message.byProto[acc.Unit]
	sv, ok := fieldValue(body, subject.desc)
	if !ok {
		return nil
	}
	prefix := ""
	if unitField != nil {
		if uv, ok := fieldValue(body, unitField.desc); ok {
			unit := p.unitWord(uv.Message())
			if unit == "" {
				return nil
			}
			prefix = unit + "."
		}
	}
	var args []text
	for _, f := range k.message.fields {
		if v, ok := fieldValue(body, f.desc); ok && f != subject && f != unitField {
			args = append(args, cat(word(f.Siml+"="), p.field(f, v, sc)))
		}
	}
	out := cat(word(prefix+acc.Head+"("), p.field(subject, sv, sc), word(")."+acc.Prop))
	if len(args) > 0 {
		out = cat(out, word("("), joinTexts(args, ", "), word(")"))
	}
	return out
}

// A kind as a call (SPEC.md §8.3 rules 3, 5 and 6): bare when no field is set, else the positional value or the
// list's items unnamed and every other field named; nil when no call can write its name.
func (p *printer) call(k *aplKind, body protoreflect.Message, sc scope, action bool) text {
	bare := !notNames[k.Kind] && !(action && keywordActions[k.Kind])
	if !bare && !callWords[k.Kind] {
		return nil
	}
	set := p.setFields(body)
	if len(set) == 0 {
		if bare {
			return word(k.Kind)
		}
		return word(k.Kind + "()")
	}
	defer p.enter(k, body)()
	var list, positional *field
	if k.ListField != "" {
		list = k.message.byProto[k.ListField]
	} else if k.Positional != "" {
		positional = k.message.byProto[k.Positional]
	}
	bools := map[string]bool{}
	for _, f := range k.message.fields {
		if f.Printed == "bool" {
			bools[f.Siml] = true
		}
	}
	var args []text
	if list != nil {
		if lv, ok := fieldValue(body, list.desc); ok {
			for i := range lv.List().Len() {
				args = append(args, p.single(list, lv.List().Get(i), sc))
			}
		}
	}
	if positional != nil {
		if pv, ok := fieldValue(body, positional.desc); ok {
			t := p.field(positional, pv, sc)
			if len(t) == 1 && bools[t[0]] {
				t = cat(word(positional.Siml+"="), t)
			}
			args = append(args, t)
		}
	}
	for _, f := range set {
		if f != list && f != positional {
			v, _ := fieldValue(body, f.desc)
			args = append(args, cat(word(f.Siml+"="), p.field(f, v, sc)))
		}
	}
	return cat(word(k.Kind+"("), joinTexts(args, ", "), word(")"))
}

// --- Fields (SPEC.md §5.2, §8.3)

// A field's whole value: one value, or a list's values in brackets.
func (p *printer) field(f *field, v protoreflect.Value, sc scope) text {
	if !f.isList() {
		return p.single(f, v, sc)
	}
	items := make([]text, v.List().Len())
	for i := range items {
		items[i] = p.single(f, v.List().Get(i), sc)
	}
	return cat(word("["), joinTexts(items, ", "), word("]"))
}

// One value of f, an element of a repeated field; an invalid value writes the field's default.
func (p *printer) single(f *field, v protoreflect.Value, sc scope) text {
	if !v.IsValid() {
		return word(p.zero(f))
	}
	switch f.Printed {
	case "value":
		return p.value(v.Message(), precTop, sc)
	case "action":
		return p.inlineAction(v.Message(), sc)
	case "action_id":
		return p.ref(v.Message(), p.kind.namespace(f), sc)
	case "unit":
		return word(p.unitText(v.Message()))
	case "bool":
		return word(strconv.FormatBool(v.Bool()))
	case "number":
		if s := numberValue(f.desc, v); s != "" {
			return word(s)
		}
	case "string":
		return word(jsonString(v.String()))
	case "enum", "flag_enum":
		if e := p.ix.enums[f.Type].byNumber[int32(v.Enum())]; e != nil {
			return word(e.Short)
		}
	case "stat":
		if e := p.ix.enums[statEnum].byNumber[int32(intOf(f.desc, v))]; e != nil {
			return word(e.Short)
		}
		return word(numberValue(f.desc, v))
	case "ref":
		if name, ok := p.idName(Namespace(f.Ref), intOf(f.desc, v), -1); ok {
			return word(name)
		}
	case "message":
		if len(p.setFields(v.Message())) == 0 {
			return word("none")
		}
	}
	return rawWord(p.ix.valueJSON(f.desc, v, f.isList()))
}

// A number field's value as SIML writes it, "" for a float that is not finite.
func numberValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) string {
	switch fd.Kind() {
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return ""
		}
		bits := 64
		if fd.Kind() == protoreflect.FloatKind {
			bits = 32
		}
		return numberText(f, bits)
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(v.Uint(), 10)
	}
	return strconv.FormatInt(v.Int(), 10)
}

// An integer field's value.
func intOf(fd protoreflect.FieldDescriptor, v protoreflect.Value) int64 {
	switch fd.Kind() {
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return int64(min(v.Uint(), math.MaxInt64))
	}
	return v.Int()
}

// A field's default value, which protojson leaves out.
func (p *printer) zero(f *field) string {
	enumZero := func(name string) string {
		if e := p.ix.enums[name].byNumber[0]; e != nil {
			return e.Short
		}
		return "0"
	}
	switch f.Printed {
	case "bool":
		return "false"
	case "number":
		return "0"
	case "string":
		return `""`
	case "enum", "flag_enum":
		return enumZero(f.Type)
	case "stat":
		return enumZero(statEnum)
	}
	return "none"
}

// An ActionID: the name of its spell id and tag in ns, item.<name>, other.<short>, none, or raw JSON (SPEC.md §7.4).
func (p *printer) ref(v protoreflect.Message, ns Namespace, sc scope) text {
	set := p.setFields(v)
	if len(set) == 0 {
		return word("none")
	}
	only := func(names ...string) bool {
		return !slices.ContainsFunc(set, func(f *field) bool { return !slices.Contains(names, f.Name) })
	}
	switch {
	case len(set) == 1 && set[0].Name == "other_id":
		if e := p.ix.enums[otherEnum].byNumber[int32(v.Get(set[0].desc).Enum())]; e != nil {
			return word("other." + e.Short)
		}
	case only("spell_id", "tag", "rank"):
		if name, ok := p.names.SpellName(p.class, ns, v.Interface().(*proto.ActionID)); ok {
			return word(name)
		}
	case only("item_id"):
		if name, ok := p.idName(NamespaceItem, v.Get(set[0].desc).Int(), -1); ok {
			return word("item." + name)
		}
	}
	return rawWord(p.ix.jsonText(v))
}

// The name of an item, enchant effect or random suffix id; an enchant's for slot, which is -1 when it is on none.
func (p *printer) idName(ns Namespace, id int64, slot proto.ItemSlot) (string, bool) {
	if id <= 0 || id > math.MaxInt32 {
		return "", false
	}
	switch ns {
	case NamespaceItem:
		return p.names.ItemName(int32(id))
	case NamespaceEnchant:
		return p.names.EnchantName(int32(id), slot)
	case NamespaceSuffix:
		return p.names.SuffixName(int32(id))
	}
	return "", false
}

// Prints the fields of the kind k, whose message is body, next; gives what restores the kind before.
func (p *printer) enter(k *aplKind, body protoreflect.Message) func() {
	outer := p.kind
	p.kind = kindScope{kind: k}
	if unit := k.unitField(); unit != nil {
		if v, ok := fieldValue(body, unit.desc); ok {
			p.kind.unit = v.Message().Interface().(*proto.UnitReference).GetType()
		}
	}
	return func() { p.kind = outer }
}

func (p *printer) unitText(v protoreflect.Message) string {
	if len(p.setFields(v)) == 0 {
		return "none"
	}
	if w := p.unitWord(v); w != "" {
		return w
	}
	return rawText(p.ix.jsonText(v))
}

// The unit word that writes v (SPEC.md §6.6), if one does: the word that reads back as v.
func (p *printer) unitWord(v protoreflect.Message) string {
	if len(p.setFields(v)) == 0 {
		return ""
	}
	u := v.Interface()
	for _, w := range p.ix.unitWordList() {
		if googleProto.Equal(p.ix.unitReference(w), u) {
			return w
		}
	}
	typ := v.Get(p.fieldOf(v, "type").desc).Enum()
	index := v.Get(p.fieldOf(v, "index").desc).Int()
	for prefix, t := range p.ix.unitPrefixes {
		if protoreflect.EnumNumber(t) != typ || index < 0 {
			continue
		}
		w := prefix + strconv.FormatInt(index+1, 10)
		if googleProto.Equal(p.ix.unitReference(w), u) {
			return w
		}
	}
	return ""
}

// --- Setup sections (SPEC.md §5.2, §5.3, §12.2)

// The entry lines of a section's writes: each run of flags as flag lists packed into lines of at most 100 columns,
// and every other field, or message of a list, on a line of its own.
func (p *printer) entries(writes []write, indent int) []*printedStmt {
	var out []*printedStmt
	type flagged struct {
		write write
		text  string
	}
	var run []flagged
	flush := func() {
		var lines [][]flagged
		width := 0
		for _, fl := range run {
			if n := len(lines); n > 0 && indent+width+2+textWidth(fl.text) <= printWidth {
				lines[n-1] = append(lines[n-1], fl)
				width += 2 + textWidth(fl.text)
			} else {
				lines = append(lines, []flagged{fl})
				width = textWidth(fl.text)
			}
		}
		for _, line := range lines {
			writes := make([]write, len(line))
			flags := make([]string, len(line))
			for i, fl := range line {
				writes[i], flags[i] = fl.write, fl.text
			}
			out = append(out, p.entryStmt(writes, []text{word(strings.Join(flags, ", "))}, indent))
		}
		run = nil
	}
	for _, w := range writes {
		if flag, ok := p.flag(w.placed.field, w.values[0]); ok && len(w.values) == 1 {
			run = append(run, flagged{w, flag})
			continue
		}
		flush()
		out = append(out, p.entryStmt([]write{w}, p.entryParts(w), indent))
	}
	flush()
	return out
}

// A field's value as a flag: a true bool by its name, a flag-value enum by its name at its flag value and as
// name=value at another; false when it is none.
func (p *printer) flag(f *field, v protoreflect.Value) (string, bool) {
	if f.isList() || topWords[f.Siml] || !v.IsValid() {
		return "", false
	}
	switch f.Printed {
	case "bool":
		return f.Siml, v.Bool()
	case "flag_enum":
		e := p.ix.enums[f.Type]
		value := e.byNumber[int32(v.Enum())]
		switch {
		case value == nil:
			return "", false
		case value.Name == e.flag:
			return f.Siml, true
		}
		return f.Siml + "=" + value.Short, true
	}
	return "", false
}

// One write's entry line: `field value`; `field a, b` for a list, broken after each comma past the width;
// `field(…)` for a message, broken after each argument's comma; and `field=value` for a field named like a
// top-level line, which `field value` would read as.
func (p *printer) entryParts(w write) []text {
	f := w.placed.field
	switch {
	case topWords[f.Siml] && f.isList():
		items := make([]text, len(w.values))
		for i, v := range w.values {
			items[i] = p.single(f, v, scope{})
		}
		return []text{cat(word(f.Siml+"=["), joinTexts(items, ", "), word("]"))}
	case topWords[f.Siml]:
		return []text{cat(word(f.Siml+"="), p.single(f, w.values[0], scope{}))}
	case f.Printed == "message" && w.values[0].IsValid():
		return p.messageParts(f, w.values[0].Message())
	case f.Printed == "indexed":
		return []text{p.indexed(f, w.values[0])}
	}
	parts := make([]text, len(w.values))
	for i, v := range w.values {
		t := p.single(f, v, scope{})
		if i == 0 {
			t = cat(word(f.Siml+" "), t)
		}
		if i < len(w.values)-1 {
			t = cat(t, word(","))
		}
		parts[i] = t
	}
	return parts
}

// An entry line, which must read back as its writes, else is written as a raw entry of the fields they set.
func (p *printer) entryStmt(writes []write, parts []text, indent int) *printedStmt {
	var values []entryWrite
	for _, w := range writes {
		for i, v := range w.values {
			values = append(values, entryWrite{placed: w.placed, index: w.indices[i], value: v})
		}
	}
	fallback := func() *printedStmt {
		var b strings.Builder
		b.WriteByte('{')
		for i, w := range writes {
			if i > 0 {
				b.WriteByte(',')
			}
			writeString(&b, w.placed.JSON)
			b.WriteByte(':')
			if !w.placed.isList() {
				b.WriteString(p.ix.valueJSON(w.placed.desc, w.values[0], false))
				continue
			}
			items := make([]string, len(w.values))
			for k, v := range w.values {
				items[k] = p.ix.valueJSON(w.placed.desc, v, true)
			}
			b.WriteString("[" + strings.Join(items, ",") + "]")
		}
		b.WriteByte('}')
		return &printedStmt{indent: indent, parts: []text{rawWord(b.String())}, role: roleEntry}
	}
	return &printedStmt{indent: indent, parts: parts, role: roleEntry, expect: &expected{json: p.ix.writtenJSON(values), fallback: fallback}}
}

// `field(arg, …)`, each argument a part of its own so that a long one breaks after each comma.
func (p *printer) messageParts(f *field, v protoreflect.Message) []text {
	args := p.messageArgs(f.Type, v)
	if len(args) < 2 {
		return []text{cat(word(f.Siml+"("), joinTexts(args, ""), word(")"))}
	}
	parts := make([]text, len(args))
	for i, a := range args {
		end := ","
		if i == len(args)-1 {
			end = ")"
		}
		parts[i] = cat(a, word(end))
	}
	parts[0] = cat(word(f.Siml+"("), parts[0])
	return parts
}

// The fields of a message as arguments: a true bool, or a flag-value enum at its flag value, by its name alone; a
// message as field(…); each message of a list as a field(…) of its own; every other field named.
func (p *printer) messageArgs(message string, v protoreflect.Message) []text {
	var out []text
	for _, f := range p.ix.messages[message].fields {
		x, ok := fieldValue(v, f.desc)
		if !ok {
			continue
		}
		if flag, ok := p.flag(f, x); ok {
			out = append(out, word(flag))
			continue
		}
		switch {
		case f.Printed == "message" && f.isList():
			for i := range x.List().Len() {
				out = append(out, p.messageText(f, x.List().Get(i).Message()))
			}
		case f.Printed == "message":
			out = append(out, p.messageText(f, x.Message()))
		case f.Printed == "indexed":
			out = append(out, p.indexed(f, x))
		default:
			out = append(out, cat(word(f.Siml+"="), p.field(f, x, scope{})))
		}
	}
	return out
}

func (p *printer) messageText(f *field, v protoreflect.Message) text {
	return cat(word(f.Siml+"("), joinTexts(p.messageArgs(f.Type, v), ", "), word(")"))
}

// `field(<short name>=<number>, …)` for an array indexed by an enum, its non-zero entries in index order, or
// `field=` raw JSON when an entry has no name or is no number.
func (p *printer) indexed(f *field, v protoreflect.Value) text {
	raw := cat(word(f.Siml+"="), rawWord(p.ix.valueJSON(f.desc, v, false)))
	if !v.IsValid() {
		return raw
	}
	e := p.ix.enums[f.Index]
	var entries []text
	for i := range v.List().Len() {
		x := v.List().Get(i)
		n := numberValue(f.desc, x)
		if n == "0" {
			continue
		}
		value := e.byNumber[int32(i)]
		if value == nil || n == "" {
			return raw
		}
		entries = append(entries, word(value.Short+"="+n))
	}
	return cat(word(f.Siml+"("), joinTexts(entries, ", "), word(")"))
}

// `<slot> <item> [<suffix>] [+ <enchant>]` by their names, or the slot and the raw ItemSpec when a part has no name
// (SPEC.md §5.3, §7.4). enchant is the name the enchant is written with, "" for none, which it keeps while that means
// its effect (SPEC.md §12.1).
func (p *printer) slotRow(slot string, item protoreflect.Message, slots *placed, index int, indent int, enchant string) *printedStmt {
	pad := padEnd(slot, slotWidth-1) + " "
	fallback := func() *printedStmt {
		return &printedStmt{indent: indent, parts: []text{cat(word(pad), rawWord(p.ix.jsonText(item)))}, role: roleSlot}
	}
	exp := &expected{json: p.ix.writtenJSON([]entryWrite{{placed: slots, index: index, value: protoreflect.ValueOfMessage(item)}}), fallback: fallback}
	raw := fallback()
	raw.expect = exp
	if !p.only(item, "id", "random_suffix", "enchant") {
		return raw
	}
	t := word(pad)
	for _, part := range []struct {
		field, before string
		ns            Namespace
	}{{"id", "", NamespaceItem}, {"random_suffix", " ", NamespaceSuffix}, {"enchant", " + ", NamespaceEnchant}} {
		v, ok := p.get(item, part.field)
		if !ok && part.ns == NamespaceItem {
			return raw
		}
		if !ok {
			continue
		}
		name, ok := p.idName(part.ns, v.Int(), proto.ItemSlot(index))
		if id, written := p.names.Enchant(enchant); part.ns == NamespaceEnchant && written && int64(id) == v.Int() {
			name, ok = enchant, true
		}
		if !ok {
			return raw
		}
		t = cat(t, word(part.before+name))
	}
	return &printedStmt{indent: indent, parts: []text{t}, role: roleSlot, expect: exp}
}

// A variable a use line's with list sets, with the value it gives.
type useBind struct {
	name   string
	braced bool
	value  *proto.APLValue
}

// `use "path" [with $v = value, …]`, its values checked as the with list's values.
func (p *printer) use(path string, binds []useBind, indent int) *printedStmt {
	line := func(value func(*proto.APLValue) text) []text {
		t := word("use " + jsonString(path))
		if len(binds) == 0 {
			return []text{t}
		}
		list := make([]text, len(binds))
		for i, b := range binds {
			name := "$" + b.name
			if b.braced {
				name = braced("$", b.name)
			}
			list[i] = cat(word(name), word(" = "), value(b.value))
		}
		return []text{cat(t, word(" with "), joinTexts(list, ", "))}
	}
	values := make([]*proto.APLValue, len(binds))
	for i, b := range binds {
		values[i] = b.value
	}
	fallback := func() *printedStmt {
		return &printedStmt{indent: indent, parts: line(func(v *proto.APLValue) text { return rawWord(p.ix.jsonText(v.ProtoReflect())) }), role: roleUse}
	}
	parts := line(func(v *proto.APLValue) text { return p.value(v.ProtoReflect(), precTop, scope{}) })
	return &printedStmt{indent: indent, parts: parts, role: roleUse, expect: &expected{json: p.ix.withsJSON(values), fallback: fallback}}
}

// --- Actions (SPEC.md §6.4)

// An action's step, without its condition; nil when only raw JSON writes it.
func (p *printer) step(a protoreflect.Message, sc scope) text {
	if !slices.ContainsFunc(p.setFields(a), func(f *field) bool { return f.Name != "condition" }) {
		return word("none")
	}
	k, body := p.kindOf(a, p.ix.actionKinds, "condition")
	if k == nil {
		return nil
	}
	if t := p.shortAction(k, body, sc); t != nil {
		return t
	}
	return p.call(k, body, sc, true)
}

func (p *printer) shortAction(k *aplKind, body protoreflect.Message, sc scope) text {
	switch k.Kind {
	case "cast_spell", "channel_spell":
		fields := []string{"spell_id", "target"}
		out := word("cast ")
		if k.Kind == "channel_spell" {
			fields = append(fields, "interrupt_if", "allow_recast")
			out = word("channel ")
		}
		spell, ok := p.get(body, "spell_id")
		if !ok || !p.only(body, fields...) {
			return nil
		}
		out = cat(out, p.ref(spell.Message(), NamespaceSpell, sc))
		if target, ok := p.get(body, "target"); ok {
			out = cat(out, word(" on "+p.unitText(target.Message())))
		}
		if until, ok := p.get(body, "interrupt_if"); ok {
			out = cat(out, word(" until "), p.value(until.Message(), precTop, sc))
		}
		if _, ok := p.get(body, "allow_recast"); ok {
			out = cat(out, word(" recast"))
		}
		return out
	case "wait", "wait_until", "move", "move_duration":
		form := valueSteps[k.Kind]
		v, ok := p.get(body, form.field)
		if !ok || !p.only(body, form.field, "recheck_while_moving") {
			return nil
		}
		out := cat(word(form.keyword), p.value(v.Message(), precTop, sc))
		if _, ok := p.get(body, "recheck_while_moving"); ok {
			out = cat(out, word(" recheck"))
		}
		return out
	case "group_reference":
		if !p.only(body, "group_name", "variables") {
			return nil
		}
		var binds []text
		if vars, ok := p.get(body, "variables"); ok {
			for i := range vars.List().Len() {
				x := vars.List().Get(i).Message()
				if !p.only(x, "name", "value") {
					return nil
				}
				value, ok := p.get(x, "value")
				if !ok {
					return nil
				}
				placeholder := placeholderText(x.Get(p.fieldOf(x, "name").desc).String())
				binds = append(binds, cat(placeholder, word(" = "), p.value(value.Message(), precTop, sc)))
			}
		}
		out := word(groupKey(body.Get(p.fieldOf(body, "group_name").desc).String()))
		if len(binds) > 0 {
			out = cat(out, word(" with "), joinTexts(binds, ", "))
		}
		return out
	case "schedule":
		inner, ok := p.get(body, "inner_action")
		if !ok || !p.only(body, "schedule", "inner_action") {
			return nil
		}
		if _, hasCondition := p.get(inner.Message(), "condition"); hasCondition {
			return nil
		}
		step := p.step(inner.Message(), sc)
		if step == nil {
			step = rawWord(p.ix.jsonText(inner.Message()))
		}
		return cat(word("schedule "+jsonString(body.Get(p.fieldOf(body, "schedule").desc).String())+" "), step)
	case "sequence":
		if !p.only(body, "name", "actions") {
			return nil
		}
		var actions []text
		if list, ok := p.get(body, "actions"); ok {
			for i := range list.List().Len() {
				actions = append(actions, p.inlineAction(list.List().Get(i).Message(), sc))
			}
		}
		return cat(word("sequence "+jsonString(body.Get(p.fieldOf(body, "name").desc).String())+" ["), joinTexts(actions, ", "), word("]"))
	}
	return nil
}

// An action written in one piece: `step if condition`, or raw JSON.
func (p *printer) inlineAction(v protoreflect.Message, sc scope) text {
	condition, hasCondition := p.get(v, "condition")
	step := p.step(v, sc)
	if step == nil {
		step = rawWord(p.ix.jsonText(without(v, p.fieldOf(v, "condition").desc)))
	}
	if !hasCondition {
		return step
	}
	return cat(step, word(" if "), p.value(condition.Message(), precTop, sc))
}

// A copy of m without its field fd.
func without(m protoreflect.Message, fd protoreflect.FieldDescriptor) protoreflect.Message {
	c := googleProto.Clone(m.Interface()).ProtoReflect()
	c.Clear(fd)
	return c
}

// --- Statements

// A row (SPEC.md §6.3): an APLPrepullAction in prepull:, an APLListItem elsewhere. A row no form writes is raw JSON
// alone, which reads as the whole row.
func (p *printer) row(item protoreflect.Message, prepull bool, sc scope, indent int) *printedStmt {
	fallback := func() *printedStmt {
		return &printedStmt{indent: indent, parts: []text{rawWord(p.ix.jsonText(item))}, role: roleRow}
	}
	exp := &expected{json: p.ix.comparedJSON(item), fallback: fallback}
	raw := fallback()
	raw.expect = exp
	if !item.IsValid() {
		return raw
	}
	allowed := []string{"hide", "notes", "action"}
	if prepull {
		allowed = []string{"action", "do_at_value", "hide"}
	}
	action, hasAction := p.get(item, "action")
	notes, hasNotes := p.get(item, "notes")
	_, hide := p.get(item, "hide")
	doAt, hasDoAt := p.get(item, "do_at_value")
	if !p.only(item, allowed...) || !hasAction || prepull && !hasDoAt || hasNotes && unprintable.MatchString(notes.String()) {
		return raw
	}
	condition, hasCondition := p.get(action.Message(), "condition")
	step := p.step(action.Message(), sc)
	if step == nil && !hide && !prepull && !hasCondition && !hasNotes {
		return raw
	}
	head := word("")
	if hide {
		head = word("# ")
	}
	if prepull {
		head = cat(head, p.value(doAt.Message(), precOperand, sc), word(" "))
	}
	if step == nil {
		step = rawWord(p.ix.jsonText(without(action.Message(), p.fieldOf(action.Message(), "condition").desc)))
	}
	parts := []text{cat(head, step)}
	if hasCondition {
		parts = append(parts, p.conditionParts(condition.Message(), sc)...)
	}
	s := &printedStmt{indent: indent, parts: parts, role: roleRow, expect: exp}
	if hasNotes {
		s.notes = strings.Split(notes.String(), "\n")
	}
	return s
}

// `if c`, and each further operand of a top-level and or or as a part of its own.
func (p *printer) conditionParts(condition protoreflect.Message, sc scope) []text {
	op, operands, ok := p.infix(condition, sc)
	if !ok {
		return []text{cat(word("if "), p.value(condition, precTop, sc))}
	}
	parts := make([]text, len(operands))
	for i, t := range operands {
		lead := op + " "
		if i == 0 {
			lead = "if "
		}
		parts[i] = cat(word(lead), t)
	}
	return parts
}

// `var $x = v` (SPEC.md §6.1), breaking after = and before each operand of a top-level and or or; `var $x` for a
// variable with no value.
func (p *printer) variable(v protoreflect.Message, sc scope, indent int) *printedStmt {
	name := ""
	var value protoreflect.Message
	if v.IsValid() {
		name = v.Get(p.fieldOf(v, "name").desc).String()
		value = v.Get(p.fieldOf(v, "value").desc).Message()
	}
	declared := "var " + braced("$", name)
	if identPattern.MatchString(name) {
		declared = "var $" + name
	}
	head := word(declared + " =")
	fallback := func() *printedStmt {
		if value == nil || !value.IsValid() {
			return &printedStmt{indent: indent, parts: []text{word(declared)}, role: roleVar}
		}
		return &printedStmt{indent: indent, parts: []text{head, rawWord(p.ix.jsonText(value))}, role: roleVar}
	}
	exp := &expected{json: p.ix.comparedJSON(v), fallback: fallback}
	if value == nil || !value.IsValid() {
		s := fallback()
		s.expect = exp
		return s
	}
	parts := []text{head}
	if op, operands, ok := p.infix(value, sc); ok {
		for i, t := range operands {
			if i > 0 {
				t = cat(word(op+" "), t)
			}
			parts = append(parts, t)
		}
	} else {
		parts = append(parts, p.value(value, precTop, sc))
	}
	return &printedStmt{indent: indent, parts: parts, role: roleVar, expect: exp}
}

// --- Layout

// The text of items.
func layout(items []*item) string {
	var out []string
	for i := 0; i < len(items); i++ {
		it := items[i]
		switch {
		case it.blank:
			out = append(out, "")
		case it.stmt != nil:
			out = append(out, statementLines(it.stmt)...)
		case it.join && i+1 < len(items) && items[i+1].stmt != nil && len(items[i+1].stmt.notes) == 0:
			next := items[i+1].stmt
			joined := it.line + " " + strings.Join(texts(next), " ")
			if textWidth(joined) > printWidth {
				out = append(out, it.line)
				continue
			}
			if next.trailing != "" {
				joined += " " + next.trailing
			}
			out = append(out, joined)
			i++
		default:
			out = append(out, it.line)
		}
	}
	return strings.Join(out, "\n") + "\n"
}

// Each part of s as text.
func texts(s *printedStmt) []string {
	out := make([]string, len(s.parts))
	for i, t := range s.parts {
		out[i] = strings.Join(t, "")
	}
	return out
}

func statementLines(s *printedStmt) []string {
	pad := strings.Repeat(" ", s.indent)
	parts := texts(s)
	body := []string{pad + strings.Join(parts, " ")}
	if textWidth(body[0]) > printWidth && len(parts) > 1 {
		body = []string{pad + parts[0]}
		for _, part := range parts[1:] {
			body = append(body, strings.Repeat(" ", s.indent+continuation)+part)
		}
	}
	if s.trailing != "" {
		body[len(body)-1] += " " + s.trailing
	}
	lines := make([]string, 0, len(s.notes)+len(body))
	for _, note := range s.notes {
		if note == "" {
			lines = append(lines, pad+"///")
		} else {
			lines = append(lines, pad+"/// "+note)
		}
	}
	return append(lines, body...)
}

// The width of s in columns, which count UTF-16 code units (SPEC.md §2).
func textWidth(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

func padEnd(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-textWidth(s)))
}

// --- Reading back

// Whether a statement is one a printed statement reads back as.
func readsBack(s stmt) bool {
	switch s.(type) {
	case *rowStmt, *varStmt, *slotRowStmt, *fieldEntryStmt, *flagListStmt, *messageEntryStmt, *rawEntryStmt, *useStmt, *unparsedStmt:
		return true
	}
	return false
}

// The JSON the read-back check compares for m: its JSON with the rank of every ActionID left out, since a name reads
// back with the rank of its entry (SPEC.md §12.3). Raw JSON reads back exactly, so only a name's rank can differ.
func (ix *schemaIndex) comparedJSON(m protoreflect.Message) string {
	if !m.IsValid() {
		return ix.jsonText(m)
	}
	m = googleProto.Clone(m.Interface()).ProtoReflect()
	rank := actionIDMessage.Fields().ByName("rank")
	walkMessages(m, func(x protoreflect.Message) {
		if x.Descriptor().FullName() == actionIDMessage.FullName() {
			x.Clear(rank)
		}
	})
	return ix.jsonText(m)
}

// The JSON the read-back check compares for the values entry lines or slot rows write, each with its field's path
// and its place in its list.
func (ix *schemaIndex) writtenJSON(values []entryWrite) string {
	parts := make([]string, len(values))
	for i, w := range values {
		value := ix.valueJSON(w.placed.desc, w.value, w.index >= 0)
		if w.placed.desc.Message() != nil && !w.placed.desc.IsMap() && w.value.IsValid() {
			value = ix.comparedJSON(w.value.Message())
		}
		parts[i] = "[" + jsonString(w.placed.path) + "," + strconv.Itoa(w.index) + "," + value + "]"
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// The JSON the read-back check compares for the values of a with list.
func (ix *schemaIndex) withsJSON(values []*proto.APLValue) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = ix.comparedJSON(v.ProtoReflect())
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// The JSON the read-back check compares for what a statement of a typed file writes, in the form a printed
// statement's expectation gives it; false for a statement that writes nothing.
func (f *typedFile) writtenJSON(s stmt) (string, bool) {
	if m, ok := f.written[s]; ok {
		return f.typer.ix.comparedJSON(m), true
	}
	if values, ok := f.entries[s]; ok {
		return f.typer.ix.writtenJSON(values), true
	}
	if values, ok := f.withs[s]; ok {
		return f.typer.ix.withsJSON(values), true
	}
	return "", false
}

// Reads the text of items back as a file of kind, and writes each statement that does not read back as what it was
// printed from as its fallback instead, until every one does or has fallen back.
func verified(p *printer, kind Kind, items []*item) string {
	for {
		out := layout(items)
		f := readFile(p.ix, p.names, out, kind, p.uses)
		var read []stmt
		for _, s := range f.tree.inFileOrder() {
			if readsBack(s) {
				read = append(read, s)
			}
		}
		var printed []*item
		for _, it := range items {
			if it.stmt != nil && it.stmt.role != roleNone {
				printed = append(printed, it)
			}
		}
		if len(read) != len(printed) {
			return out
		}
		changed := false
		for k, it := range printed {
			e := it.stmt.expect
			if e == nil {
				continue
			}
			if json, ok := f.writtenJSON(read[k]); ok && json == e.json {
				continue
			}
			trailing := it.stmt.trailing
			it.stmt = e.fallback()
			it.stmt.trailing = trailing
			changed = true
		}
		if !changed {
			return out
		}
	}
}

// --- Files (SPEC.md §12.1)

// Prints s as a file of kind. The spec is spec, or with SpecUnknown the one whose message the player holds.
// categories are the site's category letters whose sections a full file writes; with nil, it writes every section
// that holds something. A section file writes its own section. Spells, auras, items, enchants and suffixes are written
// by their names in names, and as raw JSON where it has none (SPEC.md §7.4). It panics if the protos break a rule of
// SPEC.md §8, as Specs does.
func Print(kind Kind, spec proto.Spec, s *proto.IndividualSimSettings, categories []string, names *Names) []byte {
	ix, err := loadIndex()
	if err != nil {
		panic(err)
	}
	return []byte(printFile(ix, names, kind, specFor(ix, spec, s), s, categories))
}

// The spec settings are printed for: spec, or with SpecUnknown the one whose message the player holds; nil for none.
func specFor(ix *schemaIndex, spec proto.Spec, s *proto.IndividualSimSettings) *SpecInfo {
	for _, info := range ix.specs {
		if spec == proto.Spec_SpecUnknown && s.GetPlayer().ProtoReflect().Has(info.Field) || spec != proto.Spec_SpecUnknown && info.Spec == spec {
			return &info
		}
	}
	return nil
}

// The message the path of message fields reaches from m; invalid where one is missing.
func readAt(m protoreflect.Message, path []protoreflect.FieldDescriptor) protoreflect.Message {
	for _, fd := range path {
		m = m.Get(fd).Message()
	}
	return m
}

// The settings a file is printed from, after printable, and the spec's Rotation each string that holds one decodes
// to, by the string's JSON path.
type printInput struct {
	ix       *schemaIndex
	settings protoreflect.Message
	spec     *SpecInfo
	decoded  map[string]protoreflect.Message
}

// Writes each array an enum indexes in m out to the enum's length, and leaves out one that is all zeros, which
// SPEC.md §12.4 reads as absent. An array that holds a number that is not finite is kept as it is.
func (ix *schemaIndex) printable(m protoreflect.Message) {
	msg := ix.messages[string(m.Descriptor().FullName())]
	for _, f := range msg.fields {
		v, ok := fieldValue(m, f.desc)
		switch {
		case !ok || f.desc.IsMap():
		case f.Printed == "indexed":
			l := m.Mutable(f.desc).List()
			finite, zero := true, true
			for i := range l.Len() {
				n := floatOf(f.desc, l.Get(i))
				finite = finite && !math.IsNaN(n) && !math.IsInf(n, 0)
				zero = zero && n == 0
			}
			switch {
			case !finite:
			case zero:
				m.Clear(f.desc)
			default:
				for l.Len() < ix.enumLength(f.Index) {
					l.Append(l.NewElement())
				}
			}
		case f.desc.Message() != nil && f.desc.IsList():
			for i := range v.List().Len() {
				ix.printable(m.Mutable(f.desc).List().Get(i).Message())
			}
		case f.desc.Message() != nil:
			ix.printable(m.Mutable(f.desc).Message())
		}
	}
}

// A number field's value as a float.
func floatOf(fd protoreflect.FieldDescriptor, v protoreflect.Value) float64 {
	switch fd.Kind() {
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return v.Float()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return float64(v.Uint())
	}
	return float64(v.Int())
}

// The value of a placed field; false when it is unset.
func (in *printInput) valueOf(p *placed) (protoreflect.Value, bool) {
	m := readAt(in.settings, p.at)
	if p.encoded != nil {
		key := jsonPath(p.at, p.encoded)
		decoded, ok := in.decoded[key]
		if !ok {
			decoded = in.ix.decodedRotation(m.Get(p.encoded).String(), in.spec.Rotation)
			in.decoded[key] = decoded
		}
		m = decoded
	}
	return fieldValue(m, p.desc)
}

// A section as printed: its header, its statements, and whether it is written whatever the categories ask for.
type printedSection struct {
	def     *sectionDef
	header  string
	body    []*item
	written bool
}

func printFile(ix *schemaIndex, names *Names, kind Kind, spec *SpecInfo, s *proto.IndividualSimSettings, categories []string) string {
	settings := googleProto.Clone(s).ProtoReflect()
	ix.printable(settings)
	in := &printInput{ix: ix, settings: settings, spec: spec, decoded: map[string]protoreflect.Message{}}
	own := kind.Section
	allowed := fileSections(ix, kind)
	var rotation *proto.APLRotation
	hasRotation := readAt(settings, ix.rotationPath).IsValid()
	if hasRotation {
		rotation = readAt(settings, ix.rotationPath).Interface().(*proto.APLRotation)
	}
	var declarations []declared
	for _, d := range declaredBy(rotation).variables {
		if d.scope.inGroup && allowed[sectionGroup] || !d.scope.inGroup && allowed[sectionAPL] {
			declarations = append(declarations, d)
		}
	}
	p := &printer{ix: ix, names: names, declarations: declarations}
	var given map[string]bool
	if categories != nil {
		given = map[string]bool{}
		for _, c := range categories {
			given[c] = true
		}
	}
	// A section file writes its own section's body at column 1 with no header, except an x.group.siml: a group's
	// header names it.
	headless := func(def *sectionDef) bool { return def.Name == own && own != sectionGroup }
	indent := func(def *sectionDef) int {
		if headless(def) {
			return 0
		}
		return indentWidth
	}
	items := []*item{{line: "siml 1"}}
	if spec != nil && holdsSpec(ix, kind) {
		items = append(items, &item{line: "spec " + string(spec.Field.Name())})
		p.class = spec.Class
	}
	if talents := s.GetPlayer().GetTalentsString(); own == "" && (given != nil && given[talentsCategory] || given == nil && talents != "") {
		items = append(items, &item{line: "talents " + jsonString(talents)})
	}

	// A setup section's slot rows and entries. An x.gear.siml holds the slot rows alone.
	setup := func(def *sectionDef) *printedSection {
		fields := ix.sectionFields(def, spec)
		var slots *placed
		if slices.Contains(def.Body, "slot_rows") {
			slots = slotField(fields)
		}
		out := &printedSection{def: def, header: def.Name + ":"}
		if def.Presence != nil {
			at, ok := ix.specPath(def.Presence, spec)
			out.written = ok && readAt(settings, at).IsValid()
		}
		var writes []write
		for _, pl := range fields {
			v, ok := in.valueOf(pl)
			if !ok {
				continue
			}
			out.written = true
			if pl != slots {
				writes = append(writes, writesOf(pl, v)...)
				continue
			}
			for i := range v.List().Len() {
				slot := ix.enums[slotEnum].byNumber[int32(i)]
				itemSpec := v.List().Get(i).Message()
				if slot != nil && len(p.setFields(itemSpec)) > 0 {
					out.body = append(out.body, &item{stmt: p.slotRow(slot.Short, itemSpec, slots, i, indent(def), "")})
				}
			}
		}
		if own != sectionGear {
			for _, e := range p.entries(writes, indent(def)) {
				out.body = append(out.body, &item{stmt: e})
			}
		}
		return out
	}

	var sections []*printedSection
	simple := hasRotation && rotation.Simple != nil
	for _, def := range ix.sectionList {
		if !allowed[def.Name] || given != nil && def.Name != own && !given[def.Category] {
			continue
		}
		switch def.Name {
		case sectionAPL:
			out := &printedSection{def: def, header: def.Name + ":"}
			inferred := proto.APLRotation_TypeAPL
			if simple {
				inferred = proto.APLRotation_TypeSimple
			}
			if hasRotation && rotation.Type != inferred {
				fields := ix.sectionFields(def, spec)
				typeField := fields[slices.IndexFunc(fields, func(pl *placed) bool { return pl.Name == "type" })]
				typeValue, _ := fieldValue(rotation.ProtoReflect(), typeField.desc)
				for _, e := range p.entries([]write{{placed: typeField, values: []protoreflect.Value{typeValue}, indices: []int{-1}}}, indent(def)) {
					out.body = append(out.body, &item{stmt: e})
				}
			}
			for _, v := range rotation.GetValueVariables() {
				out.body = append(out.body, &item{stmt: p.variable(v.ProtoReflect(), scope{}, indent(def))})
			}
			rest := simple || len(rotation.GetPrepullActions()) > 0 || len(rotation.GetPriorityList()) > 0 || len(rotation.GetGroups()) > 0
			out.written = hasRotation && (len(out.body) > 0 || !rest)
			sections = append(sections, out)
		case sectionPrepull, sectionPriority:
			out := &printedSection{def: def, header: def.Name + ":"}
			if hasRotation {
				rows := rotation.ProtoReflect().Get(ix.sectionFields(def, spec)[0].desc).List()
				for i := range rows.Len() {
					out.body = append(out.body, &item{stmt: p.row(rows.Get(i).Message(), def.Name == sectionPrepull, scope{}, indent(def))})
				}
			}
			out.written = len(out.body) > 0
			sections = append(sections, out)
		case sectionGroup:
			for _, g := range rotation.GetGroups() {
				sc := scope{group: g.GetName(), inGroup: true}
				out := &printedSection{def: def, header: groupKey(g.GetName()) + ":", written: true}
				for _, v := range g.GetVariables() {
					out.body = append(out.body, &item{stmt: p.variable(v.ProtoReflect(), sc, indent(def))})
				}
				for _, r := range g.GetActions() {
					out.body = append(out.body, &item{stmt: p.row(r.ProtoReflect(), false, sc, indent(def))})
				}
				sections = append(sections, out)
			}
		default:
			sections = append(sections, setup(def))
		}
	}
	// A category asked for that no section holds gets its first section, empty, so that reading the file sets it.
	for _, c := range categories {
		first := slices.IndexFunc(sections, func(s *printedSection) bool { return s.def.Category == c })
		if first >= 0 && !slices.ContainsFunc(sections, func(s *printedSection) bool { return s.written && s.def.Category == c }) {
			sections[first].written = true
		}
	}

	for _, s := range sections {
		switch {
		case headless(s.def):
			if len(s.body) > 0 {
				items = append(append(items, blankLine()), s.body...)
			}
		case s.written:
			one := len(s.body) == 1 && s.body[0].stmt != nil && s.body[0].stmt.role == roleEntry
			items = append(append(items, blankLine(), &item{line: s.header, join: one}), s.body...)
		}
	}
	return verified(p, kind, items)
}

// A field's value as writes: one value, a list's values, or each message of a list as a write of its own.
func writesOf(p *placed, v protoreflect.Value) []write {
	if !p.isList() {
		return []write{{placed: p, values: []protoreflect.Value{v}, indices: []int{-1}}}
	}
	l := v.List()
	if p.Printed == "message" {
		out := make([]write, l.Len())
		for i := range out {
			out[i] = write{placed: p, values: []protoreflect.Value{l.Get(i)}, indices: []int{i}}
		}
		return out
	}
	w := write{placed: p}
	for i := range l.Len() {
		w.values = append(w.values, l.Get(i))
		w.indices = append(w.indices, i)
	}
	return []write{w}
}
