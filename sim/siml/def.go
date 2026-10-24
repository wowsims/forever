package siml

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// Defs (SPEC.md §6.10): a def names a value, and each call of it reads as that value with every parameter replaced
// by the argument written for it, typed where the call is. A diagnostic in the value is reported at the call.

// A file's defs, as typing reads them.
type defTyping struct {
	byName   map[string]*defStmt     // The file's own defs.
	imported map[string]*importedDef // The defs of the files it uses, and of the files those use in turn (SPEC.md §9).
	from     map[*defStmt]*importedDef
	cyclic   map[*defStmt]bool             // A def that calls itself through other defs, which no call expands.
	used     map[*defStmt]bool             // Called by a row, var line, when line, with list or other def.
	values   map[*proto.APLValue]*callNode // The value each call reads as, which format writes as the call.
	count    int                           // The calls expanded so far.
}

// How many calls a file's defs expand at most, so that defs that call each other several times stay quick to read.
const maxExpansions = 1 << 16

// A def a used file declares. Its spans are offset by base, past the text of this file and of the files taken in
// before it, so a diagnostic in its value is told apart from one in the arguments a call here passes it.
type importedDef struct {
	*defStmt
	path  string   // The file that declares it.
	lines *lineMap // That file's text.
	base  int
}

// The def a call of name reads: the file's own, else one a used file declares.
func (d *defTyping) lookup(name string) *defStmt {
	if own := d.byName[name]; own != nil {
		return own
	}
	if in := d.imported[name]; in != nil {
		return in.defStmt
	}
	return nil
}

// A def call being read, within the calls being read around it.
type expansion struct {
	def   *defStmt
	from  *importedDef // Where the def comes from, when a used file declares it.
	call  span
	outer *expansion
}

// Where offset, a place in the def's value or name, is: line and column, after the path for a used file's def.
func (e *expansion) where(offset int, lines *lineMap) string {
	if e.from != nil {
		at := e.from.lines.position(offset - e.from.base)
		return fmt.Sprintf("%s:%d:%d", e.from.path, at.Line, at.Col)
	}
	at := lines.position(offset)
	return fmt.Sprintf("%d:%d", at.Line, at.Col)
}

// d as the calls being read report it: at the call when it lies in the def's value, its place in the def added to
// the message. An argument that does not fit where the def puts it keeps its place, and the message names the def.
func (e *expansion) relocate(d diag, lines *lineMap) diag {
	argument := false
	for ; e != nil; e = e.outer {
		value := e.def.value.pos()
		switch {
		case value.start <= d.start && d.end <= value.end:
			d.message += fmt.Sprintf(" (in def %s, at %s)", e.def.name.text, e.where(d.start, lines))
			d.span = e.call
		case !argument:
			argument = true
			d.message += fmt.Sprintf(" (passed to def %s, at %s)", e.def.name.text, e.where(e.def.name.start, lines))
		}
	}
	return d
}

func (d *defStmt) isParam(word string) bool {
	return slices.ContainsFunc(d.params, func(p name) bool { return p.text == word })
}

// Reads the file's defs, wherever they are, reporting what SPEC.md §6.10 refuses: a name that means something else, a
// def or parameter written twice, a parameter its value never mentions, and a def that calls itself.
func (t *typer) readDefs(tree *syntaxTree) {
	t.defs = defTyping{byName: map[string]*defStmt{}, imported: map[string]*importedDef{}, from: map[*defStmt]*importedDef{}, cyclic: map[*defStmt]bool{}, used: map[*defStmt]bool{}, values: map[*proto.APLValue]*callNode{}}
	var defs []*defStmt
	for _, s := range tree.inFileOrder() {
		d, ok := s.(*defStmt)
		if !ok {
			continue
		}
		n := d.name.text
		if what := t.meaning(n); what != "" {
			t.report("def-shadows", fmt.Sprintf("%s is %s; a def takes a name of its own", n, what), d.name.span)
			continue
		}
		if t.defs.byName[n] != nil {
			t.report("duplicate-def", "a def named "+n+" is written twice", d.name.span)
			continue
		}
		t.defs.byName[n] = d
		defs = append(defs, d)
		for i, p := range d.params {
			if slices.ContainsFunc(d.params[:i], func(q name) bool { return q.text == p.text }) {
				t.report("duplicate-parameter", fmt.Sprintf("%s names the parameter %s twice", n, p.text), p.span)
			}
		}
	}
	calls := map[*defStmt][]*defStmt{}
	for _, d := range defs {
		mentioned := map[string]bool{}
		walkTree(d.value, nil, func(n, _ any) {
			switch x := n.(type) {
			case *callNode:
				if x.args == nil && d.isParam(x.name.text) {
					mentioned[x.name.text] = true
				} else if callee := t.defs.byName[x.name.text]; callee != nil {
					calls[d] = append(calls[d], callee)
					t.defs.used[callee] = t.defs.used[callee] || callee != d
				}
			case *memberNode:
				mentioned[x.object.text] = true
			case *accessorNode:
				if x.unit != nil {
					mentioned[x.unit.text] = true
				}
			}
		})
		for _, p := range d.params {
			if !mentioned[p.text] {
				t.warn("unused-parameter", fmt.Sprintf("%s never reads its parameter %s", d.name.text, p.text), p.span)
			}
		}
	}
	for _, d := range defs {
		if reaches(d, d, func(x *defStmt) []*defStmt { return slices.Clone(calls[x]) }) {
			t.defs.cyclic[d] = true
			t.report("reference-cycle", d.name.text+" calls itself through other defs", d.name.span)
		}
	}
}

