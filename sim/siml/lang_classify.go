package siml

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// What each word of the file is, read off the syntax tree with the schema: the base of hover and semantic tokens.

// A word and what it is; role says which, and the other fields what it is about.
type wordRole struct {
	span
	role        string // section, spec, field, kind, head, prop, unit, enum, slot, name, variable, placeholder, def or parameter.
	word        string // section, spec, head, unit, enum, slot: the word.
	field       *field
	kind        *aplKind // kind, prop.
	ref         *nameRef // name.
	declaration bool     // variable, def, parameter.
}

func (d *Doc) classify() []wordRole {
	d.wordsOnce.Do(func() { d.words = d.readWords() })
	return d.words
}

func (d *Doc) readWords() []wordRole {
	tree := d.tree()
	if tree == nil {
		return nil
	}
	ix := d.ix
	var out []wordRole
	refs := map[int]*nameRef{}
	for i := range d.typed.typer.refs {
		r := &d.typed.typer.refs[i]
		if _, ok := refs[r.start]; !ok {
			refs[r.start] = r
		}
	}
	sections := sectionsByStatement(tree)
	push := func(at span, r wordRole) {
		r.span = at
		out = append(out, r)
	}
	var inDef *defStmt // The def whose value is being read, whose parameters its words may name.
	parameter := func(n name) bool {
		if inDef == nil || !inDef.isParam(n.text) {
			return false
		}
		push(n.span, wordRole{role: "parameter"})
		return true
	}
	var term func(t node, f *field, m *message)
	var expr func(e node)
	var action func(a *actionNode)
	args := func(a *argList, m *message, kind *aplKind) {
		if a == nil {
			return
		}
		var positional *field
		if m != nil && kind != nil {
			if kind.ListField != "" {
				positional = m.byProto[kind.ListField]
			}
			if positional == nil && kind.Positional != "" {
				positional = m.byProto[kind.Positional]
			}
		}
		for _, x := range a.items {
			f := positional
			if x.name != nil {
				f = nil
				if m != nil {
					f = m.byName[x.name.text]
				}
				if f != nil {
					push(x.name.span, wordRole{role: "field", field: f})
				}
			}
			term(x.value, f, m)
		}
	}
	term = func(t node, f *field, m *message) {
		switch x := t.(type) {
		case *actionNode:
			action(x)
			return
		case *listNode:
			for _, item := range x.items {
				term(item, f, m)
			}
			return
		case *callNode:
			// A bare word naming a bool field of the message sets it.
			if own := messageField(m, x.name.text); x.args == nil && own != nil && own.Printed == "bool" {
				push(x.name.span, wordRole{role: "field", field: own})
				return
			}
			// A message value is written `field(...)` with the field's own name.
			if own := messageField(m, x.name.text); own != nil && f != nil && (f.Printed == "message" || f.Printed == "indexed") && x.args != nil {
				push(x.name.span, wordRole{role: "field", field: own})
				if own.Printed == "indexed" {
					for _, a := range x.args.items {
						if a.name != nil {
							push(a.name.span, wordRole{role: "enum", word: a.name.text})
						}
					}
				} else {
					args(x.args, ix.messages[own.Type], nil)
				}
				return
			}
		}
		expr(t)
	}
	call := func(e *callNode, isAction bool) {
		if ref := refs[e.name.start]; ref != nil {
			push(e.name.span, wordRole{role: "name", ref: ref})
			return
		}
		if e.args == nil && parameter(e.name) {
			return
		}
		if d.typed.typer.defs.lookup(e.name.text) != nil {
			push(e.name.span, wordRole{role: "def"})
			args(e.args, nil, nil)
			return
		}
		var kind *aplKind
		if isAction {
			kind = ix.actionKinds[e.name.text]
		}
		if kind == nil {
			kind = ix.valueKinds[e.name.text]
		}
		switch {
		case kind != nil:
			push(e.name.span, wordRole{role: "kind", kind: kind})
			args(e.args, kind.message, kind)
		case ix.unitReference(e.name.text) != nil:
			push(e.name.span, wordRole{role: "unit", word: e.name.text})
		case e.args == nil:
			push(e.name.span, wordRole{role: "enum", word: e.name.text})
		}
	}
	expr = func(e node) {
		switch x := e.(type) {
		case *varNode:
			push(x.span, wordRole{role: "variable"})
		case *placeholderNode:
			push(x.span, wordRole{role: "placeholder"})
		case *parenNode:
			expr(x.expr)
		case *notNode:
			expr(x.operand)
		case *logicNode:
			for _, o := range x.operands {
				expr(o)
			}
		case *compareNode:
			expr(x.left)
			expr(x.right)
		case *mathNode:
			expr(x.left)
			expr(x.right)
		case *memberNode:
			if ref := refs[x.name.start]; ref != nil {
				push(x.name.span, wordRole{role: "name", ref: ref})
				return
			}
			if x.object.text == "other" {
				push(x.name.span, wordRole{role: "enum", word: x.name.text})
				return
			}
			if !parameter(x.object) && ix.unitReference(x.object.text) != nil {
				push(x.object.span, wordRole{role: "unit", word: x.object.text})
			}
			if kind := ix.valueKinds[x.name.text]; kind != nil {
				push(x.name.span, wordRole{role: "kind", kind: kind})
			}
		case *accessorNode:
			if x.unit != nil && !parameter(*x.unit) {
				push(x.unit.span, wordRole{role: "unit", word: x.unit.text})
			}
			push(x.head.span, wordRole{role: "head", word: x.head.text})
			term(x.subject, nil, nil)
			kind := ix.accessors[x.head.text+"."+x.prop.text]
			var m *message
			if kind != nil {
				push(x.prop.span, wordRole{role: "prop", kind: kind})
				m = kind.message
			}
			args(x.args, m, kind)
		case *callNode:
			call(x, false)
		}
	}
	var step func(s node)
	step = func(s node) {
		switch x := s.(type) {
		case *castStep:
			expr(x.spell)
			expr(x.target)
		case *channelStep:
			expr(x.spell)
			expr(x.target)
			expr(x.until)
		case *waitStep:
			expr(x.value)
		case *moveStep:
			expr(x.value)
		case *groupStep:
			for _, b := range x.with {
				push(b.placeholder.span, wordRole{role: "placeholder"})
				expr(b.value)
			}
		case *scheduleStep:
			step(x.step)
		case *sequenceStep:
			for _, a := range x.actions {
				action(a)
			}
		case *callNode:
			call(x, true)
		}
	}
	action = func(a *actionNode) {
		if a == nil {
			return
		}
		step(a.step)
		expr(a.condition)
	}
	entry := func(def *sectionDef, written string) *field {
		if def == nil {
			return nil
		}
		return entryField(ix, def, d.specInfo(), written)
	}
	var statement func(s stmt, def *sectionDef)
	statement = func(s stmt, def *sectionDef) {
		switch x := s.(type) {
		case *headerStmt:
			if x.group == nil {
				push(x.name.span, wordRole{role: "section", word: x.name.text})
			}
			if x.body != nil {
				statement(x.body, ix.sections[x.name.text])
			}
		case *specStmt:
			push(x.spec.span, wordRole{role: "spec", word: x.spec.text})
		case *useStmt:
			for _, b := range x.with {
				push(b.variable.span, wordRole{role: "variable"})
				expr(b.value)
			}
		case *varStmt:
			push(x.variable.span, wordRole{role: "variable", declaration: true})
			expr(x.value)
		case *rowStmt:
			expr(x.doAt)
			action(x.action)
		case *whenStmt:
			expr(x.condition)
		case *defStmt:
			push(x.name.span, wordRole{role: "def", declaration: true})
			for _, p := range x.params {
				push(p.span, wordRole{role: "parameter", declaration: true})
			}
			inDef = x
			expr(x.value)
			inDef = nil
		case *slotRowStmt:
			push(x.slot.span, wordRole{role: "slot", word: x.slot.text})
			for _, part := range []*name{x.item, x.suffix, x.enchant} {
				if part == nil {
					continue
				}
				if ref := refs[part.start]; ref != nil {
					push(ref.span, wordRole{role: "name", ref: ref})
				}
			}
		case *fieldEntryStmt:
			f := entry(def, x.name.text)
			if f != nil {
				push(x.name.span, wordRole{role: "field", field: f})
			}
			for _, v := range x.values {
				term(v, f, nil)
			}
		case *flagListStmt:
			for _, fl := range x.flags {
				f := entry(def, fl.name.text)
				if f != nil {
					push(fl.name.span, wordRole{role: "field", field: f})
				}
				if fl.value != nil {
					term(fl.value, f, nil)
				}
			}
		case *messageEntryStmt:
			f := entry(def, x.name.text)
			if f == nil {
				return
			}
			push(x.name.span, wordRole{role: "field", field: f})
			if f.Printed == "indexed" {
				for _, a := range x.args.items {
					if a.name != nil {
						push(a.name.span, wordRole{role: "enum", word: a.name.text})
					}
				}
			} else {
				args(x.args, ix.messages[f.Type], nil)
			}
		}
	}
	for _, s := range tree.statements {
		var def *sectionDef
		if sec := sections[s]; sec != nil {
			def = ix.sections[sec.name]
		}
		statement(s, def)
	}
	slices.SortStableFunc(out, func(a, b wordRole) int { return a.start - b.start })
	return out
}

