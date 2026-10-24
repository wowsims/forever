package siml

import (
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
)

// Completion: what the cursor context expects, offered from the schema, the name table, the item database and the
// file's own variables, groups, sequences and placeholders.

var actionWords = [][3]string{ // label, insert, detail
	{"cast", "cast ", "cast <spell> [on <unit>]"},
	{"channel", "channel ", "channel <spell> [on <unit>] [until <value>] [recast]"},
	{"wait", "wait ", "wait <time>"},
	{"wait until", "wait until ", "wait until <condition>"},
	{"move", "move ", "move <range from the target> [recheck]"},
	{"move for", "move for ", "move for <time>"},
	{"group", `group "`, `group "Name" [with %placeholder = <value>, …]`},
	{"schedule", `schedule "`, `schedule "0s, 60s" <action>`},
	{"sequence", `sequence "`, `sequence "Name" [<action>, …]`},
	{"none", "none", "no action"},
}

// The indexed unit words, offered at their first index.
var indexedUnits = []string{"target1", "player1", "pet1"}

var topLines = [][3]string{ // label, insert, detail
	{"siml 1", "siml 1", "the format version, the first line"},
	{"spec", "spec ", "the spec, whose class reads the file's spell names"},
	{"talents", `talents "`, "the talents string"},
	{"use", `use "`, "take the sections of another file"},
	{"var", "var $", "a rotation value variable"},
	{"def", "def ", "def name(p, …) = <value>: a value each call reads as, its parameters replaced"},
}

// The item fields of ConsumesSpec, by proto name, and the consumable types each offers; a field not listed offers
// every consumable.
var consumableTypes = map[string][]proto.ConsumableType{
	"resource_pot_id":    {proto.ConsumableType_ConsumableTypePotion},
	"defensive_pot_id":   {proto.ConsumableType_ConsumableTypePotion},
	"offensive_pot_id":   {proto.ConsumableType_ConsumableTypePotion},
	"potions":            {proto.ConsumableType_ConsumableTypePotion},
	"flask_id":           {proto.ConsumableType_ConsumableTypeFlask},
	"food_id":            {proto.ConsumableType_ConsumableTypeFood},
	"explosive_id":       {proto.ConsumableType_ConsumableTypeExplosive},
	"mhImbue_id":         {proto.ConsumableType_ConsumableTypeImbue},
	"ohImbue_id":         {proto.ConsumableType_ConsumableTypeImbue},
	"elixir_ids":         {proto.ConsumableType_ConsumableTypeElixir},
	"conjured_id":        {proto.ConsumableType_ConsumableTypeConjured},
	"conjured_items":     {proto.ConsumableType_ConsumableTypeConjured},
	"mh_weapon_imbue_id": {proto.ConsumableType_ConsumableTypeWeaponImbue},
	"oh_weapon_imbue_id": {proto.ConsumableType_ConsumableTypeWeaponImbue},
}

// Completion opens again after an item that ends where the next part starts: `cast `, `aura(`, `target.`, `item.`.
var retriggers = regexp.MustCompile(`[ (.="]$`)

type completer struct {
	d       *Doc
	ctx     *cursorContext
	c       *DatabaseCatalog
	defined []occurrence
	read    bool
	items   []Completion
	seen    map[string]bool
}

// What a completion is written with besides its label and kind.
type offerAs struct {
	insert, detail, documentation, filterText string
	rank                                      int // Where it sorts: lower first; 0 is 5.
	ranked                                    bool
}

func (d *Doc) complete(at Position, c *DatabaseCatalog) []Completion {
	if d.tree() == nil {
		return nil
	}
	ctx := d.cursorAt(d.Offset(at))
	if ctx == nil {
		return nil
	}
	b := &completer{d: d, ctx: ctx, c: c, seen: map[string]bool{}}
	for _, w := range ctx.wants {
		b.offer(w)
	}
	return b.items
}

