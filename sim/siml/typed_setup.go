package siml

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A whole file of any kind (SPEC.md §1, §5, §9): the spec and talents lines, every section typed into one
// IndividualSimSettings, what a present section sets, the categories, and the files use takes in.

const (
	talentsKey       = "talents" // The talents line, which counts as a section where files are put together.
	talentsCategory  = "t"
	rotationCategory = "r"
	categoryOrder    = "gtrcmxeu" // The site's import categories, in its order (SIM_CATEGORY_KEYS).
)

// A typed file, before its diagnostics decide whether it yields a proto.
type typedFile struct {
	kind         Kind
	tree         *syntaxTree
	spec         *SpecInfo
	settings     *proto.IndividualSimSettings
	encoded      map[string]protoreflect.Message // The spec's Rotation each spec_rotation_json holds while its section is typed, by the string's JSON path.
	present      *keySet                         // The sections present, a group as `group "Name"`, and talents for the talents line.
	categories   []string                        // In the site's order.
	typer        *typer
	written      map[stmt]protoreflect.Message // What each row and var line writes; invalid when it does not read.
	paths        map[stmt][]any                // Each row's and var line's JSON path within the rotation, as if every statement read.
	entries      map[stmt][]entryWrite         // What each entry line and slot row writes.
	withs        map[stmt][]*proto.APLValue    // The values each use line's with list gives, nil where one does not read.
	used         map[stmt]*typedFile           // The file each use line reads, typed, where one was read.
	usedPaths    map[stmt]string               // The path each of those is read at.
	usedDeclared []declared                    // The variables outside groups that used files declare, which this file's rows may name.
	ownRotation  *proto.APLRotation            // The rotation this file writes alone, when used files bring more; nil when they bring none.
	whens        map[*whenStmt]*proto.APLValue // Each when line's condition, read on its own.
	declarations []*variableDecl
	explicitType bool // Whether an apl: entry of the file, or of a file it uses, sets the rotation type.
	diags        []diag
}

// How use lines are read (SPEC.md §9): through open, with the chain of files being read for cycles, or refused by a
// host that reads no other files. A nil *uses does not follow them: format and the printer read one file's text.
type uses struct {
	refused bool
	path    string // The file being read, which the paths of its use lines are read against.
	open    Opener
	chain   []string
}

// A file's text, read and typed, its names read with names.
func readFile(ix *schemaIndex, names *Names, src string, kind Kind, u *uses) *typedFile {
	tree, diags := parseSyntax(src, ix, kind)
	f := typeFile(ix, names, tree, kind, u)
	f.diags = append(diags, f.diags...)
	return f
}

var absolutePath = regexp.MustCompile(`^([A-Za-z]:)?[\\/]`)