// The field of m written name, nil for none or no message.
func messageField(m *message, name string) *field {
	if m == nil {
		return nil
	}
	return m.byName[name]
}

// The innermost word at an offset.
func wordAt(words []wordRole, offset int) *wordRole {
	var best *wordRole
	for i := range words {
		w := &words[i]
		if w.holds(offset) && (best == nil || w.end-w.start < best.end-best.start) {
			best = w
		}
	}
	return best
}

// --- Semantic tokens.

// The token types and modifiers the analysis gives. The grammar colours keywords, literals, strings and comments;
// these colour what only the schema and the table know: sections, fields, kinds, units, enum values and names.
var (
	tokenTypes     = []string{"namespace", "class", "property", "function", "method", "variable", "parameter", "enumMember", "type"}
	tokenModifiers = []string{"declaration", "readonly", "defaultLibrary"}
)

func tokenOf(w wordRole) (string, []string) {
	switch w.role {
	case "section":
		return "namespace", nil
	case "spec":
		return "class", nil
	case "field":
		return "property", nil
	case "kind":
		return "function", nil
	case "head":
		return "function", []string{"defaultLibrary"}
	case "prop":
		return "method", nil
	case "unit":
		return "variable", []string{"readonly", "defaultLibrary"}
	case "enum", "slot":
		return "enumMember", nil
	case "variable":
		if w.declaration {
			return "variable", []string{"declaration"}
		}
		return "variable", nil
	case "placeholder":
		return "parameter", nil
	case "def", "parameter":
		typ := map[string]string{"def": "function", "parameter": "parameter"}[w.role]
		if w.declaration {
			return typ, []string{"declaration"}
		}
		return typ, nil
	case "name":
		if w.ref.ns.SpellLike() {
			return "type", nil
		}
		return "type", []string{"readonly"}
	}
	return "", nil
}

