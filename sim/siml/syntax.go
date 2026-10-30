package siml

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strconv"
	"unicode/utf8"
)

// SIML text into its syntax tree (SPEC.md §3, §4, §10): lines into statements, statements into sections, and
// each statement read on its own, so a bad one costs only itself. Typing gives the nodes their meaning.

// A node of an expression, step, term or argument list. Its concrete type says what it is.
type node interface{ pos() span }

type name struct {
	span
	text string
}

// A number with an optional unit, its text exactly as written, a leading minus included (-22s, 5.00).
type literalNode struct {
	span
	text string
}

type stringNode struct {
	span
	value string
}

type boolNode struct {
	span
	value bool
}

type noneNode struct{ span }

// `$name`, or `${Name}` when braced.
type varNode struct {
	span
	name   string
	braced bool
}

// `%name`, or `%{Name}` when braced.
type placeholderNode struct {
	span
	name   string
	braced bool
}

// Backticked JSON; json is the text between the backticks, which the parser has checked is JSON.
type rawNode struct {
	span
	json string
}

// Kept, because parentheses keep an `and` inside an `and` (or an `or` inside an `or`) apart.
type parenNode struct {
	span
	expr node
}

type notNode struct {
	span
	operand node
}

// `a and b and c`: one node for the whole run; op is "and" or "or".
type logicNode struct {
	span
	op       string
	operands []node
}

type compareNode struct {
	span
	op          string
	left, right node
}

type mathNode struct {
	span
	op          string
	left, right node
}

// `name` or `name(args)`: a kind, or a name, unit word or enum value, as its position types it.
type callNode struct {
	span
	name name
	args *argList // Nil when no parentheses follow the name.
}

// `object.name`: a unit-prefixed kind (target.current_health), item.<name> or other.<value>.
type memberNode struct {
	span
	object, name name
}

// `[unit.]head(subject).prop[(args)]`.
type accessorNode struct {
	span
	unit    *name
	head    name
	subject node
	prop    name
	args    *argList
}

// The parentheses and what they hold.
type argList struct {
	span
	items []arg
}

type arg struct {
	span
	name  *name
	value node
}

type listNode struct {
	span
	items []node
}

type actionNode struct {
	span
	step      node
	condition node
}

type castStep struct {
	span
	spell, target node
}

type channelStep struct {
	span
	spell, target, until node
	recast               bool
}

// `wait v`, or `wait until v`.
type waitStep struct {
	span
	until bool
	value node
}

// `move v`, or `move for v`.
type moveStep struct {
	span
	duration bool
	value    node
	recheck  bool
}

type placeholderBind struct {
	span
	placeholder *placeholderNode
	value       node
}

type groupStep struct {
	span
	name *stringNode
	with []placeholderBind
}

// The inner step takes no `if`; the row's `if` is the schedule's.
type scheduleStep struct {
	span
	schedule *stringNode
	step     node
}

type sequenceStep struct {
	span
	name    *stringNode
	actions []*actionNode
}

type varBind struct {
	span
	variable *varNode
	value    node
}

type stmt interface{ base() *stmtBase }

type stmtBase struct {
	span
	leading []trivia // The blank, comment and note lines between the previous statement and this one.
	inner   []trivia // The trivia on this statement's lines: trailing comments, and lines inside a continuation.
}

func (b *stmtBase) base() *stmtBase { return b }

type versionStmt struct {
	stmtBase
	version literalNode
}

type specStmt struct {
	stmtBase
	spec name
}

type talentsStmt struct {
	stmtBase
	talents stringNode
}

type useStmt struct {
	stmtBase
	path stringNode
	with []varBind
}

// `name:` or `group "Name":`, with the body line written after the colon, if any.
type headerStmt struct {
	stmtBase
	name  name
	group *stringNode
	body  stmt
}

type varStmt struct {
	stmtBase
	variable *varNode
	value    node // nil for a variable with no value.
}

type rowStmt struct {
	stmtBase
	notes  []trivia
	hide   bool
	doAt   node
	action *actionNode
	when   *whenStmt // The innermost when block it is in, if any.
}

// `when <expr>:`, whose body is the statements after it indented deeper than it (SPEC.md §6.9).
type whenStmt struct {
	stmtBase
	hide      bool
	condition node
	parent    *whenStmt // The block it is in, if any.
	body      []stmt    // Its own statements, a nested block's line among them.
}

// `def name(p, …) = value`: a macro each call reads as its value (SPEC.md §6.10).
type defStmt struct {
	stmtBase
	name   name
	params []name
	value  node
}

type slotRowStmt struct {
	stmtBase
	slot    name
	item    *name    // Nil when the row is raw ItemSpec JSON.
	raw     *rawNode // The raw ItemSpec in place of the rest of the row.
	suffix  *name
	enchant *name
}

// `name` alone is a flag; `name=value` sets a flag-value enum to another value.
type flag struct {
	span
	name  name
	value node
}

type flagListStmt struct {
	stmtBase
	flags []flag
}

type fieldEntryStmt struct {
	stmtBase
	name   name
	values []node
}

type messageEntryStmt struct {
	stmtBase
	name name
	args *argList
}

type rawEntryStmt struct {
	stmtBase
	value *rawNode
}

// A statement that could not be read, or one under a header that names no section. Its diagnostic, if any, is
// in the parse result.
type unparsedStmt struct{ stmtBase }

// A section, from its header to the next header. The section a section file implies has no header until the first
// written header of it joins it (SPEC.md §1). body holds its statements in order, a header's own body line first.
type section struct {
	name   string
	group  *stringNode
	header *headerStmt
	body   []stmt
}

type syntaxTree struct {
	statements []stmt     // Every line's statement in file order; a header's body line sits in the header.
	top        []stmt     // siml, spec, talents and use, and the var and def lines of a full file before any header, and those of its lines there that do not read.
	sections   []*section // In file order.
	trailing   []trivia   // The trivia after the last statement.
	lines      *lineMap
	blanks     map[int]bool // The 0-based lines that are blank.
}

// Every statement in file order, a header's body line right after its header.
func (t *syntaxTree) inFileOrder() []stmt {
	out := make([]stmt, 0, len(t.statements))
	for _, s := range t.statements {
		out = append(out, s)
		if h, ok := s.(*headerStmt); ok && h.body != nil {
			out = append(out, h.body)
		}
	}
	return out
}

const (
	sectionAPL      = "apl"
	sectionGear     = "gear"
	sectionGroup    = "group"
	sectionPrepull  = "prepull" // The rows section whose rows start with their time and take no notes.
	sectionPriority = "priority"
	sectionSimple   = "simple"
)

