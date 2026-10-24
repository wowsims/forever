package siml

import (
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// What may go at the cursor: the statement's tokens up to the cursor read by a partial reader of the grammar
// (SPEC.md §4), which stops where the tokens run out and says what it expected there. The parser reads whole
// statements and reports errors; this reads a prefix and reports expectations.

// What is expected at the cursor; t says which, and the other fields what it is about.
type want struct {
	t       string
	column0 bool     // line: the cursor is at the start of its line.
	prepull bool     // step.
	when    bool     // step: a when line may start here.
	words   []string // words.
	ns      Namespace
	enum    string   // enum: the enum's name.
	message *message // args.
	kind    *aplKind // args.
	used    map[string]bool
	skip    map[string]bool
	head    string      // props.
	of      string      // strings: group or sequence.
	group   scope       // placeholders.
	def     *sectionDef // entries, flags.
	field   *field      // field.
	slot    string      // slot-item, slot-rest, enchants.
	item    string      // slot-rest, enchants; "" for none.
	path    string      // use-variables.
}

// Where the cursor is: what is expected there, the word being written and the text it replaces.
type cursorContext struct {
	wants   []want
	prefix  string // The text from the start of the word, or of the string, to the cursor.
	start   int    // Where the replaced text starts.
	end     int    // Where it ends: the cursor, or the end of a string the cursor is inside.
	string  bool   // The cursor is inside a string.
	section string // The section the statement belongs to; "" for none.
	group   scope  // The group, when the section is one.
	headers int    // How many headers come before the statement.
	// The parameters of the def whose value the cursor is in.
	params []string
}

var (
	contextOperators = []string{"==", "!=", ">=", ">", "<=", "<", "+", "-", "*", "/"}
	afterValue       = []string{"and", "or"}
	blankLineBetween = regexp.MustCompile(`\n[ \t]*\r?\n`)
)

// The tokens before the cursor ran out: what was expected there.
type contextEnd struct{ wants []want }

func (d *Doc) cursorAt(offset int) *cursorContext {
	tree := d.tree()
	if tree == nil {
		return nil
	}
	text := d.src
	if insideTrivia(text, offset) {
		return nil
	}
	statement := d.statementAt(offset)
	start := lineIndentEnd(text, offset)
	if statement != nil {
		start = statement.base().start
	}
	lineEnd := strings.IndexByte(text[offset:], '\n')
	if lineEnd < 0 {
		lineEnd = len(text)
	} else {
		lineEnd += offset
	}
	tokens, _, _ := lex(text[start:lineEnd])
	for i := range tokens {
		tokens[i].start += start
		tokens[i].end += start
	}
	var at *token
	for i := range tokens {
		t := &tokens[i]
		if t.start < offset && offset <= t.end && t.kind != tokPunct && t.kind != tokBad {
			at = t
			break
		}
	}
	upTo := offset
	if at != nil {
		upTo = at.start
	}
	var before, after []token
	for _, t := range tokens {
		if t.end <= upTo {
			before = append(before, t)
		}
		if t.start >= offset {
			after = append(after, t)
		}
	}
	ctx := &cursorContext{start: offset, end: offset}
	ctx.section, ctx.group, ctx.headers = d.sectionAt(statement, start)
	if at != nil {
		ctx.prefix, ctx.start = text[at.start:offset], at.start
		if at.kind == tokString {
			ctx.end, ctx.string = at.end, true
		}
		// A ${…} or %{…} name is replaced up to its }.
		if braced := (at.kind == tokVar || at.kind == tokPlaceholder) && strings.HasPrefix(at.text[1:], "{"); braced && (at.err == nil || at.err.code != "expected-name") {
			ctx.end = at.end
		}
	}
	column0 := upTo == lineStartOf(text, offset) && len(before) == 0
	ctx.wants, ctx.params = d.readContext(before, after, ctx.section, column0)
	has := func(ts ...string) bool {
		return slices.ContainsFunc(ctx.wants, func(w want) bool { return slices.Contains(ts, w.t) })
	}
	// A variable or placeholder being written is one wherever it stands.
	if at != nil && at.kind == tokVar && has("value", "time", "ref") {
		ctx.wants = []want{{t: "variables"}}
	}
	if at != nil && at.kind == tokPlaceholder {
		i := slices.IndexFunc(ctx.wants, func(w want) bool { return w.t == "placeholders" })
		switch {
		case i >= 0:
			ctx.wants = []want{ctx.wants[i]}
		case has("value", "time"):
			ctx.wants = []want{{t: "placeholders", group: ctx.group}}
		default:
			ctx.wants = nil
		}
	}
	if at != nil && at.kind == tokVar && !has("variables", "use-variables") {
		ctx.wants = nil
	}
	return ctx
}

// Runs the reader over the tokens before the cursor and gives what it expected where they ran out, with the
// parameters of the def whose value they ran out in.
func (d *Doc) readContext(before, after []token, section string, column0 bool) (wants []want, params []string) {
	r := &contextReader{d: d, ix: d.ix, ts: before, after: after, section: section}
	defer func() {
		if x := recover(); x != nil {
			end, ok := x.(contextEnd)
			if !ok {
				panic(x)
			}
			wants, params = end.wants, r.params
		}
	}()
	r.statement(column0)
	return nil, nil
}

func lineIndentEnd(text string, offset int) int {
	i := lineStartOf(text, offset)
	for i < offset && (text[i] == ' ' || text[i] == '\t') {
		i++
	}
	return i
}

// Whether the cursor is in a `//` comment or a `///` note.
func insideTrivia(text string, offset int) bool {
	_, triv, _ := lex(text[lineStartOf(text, offset):offset])
	return slices.ContainsFunc(triv, func(t trivia) bool { return t.kind == triviaComment || t.kind == triviaNote })
}

// The statement the cursor is in: one it lies inside, or the last one before it that it continues, on its last line
// or on a deeper-indented line below with no blank line between.
func (d *Doc) statementAt(offset int) stmt {
	text := d.src
	var last stmt
	for _, s := range d.tree().statements {
		if s.base().start >= offset {
			break
		}
		last = s
	}
	if last == nil {
		return nil
	}
	b := last.base()
	if b.holds(offset) {
		return last
	}
	between := text[b.end:offset]
	if !strings.Contains(between, "\n") {
		return last
	}
	if _, header := last.(*headerStmt); header || blankLineBetween.MatchString(between) {
		return nil
	}
	indentOf := func(at int) int {
		from := lineStartOf(text, at)
		i := from
		for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
			i++
		}
		return i - from
	}
	if min(indentOf(offset), offset-lineStartOf(text, offset)) > indentOf(b.start) {
		return last
	}
	return nil
}