// The words the schema and the table give a meaning, in file order: each on one line, none empty, none overlapping.
func (d *Doc) tokens() []Token {
	var out []Token
	reached := 0 // The offset the last token ends at.
	for _, w := range d.classify() {
		typ, modifiers := tokenOf(w)
		r := d.rangeOf(w.span)
		if typ == "" || r.Start.Line != r.End.Line || r.End.Col <= r.Start.Col || w.start < reached {
			continue
		}
		reached = w.end
		out = append(out, Token{Range: r, Type: typ, Modifiers: modifiers})
	}
	return out
}

// --- Hover.

func (d *Doc) hover(at Position, c *DatabaseCatalog) (Hover, bool) {
	if d.tree() == nil {
		return Hover{}, false
	}
	offset := d.Offset(at)
	w := wordAt(d.classify(), offset)
	if o, ok := d.occurrenceAt(offset); ok && (w == nil || slices.Contains([]string{"variable", "placeholder", "def", "parameter"}, w.role)) {
		what := fmt.Sprintf("%s **%s**", o.kind, o.name)
		switch {
		case o.kind == "parameter":
			what += " of def " + o.scope.group
		case o.scope.inGroup:
			what += fmt.Sprintf(` of group "%s"`, o.scope.group)
		}
		parts := []string{what}
		if home, declaration, ok := d.declarationOf(o); ok {
			text := home.lineText(declaration.start)
			if def := home.typed.typer.defs.byName[o.name]; o.kind == "def" && def != nil {
				text = strings.Join(statementLines(defLine(def, home.src, 0)), "\n")
			}
			parts = append(parts, fence(text, "siml"))
			if expansion := d.expansionAt(o.span); o.kind == "def" && expansion != "" {
				parts = append(parts, "Reads here as", fence(expansion, "siml"))
			}
		} else if o.file != "" {
			parts = append(parts, fmt.Sprintf("Not declared in %s.", filepath.Base(o.file)))
		} else {
			parts = append(parts, "Not declared in this file.")
		}
		return Hover{Markdown: strings.Join(parts, "\n\n"), Range: d.rangeOf(o.span)}, true
	}
	if w == nil {
		return d.whenHover(offset)
	}
	text, spellID := d.describe(*w, c)
	if text == "" {
		return Hover{}, false
	}
	return Hover{Markdown: text, Range: d.rangeOf(w.span), SpellID: spellID}, true
}