// Words that end an expression or start an action, and so never read as a name or value.
var stopWords = map[string]bool{
	"if": true, "on": true, "until": true, "recast": true, "recheck": true, "for": true, "with": true, "and": true, "or": true,
	"not": true, "cast": true, "channel": true, "wait": true, "move": true, "group": true, "when": true,
}

// Words that start a top-level line, a var line or a def line, unless `=` follows, which makes the line an entry.
var topWords = map[string]bool{"siml": true, "spec": true, "talents": true, "use": true, "var": true, "def": true}

var (
	compareOps = []string{"==", "!=", ">=", ">", "<=", "<"}
	argEnds    = []string{",", ")"}
	listEnds   = []string{",", "]"}
	entryEnds  = []string{","}
)

func join(a, b node) span { return span{a.pos().start, b.pos().end} }

func isInt(s string) bool {
	if s == "" {
		return false
	}
	for i := range len(s) {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}

func parseSyntax(src string, ix *schemaIndex, kind Kind) (*syntaxTree, []diag) {
	tokens, triv, lines := lex(src)
	blanks := blankLines(triv)
	f := &fileParser{
		ix: ix, kind: kind, tokens: tokens, trivia: triv, lines: lines,
		allowed: fileSections(ix, kind), tree: &syntaxTree{lines: lines, blanks: blanks}, present: map[string]bool{},
	}
	return f.run()
}

// A fatal error in a statement: the statement is left unread.
type failure struct {
	code, message string
	span
	atEnd bool // At the end of the statement's readable tokens rather than at a token.
}

// The tokens of each statement (SPEC.md §3.1). A statement continues while a bracket is open, and, unless it is a
// header, onto a later line indented deeper than its first, with no blank line between, that does not itself start
// with a header; comment lines never decide. A when line ends at the first : that ends a line.
func statementsOf(tokens []token, blank map[int]bool, lines *lineMap) [][]token {
	var out [][]token
	i := 0
	for i < len(tokens) {
		indent := lines.column(tokens[i].start)
		header, when := headerShape(tokens, i, lines), whenShape(tokens, i)
		var open []int
		j := i
		for ; ; j++ {
			t := tokens[j]
			if t.kind == tokPunct && (t.text == "(" || t.text == "[") {
				open = append(open, j)
			} else if t.kind == tokPunct && (t.text == ")" || t.text == "]") && len(open) > 0 {
				open = open[:len(open)-1]
			}
			if j+1 >= len(tokens) {
				break
			}
			next := tokens[j+1]
			endLine := lines.line(t.end - 1)
			nextLine := lines.line(next.start)
			if nextLine == endLine || len(open) > 0 {
				continue
			}
			if when && t.kind == tokPunct && t.text == ":" {
				break
			}
			gap := false
			for l := endLine + 1; l < nextLine && !gap; l++ {
				gap = blank[l]
			}
			if gap || header || next.start-lines.starts[nextLine] <= indent || headerShape(tokens, j+1, lines) {
				break
			}
		}
		if len(open) > 0 {
			t := &tokens[open[0]]
			t.err = &tokenError{"unclosed-bracket", fmt.Sprintf("this '%s' is never closed", t.text), span{t.start, t.end}}
		}
		out = append(out, tokens[i:j+1])
		i = j + 1
	}
	return out
}

// The lines of the blank trivia.
func blankLines(triv []trivia) map[int]bool {
	out := map[int]bool{}
	for _, t := range triv {
		if t.kind == triviaBlank {
			out[t.line] = true
		}
	}
	return out
}

// `name:` or `group "Name":` on the statement's first line. A header-shaped statement is a header wherever it
// starts; one that does not start at column 1 is indented-header.
func headerShape(tokens []token, i int, lines *lineMap) bool {
	a := tokens[i]
	if a.kind != tokIdent || a.err != nil {
		return false
	}
	colon := func(k int) bool {
		return k < len(tokens) && tokens[k].kind == tokPunct && tokens[k].text == ":" && lines.line(tokens[k].start) == lines.line(a.start)
	}
	if colon(i + 1) {
		return true
	}
	return a.text == sectionGroup && i+1 < len(tokens) && tokens[i+1].kind == tokString && tokens[i+1].err == nil && colon(i+2)
}

// `when` or `# when` at tokens[i], which starts a when line unless `=` follows the word (SPEC.md §6.9).
func whenShape(tokens []token, i int) bool {
	if i < len(tokens) && tokens[i].kind == tokPunct && tokens[i].text == "#" {
		i++
	}
	word := func(k int, text string) bool {
		return k < len(tokens) && tokens[k].kind == tokIdent && tokens[k].text == text
	}
	return word(i, "when") && !(i+1 < len(tokens) && tokens[i+1].kind == tokPunct && tokens[i+1].text == "=")
}

// Whether a file of kind holds a spec line: a full file, and a section file whose section reads the spec or holds a
// rotation, whose names the spec's class reads (SPEC.md §3.3). A kind that names no section is not read past it.
func holdsSpec(ix *schemaIndex, kind Kind) bool {
	def := ix.sections[kind.Section]
	return kind.Section == "" || def == nil || def.Category == rotationCategory ||
		slices.ContainsFunc(def.Sources, func(src sectionSource) bool { return src.Flatten != "" || slices.Contains(src.Path, "*") })
}

// The sections a file kind holds: all of them in a full file, and its own in a section file. An x.apl.siml holds
// the rotation, every section whose fields all lie in the apl section's message.
func fileSections(ix *schemaIndex, kind Kind) map[string]bool {
	allowed := map[string]bool{}
	if kind.Section == "" {
		for name := range ix.sections {
			allowed[name] = true
		}
		return allowed
	}
	own := ix.sections[kind.Section]
	if own == nil {
		return allowed
	}
	allowed[own.Name] = true
	if own.Name != sectionAPL {
		return allowed
	}
	root := own.Sources[0].Path
	within := func(def *sectionDef) bool {
		return !slices.ContainsFunc(def.Sources, func(s sectionSource) bool {
			return len(s.Path) < len(root) || !slices.Equal(s.Path[:len(root)], root)
		})
	}
	for name, def := range ix.sections {
		if within(def) {
			allowed[name] = true
		}
	}
	return allowed
}

type fileParser struct {
	ix           *schemaIndex
	kind         Kind
	tokens       []token
	trivia       []trivia
	lines        *lineMap
	allowed      map[string]bool
	diags        []diag
	tree         *syntaxTree
	present      map[string]bool
	nextTrivia   int
	implied      *section
	section      *section
	def          *sectionDef
	skip         bool // The current header names no section, so its statements are not read.
	seenHeader   bool
	seenSpec     bool
	seenTalents  bool
	versionFirst bool
	blocks       []openBlock // The when blocks open at the current statement, outermost first.
}

// A when line and the column it starts at, whose body is the statements indented deeper. A misplaced when's body is
// not read; a when line that does not read still holds its body, which is read on its own.
type openBlock struct {
	indent int
	when   *whenStmt // Nil for a when line that did not read.
	skip   bool
}

func (f *fileParser) run() (*syntaxTree, []diag) {
	f.openImplied()
	for n, toks := range statementsOf(f.tokens, f.tree.blanks, f.lines) {
		f.physical(toks, n == 0)
	}
	if !f.versionFirst {
		at := span{}
		if len(f.tree.statements) > 0 {
			at = f.tree.statements[0].base().span
		}
		f.diags = append(f.diags, errorAt("missing-version", "a file starts with siml 1", at))
	}
	f.closeBlocks(-1)
	f.tree.trailing = f.trivia[f.nextTrivia:]
	if !f.skip {
		for _, t := range f.tree.trailing {
			if t.kind == triviaNote {
				f.misplacedNote(t)
			}
		}
	}
	return f.tree, f.diags
}

func (f *fileParser) openImplied() {
	name := f.kind.Section
	if name == "" {
		return
	}
	def := f.ix.sections[name]
	if def == nil {
		f.diags = append(f.diags, errorAt("unknown-section", "the file name names no section: "+name, span{}))
	}
	f.implied = &section{name: name}
	f.tree.sections = append(f.tree.sections, f.implied)
	f.present[name] = true
	f.section = f.implied
	f.def = def
	f.skip = def == nil
}

func (f *fileParser) physical(toks []token, first bool) {
	p := newParser(toks)
	start, end := toks[0].start, toks[len(toks)-1].end
	leading := f.triviaBefore(f.lines.line(start))
	inner := f.triviaBefore(f.lines.line(end-1) + 1)
	isHeader := headerShape(toks, 0, f.lines)
	if isHeader {
		f.closeBlocks(-1)
	} else {
		f.closeBlocks(f.lines.column(start))
	}
	block := f.block()
	var statement stmt
	row := false
	switch {
	case isHeader:
		statement = f.header(p, end)
	case block != nil && block.skip:
		statement = &unparsedStmt{}
		f.place(f.section, statement)
	case p.atEnd():
		// Its first token has a lexer error, which is all there is to report.
		statement = &unparsedStmt{}
		f.place(f.section, statement)
	default:
		read := f.line(p, first)
		row = read.row
		statement = read.statement
		f.place(read.place, statement)
	}
	b := statement.base()
	b.span, b.leading, b.inner = span{start, end}, leading, inner
	if block != nil && block.when != nil {
		block.when.body = append(block.when.body, statement)
		if r, ok := statement.(*rowStmt); ok {
			r.when = block.when
		}
	}
	if block == nil || !block.skip {
		f.attachNotes(statement, row)
	}
	f.tree.statements = append(f.tree.statements, statement)
	for k, t := range toks {
		if t.err != nil && k < p.suppressFrom {
			f.diags = append(f.diags, errorAt(t.err.code, t.err.message, t.err.span))
		}
	}
	f.diags = append(f.diags, p.diags...)
}

// The innermost open when block, or nil.
func (f *fileParser) block() *openBlock {
	if len(f.blocks) == 0 {
		return nil
	}
	return &f.blocks[len(f.blocks)-1]
}

// Closes the blocks a statement at column indent ends, those that start at it or deeper; -1 closes all. An empty one
// is empty-when.
func (f *fileParser) closeBlocks(indent int) {
	for len(f.blocks) > 0 && (indent < 0 || f.blocks[len(f.blocks)-1].indent >= indent) {
		if w := f.blocks[len(f.blocks)-1].when; w != nil && len(w.body) == 0 {
			f.diags = append(f.diags, diag{span: w.span, severity: Warning, code: "empty-when", message: "the when block holds no row; its rows go on the lines below it, indented deeper"})
		}
		f.blocks = f.blocks[:len(f.blocks)-1]
	}
}

func (f *fileParser) place(s *section, statement stmt) {
	if s == nil {
		f.tree.top = append(f.tree.top, statement)
	} else {
		s.body = append(s.body, statement)
	}
}

func (f *fileParser) triviaBefore(line int) []trivia {
	from := f.nextTrivia
	for f.nextTrivia < len(f.trivia) && f.trivia[f.nextTrivia].line < line {
		f.nextTrivia++
	}
	return f.trivia[from:f.nextTrivia]
}

func (f *fileParser) header(p *parser, end int) *headerStmt {
	nameTok := p.take()
	h := &headerStmt{name: nameTok.name()}
	if nameTok.text == sectionGroup && p.peekKind(0, tokString) {
		h.group = p.take().stringNode()
	}
	p.take()
	if f.lines.column(nameTok.start) != 0 {
		p.report("indented-header", "a section header starts at column 1", h.name.span)
	}
	f.open(p, h)
	if !p.atEnd() {
		read := f.line(p, false)
		h.body = read.statement
		h.body.base().span = span{read.start, end}
		f.place(read.place, h.body)
	}
	return h
}

func (f *fileParser) open(p *parser, h *headerStmt) {
	def := f.ix.sections[h.name.text]
	f.seenHeader = true
	skip := false
	switch {
	case def == nil:
		p.report("unknown-section", "no section is named "+h.name.text, h.name.span)
		skip = true
	case def.Name == sectionGroup && h.group == nil:
		p.report("parse", `a group header names its group: group "Name":`, h.name.span)
		skip = true
	default:
		if !f.allowed[def.Name] {
			p.report("section-not-allowed", fmt.Sprintf("an x.%s.siml file holds no %s: section", f.kind.Section, def.Name), h.name.span)
		}
		// A group's header names its group, so it opens a section of its own.
		if f.implied != nil && f.implied.name == def.Name && f.implied.header == nil && h.group == nil {
			f.implied.header = h
			f.section, f.def, f.skip = f.implied, def, false
			return
		}
		key := def.Name
		if h.group != nil {
			key = sectionGroup + " " + h.group.value
		}
		if f.present[key] {
			if h.group != nil {
				p.report("duplicate-group", fmt.Sprintf(`group "%s" is written twice`, h.group.value), h.group.span)
			} else {
				p.report("duplicate-section", def.Name+": is written twice", h.name.span)
			}
		}
		f.present[key] = true
	}
	f.section = &section{name: h.name.text, group: h.group, header: h}
	f.tree.sections = append(f.tree.sections, f.section)
	f.def, f.skip = def, skip
	if skip {
		f.def = nil
	}
}

type lineRead struct {
	statement stmt
	place     *section
	row       bool
	start     int
}

// One statement that is not a header, in the current section.
func (f *fileParser) line(p *parser, first bool) lineRead {
	t := p.peek(0)
	word := ""
	if t.kind == tokIdent && !(topWords[t.text] && p.isPunct("=", 1)) {
		word = t.text
	}
	top := func(s stmt) lineRead { return lineRead{statement: s, start: t.start} }
	here := func(s stmt, row bool) lineRead {
		return lineRead{statement: s, place: f.section, row: row, start: t.start}
	}
	switch word {
	case "siml":
		if first {
			f.versionFirst = true
		} else {
			p.report("misplaced-statement", "siml 1 is the first line of a file", t.at())
		}
		return top(attempt(p, p.version))
	case "spec", "talents", "use":
		switch {
		case word == "spec" && f.seenSpec || word == "talents" && f.seenTalents:
			p.report("misplaced-statement", word+" is written once", t.at())
		case f.seenHeader:
			p.report("misplaced-statement", word+" goes before the first section header", t.at())
		case word == "spec" && !holdsSpec(f.ix, f.kind):
			p.report("misplaced-statement", fmt.Sprintf("an x.%s.siml file holds no spec line, as nothing in it reads the spec", f.kind.Section), t.at())
		case word == "talents" && f.kind.Section != "":
			p.report("misplaced-statement", fmt.Sprintf("an x.%s.siml file holds no talents line, which only an x.siml file holds", f.kind.Section), t.at())
		}
		f.seenSpec = f.seenSpec || word == "spec"
		f.seenTalents = f.seenTalents || word == "talents"
		read := p.use
		switch word {
		case "spec":
			read = p.spec
		case "talents":
			read = p.talents
		}
		return top(attempt(p, read))
	}
	if f.skip {
		return here(&unparsedStmt{}, false)
	}
	headless := f.section == f.implied && f.kind.Section == sectionGroup
	if headless {
		p.report("misplaced-statement", `in an x.group.siml, every line but siml, spec and use goes under a group "Name": header`, t.at())
	}
	if whenShape(p.toks, p.i) {
		return f.when(p, here)
	}
	if word == "def" {
		if f.section != nil && !headless && (f.def == nil || f.def.Name != sectionAPL) {
			p.report("misplaced-def", "def goes at the top of a file or under apl:", t.at())
		}
		s := attempt(p, p.defDecl)
		if f.section == nil {
			return top(s)
		}
		return here(s, false)
	}
	if word == "var" {
		switch {
		case f.def != nil && !slices.Contains(f.def.Body, "var"):
			p.report("misplaced-var", "var goes at the top of a file, under apl: or under a group", t.at())
		case len(f.blocks) > 0:
			p.report("misplaced-var", "a var line goes outside when blocks", t.at())
		}
		s := attempt(p, p.varDecl)
		if f.section == nil {
			return top(s)
		}
		return here(s, false)
	}
	if f.def == nil {
		return top(attempt(p, func() stmt {
			p.fail("this line goes under a section header, such as priority: or settings:", "parse")
			return nil
		}))
	}
	body := f.def.Body
	if slices.Contains(body, "rows") {
		return here(f.row(p, f.def.Name == sectionPrepull), true)
	}
	slotRowsOnly := f.kind.Section == sectionGear && f.section == f.implied // An x.gear.siml holds slot rows only.
	if slices.Contains(body, "slot_rows") && (slotRowsOnly || f.ix.slots[word]) {
		return here(attempt(p, p.slotRow), false)
	}
	return here(attempt(p, p.entry), false)
}

// A when line, which opens a block (SPEC.md §6.9). One outside priority: and the groups is misplaced-when, and
// neither it nor its body is read.
func (f *fileParser) when(p *parser, here func(stmt, bool) lineRead) lineRead {
	t := p.peek(0)
	block := openBlock{indent: f.lines.column(t.start)}
	if f.def == nil || !slices.Contains(f.def.Body, "rows") || f.def.Name == sectionPrepull {
		p.report("misplaced-when", "a when block goes under priority: or a group", t.at())
		block.skip = true
		f.blocks = append(f.blocks, block)
		return here(&unparsedStmt{}, false)
	}
	s := attempt(p, p.when)
	if w, ok := s.(*whenStmt); ok {
		if parent := f.block(); parent != nil {
			w.parent = parent.when
		}
		block.when = w
	}
	f.blocks = append(f.blocks, block)
	return here(s, false)
}

// A row; a hidden row that does not read is bad-hidden-row alone, since a `#` line is most often a comment.
func (f *fileParser) row(p *parser, prepull bool) stmt {
	read := func() stmt { return p.row(prepull) }
	if !p.isPunct("#", 0) {
		return attempt(p, read)
	}
	from := p.i
	mark := len(p.diags)
	hash := p.peek(0)
	s, failed := try(read)
	if _, isRow := s.(*rowStmt); failed == nil && isRow && !p.truncated {
		return s
	}
	p.diags = p.diags[:mark]
	p.suppressFrom = from
	p.report("bad-hidden-row", "# hides a row; write // for a comment", span{hash.start, p.lastEnd})
	return &unparsedStmt{}
}

// `///` lines directly above a priority or group row are its notes; any other `///` is misplaced-note.
func (f *fileParser) attachNotes(statement stmt, row bool) {
	if f.skip {
		return
	}
	b := statement.base()
	k := len(b.leading)
	for k > 0 && b.leading[k-1].kind != triviaBlank {
		k--
	}
	var adjacent []trivia
	adjacentAt := map[span]bool{}
	for _, t := range b.leading[k:] {
		if t.kind == triviaNote {
			adjacent = append(adjacent, t)
			adjacentAt[t.span] = true
		}
	}
	attachable := row && (f.def == nil || f.def.Name != sectionPrepull)
	isAdjacent := func(t trivia) bool { return t.kind == triviaNote && adjacentAt[t.span] }
	if r, ok := statement.(*rowStmt); ok && attachable {
		r.notes = adjacent
		b.leading = slices.DeleteFunc(slices.Clone(b.leading), isAdjacent)
	}
	for _, t := range slices.Concat(b.leading, b.inner) {
		if t.kind == triviaNote && !(attachable && isAdjacent(t)) {
			f.misplacedNote(t)
		}
	}
}

func (f *fileParser) misplacedNote(t trivia) {
	f.diags = append(f.diags, errorAt("misplaced-note", "/// notes go directly above a priority or group row; write // for a comment", t.span))
}

// Runs a statement's reader. A failure leaves the statement unread with its diagnostic, and so does a lexer error,
// in front of which the reader stopped.
func attempt(p *parser, read func() stmt) stmt {
	s, f := try(read)
	if f != nil {
		p.failed(*f)
		return &unparsedStmt{}
	}
	if p.truncated {
		return &unparsedStmt{}
	}
	return s
}

// Runs read, recovering the failure that ends the statement.
func try(read func() stmt) (s stmt, fail *failure) {
	defer func() {
		if r := recover(); r != nil {
			f, ok := r.(failure)
			if !ok {
				panic(r)
			}
			fail = &f
		}
	}()
	return read(), nil
}

// Reads one statement's tokens. They stop in front of the first token with a lexer error, whose diagnostic
// stands for the statement: a failure at the end of the readable tokens then adds none.
type parser struct {
	diags        []diag
	truncated    bool
	lastEnd      int // The end of the statement's last token, readable or not.
	suppressFrom int // The index from which the tokens' lexer errors are not reported.
	toks         []token
	end          int // The end of the last readable token.
	i            int
	depth        int // How deep the statement nests at the next token (nest).
}

func newParser(all []token) *parser {
	p := &parser{toks: all, lastEnd: all[len(all)-1].end, suppressFrom: math.MaxInt}
	if bad := slices.IndexFunc(all, func(t token) bool { return t.err != nil }); bad >= 0 {
		p.toks, p.truncated = all[:bad], true
	}
	p.end = all[0].start
	if len(p.toks) > 0 {
		p.end = p.toks[len(p.toks)-1].end
	}
	return p
}

// --- Tokens

func (t token) at() span { return span{t.start, t.end} }

func (t token) name() name { return name{span: t.at(), text: t.text} }

func (t token) stringNode() *stringNode { return &stringNode{span: t.at(), value: t.value} }

// The token k ahead, or the zero token, whose kind matches nothing a reader asks for, past the end.
func (p *parser) peek(k int) token {
	if p.i+k < len(p.toks) {
		return p.toks[p.i+k]
	}
	return token{kind: -1}
}

func (p *parser) has(k int) bool { return p.i+k < len(p.toks) }

func (p *parser) atEnd() bool { return p.i >= len(p.toks) }

func (p *parser) peekKind(k int, kind tokenKind) bool { return p.has(k) && p.toks[p.i+k].kind == kind }

func (p *parser) isPunct(text string, k int) bool {
	return p.peekKind(k, tokPunct) && p.toks[p.i+k].text == text
}

func (p *parser) isWord(text string, k int) bool {
	return p.peekKind(k, tokIdent) && p.toks[p.i+k].text == text
}

// Takes the next token when it is the word text, and says whether it was.
func (p *parser) acceptWord(text string) bool {
	ok := p.isWord(text, 0)
	if ok {
		p.take()
	}
	return ok
}

// Takes the next token when it is the punctuation text, and says whether it was.
func (p *parser) acceptPunct(text string) bool {
	ok := p.isPunct(text, 0)
	if ok {
		p.take()
	}
	return ok
}

// The items of a list whose open bracket is taken, each read by item, up to close, which it takes.
func (p *parser) items(open token, close string, item func()) {
	p.nest(open, 1)
	if !p.isPunct(close, 0) {
		for {
			item()
			if !p.acceptPunct(",") {
				break
			}
		}
	}
	p.expectPunct(close, "',' or '"+close+"'")
	p.depth--
}

func (p *parser) take() token {
	p.i++
	return p.toks[p.i-1]
}

// From start to the end of the last token taken.
func (p *parser) from(start int) span { return span{start, p.toks[p.i-1].end} }

func (p *parser) report(code, message string, at span) {
	p.diags = append(p.diags, errorAt(code, message, at))
}

func (p *parser) failed(f failure) {
	if !(f.atEnd && p.truncated) {
		p.report(f.code, f.message, f.span)
	}
}

// Ends the statement with a failure at the next token, or at the end of the readable tokens.
func (p *parser) fail(message, code string) {
	if !p.has(0) {
		panic(failure{code: code, message: message, span: span{p.end, p.end}, atEnd: true})
	}
	panic(failure{code: code, message: message, span: p.peek(0).at()})
}

func (p *parser) failAt(message, code string, at span) {
	panic(failure{code: code, message: message, span: at})
}

// How deep a statement nests at most (SPEC.md §4 note 10), which keeps reading and printing it within a browser's
// stack.
const maxDepth = 256

// Goes levels deeper at t, failing past maxDepth: a bracket, not or schedule goes one deeper until what it holds ends,
// an arithmetic operator one deeper for the rest of the statement, and raw JSON as deep as its own brackets.
func (p *parser) nest(t token, levels int) {
	p.depth += levels
	if p.depth > maxDepth {
		p.failAt(fmt.Sprintf("the statement nests deeper than %d", maxDepth), "parse", t.at())
	}
}

func (p *parser) expected(what, code string) {
	if !p.has(0) {
		p.fail("expected "+what, code)
	}
	found := p.peek(0).text
	if utf8.RuneCountInString(found) > 24 {
		found = string([]rune(found)[:21]) + "..."
	}
	p.fail(fmt.Sprintf("expected %s, found '%s'", what, found), code)
}

func (p *parser) expectPunct(text, what string) token {
	if !p.isPunct(text, 0) {
		p.expected(what, "parse")
	}
	return p.take()
}

func (p *parser) finish(what string) {
	if !p.atEnd() {
		p.expected(what, "parse")
	}
}

func (p *parser) string(what string) *stringNode {
	if !p.peekKind(0, tokString) {
		p.expected(what, "parse")
	}
	return p.take().stringNode()
}

// --- Top-level lines (SPEC.md §3.3)

func (p *parser) version() stmt {
	p.take()
	t := p.peek(0)
	if t.kind != tokNumber || !isInt(t.text) {
		p.expected("the major version: siml 1", "parse")
	}
	p.take()
	p.finish("the end of the siml line")
	if n, err := strconv.Atoi(t.text); err != nil || n != 1 {
		p.report("unsupported-version", fmt.Sprintf("this reads SIML 1, and the file is SIML %s", t.text), t.at())
	}
	return &versionStmt{version: literalNode{span: t.at(), text: t.text}}
}

func (p *parser) spec() stmt {
	p.take()
	t := p.peek(0)
	if t.kind != tokIdent {
		p.expected("the spec, such as spec hunter", "parse")
	}
	p.take()
	p.finish("the end of the spec line")
	return &specStmt{spec: t.name()}
}

func (p *parser) talents() stmt {
	p.take()
	talents := p.string("the talents in quotes")
	p.finish("the end of the talents line")
	return &talentsStmt{talents: *talents}
}

func (p *parser) use() stmt {
	p.take()
	s := &useStmt{path: *p.string("the used file's path in quotes")}
	if p.acceptWord("with") {
		for {
			s.with = append(s.with, p.varBind())
			if !p.acceptPunct(",") {
				break
			}
		}
	}
	if len(s.with) > 0 {
		p.finish("',' or the end of the use line")
	} else {
		p.finish("with or the end of the use line")
	}
	return s
}

func (p *parser) varBind() varBind {
	variable := p.variable()
	p.expectPunct("=", "'='")
	value := p.expr()
	return varBind{span: join(variable, value), variable: variable, value: value}
}

func (p *parser) variable() *varNode {
	t := p.peek(0)
	if t.kind != tokVar {
		p.expected("a variable, $name or ${Name}", "parse")
	}
	p.take()
	return &varNode{span: t.at(), name: t.value, braced: t.text[1] == '{'}
}

func (p *parser) placeholder(t token) *placeholderNode {
	return &placeholderNode{span: t.at(), name: t.value, braced: t.text[1] == '{'}
}

// `var $x = value`, or `var $x` for a variable with no value.
func (p *parser) varDecl() stmt {
	p.take()
	s := &varStmt{variable: p.variable()}
	if p.atEnd() {
		return s
	}
	p.expectPunct("=", "'=' or the end of the var line")
	s.value = p.expr()
	p.finish("an operator or the end of the var line")
	return s
}

// `def name(p, …) = value`, or `def name = value` for a def with no parameters.
func (p *parser) defDecl() stmt {
	p.take()
	t := p.peek(0)
	if t.kind != tokIdent {
		p.expected("the def's name, such as def fits(x) = …", "parse")
	}
	p.take()
	s := &defStmt{name: t.name()}
	if p.acceptPunct("(") {
		for !p.isPunct(")", 0) {
			if len(s.params) > 0 {
				p.expectPunct(",", "',' or ')'")
			}
			if !p.peekKind(0, tokIdent) {
				p.expected("a parameter's name", "parse")
			}
			s.params = append(s.params, p.take().name())
		}
		p.take()
	}
	p.expectPunct("=", "'=' and the def's value")
	s.value = p.expr()
	p.finish("an operator or the end of the def line")
	return s
}

// `[#] when <expr>:`, the colon ending its line.
func (p *parser) when() stmt {
	s := &whenStmt{hide: p.acceptPunct("#")}
	p.take()
	s.condition = p.expr()
	if !p.isPunct(":", 0) || p.has(1) {
		p.acceptPunct(":")
		p.fail("a when line ends with :", "parse")
	}
	p.take()
	return s
}

// --- Setup sections (SPEC.md §5)

func (p *parser) slotRow() stmt {
	if !p.peekKind(0, tokIdent) {
		p.expected("a slot, such as head", "parse")
	}
	s := &slotRowStmt{slot: p.take().name()}
	if p.peekKind(0, tokRaw) {
		s.raw = p.raw()
		p.finish("the end of the slot row")
		return s
	}
	if !p.peekKind(0, tokIdent) {
		p.expected("the item's name", "parse")
	}
	item := p.take().name()
	s.item = &item
	if p.peekKind(0, tokIdent) {
		suffix := p.take().name()
		s.suffix = &suffix
	}
	if p.acceptPunct("+") {
		if !p.peekKind(0, tokIdent) {
			p.expected("the enchant's name", "parse")
		}
		enchant := p.take().name()
		s.enchant = &enchant
	}
	if s.enchant != nil {
		p.finish("the end of the slot row")
	} else {
		p.finish("'+ enchant' or the end of the slot row")
	}
	return s
}

// A flag list, a field and its values, a field and its message's arguments, or raw JSON (SPEC.md §4 note 5).
func (p *parser) entry() stmt {
	t := p.peek(0)
	if t.kind == tokRaw {
		value := p.raw()
		p.finish("the end of the entry")
		return &rawEntryStmt{value: value}
	}
	if t.kind != tokIdent {
		p.expected("a field name", "parse")
	}
	if !p.has(1) || p.isPunct(",", 1) || p.isPunct("=", 1) {
		return p.flags()
	}
	p.take()
	if p.isPunct("(", 0) {
		args := p.args()
		p.finish("the end of the entry")
		return &messageEntryStmt{name: t.name(), args: args}
	}
	s := &fieldEntryStmt{name: t.name(), values: []node{p.term(entryEnds)}}
	for p.acceptPunct(",") {
		s.values = append(s.values, p.term(entryEnds))
	}
	p.finish("',' or the end of the entry")
	return s
}

func (p *parser) flags() stmt {
	s := &flagListStmt{}
	for {
		t := p.peek(0)
		if t.kind != tokIdent {
			p.expected("a field name", "parse")
		}
		p.take()
		fl := flag{name: t.name()}
		if p.acceptPunct("=") {
			fl.value = p.term(entryEnds)
		}
		fl.span = p.from(t.start)
		s.flags = append(s.flags, fl)
		if !p.acceptPunct(",") {
			break
		}
	}
	p.finish("',' or the end of the flags")
	return s
}

// --- Rows (SPEC.md §6.3)

func (p *parser) row(prepull bool) stmt {
	s := &rowStmt{hide: p.acceptPunct("#")}
	// Raw JSON alone is the whole row, its time included (SPEC.md §6.3).
	rawRow := !s.hide && p.peekKind(0, tokRaw) && !p.has(1)
	if prepull && !rawRow {
		s.doAt = p.doAt()
	} else if p.startsTime() {
		s.doAt = p.unary()
		p.report("unexpected-time", "only a prepull row starts with a time", s.doAt.pos())
	}
	s.action = p.action(nil)
	p.finish("an operator or the end of the row")
	return s
}

// A prepull row's time. A row that holds only an action is missing-time.
func (p *parser) doAt() node {
	at := span{p.end, p.end}
	if p.has(0) {
		at = p.peek(0).at()
		if !p.startsAction() {
			from, depth := p.i, p.depth
			value := p.unary()
			if !p.atEnd() && !p.isWord("if", 0) {
				return value
			}
			p.i, p.depth = from, depth
		}
	}
	p.report("missing-time", "a prepull row starts with its time, such as -1.5s", at)
	return nil
}

// A token that starts a value but never an action.
func (p *parser) startsTime() bool {
	t := p.peek(0)
	switch t.kind {
	case tokNumber, tokString, tokVar, tokPlaceholder:
		return true
	case tokPunct:
		return t.text == "-" || t.text == "("
	case tokIdent:
		return t.text == "true" || t.text == "false"
	}
	return false
}

// A keyword that starts an action and never a value.
func (p *parser) startsAction() bool {
	t := p.peek(0)
	if t.kind != tokIdent {
		return false
	}
	switch t.text {
	case "cast", "channel", "wait", "move", "group":
		return true
	case "schedule", "sequence":
		return p.peekKind(1, tokString)
	}
	return false
}

// A step and its condition. Tokens after the step that are not `if` or one of ends are expected-if.
func (p *parser) action(ends []string) *actionNode {
	start := p.end
	if p.has(0) {
		start = p.peek(0).start
	}
	a := &actionNode{step: p.step()}
	if p.has(0) && !p.isWord("if", 0) && !(p.peek(0).kind == tokPunct && slices.Contains(ends, p.peek(0).text)) {
		p.fail("a condition starts with if", "expected-if")
	}
	if p.acceptWord("if") {
		a.condition = p.expr()
	}
	a.span = p.from(start)
	return a
}

func (p *parser) step() node {
	t := p.peek(0)
	if t.kind == tokRaw {
		return p.raw()
	}
	if t.kind != tokIdent {
		p.expected("an action", "parse")
	}
	start := t.start
	switch t.text {
	case "none":
		p.take()
		return &noneNode{span: p.from(start)}
	case "cast":
		p.take()
		s := &castStep{spell: p.ref()}
		s.target = p.on()
		s.span = p.from(start)
		return s
	case "channel":
		p.take()
		s := &channelStep{spell: p.ref()}
		s.target = p.on()
		if p.acceptWord("until") {
			s.until = p.expr()
		}
		s.recast = p.acceptWord("recast")
		s.span = p.from(start)
		return s
	case "wait":
		p.take()
		s := &waitStep{until: p.acceptWord("until")}
		s.value = p.expr()
		s.span = p.from(start)
		return s
	case "move":
		p.take()
		s := &moveStep{duration: p.acceptWord("for")}
		s.value = p.expr()
		s.recheck = !s.duration && p.acceptWord("recheck")
		s.span = p.from(start)
		return s
	case "group":
		return p.groupStep()
	case "schedule":
		if !p.peekKind(1, tokString) {
			break
		}
		p.nest(p.take(), 1)
		s := &scheduleStep{schedule: p.take().stringNode()}
		s.step = p.step()
		p.depth--
		s.span = p.from(start)
		return s
	case "sequence":
		if !p.peekKind(1, tokString) {
			break
		}
		p.take()
		s := &sequenceStep{name: p.take().stringNode()}
		p.items(p.expectPunct("[", "'[' and the sequence's actions"), "]", func() { s.actions = append(s.actions, p.action(listEnds)) })
		s.span = p.from(start)
		return s
	}
	if !p.callable() {
		p.expected("an action", "parse")
	}
	p.take()
	return p.call(t)
}

// `group "Name" [with %p = v, …]`; the list goes past a comma only while a `%p =` follows (SPEC.md §4 note 1).
func (p *parser) groupStep() node {
	start := p.take().start
	s := &groupStep{name: p.string("the group's name in quotes")}
	if p.acceptWord("with") {
		for {
			if len(s.with) > 0 {
				p.take()
			}
			t := p.peek(0)
			if t.kind != tokPlaceholder {
				p.expected("a placeholder, %name or %{Name}", "parse")
			}
			p.take()
			p.expectPunct("=", "'='")
			placeholder := p.placeholder(t)
			value := p.expr()
			s.with = append(s.with, placeholderBind{span: join(placeholder, value), placeholder: placeholder, value: value})
			if !(p.isPunct(",", 0) && p.peekKind(1, tokPlaceholder) && p.isPunct("=", 2)) {
				break
			}
		}
	}
	s.span = p.from(start)
	return s
}

// A name, item.<name>, other.<value>, a variable, none or raw JSON.
func (p *parser) ref() node {
	t := p.peek(0)
	switch t.kind {
	case tokRaw:
		return p.raw()
	case tokVar:
		return p.variable()
	}
	if t.kind != tokIdent || stopWords[t.text] {
		p.expected("a spell name, item.<name> or other.<name>", "parse")
	}
	p.take()
	n := t.name()
	if t.text == "none" {
		return &noneNode{span: n.span}
	}
	if (t.text == "other" || t.text == "item") && p.isPunct(".", 0) && p.peekKind(1, tokIdent) {
		p.take()
		member := p.take().name()
		return &memberNode{span: p.from(t.start), object: n, name: member}
	}
	return &callNode{span: n.span, name: n}
}

func (p *parser) on() node {
	if !p.isWord("on", 0) {
		return nil
	}
	p.take()
	t := p.peek(0)
	if t.kind == tokRaw {
		return p.raw()
	}
	if t.kind != tokIdent || stopWords[t.text] {
		p.expected("a unit, such as target", "parse")
	}
	p.take()
	n := t.name()
	if t.text == "none" {
		return &noneNode{span: n.span}
	}
	return &callNode{span: n.span, name: n}
}

func (p *parser) raw() *rawNode {
	t := p.take()
	depth := jsonDepth(t.value)
	p.nest(t, depth)
	p.depth -= depth
	if err := checkJSON(t.value); err != nil {
		p.failAt("the raw JSON does not read: "+err.Error(), "bad-raw", t.at())
	}
	return &rawNode{span: t.at(), json: t.value}
}

// --- Values (SPEC.md §6.5)

func (p *parser) expr() node { return p.logic("or", p.conj) }

func (p *parser) conj() node { return p.logic("and", p.neg) }

func (p *parser) logic(op string, next func() node) node {
	first := next()
	if !p.isWord(op, 0) {
		return first
	}
	operands := []node{first}
	for p.isWord(op, 0) {
		p.take()
		operands = append(operands, next())
	}
	return &logicNode{span: join(first, operands[len(operands)-1]), op: op, operands: operands}
}

func (p *parser) neg() node {
	if !p.isWord("not", 0) || p.callable() {
		return p.cmp()
	}
	t := p.take()
	p.nest(t, 1)
	operand := p.neg()
	p.depth--
	return &notNode{span: span{t.start, operand.pos().end}, operand: operand}
}

func (p *parser) isCompare(k int) bool {
	return p.peekKind(k, tokPunct) && slices.Contains(compareOps, p.toks[p.i+k].text)
}

func (p *parser) cmp() node {
	left := p.sum()
	if !p.isCompare(0) {
		p.noAssign()
		return left
	}
	op := p.take().text
	right := p.sum()
	if p.isCompare(0) {
		p.fail("comparisons do not chain: write a < b and b < c", "chained-comparison")
	}
	p.noAssign()
	return &compareNode{span: join(left, right), op: op, left: left, right: right}
}

func (p *parser) noAssign() {
	if p.isPunct("=", 0) {
		p.fail("write == to compare", "write-eq")
	}
}

func (p *parser) sum() node { return p.math("+", "-", p.prod) }

func (p *parser) prod() node { return p.math("*", "/", p.unary) }

func (p *parser) math(a, b string, next func() node) node {
	left := next()
	for p.isPunct(a, 0) || p.isPunct(b, 0) {
		op := p.take()
		p.nest(op, 1)
		right := next()
		left = &mathNode{span: join(left, right), op: op.text, left: left, right: right}
	}
	return left
}

// A minus goes only directly before a number, and is part of its literal.
func (p *parser) unary() node {
	if !p.isPunct("-", 0) {
		return p.operand()
	}
	minus := p.take()
	t := p.peek(0)
	if t.kind != tokNumber {
		p.expected("a number after -", "parse")
	}
	p.take()
	return &literalNode{span: span{minus.start, t.end}, text: "-" + t.text}
}

func (p *parser) operand() node {
	if !p.has(0) {
		p.expected("a value", "parse")
	}
	t := p.peek(0)
	switch t.kind {
	case tokNumber:
		p.take()
		return &literalNode{span: t.at(), text: t.text}
	case tokString:
		return p.take().stringNode()
	case tokVar:
		return p.variable()
	case tokPlaceholder:
		return p.placeholder(p.take())
	case tokRaw:
		return p.raw()
	case tokIdent:
		return p.word()
	}
	if !p.isPunct("(", 0) {
		p.expected("a value", "parse")
	}
	p.nest(p.take(), 1)
	e := p.expr()
	p.expectPunct(")", "an operator or ')'")
	p.depth--
	return &parenNode{span: p.from(t.start), expr: e}
}

// Whether the word ahead reads as a name: any but a stop word, and `and`, `or` and `not` directly followed by
// `(`, which are calls.
func (p *parser) callable() bool {
	t := p.peek(0)
	if !stopWords[t.text] {
		return true
	}
	return callWords[t.text] && p.isPunct("(", 1) && p.peek(1).start == t.end
}

func (p *parser) word() node {
	t := p.peek(0)
	switch t.text {
	case "true", "false":
		p.take()
		return &boolNode{span: t.at(), value: t.text == "true"}
	case "none":
		if !p.isPunct(".", 1) {
			p.take()
			return &noneNode{span: t.at()}
		}
	}
	if !p.callable() {
		p.expected("a value", "parse")
	}
	p.take()
	n := t.name()
	if p.isPunct(".", 0) && p.peekKind(1, tokIdent) {
		p.take()
		member := p.take().name()
		if !p.isPunct("(", 0) {
			return &memberNode{span: p.from(t.start), object: n, name: member}
		}
		args := p.args()
		if !p.isPunct(".", 0) {
			p.expected(fmt.Sprintf(".property after %s(…)", member.text), "parse")
		}
		return p.accessor(&n, member, args, t.start)
	}
	if !p.isPunct("(", 0) {
		return &callNode{span: n.span, name: n}
	}
	args := p.args()
	if p.isPunct(".", 0) {
		return p.accessor(nil, n, args, t.start)
	}
	return &callNode{span: p.from(t.start), name: n, args: args}
}

func (p *parser) call(t token) node {
	c := &callNode{name: t.name()}
	if p.isPunct("(", 0) {
		c.args = p.args()
	}
	c.span = p.from(t.start)
	return c
}

// `[unit.]head(subject).prop[(args)]`; the head's parentheses hold the subject and nothing else.
func (p *parser) accessor(unit *name, head name, headArgs *argList, start int) node {
	for k, a := range headArgs.items {
		if a.name != nil || k > 0 {
			p.failAt(fmt.Sprintf("arguments go on the property: %s(x).p(f=v)", head.text), "args-on-head", a.span)
		}
	}
	if len(headArgs.items) == 0 {
		p.failAt(fmt.Sprintf("expected what %s() reads, such as %[1]s(name)", head.text), "parse", headArgs.span)
	}
	p.take()
	prop := p.peek(0)
	if prop.kind != tokIdent {
		p.expected("a property after the .", "parse")
	}
	p.take()
	a := &accessorNode{unit: unit, head: head, subject: headArgs.items[0].value, prop: prop.name()}
	if p.isPunct("(", 0) {
		a.args = p.args()
	}
	a.span = p.from(start)
	return a
}

func (p *parser) args() *argList {
	open := p.take()
	l := &argList{}
	p.items(open, ")", func() { l.items = append(l.items, p.arg()) })
	l.span = p.from(open.start)
	return l
}

func (p *parser) arg() arg {
	t := p.peek(0)
	if t.kind == tokIdent && p.isPunct("=", 1) {
		p.take()
		p.take()
		n := t.name()
		value := p.term(argEnds)
		return arg{span: p.from(t.start), name: &n, value: value}
	}
	value := p.term(argEnds)
	return arg{span: value.pos(), value: value}
}

// An argument or entry value: a list, an action, or a value, which `if` makes an action.
func (p *parser) term(ends []string) node {
	if p.isPunct("[", 0) {
		return p.list()
	}
	if p.startsAction() {
		return p.action(ends)
	}
	value := p.expr()
	if !p.isWord("if", 0) {
		return value
	}
	switch value.(type) {
	case *callNode, *rawNode, *noneNode:
	default:
		p.fail("if follows an action, such as cast x if y", "parse")
	}
	p.take()
	condition := p.expr()
	return &actionNode{span: join(value, condition), step: value, condition: condition}
}

func (p *parser) list() node {
	open := p.take()
	l := &listNode{}
	p.items(open, "]", func() { l.items = append(l.items, p.term(listEnds)) })
	l.span = p.from(open.start)
	return l
}

// How deep JSON text nests its arrays and objects; the text need not be JSON.
func jsonDepth(text string) int {
	depth, deepest, inString := 0, 0, false
	for i := 0; i < len(text); i++ {
		switch c := text[i]; {
		case inString && c == '\\':
			i++
		case c == '"':
			inString = !inString
		case inString:
		case c == '[' || c == '{':
			depth++
			deepest = max(deepest, depth)
		case c == ']' || c == '}':
			depth--
		}
	}
	return deepest
}

// Why the text between a raw token's backticks is no JSON, or nil.
func checkJSON(text string) error {
	var raw json.RawMessage
	return json.Unmarshal([]byte(text), &raw)
}

func errorAt(code, message string, at span) diag {
	return diag{span: at, severity: Error, code: code, message: message}
}