// The file's variables, groups, sequences and placeholders. A bracket left open, as one is while its call is typed,
// joins every line below it into its statement; the file is then read without the lines that open one, as often as a
// continuation line left behind opens another.
func (b *completer) definitions() []occurrence {
	if b.read {
		return b.defined
	}
	b.read = true
	read, src := b.d, b.d.src
	for open := unclosedLines(src); len(open) > 0; open = unclosedLines(src) {
		lines := strings.Split(src, "\n")
		for i := range lines {
			if open[i] {
				lines[i] = ""
			}
		}
		src = strings.Join(lines, "\n")
	}
	if src != b.d.src {
		read = b.d.analyze(b.d.path, src)
	}
	b.defined = read.occurrences()
	return b.defined
}

// The 0-based lines of the brackets the text never closes. A statement ends only once its brackets close, so those
// are the brackets still open when the file's tokens are read in order, as statementsOf reads them.
func unclosedLines(src string) map[int]bool {
	tokens, _, lines := lex(src)
	var open []token
	for _, t := range tokens {
		if t.kind == tokPunct && (t.text == "(" || t.text == "[") {
			open = append(open, t)
		} else if t.kind == tokPunct && (t.text == ")" || t.text == "]") && len(open) > 0 {
			open = open[:len(open)-1]
		}
	}
	out := map[int]bool{}
	for _, t := range open {
		out[lines.line(t.start)] = true
	}
	return out
}

func (b *completer) add(label string, kind CompletionKind, o offerAs) {
	key := fmt.Sprint(kind) + "\x00" + label
	if b.seen[key] {
		return
	}
	b.seen[key] = true
	text := o.insert
	if text == "" {
		text = label
	}
	rank := o.rank
	if !o.ranked {
		rank = 5
	}
	if b.ctx.string && !strings.HasPrefix(text, `"`) {
		return
	}
	b.items = append(b.items, Completion{
		Label: label, Kind: kind, Detail: o.detail, Documentation: o.documentation, InsertText: text,
		Replace:  b.d.rangeOf(span{b.ctx.start, b.ctx.end}),
		SortText: fmt.Sprintf("%d%s", rank, label), FilterText: o.filterText, Retrigger: retriggers.MatchString(text),
	})
}

func ranked(rank int) offerAs { return offerAs{rank: rank, ranked: true} }

func (o offerAs) with(insert, detail string) offerAs {
	o.insert, o.detail = insert, detail
	return o
}

func (b *completer) offer(w want) {
	ix := b.d.ix
	switch w.t {
	case "line":
		b.line(w.column0)
	case "step":
		b.steps(w.when)
	case "time":
		b.add("-1s", CompletionValue, ranked(1).with("-1s ", "a time before the pull"))
		b.variables()
	case "value":
		b.values()
	case "words":
		for _, word := range w.words {
			insert := word + " "
			if word == "1" || word == "+" || word == ":" {
				insert = word
			}
			b.add(word, CompletionKeyword, ranked(1).with(insert, ""))
		}
	case "ref":
		b.refs(w.ns)
	case "usable":
		b.usable()
	case "other":
		for _, v := range ix.enums[otherEnum].values {
			if v.Number != 0 {
				b.add(v.Short, CompletionEnumMember, offerAs{detail: v.Name})
			}
		}
	case "unit":
		b.units(false)
	case "member":
		b.members()
	case "enum":
		if e := ix.enums[w.enum]; e != nil {
			for _, v := range e.values {
				b.add(v.Short, CompletionEnumMember, offerAs{detail: v.Name})
			}
		}
	case "bool":
		b.add("true", CompletionKeyword, offerAs{})
		b.add("false", CompletionKeyword, offerAs{})
	case "args":
		b.args(w.message, w.kind, w.used, w.skip)
	case "props":
		for _, k := range ix.headKinds(w.head) {
			b.add(k.Accessor.Prop, CompletionMethod, offerAs{detail: k.Kind, documentation: fieldsText(k.message, k.Accessor.Subject)})
		}
	case "strings":
		for _, o := range b.definitions() {
			if o.kind == w.of && o.definition {
				b.add(jsonString(o.name), CompletionModule, offerAs{detail: w.of})
			}
		}
	case "placeholders":
		for _, o := range b.definitions() {
			if o.kind == "placeholder" && o.definition && o.scope == w.group {
				b.add(placeholderRef(o.name), CompletionTypeParameter, offerAs{})
			}
		}
	case "variables":
		b.variables()
	case "entries":
		b.entries(w.def.Name, false)
	case "flags":
		b.entries(w.def.Name, true)
	case "field":
		b.fieldValues(w.field)
	case "slot-item":
		b.slotItems(w.slot)
	case "slot-rest":
		b.slotRest(w.slot, w.item)
	case "enchants":
		b.enchants(w.slot, w.item, "")
	case "specs":
		for _, spec := range Specs() {
			b.add(string(spec.Field.Name()), CompletionClass, offerAs{detail: spec.Spec.String()})
		}
	case "use-path":
		b.pathsForUse()
	case "use-variables":
		b.usedVariables(w.path)
	}
}