// What a def's name means already (SPEC.md §6.10), or "".
func (t *typer) meaning(n string) string {
	switch {
	case t.ix.keywords[n]:
		return "a reserved word"
	case slices.Contains(t.ix.heads, n):
		return "an accessor head"
	case t.ix.unitReference(n) != nil:
		return "a unit"
	case t.ix.valueKinds[n] != nil:
		return "a value"
	case t.ix.actionKinds[n] != nil:
		return "an action"
	}
	if held := t.names.Holding(t.class, n); len(held) > 0 {
		return namespaceThings[held[0]]
	}
	return ""
}

// unused-def for each def that no row, var line, when line, with list or other def calls.
func (t *typer) checkDefsUsed() {
	for _, d := range t.defs.byName {
		if !t.defs.used[d] {
			t.warn("unused-def", "nothing calls the def "+d.name.text, d.name.span)
		}
	}
}

// Reads what follows as d's value at the call c, so a diagnostic in it moves to the call; gives what ends the call.
func (t *typer) callSite(d *defStmt, c *callNode) func() {
	t.site = &expansion{def: d, from: t.defs.from[d], call: c.span, outer: t.site}
	return func() { t.site = t.site.outer }
}

// A call of d: its value with each parameter replaced by the call's argument for it, typed where the call is.
func (t *typer) expand(c *callNode, d *defStmt) *proto.APLValue {
	var args []arg
	if c.args != nil {
		args = c.args.items
	}
	for _, a := range args {
		if a.name != nil {
			t.report("positional-only", fmt.Sprintf("%s takes its arguments in order, with no names", d.name.text), a.span)
			return nil
		}
	}
	if len(args) != len(d.params) {
		count := "no arguments"
		switch n := len(d.params); {
		case n == 1:
			count = "1 argument"
		case n > 1:
			count = strconv.Itoa(n) + " arguments"
		}
		t.report("def-arity", d.name.text+" takes "+count, c.span)
		return nil
	}
	t.defs.used[d] = true
	if t.defs.cyclic[d] {
		return nil
	}
	if t.defs.count++; t.defs.count > maxExpansions {
		t.report("parse", fmt.Sprintf("the defs expand past %d calls", maxExpansions), c.span)
		return nil
	}
	bound := map[string]node{}
	for i, p := range d.params {
		if _, ok := bound[p.text]; !ok {
			bound[p.text] = args[i].value
		}
	}
	defer t.callSite(d, c)()
	body, ok := t.substitute(d.value, bound)
	if !ok {
		return nil
	}
	v := t.value(body)
	if v != nil {
		t.defs.values[v] = c
	}
	return v
}

// What a bare call of a def with no parameters stands for where a name goes, a spell's or an aura's (SPEC.md §6.10):
// the def's value, read as if written at the call, and what ends the call. ok is false when n is no such call; value is
// nil when the def takes parameters, def-arity, or calls itself, which readDefs reports.
func (t *typer) named(n node) (value node, done func(), ok bool) {
	c, isCall := n.(*callNode)
	if !isCall || c.args != nil {
		return nil, nil, false
	}
	d := t.defs.lookup(c.name.text)
	if d == nil {
		return nil, nil, false
	}
	t.defs.used[d] = true
	if len(d.params) > 0 {
		t.report("def-arity", d.name.text+" takes arguments, so it stands for no name", c.span)
		return nil, func() {}, true
	}
	if t.defs.cyclic[d] {
		return nil, func() {}, true
	}
	return d.value, t.callSite(d, c), true
}