// The section a statement belongs to: the last header at or before it, else the section a section file implies.
func (d *Doc) sectionAt(statement stmt, start int) (name string, group scope, headers int) {
	tree := d.tree()
	if kind := d.Kind().Section; kind != sectionGroup {
		name = kind
	}
	for _, s := range tree.sections {
		if s.header != nil && s.header.start <= start {
			name, group, headers = s.name, scopeOf(s), headers+1
		}
	}
	if h, ok := statement.(*headerStmt); ok {
		for _, s := range tree.sections {
			if s.header == h {
				name, group = s.name, scopeOf(s)
			}
		}
	}
	return name, group, headers
}

type contextReader struct {
	d       *Doc
	ix      *schemaIndex
	i       int
	ts      []token // The statement's tokens before the word at the cursor.
	after   []token // The tokens after the cursor, which only say what an accessor goes on to.
	section string
	params  []string // The parameters of the def line being read.
}

func (r *contextReader) tok() *token { return r.peek(0) }

func (r *contextReader) peek(k int) *token {
	if r.i+k < len(r.ts) {
		return &r.ts[r.i+k]
	}
	return nil
}

func isText(t *token, text string) bool {
	return t != nil && (t.kind == tokPunct || t.kind == tokIdent) && t.text == text
}

func (r *contextReader) is(text string) bool { return isText(r.tok(), text) }

// Throws what is expected when the tokens have run out.
func (r *contextReader) need(wants ...want) {
	if r.i >= len(r.ts) {
		panic(contextEnd{wants})
	}
}

// The tokens before the cursor do not read: nothing is offered.
func (r *contextReader) fail() { panic(contextEnd{}) }

func (r *contextReader) take(text ...string) token {
	t := r.tok()
	if t == nil || len(text) > 0 && !r.is(text[0]) {
		r.fail()
	}
	r.i++
	return *t
}