// --- Lines.

func (b *completer) line(column0 bool) {
	ix := b.d.ix
	kind := b.d.Kind()
	if column0 {
		allowed := fileSections(ix, kind)
		for _, def := range ix.sectionList {
			switch {
			case !allowed[def.Name]:
			case def.Name == sectionGroup:
				b.add(`group "Name":`, CompletionModule, offerAs{insert: `group "`, filterText: "group", detail: "a rotation group", rank: 2, ranked: true})
			case def.Name != kind.Section:
				b.add(def.Name+":", CompletionModule, ranked(2).with("", "section"))
			}
		}
	}
	if b.ctx.headers == 0 && (column0 || kind.Section != "") {
		written := map[string]bool{}
		for _, s := range b.d.tree().top {
			switch s.(type) {
			case *versionStmt:
				written["siml 1"] = true
			case *specStmt:
				written["spec"] = true
			case *talentsStmt:
				written["talents"] = true
			}
		}
		for _, l := range topLines {
			if written[l[0]] || (l[0] == "var" || l[0] == "def") && kind.Section != "" && kind.Section != sectionAPL {
				continue
			}
			b.add(l[0], CompletionKeyword, ranked(3).with(l[1], l[2]))
		}
	}
}

// The action kinds in APLAction's declaration order.
func (ix *schemaIndex) actionKindList() []*aplKind {
	out := make([]*aplKind, 0, len(ix.actionKinds))
	for _, k := range ix.actionKinds {
		out = append(out, k)
	}
	slices.SortFunc(out, func(a, b *aplKind) int { return a.field.Index() - b.field.Index() })
	return out
}

func (b *completer) steps(when bool) {
	for _, w := range actionWords {
		b.add(w[0], CompletionKeyword, ranked(1).with(w[1], w[2]))
	}
	if when {
		b.add("when", CompletionKeyword, ranked(1).with("when ", "when <condition>: the rows below it, indented deeper, each read with the condition"))
	}
	for _, k := range b.d.ix.actionKindList() {
		if k.Form != "short" {
			b.kindItem(k)
		}
	}
}

func (b *completer) kindItem(k *aplKind) {
	insert := k.Kind + "("
	if k.Form == "bare" || k.Form == "unit_prefix" || len(k.message.fields) == 0 {
		insert = k.Kind
	}
	b.add(k.Kind, CompletionFunction, offerAs{insert: insert, detail: formText(k), documentation: fieldsText(k.message), rank: 3, ranked: true})
}

func formText(k *aplKind) string {
	switch k.Form {
	case "bare":
		return "a value with no fields"
	case "unit_prefix":
		return "[unit.]kind"
	case "list_args":
		return "kind(a, b, …)"
	}
	if k.Positional != "" {
		return fmt.Sprintf("kind(%s, field=value, …)", k.Positional)
	}
	return "kind(field=value, …)"
}

// A message's fields and their printed types, one per line, but those named in leave.
func fieldsText(m *message, leave ...string) string {
	var lines []string
	for _, f := range m.fields {
		if slices.Contains(leave, f.Name) {
			continue
		}
		list := ""
		if f.Repeated {
			list = " list"
		}
		lines = append(lines, fmt.Sprintf("%s: %s%s", f.Siml, f.Printed, list))
	}
	return strings.Join(lines, "\n")
}