// The path target names, read against the folder of from, with / between its parts and . and .. resolved.
func resolveUse(from, target string) string {
	target = strings.ReplaceAll(target, `\`, "/")
	if !absolutePath.MatchString(target) {
		target = path.Join(path.Dir(strings.ReplaceAll(from, `\`, "/")), target)
	}
	if target = path.Clean(target); target == "." {
		return ""
	}
	return target
}

// Types a file's syntax tree.
func typeFile(ix *schemaIndex, names *Names, tree *syntaxTree, kind Kind, u *uses) *typedFile {
	t := newTyper(ix, names)
	f := &typedFile{
		kind:      kind,
		tree:      tree,
		settings:  &proto.IndividualSimSettings{},
		encoded:   map[string]protoreflect.Message{},
		present:   presentSections(ix, tree),
		typer:     t,
		written:   map[stmt]protoreflect.Message{},
		paths:     map[stmt][]any{},
		entries:   map[stmt][]entryWrite{},
		withs:     map[stmt][]*proto.APLValue{},
		used:      map[stmt]*typedFile{},
		usedPaths: map[stmt]string{},
		whens:     map[*whenStmt]*proto.APLValue{},
	}
	f.spec = typeSpec(t, tree, kind)
	if f.spec != nil {
		t.class = f.spec.Class
		t.spec = string(f.spec.Field.Name())
		player := f.player()
		player.Class = f.spec.Class
		player.ProtoReflect().Mutable(f.spec.Field)
	}
	t.lines = tree.lines
	t.readDefs(tree)
	for _, s := range tree.top {
		if s, ok := s.(*talentsStmt); ok {
			f.player().TalentsString = s.talents.value
			pop := t.in("player", "talentsString")
			t.mark(s.talents.span)
			pop()
			break
		}
	}

	used := f.readUses(u)
	f.importDefs()
	t.checkDefCycles()
	for _, typeWith := range used.withs {
		typeWith()
	}
	// A section file's own section is present when it writes it, and otherwise when no used file brings it.
	if own := ix.sections[kind.Section]; own != nil && !used.present.has(own.Name) {
		f.present.add(own.Name)
	}
	var rotation rotationTyping
	if f.hasRotationStatements() {
		rotation = f.typeRotation()
	}
	for _, s := range tree.sections {
		f.typeSection(s)
	}
	for _, key := range f.present.keys {
		f.setPresent(key)
	}
	f.encodeSimple()

	t.sequenceDefs = maps.Clone(used.declared.sequences)
	sequencesIn(f.settings.GetPlayer().GetRotation(), t.sequenceDefs)
	t.checkReferences(rotation.declarations, rotation.groups, used.declared)
	if f.writesRows() {
		t.checkDefsUsed()
	}
	f.usedDeclared = slices.DeleteFunc(slices.Clone(used.declared.variables), func(d declared) bool { return d.scope.inGroup })

	if len(used.settings) > 0 {
		f.ownRotation = &proto.APLRotation{}
		if r := f.settings.GetPlayer().GetRotation(); r != nil {
			f.ownRotation = googleProto.Clone(r).(*proto.APLRotation)
		}
	}
	for _, s := range used.settings {
		merge(f.settings.ProtoReflect(), s.ProtoReflect())
	}
	f.checkVariableCycles(rotation.declarations, used)
	for _, key := range used.present.keys {
		f.present.add(key)
	}
	f.explicitType = rotation.explicitType || used.explicitType
	if kind.Section == sectionAPL || slices.ContainsFunc(f.present.keys, func(key string) bool { return categoryOf(ix, key) == rotationCategory && key != "cooldowns" }) {
		r := f.rotation()
		if !f.explicitType {
			r.Type = proto.APLRotation_TypeAPL
			if f.present.has(sectionSimple) {
				r.Type = proto.APLRotation_TypeSimple
			}
		}
		if f.ownRotation != nil {
			f.ownRotation.Type = r.Type
		}
	}
	for _, key := range f.present.keys {
		if c := categoryOf(ix, key); c != "" && !slices.Contains(f.categories, c) {
			f.categories = append(f.categories, c)
		}
	}
	slices.SortFunc(f.categories, func(a, b string) int { return strings.Index(categoryOrder, a) - strings.Index(categoryOrder, b) })
	f.declarations = rotation.declarations
	f.diags = t.diags
	return f
}

// The rotation as this file writes it, without what used files bring.
func (f *typedFile) writtenRotation() *proto.APLRotation {
	if f.ownRotation != nil {
		return f.ownRotation
	}
	return f.settings.GetPlayer().GetRotation()
}

func (f *typedFile) player() *proto.Player {
	if f.settings.Player == nil {
		f.settings.Player = &proto.Player{}
	}
	return f.settings.Player
}

// The spec line, which a rotation file and a full file must hold (SPEC.md §3.3).
func typeSpec(t *typer, tree *syntaxTree, kind Kind) *SpecInfo {
	for _, s := range tree.top {
		s, ok := s.(*specStmt)
		if !ok {
			continue
		}
		info, ok := t.ix.specs[s.spec.text]
		if !ok {
			var names []string
			for name := range t.ix.specs {
				names = append(names, name)
			}
			t.report("unknown-spec", "no spec is named "+s.spec.text+didYouMean(s.spec.text, names), s.spec.span)
			return nil
		}
		return &info
	}
	if kind.Section != "" && kind.Section != sectionAPL {
		return nil
	}
	at := span{}
	for _, s := range tree.top {
		if s, ok := s.(*versionStmt); ok {
			at = s.span
			break
		}
	}
	what := "a setup file"
	if kind.Section == sectionAPL {
		what = "a rotation file"
	}
	t.report("spec-required", what+" names its spec: spec <field>, such as spec hunter", at)
	return nil
}

// Keys in the order they are first added.
type keySet struct {
	keys []string
	set  map[string]bool
}

func (k *keySet) add(key string) {
	if !k.set[key] {
		k.set[key] = true
		k.keys = append(k.keys, key)
	}
}

func (k *keySet) has(key string) bool { return k.set[key] }

func newKeySet() *keySet { return &keySet{set: map[string]bool{}} }

// The sections a file writes, a group by its name, and talents when it has a talents line. A file's top-level var
// lines are its apl: section; its def lines write no section.
func presentSections(ix *schemaIndex, tree *syntaxTree) *keySet {
	present := newKeySet()
	for _, s := range tree.sections {
		written := s.header != nil || slices.ContainsFunc(s.body, func(b stmt) bool { _, isDef := b.(*defStmt); return !isDef })
		if ix.sections[s.name] == nil || !written {
			continue
		}
		if s.group != nil {
			present.add(groupKey(s.group.value))
		} else {
			present.add(s.name)
		}
	}
	for _, s := range tree.top {
		switch s.(type) {
		case *varStmt:
			present.add(sectionAPL)
		case *talentsStmt:
			present.add(talentsKey)
		}
	}
	return present
}

// Whether the file, or a file it uses, writes an apl: entry, such as type, beyond var and def lines.
func (f *typedFile) writesAPLEntries() bool {
	for _, s := range f.tree.sections {
		if s.name == sectionAPL && s.group == nil && slices.ContainsFunc(s.body, isEntry) {
			return true
		}
	}
	for _, used := range f.used {
		if used.writesAPLEntries() {
			return true
		}
	}
	return false
}

// Whether the file writes a rotation row. A file with none holds defs and variables for the files that use it, so a
// def no line of its own calls is no warning there.
func (f *typedFile) writesRows() bool {
	return slices.ContainsFunc(f.tree.statements, func(s stmt) bool { _, ok := s.(*rowStmt); return ok })
}

func groupKey(name string) string { return sectionGroup + " " + jsonString(name) }

// The section a key of present names: group for a group's key.
func sectionOfKey(key string) string {
	if strings.HasPrefix(key, sectionGroup+" ") {
		return sectionGroup
	}
	return key
}

// The import category of a key of present, "" for none.
func categoryOf(ix *schemaIndex, key string) string {
	if key == talentsKey {
		return talentsCategory
	}
	if def := ix.sections[sectionOfKey(key)]; def != nil {
		return def.Category
	}
	return ""
}

// Whether the rotation's typing reads a section: one that holds rows or var lines.
func typedByRotation(def *sectionDef) bool {
	return def != nil && (slices.Contains(def.Body, "rows") || slices.Contains(def.Body, "var"))
}

// Whether the file holds a statement the rotation's typing reads: a top-level var line, or a section of rows or var
// lines.
func (f *typedFile) hasRotationStatements() bool {
	for _, s := range f.tree.top {
		if _, ok := s.(*varStmt); ok {
			return true
		}
	}
	return slices.ContainsFunc(f.tree.sections, func(s *section) bool { return typedByRotation(f.typer.ix.sections[s.name]) })
}

// --- Setup sections (SPEC.md §5)

// The message that holds a placed field, made where missing: a field of the spec's Rotation goes into the Rotation
// its string is encoded from.
func (f *typedFile) holder(p *placed) protoreflect.Message {
	m := mutableAt(f.settings.ProtoReflect(), p.at)
	if p.encoded == nil {
		return m
	}
	key := jsonPath(p.at, p.encoded)
	r := f.encoded[key]
	if r == nil {
		r = newMessage(f.spec.Rotation)
		f.encoded[key] = r
	}
	return r
}

// The entries and slot rows of a setup section.
func (f *typedFile) typeSection(s *section) {
	ix := f.typer.ix
	def := ix.sections[s.name]
	if def == nil || typedByRotation(def) {
		return
	}
	fields := ix.sectionFields(def, f.spec)
	var slots *placed
	if slices.Contains(def.Body, "slot_rows") {
		slots = slotField(fields)
	}
	fields = slices.DeleteFunc(fields, func(p *placed) bool { return p == slots })
	set := map[*field]bool{}
	taken := map[int32]bool{}
	for _, b := range s.body {
		if row, ok := b.(*slotRowStmt); ok && slots != nil {
			f.entries[b] = f.slotRow(row, slots, taken)
		} else if isEntry(b) {
			f.entries[b] = f.typer.entry(b, fields, f.holder, set)
		}
	}
}

// `<slot> <item> [<suffix>] [+ <enchant>]`, or the slot and a raw ItemSpec (SPEC.md §5.3).
func (f *typedFile) slotRow(s *slotRowStmt, slots *placed, taken map[int32]bool) []entryWrite {
	t := f.typer
	slotWords := t.ix.enums[slotEnum]
	slot := slotWords.byName[s.slot.text]
	if slot == nil {
		t.report("unknown-slot", "no slot is named "+s.slot.text+didYouMean(s.slot.text, slotWords.shortNames()), s.slot.span)
		return nil
	}
	if taken[slot.Number] {
		t.report("duplicate-slot", slot.Short+" is written twice", s.slot.span)
		return nil
	}
	taken[slot.Number] = true
	itemSpec := t.ix.messages[itemSpecName]
	var item protoreflect.Message
	if s.raw != nil {
		item = t.raw(itemSpec.desc, s.raw)
	} else {
		item = newMessage(itemSpec.desc)
		row := &SlotRow{Slot: proto.ItemSlot(slot.Number)}
		at := append(slots.keys(), int(slot.Number))
		for _, part := range []struct {
			field    string
			written  *name
			ns       Namespace
			position string
		}{
			{"id", s.item, NamespaceItem, "a slot"},
			{"random_suffix", s.suffix, NamespaceSuffix, "a slot row's suffix"},
			{"enchant", s.enchant, NamespaceEnchant, "+"},
		} {
			if part.written == nil {
				continue
			}
			fd := itemSpec.byProto[part.field].desc
			pop := t.in(slices.Concat(at, []any{fd.JSONName()})...)
			r, ok := t.lookup(*part.written, part.ns, part.position, false)
			if ok {
				t.mark(part.written.span)
			}
			pop()
			if !ok {
				item = nil
				continue
			}
			r.row = row
			t.addRef(r)
			if part.ns == NamespaceItem {
				row.Item = r.id
			}
			if item != nil {
				item.Set(fd, intValue(fd, int64(r.id)))
			}
		}
	}
	if item == nil {
		return nil
	}
	pop := t.in(append(slots.keys(), int(slot.Number))...)
	t.mark(s.span)
	pop()
	items := mutableAt(f.settings.ProtoReflect(), slots.at).Mutable(slots.desc).List()
	for items.Len() <= int(slot.Number) {
		items.Append(items.NewElement())
	}
	items.Set(int(slot.Number), protoreflect.ValueOfMessage(item))
	return []entryWrite{{placed: slots, index: int(slot.Number), value: protoreflect.ValueOfMessage(item)}}
}

// What a present section sets (SPEC.md §5.1): its message, even when its body is empty; the class options of the
// spec's options; and all of a section's items, a slot not written being empty.
func (f *typedFile) setPresent(key string) {
	ix := f.typer.ix
	def := ix.sections[sectionOfKey(key)]
	if def == nil || def.Presence == nil {
		return
	}
	at, ok := ix.specPath(def.Presence, f.spec)
	if !ok {
		return
	}
	settings := f.settings.ProtoReflect()
	mutableAt(settings, at)
	for _, src := range def.Sources {
		if src.Flatten != "spec_options" || f.spec == nil || f.spec.ClassOptions == nil {
			continue
		}
		if options, ok := ix.specPath(src.Path, f.spec); ok {
			mutableAt(settings, options).Mutable(f.spec.ClassOptions)
		}
	}
	if !slices.Contains(def.Body, "slot_rows") {
		return
	}
	if slots := slotField(ix.sectionFields(def, f.spec)); slots != nil {
		items := mutableAt(settings, slots.at).Mutable(slots.desc).List()
		for n := ix.enumLength(slotEnum); items.Len() < n; {
			items.Append(items.NewElement())
		}
	}
}

// Encodes the spec's Rotation each present section writes into its string, as the canonical rotation JSON of
// SPEC.md §5.4.
func (f *typedFile) encodeSimple() {
	ix := f.typer.ix
	if f.spec == nil {
		return
	}
	for _, key := range f.present.keys {
		def := ix.sections[key]
		if def == nil {
			continue
		}
		for _, src := range def.Sources {
			if src.Flatten != "spec_rotation_json" {
				continue
			}
			at, _ := ix.specPath(src.Path, f.spec)
			holder := mutableAt(f.settings.ProtoReflect(), at)
			for _, name := range src.Fields {
				fd := ix.messageAt(at).byJSON[name].desc
				r := f.encoded[jsonPath(at, fd)]
				if r == nil {
					r = newMessage(f.spec.Rotation)
				}
				holder.Set(fd, protoreflect.ValueOfString(ix.jsonText(r)))
			}
		}
	}
}

// --- Composition (SPEC.md §9)

// What the files a file uses bring.
type usedFiles struct {
	settings     []*proto.IndividualSimSettings // In the order of the use lines.
	present      *keySet
	declared     usedDeclarations
	explicitType bool
	withs        []func()        // Types each with list and takes in its use line's variables, once the used files' defs are in, which a with value may call.
	withSet      map[string]span // The variables the with lists set, each at the variable that sets it first.
}

// The variables, groups and sequences used rotations declare, which a file's references may name.
type usedDeclarations struct {
	variables []declared
	from      map[string]string          // The used file that declares each variable outside groups.
	groups    map[string]map[string]bool // Each group's placeholders.
	sequences map[string]bool            // Each sequence's name, as sequencesIn gives it.
}

// Reads each use line's file and checks it against this one: a section in both is section-conflict, and the with
// list sets the used rotation's variables. Its values are typed in this file, whose class their names are read with,
// by the withs it gives.
func (f *typedFile) readUses(u *uses) usedFiles {
	t := f.typer
	ix := t.ix
	allowed := fileSections(ix, f.kind)
	used := usedFiles{present: newKeySet(), declared: usedDeclarations{from: map[string]string{}, groups: map[string]map[string]bool{}, sequences: map[string]bool{}}, withSet: map[string]span{}}
	broughtBy := map[string]string{} // The used file that brought each key of used.present.
	aplEntriesFrom := ""             // The used file that brought apl: entries.
	type broughtVariable struct {
		file  string // The file that declares it.
		value *proto.APLValueVariable
	}
	brought := map[string]broughtVariable{} // The first value variable of each name the use lines bring.
	for _, s := range f.tree.top {
		use, ok := s.(*useStmt)
		if !ok {
			continue
		}
		if u == nil {
			used.withs = append(used.withs, func() {
				values := make([]*proto.APLValue, len(use.with))
				for i, b := range use.with {
					values[i] = t.value(b.value)
				}
				f.withs[use] = values
			})
			continue
		}
		at := use.path.span
		if u.refused {
			t.report("use-refused", "this host reads no other files: write the used sections into this file", at)
			continue
		}
		path := resolveUse(u.path, use.path.value)
		if i := slices.Index(u.chain, path); i >= 0 {
			cycle := append(slices.Clone(u.chain[i+1:]), path)
			t.report("use-cycle", fmt.Sprintf("%s uses this file, through %s", path, strings.Join(cycle, " → ")), at)
			continue
		}
		text, err := u.open(path)
		if err != nil {
			message := "no file is at " + path
			if !errors.Is(err, fs.ErrNotExist) {
				message = err.Error()
			}
			t.report("use-not-found", message, at)
			continue
		}
		file := readFile(ix, t.names, sourceText(text), kindOf(path), &uses{path: path, open: u.open, chain: append(slices.Clone(u.chain), path)})
		f.used[use] = file
		f.usedPaths[use] = path
		// The used file's first error stands for all of them; the file itself shows the rest.
		diags := file.tree.lines.diagnostics(path, file.diags)
		if i := slices.IndexFunc(diags, func(d Diagnostic) bool { return d.Severity == Error }); i >= 0 {
			t.report(diags[i].Code, fmt.Sprintf("%s:%d:%d: %s", path, diags[i].Start.Line, diags[i].Start.Col, diags[i].Message), at)
			t.diags[len(t.diags)-1].related = &RelatedLocation{Path: path, Range: Range{diags[i].Start, diags[i].End}}
			continue
		}
		if f.spec != nil && file.spec != nil && f.spec.Field != file.spec.Field {
			t.report("spec-mismatch", fmt.Sprintf("%s is for spec %s, and this file for %s", path, file.spec.Field.Name(), f.spec.Field.Name()), at)
		}
		var foreign []string
		sources := newKeySet()           // Where the keys already present came from: this file, or the used file that brought one.
		clashes := map[string][]string{} // The keys of each source.
		for _, key := range file.present.keys {
			// Variables of several files go together; only apl: entries such as type clash.
			if key == sectionAPL && !(file.writesAPLEntries() && (f.writesAPLEntries() || aplEntriesFrom != "")) {
				continue
			}
			if f.present.has(key) || used.present.has(key) {
				source := "this file"
				if !f.present.has(key) {
					source = broughtBy[key]
				}
				sources.add(source)
				clashes[source] = append(clashes[source], key)
			}
			if key == talentsKey && f.kind.Section != "" || key != talentsKey && !allowed[sectionOfKey(key)] {
				foreign = append(foreign, key)
			}
		}
		if len(sources.keys) > 0 {
			conflicts := make([]string, len(sources.keys))
			for i, source := range sources.keys {
				verb := "is"
				if len(clashes[source]) > 1 {
					verb = "are"
				}
				conflicts[i] = fmt.Sprintf("%s %s in both %s and %s", listed(clashes[source]), verb, source, path)
			}
			t.report("section-conflict", strings.Join(conflicts, "; "), at)
		}
		if len(foreign) > 0 {
			t.report("section-not-allowed", fmt.Sprintf("%s brings %s, which an x.%s.siml file does not hold", path, listed(foreign), f.kind.Section), at)
		}
		for _, key := range file.present.keys {
			if !used.present.has(key) {
				broughtBy[key] = path
			}
			used.present.add(key)
		}
		rotation := file.settings.GetPlayer().GetRotation()
		if len(use.with) > 0 && rotation == nil {
			t.report("with-not-rotation", path+" holds no rotation whose variables with could set", use.with[0].variable.span)
		} else {
			used.withs = append(used.withs, func() {
				f.withs[use] = f.setVariables(rotation, use.with, path, used.withSet)
				// A variable an earlier use line brought is brought again by a file both reach, from that file and
				// with the same value, and is taken in once; any other is declared twice.
				declaring := map[string]string{}
				for _, v := range variablesOf(declaredByFile(file, path)) {
					declaring[v.name] = cmp.Or(declaring[v.name], v.file)
				}
				var twice []string
				variables := slices.DeleteFunc(rotation.GetValueVariables(), func(v *proto.APLValueVariable) bool {
					first, ok := brought[v.Name]
					switch {
					case !ok:
						brought[v.Name] = broughtVariable{declaring[v.Name], v}
						used.declared.from[v.Name] = path
						return false
					case first.file == declaring[v.Name] && googleProto.Equal(first.value, v):
						return true
					}
					twice = append(twice, fmt.Sprintf("%s is declared in both %s and %s", v.Name, used.declared.from[v.Name], path))
					return false
				})
				if rotation != nil {
					rotation.ValueVariables = variables
				}
				if len(twice) > 0 {
					t.report("duplicate-variable", strings.Join(twice, "; "), at)
				}
			})
		}
		if file.writesAPLEntries() && aplEntriesFrom == "" {
			aplEntriesFrom = path
		}
		used.settings = append(used.settings, file.settings)
		used.explicitType = used.explicitType || file.explicitType
		declared := declaredBy(rotation)
		used.declared.variables = append(used.declared.variables, declared.variables...)
		for group, placeholders := range declared.groups {
			used.declared.groups[group] = placeholders
		}
		sequencesIn(rotation, used.declared.sequences)
	}
	return used
}

// Types the values of a with list into the variables of the used rotation they name, each at its variable in setAt
// unless one is there, and gives the values.
func (f *typedFile) setVariables(rotation *proto.APLRotation, with []varBind, path string, setAt map[string]span) []*proto.APLValue {
	t := f.typer
	names := make([]string, len(rotation.GetValueVariables()))
	for i, v := range rotation.GetValueVariables() {
		names[i] = v.GetName()
	}
	var values []*proto.APLValue
	set := map[string]bool{}
	for _, b := range with {
		res := resolveVariable(b.variable.name, b.variable.braced, names)
		switch {
		case t.reportAmbiguous(b.variable, res):
		case !res.found:
			t.report("with-unknown-variable", fmt.Sprintf("%s declares no variable %s%s", path, nodeText(b.variable, ""), didYouMean(b.variable.name, names)), b.variable.span)
		case set[res.name]:
			t.report("duplicate-variable", fmt.Sprintf("with sets %s twice", nodeText(b.variable, "")), b.variable.span)
		default:
			set[res.name] = true
			if _, ok := setAt[res.name]; !ok {
				setAt[res.name] = b.variable.span
			}
			k := slices.Index(names, res.name)
			value := t.valueIn(b.value, slices.Concat(jsonKeys(t.ix.rotationPath), []any{"valueVariables", k, "value"})...)
			if value != nil {
				rotation.ValueVariables[k].Value = value
			}
			values = append(values, value)
		}
	}
	return values
}

// `a`, `a and b`, `a, b and c` of the sections keys name.
func listed(keys []string) string {
	names := make([]string, len(keys))
	for i, key := range keys {
		names[i] = key + ":"
		if key == talentsKey {
			names[i] = "the talents line"
		}
	}
	if len(names) == 1 {
		return names[0]
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// The variables, groups and placeholders a rotation declares.
func declaredBy(rotation *proto.APLRotation) usedDeclarations {
	out := usedDeclarations{groups: map[string]map[string]bool{}}
	for _, v := range rotation.GetValueVariables() {
		out.variables = append(out.variables, declared{name: v.GetName()})
	}
	for _, g := range rotation.GetGroups() {
		for _, v := range g.GetVariables() {
			out.variables = append(out.variables, declared{name: v.GetName(), scope: scope{group: g.GetName(), inGroup: true}})
		}
		placeholders := map[string]bool{}
		walkMessages(g.ProtoReflect(), func(m protoreflect.Message) {
			if p, ok := m.Interface().(*proto.APLValueVariablePlaceholder); ok && p.Name != "" {
				placeholders[p.Name] = true
			}
		})
		out.groups[g.GetName()] = placeholders
	}
	return out
}

// Adds the names of the sequences the rows of rotation hold to into: true for one a live row of the priority list or
// of a group holds, which the sim finds, and false for one that only hidden and prepull rows hold.
func sequencesIn(rotation *proto.APLRotation, into map[string]bool) {
	add := func(item googleProto.Message, live bool) {
		walkMessages(item.ProtoReflect(), func(m protoreflect.Message) {
			if s, ok := m.Interface().(*proto.APLActionSequence); ok {
				into[s.Name] = into[s.Name] || live
			}
		})
	}
	for _, item := range rotation.GetPriorityList() {
		add(item, !item.Hide)
	}
	for _, g := range rotation.GetGroups() {
		for _, item := range g.Actions {
			add(item, !item.Hide)
		}
	}
	for _, item := range rotation.GetPrepullActions() {
		add(item, false)
	}
}

// Calls visit on m and on every message it holds.
func walkMessages(m protoreflect.Message, visit func(protoreflect.Message)) {
	visit(m)
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList() && fd.Message() != nil:
			for i := range v.List().Len() {
				walkMessages(v.List().Get(i).Message(), visit)
			}
		case fd.IsMap() && fd.MapValue().Message() != nil:
			v.Map().Range(func(_ protoreflect.MapKey, x protoreflect.Value) bool {
				walkMessages(x.Message(), visit)
				return true
			})
		case !fd.IsList() && !fd.IsMap() && fd.Message() != nil:
			walkMessages(v.Message(), visit)
		}
		return true
	})
}

// Merges from into into: messages field by field, lists one after the other, and a value into already has kept.
// Files that are put together write no field twice, except the class and the spec's message, which agree.
func merge(into, from protoreflect.Message) {
	from.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			l := into.Mutable(fd).List()
			for i := range v.List().Len() {
				l.Append(v.List().Get(i))
			}
		case fd.IsMap():
			m := into.Mutable(fd).Map()
			v.Map().Range(func(k protoreflect.MapKey, x protoreflect.Value) bool {
				switch {
				case !m.Has(k):
					m.Set(k, x)
				case fd.MapValue().Message() != nil:
					merge(m.Mutable(k).Message(), x.Message())
				}
				return true
			})
		case fd.Message() != nil && into.Has(fd):
			merge(into.Mutable(fd).Message(), v.Message())
		case into.Has(fd):
		case fd.ContainingOneof() != nil && into.WhichOneof(fd.ContainingOneof()) != nil:
		default:
			into.Set(fd, v)
		}
		return true
	})
}