// The value the def call whose name is at reads as, printed; "" unless the call is read once, as a call in a def's
// value is read for each call of that def.
func (d *Doc) expansionAt(at span) string {
	var value *proto.APLValue
	count := 0
	for v, c := range d.typed.typer.defs.values {
		if c.name.span == at {
			value, count = v, count+1
		}
	}
	if count != 1 {
		return ""
	}
	return d.printValue(value, at.start)
}

// A value as the printer writes it, its variables as the statement at offset sees them.
func (d *Doc) printValue(v *proto.APLValue, offset int) string {
	tree := d.tree()
	sc := scope{}
	sections := sectionsByStatement(tree)
	for _, s := range tree.statements {
		if s.base().holds(offset) {
			sc = scopeOf(sections[s])
		}
	}
	p := &printer{ix: d.ix, names: d.names, class: d.typed.typer.class, declarations: declaredNames(d.typed.declarations)}
	return strings.Join(p.value(v.ProtoReflect(), precTop, sc), "")
}

// On the word when of a when line: the condition its rows join theirs to, the blocks around it included (SPEC.md
// §6.9).
func (d *Doc) whenHover(offset int) (Hover, bool) {
	for _, s := range d.tree().statements {
		w, ok := s.(*whenStmt)
		if !ok || offset < w.start || offset > w.condition.pos().start {
			continue
		}
		var blocks []*whenStmt
		for b := w; b != nil; b = b.parent {
			blocks = append([]*whenStmt{b}, blocks...)
		}
		joined := blockConjuncts(d.typed.whens, blocks)
		if slices.Contains(joined, nil) {
			return Hover{}, false
		}
		condition := andOf(joined)
		if len(joined) == 1 {
			condition = joined[0]
		}
		text := "**when** block: each row in it reads with this condition before its own"
		if slices.ContainsFunc(blocks, func(b *whenStmt) bool { return b.hide }) {
			text += ", hidden"
		}
		parts := []string{text, fence(d.printValue(condition, w.start), "siml")}
		return Hover{Markdown: strings.Join(parts, "\n\n"), Range: d.rangeOf(span{w.start, w.condition.pos().start})}, true
	}
	return Hover{}, false
}