func (b *completer) values() {
	ix := b.d.ix
	b.variables()
	b.defs()
	if b.ctx.group.inGroup {
		for _, o := range b.definitions() {
			if o.kind == "placeholder" && o.scope == b.ctx.group {
				b.add(placeholderRef(o.name), CompletionTypeParameter, offerAs{})
			}
		}
	}
	for _, word := range []string{"true", "false", "not"} {
		insert := word
		if word == "not" {
			insert = "not "
		}
		b.add(word, CompletionKeyword, ranked(4).with(insert, ""))
	}
	for _, head := range ix.heads {
		b.add(head, CompletionFunction, ranked(2).with(head+"(", head+"(…).<property>"))
	}
	for _, k := range ix.valueKindList {
		if k.Form != "short" && k.Form != "accessor" {
			b.kindItem(k)
		}
	}
	b.units(true)
}

// The file's defs and those of the files it uses, each with its parameters, and the parameters of the def whose value
// the cursor is in.
func (b *completer) defs() {
	for _, p := range b.ctx.params {
		b.add(p, CompletionTypeParameter, ranked(0).with("", "parameter"))
	}
	params := map[string][]string{}
	for _, o := range b.definitions() {
		if o.kind == "parameter" && o.definition {
			params[o.scope.group] = append(params[o.scope.group], o.name)
		}
	}
	for _, o := range b.definitions() {
		if o.kind != "def" || !o.definition {
			continue
		}
		signature, insert := o.name, o.name
		if p := params[o.name]; len(p) > 0 {
			signature, insert = o.name+"("+strings.Join(p, ", ")+")", o.name+"("
		}
		b.add(o.name, CompletionFunction, ranked(1).with(insert, "def "+signature))
	}
	if b.d.typed == nil {
		return
	}
	imported := b.d.typed.typer.defs.imported
	for _, name := range slices.Sorted(maps.Keys(imported)) {
		in := imported[name]
		signature, insert := name, name
		if len(in.params) > 0 {
			p := make([]string, len(in.params))
			for i, x := range in.params {
				p[i] = x.text
			}
			signature, insert = name+"("+strings.Join(p, ", ")+")", name+"("
		}
		b.add(name, CompletionFunction, ranked(1).with(insert, "def "+signature+", from "+filepath.Base(in.path)))
	}
}

func (b *completer) variables() {
	for _, o := range b.definitions() {
		if o.kind != "variable" || !o.definition || o.scope.inGroup && o.scope != b.ctx.group {
			continue
		}
		detail := "variable"
		if o.scope.inGroup {
			detail = fmt.Sprintf(`variable of group "%s"`, o.scope.group)
		}
		b.add(variableRef(o.name), CompletionVariable, ranked(0).with("", detail))
	}
}

// The unit words; as prefixes, each followed by its dot.
func (b *completer) units(prefixes bool) {
	for _, word := range slices.Concat(b.d.ix.unitWordList(), indexedUnits) {
		if prefixes {
			b.add(word+".", CompletionReference, ranked(4).with("", "unit"))
		} else {
			b.add(word, CompletionReference, offerAs{detail: "unit"})
		}
	}
	if !prefixes {
		b.add("none", CompletionKeyword, offerAs{})
	}
}

// After `unit.`: the heads whose kinds take a unit, and the kinds a unit prefixes.
func (b *completer) members() {
	ix := b.d.ix
	var heads []string
	for _, k := range ix.valueKindList {
		if k.Accessor != nil && k.Accessor.Unit != "" && !slices.Contains(heads, k.Accessor.Head) {
			heads = append(heads, k.Accessor.Head)
		}
	}
	for _, head := range heads {
		b.add(head, CompletionFunction, offerAs{insert: head + "(", detail: head + "(…).<property>"})
	}
	for _, k := range ix.valueKindList {
		if k.Form == "unit_prefix" {
			b.kindItem(k)
		}
	}
}