func (r *contextReader) statement(column0 bool) {
	def := r.ix.sections[r.section]
	if r.i >= len(r.ts) {
		panic(contextEnd{append([]want{{t: "line", column0: column0}}, r.bodyStart(def)...)})
	}
	a, b, c := r.tok(), r.peek(1), r.peek(2)
	// A header, and the body line written after its colon.
	if a.kind == tokIdent && isText(b, ":") {
		r.i += 2
		r.body(r.ix.sections[a.text])
		return
	}
	if isText(a, sectionGroup) && b != nil && b.kind == tokString && isText(c, ":") {
		r.i += 3
		r.body(r.ix.sections[sectionGroup])
		return
	}
	if a.kind == tokIdent && topWords[a.text] && !isText(b, "=") && (a.text != "var" || r.allowsVar(def)) &&
		(a.text != "def" || def == nil || def.Name == sectionAPL) {
		r.top()
		return
	}
	if def == nil {
		r.fail()
	}
	r.body(def)
}

func (r *contextReader) allowsVar(def *sectionDef) bool {
	return def == nil || def.Name == sectionAPL || def.Name == sectionGroup || r.section == ""
}

// What a body line of the section starts with.
func (r *contextReader) bodyStart(def *sectionDef) []want {
	switch {
	case def == nil:
		return nil
	case slices.Contains(def.Body, "rows") && def.Name == sectionPrepull:
		return []want{{t: "time"}}
	case slices.Contains(def.Body, "rows"):
		return []want{{t: "step", when: true}}
	}
	return []want{{t: "entries", def: def}}
}

func (r *contextReader) body(def *sectionDef) {
	if def == nil {
		r.fail()
	}
	r.need(r.bodyStart(def)...)
	switch {
	case slices.Contains(def.Body, "rows") && r.is("var"):
		r.varLine()
	case slices.Contains(def.Body, "rows"):
		r.row(def.Name == sectionPrepull)
	default:
		r.entry(def)
	}
}

// --- Top-level lines.

func (r *contextReader) top() {
	switch r.take().text {
	case "siml":
		r.need(want{t: "words", words: []string{"1"}})
	case "spec":
		r.need(want{t: "specs"})
	case "talents":
		r.need()
	case "use":
		r.need(want{t: "use-path"})
		path := r.take().value
		r.need(want{t: "words", words: []string{"with"}})
		r.take("with")
		for {
			r.need(want{t: "use-variables", path: path})
			r.take()
			r.need()
			r.take("=")
			r.expr(nil)
			r.take(",")
		}
	case "var":
		r.i--
		r.varLine()
	case "def":
		r.need()
		r.take()
		if r.is("(") {
			r.i++
			for r.need(); !r.is(")"); r.need() {
				if t := r.take(); t.kind == tokIdent {
					r.params = append(r.params, t.text)
				}
			}
			r.i++
		}
		r.need(want{t: "words", words: []string{"="}})
		r.take("=")
		r.expr(nil)
	}
}

func (r *contextReader) varLine() {
	r.take("var")
	r.need()
	r.take()
	r.need()
	r.take("=")
	r.expr(nil)
}

// --- Rows and actions.

func (r *contextReader) row(prepull bool) {
	if r.is("#") {
		r.i++
		if !prepull {
			r.need(want{t: "step", when: true})
		}
	}
	if !prepull && r.is("when") {
		r.i++
		r.expr([]string{":"})
		return
	}
	if prepull {
		r.need(want{t: "time"})
		r.unary()
	}
	r.action(prepull)
}

func (r *contextReader) action(prepull bool) {
	trailing := r.step(prepull)
	r.need(want{t: "words", words: append(trailing, "if")})
	if r.is("if") {
		r.i++
		r.expr(nil)
	}
}