// The text of the line holding offset, trimmed.
func (d *Doc) lineText(offset int) string {
	line := d.lines.line(offset)
	end := len(d.src)
	if line+1 < len(d.lines.starts) {
		end = d.lines.starts[line+1] - 1
	}
	return strings.TrimSpace(d.src[d.lines.starts[line]:end])
}

func fence(text, language string) string { return "```" + language + "\n" + text + "\n```" }

// What a word is, and the spell whose game files row goes with it.
func (d *Doc) describe(w wordRole, c *DatabaseCatalog) (string, int32) {
	ix := d.ix
	switch w.role {
	case "name":
		return d.describeName(w.ref, c)
	case "kind", "prop":
		return d.kindText(w.kind), 0
	case "head":
		var props []string
		for _, k := range ix.headKinds(w.word) {
			props = append(props, k.Accessor.Prop)
		}
		return fmt.Sprintf("accessor head **%s**\n\n`%[1]s(…).%s`", w.word, strings.Join(props, "` · `")), 0
	case "field":
		f := w.field
		return fmt.Sprintf("field **%s**: %s\n\n`%s` in the JSON", f.Siml, fieldType(f), f.JSON), 0
	case "unit":
		u := ix.unitReference(w.word)
		if u == nil {
			return "", 0
		}
		return fmt.Sprintf("unit **%s**\n\n%s", w.word, fence(ix.jsonText(u.ProtoReflect()), "json")), 0
	case "slot":
		name := ""
		if v := ix.enums[slotEnum].byName[w.word]; v != nil {
			name = v.Name
		}
		return fmt.Sprintf("slot **%s**: %s", w.word, name), 0
	case "section":
		def := ix.sections[w.word]
		if def == nil {
			return "", 0
		}
		category := ""
		if def.Category != "" {
			category = ", category " + def.Category
		}
		return fmt.Sprintf("section **%s**%s\n\n%s", def.Name, category, strings.Join(def.Body, ", ")), 0
	case "spec":
		spec, ok := ix.specs[w.word]
		if !ok {
			return "", 0
		}
		return fmt.Sprintf("spec **%s**: %s, %s", spec.Field.Name(), spec.Spec, spec.Class), 0
	}
	return "", 0
}