// --- Names.

// The names of a spell-like namespace the file's spec registers: the class's first, then the shared ones.
func (b *completer) refs(ns Namespace) {
	for _, p := range b.ctx.params {
		b.add(p, CompletionTypeParameter, ranked(0).with("", "parameter"))
	}
	for _, n := range b.d.names.Registered(b.d.class(), b.d.specName(), ns) {
		game := ""
		if b.c != nil {
			if name, ok := b.c.SpellName(n.Entry.SpellID); ok && name != "" {
				game = name + " "
			}
		}
		tag := ""
		if n.Entry.Tag != 0 {
			tag = fmt.Sprintf(" tag %d", n.Entry.Tag)
		}
		from, rank := "", 1
		if n.Shared {
			from, rank = ", from gear, consumables or encounters", 2
		}
		b.add(n.Name, CompletionConstant, ranked(rank).with("", fmt.Sprintf("%s(%d%s)%s", game, n.Entry.SpellID, tag, from)))
	}
	if b.d.typed != nil {
		defs := b.d.typed.typer.defs
		for _, name := range slices.Sorted(maps.Keys(defs.byName)) {
			if d := defs.byName[name]; len(d.params) == 0 && b.names(d.value, ns) {
				b.add(name, CompletionFunction, ranked(1).with("", "def "+name+" = "+nodeText(d.value, b.d.src)))
			}
		}
		for _, name := range slices.Sorted(maps.Keys(defs.imported)) {
			if d := defs.imported[name]; len(d.params) == 0 && b.names(d.value, ns) {
				b.add(name, CompletionFunction, ranked(1).with("", "def "+name+" = "+nodeText(d.value, "")+", from "+filepath.Base(d.path)))
			}
		}
	}
	b.add("item.", CompletionKeyword, ranked(6).with("", "an item"))
	b.add("other.", CompletionKeyword, ranked(6).with("", "another action, such as other.offensive_potion"))
	b.add("none", CompletionKeyword, ranked(7))
}

// Whether n is a name of ns for the file, item.<name> or other.<name>: what a def with no parameters that stands for a
// name may be (SPEC.md §6.10).
func (b *completer) names(n node, ns Namespace) bool {
	switch x := n.(type) {
	case *callNode:
		_, ok := b.d.names.Spell(b.d.class(), ns, x.name.text)
		return x.args == nil && ok
	case *memberNode:
		if x.object.text == "item" {
			_, ok := b.d.names.Item(x.name.text)
			return ok
		}
		return x.object.text == "other" && b.d.ix.enums[otherEnum].byName[x.name.text] != nil
	}
	return false
}

// The name of an item in the catalog, "" for none.
func (b *completer) catalogName(id int32) string {
	if b.c == nil {
		return ""
	}
	name, _ := b.c.ItemName(id)
	return name
}

// Items with a use, gear and consumables, which `cast item.<name>` and the other ActionID positions take.
func (b *completer) usable() {
	if b.c == nil {
		return
	}
	var ids []int32
	for _, item := range b.c.Items() {
		if item.OnUse {
			ids = append(ids, item.ID)
		}
	}
	if b.c != nil {
		for _, x := range b.c.Consumables() {
			if x.OnUse {
				ids = append(ids, x.ID)
			}
		}
	}
	for _, id := range ids {
		if name, ok := b.d.names.ItemName(id); ok {
			b.add(name, CompletionValue, offerAs{detail: fmt.Sprintf("%s (%d)", b.catalogName(id), id)})
		}
	}
}

func (b *completer) itemItem(id int32) {
	name, ok := b.d.names.ItemName(id)
	if !ok {
		return
	}
	ilvl := ""
	if item, ok := b.c.Item(id); ok && item.Ilvl != 0 {
		ilvl = fmt.Sprintf(", ilvl %d", item.Ilvl)
	}
	b.add(name, CompletionValue, offerAs{detail: fmt.Sprintf("%s (%d%s)", b.catalogName(id), id, ilvl)})
}

// --- Entries and fields.