// Reads a step, and gives the words that may still follow it before its `if`.
func (r *contextReader) step(prepull bool) []string {
	r.need(want{t: "step", prepull: prepull})
	t := *r.tok()
	if t.kind == tokRaw {
		r.i++
		return nil
	}
	if t.kind != tokIdent {
		r.fail()
	}
	r.i++
	switch t.text {
	case "cast", "channel":
		r.ref(NamespaceSpell)
		left := []string{"on"}
		if t.text == "channel" {
			left = []string{"on", "until", "recast"}
		}
		for {
			r.need(want{t: "words", words: append(slices.Clone(left), "if")})
			next := r.tok()
			if !slices.Contains(left, next.text) || next.kind != tokIdent {
				return left
			}
			r.i++
			left = slices.DeleteFunc(left, func(w string) bool { return w == next.text })
			switch next.text {
			case "on":
				r.need(want{t: "unit"})
				r.i++
			case "until":
				r.expr(append(slices.Clone(left), "if"))
			}
		}
	case "wait", "move":
		word := "for"
		if t.text == "wait" {
			word = "until"
		}
		r.need(want{t: "value"}, want{t: "words", words: []string{word}})
		after := []string{"if"}
		if r.is(word) {
			r.i++
		} else if t.text == "move" {
			after = []string{"recheck", "if"}
		}
		r.expr(after)
		if len(after) > 1 && r.is("recheck") {
			r.i++
			r.need(want{t: "words", words: []string{"if"}})
		}
		return nil
	case "group":
		r.need(want{t: "strings", of: "group"})
		name := r.take()
		if name.kind != tokString {
			r.fail()
		}
		r.need(want{t: "words", words: []string{"with", "if"}})
		if !r.is("with") {
			return []string{"with"}
		}
		r.i++
		for {
			r.need(want{t: "placeholders", group: scope{group: name.value, inGroup: true}})
			r.take()
			r.need()
			r.take("=")
			r.expr([]string{"if"})
			if next := r.peek(1); !r.is(",") || next == nil || next.kind != tokPlaceholder {
				return nil
			}
			r.i++
		}
	case "schedule":
		if t := r.tok(); t == nil || t.kind != tokString {
			return r.call("schedule", true)
		}
		r.i++
		return r.step(prepull)
	case "sequence":
		if t := r.tok(); t == nil || t.kind != tokString {
			return r.call("sequence", true)
		}
		r.i++
		r.need()
		r.take("[")
		for {
			if r.is("]") {
				r.i++
				return nil
			}
			r.action(false)
			r.need()
			if r.is("]") {
				continue
			}
			r.take(",")
		}
	case "none":
		return nil
	}
	return r.call(t.text, true)
}

// A kind written by name, with its arguments when a `(` follows.
func (r *contextReader) call(name string, action bool) []string {
	kinds := r.ix.valueKinds
	if action {
		kinds = r.ix.actionKinds
	}
	kind := kinds[name]
	if r.is("(") {
		if kind == nil {
			r.fail()
		}
		r.args(kind.message, kind, nil, map[string]bool{})
	}
	return nil
}

// --- Expressions.

// An expression, and the words `after` the caller takes once it is complete.
func (r *contextReader) expr(after []string) {
	for {
		r.unary()
		r.need(want{t: "words", words: slices.Concat(afterValue, after)})
		t := r.tok()
		if t.kind == tokPunct && slices.Contains(contextOperators, t.text) || r.is("and") || r.is("or") {
			r.i++
			continue
		}
		return
	}
}

func (r *contextReader) unary() {
	r.need(want{t: "value"})
	if next := r.peek(1); r.is("not") && !(isText(next, "(") && next.start == r.tok().end) {
		r.i++
		r.unary()
		return
	}
	if r.is("-") {
		r.i++
		r.need()
	}
	if r.is("(") {
		r.i++
		r.expr(nil)
		r.need()
		r.take(")")
		return
	}
	r.operand()
}

func (r *contextReader) operand() {
	t, next := r.tok(), r.peek(1)
	switch t.kind {
	case tokNumber, tokString, tokRaw, tokVar, tokPlaceholder:
		r.i++
		return
	}
	if r.is("[") {
		r.skipTerm()
		return
	}
	if t.kind != tokIdent {
		r.fail()
	}
	word := t.text
	switch {
	case word == "true" || word == "false" || word == "none":
		r.i++
	case (word == "item" || word == "other") && isText(next, "."):
		r.i += 2
		if word == "item" {
			r.need(want{t: "usable"})
		} else {
			r.need(want{t: "other"})
		}
		r.i++
	case isText(next, ".") && r.ix.unitReference(word) != nil:
		r.i += 2
		r.need(want{t: "member"})
		member := r.take()
		if slices.Contains(r.ix.heads, member.text) && r.is("(") {
			r.accessor(member.text, word)
		}
	case slices.Contains(r.ix.heads, word) && isText(next, "("):
		r.i++
		r.accessor(word, "")
	case (word == "and" || word == "or" || word == "not") && isText(next, "(") && next.start == t.end:
		kind := r.ix.valueKinds[word]
		r.i++
		r.args(kind.message, kind, nil, map[string]bool{})
	default:
		r.i++
		r.call(word, false)
	}
}