// n with each parameter replaced by its argument, in a ref, unit, operand or argument position. A unit written before
// a dot takes a unit word only; false, and wrong-type, when its argument is something else.
func (t *typer) substitute(n node, args map[string]node) (node, bool) {
	ok := true
	unit := func(u name) name {
		a, isParam := args[u.text]
		if !isParam {
			return u
		}
		if c, bare := a.(*callNode); bare && c.args == nil {
			return c.name
		}
		if n, isNone := a.(*noneNode); isNone {
			return name{span: n.span, text: "none"}
		}
		t.report("wrong-type", "a unit word goes here, such as target", a.pos())
		ok = false
		return u
	}
	var sub func(n node) node
	subArgs := func(l *argList) *argList {
		if l == nil {
			return nil
		}
		out := &argList{span: l.span, items: make([]arg, len(l.items))}
		for i, a := range l.items {
			out.items[i] = arg{span: a.span, name: a.name, value: sub(a.value)}
		}
		return out
	}
	subAll := func(nodes []node) []node {
		out := make([]node, len(nodes))
		for i, x := range nodes {
			out[i] = sub(x)
		}
		return out
	}
	sub = func(n node) node {
		switch x := n.(type) {
		case *callNode:
			if a, isParam := args[x.name.text]; isParam && x.args == nil {
				return a
			}
			return &callNode{span: x.span, name: x.name, args: subArgs(x.args)}
		case *memberNode:
			return &memberNode{span: x.span, object: unit(x.object), name: x.name}
		case *accessorNode:
			out := &accessorNode{span: x.span, head: x.head, subject: sub(x.subject), prop: x.prop, args: subArgs(x.args)}
			if x.unit != nil {
				u := unit(*x.unit)
				out.unit = &u
			}
			return out
		case *parenNode:
			return &parenNode{span: x.span, expr: sub(x.expr)}
		case *notNode:
			return &notNode{span: x.span, operand: sub(x.operand)}
		case *logicNode:
			return &logicNode{span: x.span, op: x.op, operands: subAll(x.operands)}
		case *compareNode:
			return &compareNode{span: x.span, op: x.op, left: sub(x.left), right: sub(x.right)}
		case *mathNode:
			return &mathNode{span: x.span, op: x.op, left: sub(x.left), right: sub(x.right)}
		case *listNode:
			return &listNode{span: x.span, items: subAll(x.items)}
		}
		return n
	}
	out := sub(n)
	return out, ok
}

// Takes in the defs of the files the use lines read, and of the files those use in turn, each file once (SPEC.md §9).
// A def named as one of this file's, or as one another used file declares, is duplicate-def at the use line that
// brings it.
func (f *typedFile) importDefs() {
	t := f.typer
	base := len(t.lines.src) + 1
	taken := map[string]bool{}
	for _, s := range f.tree.top {
		use, ok := s.(*useStmt)
		if !ok || f.used[use] == nil {
			continue
		}
		var clashes []string
		var take func(u *typedFile, path string)
		take = func(u *typedFile, path string) {
			if taken[path] {
				return
			}
			taken[path] = true
			for _, name := range slices.Sorted(maps.Keys(u.typer.defs.byName)) {
				d := u.typer.defs.byName[name]
				if t.defs.byName[name] != nil {
					clashes = append(clashes, fmt.Sprintf("%s is in both this file and %s", name, path))
					continue
				}
				if other := t.defs.imported[name]; other != nil {
					clashes = append(clashes, fmt.Sprintf("%s is in both %s and %s", name, other.path, path))
					continue
				}
				in := &importedDef{defStmt: shiftedDef(d, base), path: path, lines: u.typer.lines, base: base}
				t.defs.imported[name] = in
				t.defs.from[in.defStmt] = in
				t.defs.cyclic[in.defStmt] = u.typer.defs.cyclic[d]
			}
			base += len(u.typer.lines.src) + 1
			for _, s := range u.tree.top {
				if x, ok := s.(*useStmt); ok && u.used[x] != nil {
					take(u.used[x], resolveUse(path, x.path.value))
				}
			}
		}
		take(f.used[use], f.usedPaths[use])
		if len(clashes) > 0 {
			t.report("duplicate-def", "the def "+strings.Join(clashes, "; the def "), use.path.span)
		}
	}
}