func (d *Doc) kindText(kind *aplKind) string {
	parts := []string{fmt.Sprintf("kind **%s** (%s)", kind.Kind, kind.Form)}
	var fields []string
	for _, f := range kind.message.fields {
		list := ""
		if f.isList() {
			list = " list"
		}
		fields = append(fields, fmt.Sprintf("- `%s`: %s%s", f.Siml, f.Printed, list))
	}
	if len(fields) > 0 {
		parts = append(parts, strings.Join(fields, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

// What a namespace holds, without its article: "aura on a target".
func thingOf(ns Namespace) string {
	thing, _ := strings.CutPrefix(namespaceThings[ns], "a ")
	thing, _ = strings.CutPrefix(thing, "an ")
	return thing
}

func (d *Doc) describeName(ref *nameRef, c *DatabaseCatalog) (string, int32) {
	names := d.names
	if ref.ns.SpellLike() {
		entry, ok := names.spellEntry(d.class(), ref.ns, ref.name)
		if !ok {
			return fmt.Sprintf("no %s is named **%s** for this file", thingOf(ref.ns), ref.name), 0
		}
		game := ref.name
		if c != nil {
			if name, ok := c.SpellName(entry.SpellID); ok && name != "" {
				game = name
			}
		}
		parts := []string{fmt.Sprintf("spell %d", entry.SpellID)}
		if entry.Tag != 0 {
			parts = append(parts, fmt.Sprintf("tag %d", entry.Tag))
		}
		if entry.Rank != 0 {
			parts = append(parts, fmt.Sprintf("rank %d", entry.Rank))
		}
		registered := "registered by no build"
		if len(entry.RegisteredBy) > 0 {
			registered = "registered by " + strings.Join(entry.RegisteredBy, ", ")
		}
		return fmt.Sprintf("**%s** · %s · %s\n\n%s", game, namespaceThings[ref.ns], strings.Join(parts, ", "), registered), entry.SpellID
	}
	switch ref.ns {
	case NamespaceItem:
		id, ok := names.Item(ref.name)
		if !ok {
			return fmt.Sprintf("no item is named **%s**", ref.name), 0
		}
		return d.itemText(c, id), 0
	case NamespaceEnchant:
		entry, ok := names.Enchants[ref.name]
		if !ok {
			return fmt.Sprintf("no enchant is named **%s**", ref.name), 0
		}
		var rows []CatalogEnchant
		if c != nil {
			rows = c.EnchantRows(entry.EffectID)
		}
		title := ref.name
		if len(rows) > 0 {
			title = rows[0].Name
		}
		parts := []string{fmt.Sprintf("**%s** · enchant effect %d", title, entry.EffectID), "fits " + strings.Join(d.slotWords(entry.Slots), ", ")}
		var seen []string
		for _, r := range rows[min(1, len(rows)):] {
			if !slices.Contains(seen, r.Name) {
				seen = append(seen, r.Name)
			}
		}
		return strings.Join(append(parts, seen...), "\n\n"), 0
	case NamespaceSuffix:
		id, ok := names.Suffix(ref.name)
		if !ok {
			return fmt.Sprintf("no random suffix is named **%s**", ref.name), 0
		}
		title := ref.name
		var stats []string
		if c != nil {
			if s, ok := c.Suffix(id); ok {
				title = s.Name
				for i, v := range s.Stats {
					if v == 0 {
						continue
					}
					stat := strconv.Itoa(i)
					if e := d.ix.enums[statEnum].byNumber[int32(i)]; e != nil {
						stat = e.Short
					}
					stats = append(stats, stat+" "+strconv.FormatFloat(v, 'f', -1, 64))
				}
			}
		}
		parts := []string{fmt.Sprintf("**%s** · random suffix %d", title, id)}
		if len(stats) > 0 {
			parts = append(parts, strings.Join(stats, ", "))
		}
		return strings.Join(parts, "\n\n"), 0
	}
	return ref.name, 0
}

// The short names of slots, such as head.
func (d *Doc) slotWords(slots []proto.ItemSlot) []string {
	out := make([]string, len(slots))
	for i, s := range slots {
		out[i] = d.slotWord(int32(s))
	}
	return out
}

func (d *Doc) itemText(c *DatabaseCatalog, id int32) string {
	var item CatalogItem
	found := false
	if c != nil {
		item, found = c.Item(id)
	}
	title := strconv.Itoa(int(id))
	if c != nil {
		if name, ok := c.ItemName(id); ok {
			title = name
		}
	}
	parts := []string{fmt.Sprintf("**%s** · item %d", title, id)}
	if found {
		slots := d.slotWords(item.slots())
		var facts []string
		if item.Ilvl != 0 {
			facts = append(facts, fmt.Sprintf("ilvl %d", item.Ilvl))
		}
		if len(slots) > 0 {
			facts = append(facts, strings.Join(slots, " or "))
		}
		if item.Phase != 0 {
			facts = append(facts, fmt.Sprintf("phase %d", item.Phase))
		}
		if len(facts) > 0 {
			parts = append(parts, strings.Join(facts, " · "))
		}
		if item.SetName != "" {
			parts = append(parts, "set: "+item.SetName)
		}
		if len(item.Effects) > 0 {
			parts = append(parts, strings.Join(item.Effects, ", "))
		}
		if n := len(c.Suffixes(id)); n > 0 {
			parts = append(parts, fmt.Sprintf("%d random suffixes", n))
		}
	}
	return strings.Join(parts, "\n\n")
}

// A slot's short name, such as head; its number when the enum has none.
func (d *Doc) slotWord(number int32) string {
	if s := shortName(slotEnum, number); s != "" {
		return s
	}
	return strconv.Itoa(int(number))
}