// The accessor kinds of a head, in APLValue's declaration order.
func (ix *schemaIndex) headKinds(head string) []*aplKind {
	var out []*aplKind
	for _, k := range ix.valueKindList {
		if k.Accessor != nil && k.Accessor.Head == head {
			out = append(out, k)
		}
	}
	return out
}

// `head(subject).prop(args)`, after the head; unitWord is the unit prefix, "" for none.
func (r *contextReader) accessor(head, unitWord string) {
	r.take("(")
	kinds := r.ix.headKinds(head)
	prop := r.lookaheadProp()
	var kind *aplKind
	for _, k := range kinds {
		if k.Accessor.Prop == prop {
			kind = k
			break
		}
	}
	if kind == nil && len(kinds) > 0 {
		kind = kinds[0]
	}
	if kind == nil {
		r.fail()
	}
	subject := kind.message.byProto[kind.Accessor.Subject]
	if subject.Printed == "string" {
		r.need(want{t: "strings", of: "sequence"})
		r.i++
	} else {
		var unit *proto.UnitReference
		if unitWord != "" {
			unit = r.ix.unitReference(unitWord)
		} else {
			unit = r.lookaheadUnit(kind)
		}
		r.ref(fieldNamespace(kind, subject, unit))
	}
	r.need()
	r.take(")")
	r.need()
	r.take(".")
	r.need(want{t: "props", head: head})
	propKind := r.ix.accessors[head+"."+r.take().text]
	if !r.is("(") {
		return
	}
	if propKind == nil {
		r.fail()
	}
	skip := map[string]bool{propKind.Accessor.Subject: true}
	var unit *proto.UnitReference
	if unitWord != "" {
		unit = r.ix.unitReference(unitWord)
		if propKind.Accessor.Unit != "" {
			skip[propKind.Accessor.Unit] = true
		}
	}
	r.args(propKind.message, propKind, unit, skip)
}

// The namespace an ActionID field of a kind reads, unit being what the kind's unit field holds.
func fieldNamespace(kind *aplKind, f *field, unit *proto.UnitReference) Namespace {
	return kindScope{kind: kind, unit: unit.GetType()}.namespace(f)
}

// The property an accessor whose subject the cursor is in goes on to, read from the text after the cursor.
func (r *contextReader) lookaheadProp() string {
	close := closingBracket(r.after, 0)
	if close < 0 || close+2 >= len(r.after) {
		return ""
	}
	if dot, prop := r.after[close+1], r.after[close+2]; dot.text == "." && prop.kind == tokIdent {
		return prop.text
	}
	return ""
}

// The unit an accessor's property argument gives, read from the text after the cursor: an aura's namespace follows
// its unit, written before or after the aura (SPEC.md §7.2).
func (r *contextReader) lookaheadUnit(kind *aplKind) *proto.UnitReference {
	unitField := kind.unitField()
	close := closingBracket(r.after, 0)
	if unitField == nil || close < 0 || close+3 >= len(r.after) || r.after[close+3].text != "(" {
		return nil
	}
	end := closingBracket(r.after, close+4)
	if end < 0 {
		end = len(r.after)
	}
	for k := close + 4; k+2 < end; k++ {
		name, eq, value := r.after[k], r.after[k+1], r.after[k+2]
		if name.kind == tokIdent && (name.text == unitField.Siml || slices.Contains(unitField.Aliases, name.text)) && eq.text == "=" && value.kind == tokIdent {
			return r.ix.unitReference(value.text)
		}
	}
	return nil
}

// The index in tokens of the `)` that closes the bracket open before from, or -1.
func closingBracket(tokens []token, from int) int {
	depth := 0
	for k := from; k < len(tokens); k++ {
		if tokens[k].kind != tokPunct {
			continue
		}
		switch tokens[k].text {
		case "(", "[":
			depth++
		case ")", "]":
			if depth == 0 {
				if tokens[k].text == ")" {
					return k
				}
				return -1
			}
			depth--
		}
	}
	return -1
}