func (b *completer) entries(sectionName string, flagsOnly bool) {
	ix := b.d.ix
	def := ix.sections[sectionName]
	if def == nil {
		return
	}
	written := b.writtenFields(sectionName)
	slots := slices.Contains(def.Body, "slot_rows")
	if slots && !flagsOnly {
		for _, v := range ix.enums[slotEnum].values {
			if !written[v.Short] {
				b.add(v.Short, CompletionField, ranked(1).with(v.Short+" ", "slot"))
			}
		}
		if b.d.Kind().Section == sectionGear {
			return
		}
	}
	if slices.Contains(def.Body, "var") && !flagsOnly {
		b.add("var", CompletionKeyword, offerAs{insert: "var $", detail: "a rotation value variable"})
	}
	if def.Name == sectionAPL && !flagsOnly {
		b.add("def", CompletionKeyword, offerAs{insert: "def ", detail: "def name(p, …) = <value>: a value each call reads as, its parameters replaced"})
	}
	for _, p := range ix.sectionFields(def, b.d.specInfo()) {
		f := p.field
		// A section of slot rows writes its items as rows.
		if slots && f.Repeated && f.Type == itemSpecName {
			continue
		}
		if written[f.Siml] && !(f.Repeated && f.Printed == "message") {
			continue
		}
		flag := f.Printed == "bool" || f.Printed == "flag_enum"
		if flagsOnly && !flag {
			continue
		}
		insert := f.Siml + " "
		switch {
		case flag:
			insert = f.Siml
		case f.Printed == "message" || f.Printed == "indexed":
			insert = f.Siml + "("
		}
		b.add(f.Siml, CompletionField, ranked(2).with(insert, fieldType(f)))
	}
}

// The fields and slots the section's other lines already write.
func (b *completer) writtenFields(sectionName string) map[string]bool {
	out := map[string]bool{}
	var sec *section
	for _, s := range b.d.tree().sections {
		if s.name == sectionName && slices.ContainsFunc(s.body, func(st stmt) bool { return st.base().start <= b.ctx.start }) {
			sec = s
			break
		}
	}
	if sec == nil {
		return out
	}
	for _, s := range sec.body {
		if s.base().holds(b.ctx.start) {
			continue
		}
		switch x := s.(type) {
		case *fieldEntryStmt:
			out[x.name.text] = true
		case *messageEntryStmt:
			out[x.name.text] = true
		case *flagListStmt:
			for _, f := range x.flags {
				out[f.name.text] = true
			}
		case *slotRowStmt:
			out[x.slot.text] = true
		}
	}
	return out
}

func fieldType(f *field) string {
	what := f.Printed
	switch {
	case f.Printed == "ref" && f.Ref != "":
		what = f.Ref
	case f.Printed == "enum" || f.Printed == "message":
		what = strings.TrimPrefix(f.Type, "proto.")
	}
	if f.isList() {
		return what + " list"
	}
	return what
}

func (b *completer) fieldValues(f *field) {
	if f.Printed != "ref" || b.c == nil {
		return
	}
	names := b.d.names
	switch f.Ref {
	case "item":
		if b.c == nil {
			return
		}
		types, typed := consumableTypes[f.Name]
		for _, x := range b.c.Consumables() {
			if !typed || slices.Contains(types, x.Type) {
				b.itemItem(x.ID)
			}
		}
	case "enchant":
		for _, name := range names.Visible(proto.Class_ClassUnknown, NamespaceEnchant) {
			b.add(name, CompletionValue, offerAs{detail: b.enchantDetail(name)})
		}
	case "suffix":
		for _, name := range names.Visible(proto.Class_ClassUnknown, NamespaceSuffix) {
			detail := ""
			if id, ok := names.Suffix(name); ok && b.c != nil {
				if s, ok := b.c.Suffix(id); ok {
					detail = s.Name
				}
			}
			b.add(name, CompletionValue, offerAs{detail: detail})
		}
	}
}