// A copy of d with every span of its name, parameters and value moved on by by.
func shiftedDef(d *defStmt, by int) *defStmt {
	at := func(s span) span { return span{start: s.start + by, end: s.end + by} }
	moved := func(n name) name { return name{span: at(n.span), text: n.text} }
	var shift func(n node) node
	args := func(l *argList) *argList {
		if l == nil {
			return nil
		}
		out := &argList{span: at(l.span), items: make([]arg, len(l.items))}
		for i, a := range l.items {
			out.items[i] = arg{span: at(a.span), value: shift(a.value)}
			if a.name != nil {
				n := moved(*a.name)
				out.items[i].name = &n
			}
		}
		return out
	}
	all := func(nodes []node) []node {
		out := make([]node, len(nodes))
		for i, x := range nodes {
			out[i] = shift(x)
		}
		return out
	}
	shift = func(n node) node {
		switch x := n.(type) {
		case *literalNode:
			return &literalNode{span: at(x.span), text: x.text}
		case *stringNode:
			return &stringNode{span: at(x.span), value: x.value}
		case *boolNode:
			return &boolNode{span: at(x.span), value: x.value}
		case *noneNode:
			return &noneNode{span: at(x.span)}
		case *varNode:
			return &varNode{span: at(x.span), name: x.name, braced: x.braced}
		case *placeholderNode:
			return &placeholderNode{span: at(x.span), name: x.name, braced: x.braced}
		case *rawNode:
			return &rawNode{span: at(x.span), json: x.json}
		case *parenNode:
			return &parenNode{span: at(x.span), expr: shift(x.expr)}
		case *notNode:
			return &notNode{span: at(x.span), operand: shift(x.operand)}
		case *logicNode:
			return &logicNode{span: at(x.span), op: x.op, operands: all(x.operands)}
		case *compareNode:
			return &compareNode{span: at(x.span), op: x.op, left: shift(x.left), right: shift(x.right)}
		case *mathNode:
			return &mathNode{span: at(x.span), op: x.op, left: shift(x.left), right: shift(x.right)}
		case *callNode:
			return &callNode{span: at(x.span), name: moved(x.name), args: args(x.args)}
		case *memberNode:
			return &memberNode{span: at(x.span), object: moved(x.object), name: moved(x.name)}
		case *accessorNode:
			out := &accessorNode{span: at(x.span), head: moved(x.head), subject: shift(x.subject), prop: moved(x.prop), args: args(x.args)}
			if x.unit != nil {
				u := moved(*x.unit)
				out.unit = &u
			}
			return out
		case *listNode:
			return &listNode{span: at(x.span), items: all(x.items)}
		}
		return n
	}
	out := *d
	out.name = moved(d.name)
	out.params = make([]name, len(d.params))
	for i, p := range d.params {
		out.params[i] = moved(p)
	}
	out.value = shift(d.value)
	return &out
}

// A node as it is written, laid out as SIML lays out an expression (SPEC.md §12.3): single spaces around operators,
// `, ` between arguments. src is the text, for an action, which is written as its text with its spacing collapsed.
func nodeText(n node, src string) string {
	all := func(nodes []node, sep string) string {
		texts := make([]string, len(nodes))
		for i, x := range nodes {
			texts[i] = nodeText(x, src)
		}
		return strings.Join(texts, sep)
	}
	args := func(l *argList) string {
		if l == nil {
			return ""
		}
		texts := make([]string, len(l.items))
		for i, a := range l.items {
			texts[i] = nodeText(a.value, src)
			if a.name != nil {
				texts[i] = a.name.text + "=" + texts[i]
			}
		}
		return "(" + strings.Join(texts, ", ") + ")"
	}
	switch x := n.(type) {
	case *literalNode:
		return x.text
	case *stringNode:
		return jsonString(x.value)
	case *boolNode:
		return strconv.FormatBool(x.value)
	case *noneNode:
		return "none"
	case *varNode:
		if x.braced {
			return braced("$", x.name)
		}
		return "$" + x.name
	case *placeholderNode:
		if x.braced {
			return braced("%", x.name)
		}
		return "%" + x.name
	case *rawNode:
		return rawText(x.json)
	case *parenNode:
		return "(" + nodeText(x.expr, src) + ")"
	case *notNode:
		return "not " + nodeText(x.operand, src)
	case *logicNode:
		return all(x.operands, " "+x.op+" ")
	case *compareNode:
		return nodeText(x.left, src) + " " + x.op + " " + nodeText(x.right, src)
	case *mathNode:
		return nodeText(x.left, src) + " " + x.op + " " + nodeText(x.right, src)
	case *callNode:
		return x.name.text + args(x.args)
	case *memberNode:
		return x.object.text + "." + x.name.text
	case *accessorNode:
		text := x.head.text + "(" + nodeText(x.subject, src) + ")." + x.prop.text + args(x.args)
		if x.unit != nil {
			text = x.unit.text + "." + text
		}
		return text
	case *listNode:
		return "[" + all(x.items, ", ") + "]"
	}
	at := n.pos()
	return strings.Join(strings.Fields(src[at.start:at.end]), " ")
}