// A ref: a name of ns, `item.<name>`, `other.<short>`, `none` or raw JSON.
func (r *contextReader) ref(ns Namespace) {
	r.need(want{t: "ref", ns: ns})
	if t := r.tok(); (t.text == "item" || t.text == "other") && isText(r.peek(1), ".") {
		r.i += 2
		if t.text == "item" {
			r.need(want{t: "usable"})
		} else {
			r.need(want{t: "other"})
		}
	}
	r.i++
}

// The arguments of a kind or a message, from its `(`. skip holds fields the form fills elsewhere: an accessor's
// subject and the unit its prefix gives.
func (r *contextReader) args(m *message, kind *aplKind, unit *proto.UnitReference, skip map[string]bool) {
	r.take("(")
	used := map[string]bool{}
	var list, positional *field
	if kind != nil && kind.ListField != "" {
		list = m.byProto[kind.ListField]
	}
	positional = list
	if positional == nil && kind != nil && kind.Positional != "" && !skip[kind.Positional] {
		positional = m.byProto[kind.Positional]
	}
	for {
		wants := []want{{t: "args", message: m, kind: kind, used: used, skip: skip}}
		if positional != nil && !used[positional.Name] {
			wants = append(wants, r.fieldWants(positional, kind, unit)...)
		}
		r.need(wants...)
		if r.is(")") {
			r.i++
			return
		}
		t, next := r.tok(), r.peek(1)
		var named, bare, nested *field
		if t.kind == tokIdent {
			switch {
			case isText(next, "="):
				named = m.byName[t.text]
			case isText(next, "("):
				nested = m.byName[t.text]
			}
			if !isText(next, "(") {
				bare = m.byName[t.text]
			}
		}
		switch {
		case named != nil:
			r.i += 2
			r.need(r.fieldWants(named, kind, unit)...)
			r.term(named, kind, unit)
			used[named.Name] = true
			if last := r.ts[r.i-1]; named.Printed == "unit" && last.kind == tokIdent {
				if u := r.ix.unitReference(last.text); u != nil {
					unit = u
				}
			}
		case bare != nil && bare.Printed == "bool":
			r.i++
			used[bare.Name] = true
		case nested != nil && (nested.Printed == "message" || nested.Printed == "indexed"):
			r.i++
			r.messageValue(nested)
			if !nested.Repeated {
				used[nested.Name] = true
			}
		case positional != nil:
			r.term(positional, kind, unit)
			if positional != list {
				used[positional.Name] = true
				positional = nil
			}
		default:
			r.fail()
		}
		r.need()
		if r.is(")") {
			continue
		}
		r.take(",")
	}
}

// A message field's value written `field(...)`, after the field's name.
func (r *contextReader) messageValue(f *field) {
	if f.Printed == "indexed" {
		r.take("(")
		for {
			r.need(want{t: "enum", enum: f.Index})
			if r.is(")") {
				r.i++
				return
			}
			r.take()
			r.need()
			r.take("=")
			r.need()
			r.skipTerm()
			r.need()
			if r.is(")") {
				continue
			}
			r.take(",")
		}
	}
	r.args(r.ix.messages[f.Type], nil, nil, map[string]bool{})
}

// What a value of f may be.
func (r *contextReader) fieldWants(f *field, kind *aplKind, unit *proto.UnitReference) []want {
	switch f.Printed {
	case "value":
		return []want{{t: "value"}}
	case "action":
		return []want{{t: "step"}}
	case "action_id":
		return []want{{t: "ref", ns: fieldNamespace(kind, f, unit)}}
	case "unit":
		return []want{{t: "unit"}}
	case "enum", "flag_enum":
		return []want{{t: "enum", enum: f.Type}}
	case "stat":
		return []want{{t: "enum", enum: statEnum}}
	case "bool":
		return []want{{t: "bool"}}
	case "ref", "message":
		return []want{{t: "field", field: f}}
	case "string":
		if of := r.ix.namedBy(kind, f.Name); of != "" {
			return []want{{t: "strings", of: of}}
		}
	}
	return nil
}

// A term of f: a list in brackets for a list field, else one value.
func (r *contextReader) term(f *field, kind *aplKind, unit *proto.UnitReference) {
	if f.isList() && r.is("[") {
		r.i++
		for {
			r.need(r.fieldWants(f, kind, unit)...)
			if r.is("]") {
				r.i++
				return
			}
			r.single(f, kind, unit)
			r.need()
			if r.is("]") {
				continue
			}
			r.take(",")
		}
	}
	r.single(f, kind, unit)
}