func (b *completer) enchantDetail(name string) string {
	entry := b.d.names.Enchants[name]
	var rows []string
	for _, r := range b.c.EnchantRows(entry.EffectID) {
		if !slices.Contains(rows, r.Name) {
			rows = append(rows, r.Name)
		}
	}
	return fmt.Sprintf("%s (effect %d; %s)", strings.Join(rows, " / "), entry.EffectID, strings.Join(b.d.slotWords(entry.Slots), ", "))
}

// --- Slot rows.

func (b *completer) slotItems(slot string) {
	v := b.d.ix.enums[slotEnum].byName[slot]
	if v == nil || b.c == nil {
		return
	}
	for _, item := range b.c.Items() {
		if slices.Contains(item.slots(), proto.ItemSlot(v.Number)) {
			b.itemItem(item.ID)
		}
	}
}

// The catalog's item a written item name means.
func (b *completer) catalogItem(name string) (CatalogItem, bool) {
	id, ok := b.d.names.Item(name)
	if !ok || b.c == nil {
		return CatalogItem{}, false
	}
	return b.c.Item(id)
}

func (b *completer) slotRest(slot, itemName string) {
	item, found := b.catalogItem(itemName)
	var suffixes []int32
	if found {
		suffixes = b.c.Suffixes(item.ID)
	}
	for _, id := range suffixes {
		name, ok := b.d.names.SuffixName(id)
		if !ok {
			continue
		}
		what := fmt.Sprint(id)
		if b.c != nil {
			if s, ok := b.c.Suffix(id); ok {
				what = s.Name
			}
		}
		b.add(name, CompletionValue, ranked(1).with("", "random suffix: "+what))
	}
	b.add("+", CompletionKeyword, ranked(2).with("+ ", "an enchant"))
	if len(suffixes) == 0 {
		b.enchants(slot, itemName, "+ ")
	}
}

func (b *completer) enchants(slot, itemName, before string) {
	if b.c == nil {
		return
	}
	v := b.d.ix.enums[slotEnum].byName[slot]
	item, found := CatalogItem{}, false
	if itemName != "" {
		item, found = b.catalogItem(itemName)
	}
	for _, name := range sortedKeys(b.d.names.Enchants) {
		entry := b.d.names.Enchants[name]
		if v == nil || !slices.Contains(entry.Slots, proto.ItemSlot(v.Number)) {
			continue
		}
		if found && !enchantEffectFits(b.c, entry.EffectID, item) {
			continue
		}
		label := name
		if before != "" {
			label = "+ " + name
		}
		b.add(label, CompletionValue, offerAs{insert: before + name, filterText: label, detail: b.enchantDetail(name), rank: 3, ranked: true})
	}
}

// --- Composition.

// The .siml files near the file, by their path from its folder.
func (b *completer) pathsForUse() {
	if b.d.path == "" {
		return
	}
	for _, file := range nearbyFiles(b.d.path) {
		rel, err := filepath.Rel(filepath.Dir(b.d.path), file)
		if err != nil {
			continue
		}
		b.add(jsonString(filepath.ToSlash(rel)), CompletionFile, offerAs{detail: "file"})
	}
}

// The variables a with list of a use line naming path sets.
func (b *completer) usedVariables(path string) {
	if b.d.path == "" || path == "" {
		return
	}
	for _, v := range b.d.settableVariables(path) {
		b.add(variableRef(v.name), CompletionVariable, offerAs{insert: variableRef(v.name) + " = "})
	}
}

// --- Arguments.

func (b *completer) args(m *message, kind *aplKind, used, skip map[string]bool) {
	for _, f := range m.fields {
		if used[f.Name] || skip[f.Name] || kind != nil && f.Name == kind.ListField {
			continue
		}
		if f.Printed == "bool" {
			b.add(f.Siml, CompletionProperty, ranked(2).with("", "bool"))
		}
		written := f.Siml + "="
		if f.Printed == "message" || f.Printed == "indexed" {
			written = f.Siml + "("
		}
		b.add(written, CompletionProperty, offerAs{filterText: f.Siml, detail: fieldType(f), rank: 2, ranked: true})
	}
}