func (r *contextReader) single(f *field, kind *aplKind, unit *proto.UnitReference) {
	switch f.Printed {
	case "value":
		r.expr(nil)
	case "action":
		r.action(false)
	case "action_id":
		r.ref(fieldNamespace(kind, f, unit))
	case "message", "indexed":
		if t := r.tok(); t != nil && t.kind == tokIdent && isText(r.peek(1), "(") {
			r.i++
			r.messageValue(f)
			return
		}
		r.skipTerm()
	default:
		r.skipTerm()
	}
}

// Steps over a term whose parts offer nothing: a value, with its arguments or brackets, up to the `,` or closing
// bracket that ends it.
func (r *contextReader) skipTerm() {
	depth := 0
	for {
		r.need()
		if depth == 0 && (r.is(",") || r.is(")") || r.is("]")) {
			return
		}
		t := r.take()
		punct := t.kind == tokPunct
		switch {
		case punct && (t.text == "(" || t.text == "["):
			depth++
		case punct && (t.text == ")" || t.text == "]"):
			depth--
		}
		if depth > 0 || punct && (t.text == "-" || t.text == ".") {
			continue
		}
		if r.is("(") || r.is("[") || r.is(".") {
			continue
		}
		return
	}
}

// --- Entries and slot rows.

func (r *contextReader) entry(def *sectionDef) {
	t := r.take()
	if t.kind != tokIdent {
		r.fail()
	}
	if slices.Contains(def.Body, "slot_rows") && r.ix.slots[t.text] {
		r.slotRow(t.text)
		return
	}
	if t.text == "var" && slices.Contains(def.Body, "var") {
		r.i--
		r.varLine()
		return
	}
	f := r.entryField(def, t.text)
	if f == nil {
		r.fail()
	}
	if r.is("(") && (f.Printed == "message" || f.Printed == "indexed") {
		r.messageValue(f)
		return
	}
	if r.is("=") || r.is(",") {
		r.flags(def, f)
		return
	}
	for {
		r.need(r.fieldWants(f, nil, nil)...)
		r.single(f, nil, nil)
		r.need()
		r.take(",")
	}
}

// A flag list, after its first flag's name.
func (r *contextReader) flags(def *sectionDef, first *field) {
	f := first
	for {
		if r.is("=") {
			r.i++
			r.need(r.fieldWants(f, nil, nil)...)
			r.single(f, nil, nil)
		}
		r.need()
		r.take(",")
		r.need(want{t: "flags", def: def})
		if f = r.entryField(def, r.take().text); f == nil {
			r.fail()
		}
	}
}

func (r *contextReader) entryField(def *sectionDef, name string) *field {
	return entryField(r.ix, def, r.d.specInfo(), name)
}

// The field of a section an entry's written name sets, nil for none.
func entryField(ix *schemaIndex, def *sectionDef, spec *SpecInfo, name string) *field {
	for _, p := range ix.sectionFields(def, spec) {
		if p.Siml == name || slices.Contains(p.Aliases, name) {
			return p.field
		}
	}
	return nil
}

func (r *contextReader) slotRow(slot string) {
	r.need(want{t: "slot-item", slot: slot})
	item := r.take()
	if item.kind != tokIdent {
		r.fail()
	}
	r.need(want{t: "slot-rest", slot: slot, item: item.text})
	if r.tok().kind == tokIdent {
		r.i++
		r.need(want{t: "words", words: []string{"+"}})
	}
	r.take("+")
	r.need(want{t: "enchants", slot: slot, item: item.text})
	r.fail()
}

// The kinds whose string field names a group, by that field's proto name.
var groupFields = map[string]string{"group_reference": "group_name", "action_group_used": "name"}

// What a kind's string field names: a group, a sequence (a field named like the sequence head's subject, SPEC.md
// §7.5), or nothing the file defines ("").
func (ix *schemaIndex) namedBy(kind *aplKind, fieldName string) string {
	if kind == nil || fieldName == "" {
		return ""
	}
	if groupFields[kind.Kind] == fieldName {
		return "group"
	}
	sequence := slices.ContainsFunc(ix.headKinds("sequence"), func(k *aplKind) bool { return k.Accessor.Subject == fieldName })
	if f := kind.message.byProto[fieldName]; sequence && f != nil && f.Printed == "string" {
		return "sequence"
	}
	return ""
}
