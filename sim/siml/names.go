package siml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/wowsims/forever/sim/core/proto"
)

// The name table of SPEC.md §7: what every name of a .siml file means. sim/siml/cmd/simlnames writes it from what
// the sim registers and the database, and it is committed as names.json.

// Where a name is looked up. The position decides the namespace (SPEC.md §7.2). The five spell-like namespaces
// are per class, and items, enchants and suffixes are one namespace each for every file.
type Namespace string

const (
	NamespaceSpell       Namespace = "spell"        // What the player casts.
	NamespaceAura        Namespace = "aura"         // Auras on the player.
	NamespaceTargetAura  Namespace = "target_aura"  // Auras on a target.
	NamespaceTargetSpell Namespace = "target_spell" // What a target casts, which boss_spell reads.
	NamespacePetAura     Namespace = "pet_aura"     // Auras on a pet.
	NamespaceItem        Namespace = "item"
	NamespaceEnchant     Namespace = "enchant"
	NamespaceSuffix      Namespace = "suffix"
)

// Every namespace, in the order SPEC.md §7.1 lists them, which is the order Holding reports them in.
func AllNamespaces() []Namespace {
	return []Namespace{NamespaceSpell, NamespaceAura, NamespaceTargetAura, NamespaceTargetSpell, NamespacePetAura,
		NamespaceItem, NamespaceEnchant, NamespaceSuffix}
}

// Whether names in ns are spells and auras, looked up per class.
func (ns Namespace) SpellLike() bool {
	switch ns {
	case NamespaceSpell, NamespaceAura, NamespaceTargetAura, NamespaceTargetSpell, NamespacePetAura:
		return true
	}
	return false
}

// A spell-like entry: the spell id, tag and rank of an ActionID. A file means the spell id and tag; the rank is
// what the sim registers, which the site shows (SPEC.md §7.3).
type SpellName struct {
	SpellID, Tag, Rank int32
	RegisteredBy       []string // The specs that register it, by Player spec field name in byte order; empty when no build does.
	ICDOnlyBy          []string // The specs of RegisteredBy that know the aura only as an internal cooldown trigger, which only the ICD kinds find.
	Uncastable         bool     // A spell no build lets a rotation cast (the sim's APL flag), as an inner spell.
}

// An enchant entry: an effect id, and the slots of the item types its rows fit.
type EnchantName struct {
	EffectID int32
	Slots    []proto.ItemSlot // In ItemSlot order.
}

// The spell-like entries of a class, or the shared ones, by namespace and then name.
type SpellNames map[Namespace]map[string]SpellName

// A name table. Lookups index it on first use, so the maps must not change after the first lookup.
type Names struct {
	Classes  map[proto.Class]SpellNames
	Shared   SpellNames // Looked up after a class's own entries, and for a file with no spec line.
	Items    map[string]int32
	Enchants map[string]EnchantName
	Suffixes map[string]int32
	// The specs that know an item's effect only as an internal cooldown trigger, which only the ICD kinds find, by
	// the item's name.
	ICDOnlyItems map[string][]string

	once  sync.Once
	index *namesIndex
}

// The committed table, names.json, which a build holds only with the with_db tag.
func DefaultNames() (*Names, error) { return defaultNames() }

type spellKey struct{ id, tag int32 }

type namesIndex struct {
	classes  map[proto.Class]map[Namespace]map[spellKey]string
	shared   map[Namespace]map[spellKey]string
	items    map[int32]string
	enchants map[int32][]string // The names of an effect id, in byte order.
	suffixes map[int32]string
}

func (t *Names) indexed() *namesIndex {
	t.once.Do(func() {
		ix := &namesIndex{
			classes:  map[proto.Class]map[Namespace]map[spellKey]string{},
			shared:   reverseSpells(t.Shared),
			items:    reverseIDs(t.Items),
			enchants: map[int32][]string{},
			suffixes: reverseIDs(t.Suffixes),
		}
		for class, names := range t.Classes {
			ix.classes[class] = reverseSpells(names)
		}
		for _, name := range slices.Sorted(maps.Keys(t.Enchants)) {
			id := t.Enchants[name].EffectID
			ix.enchants[id] = append(ix.enchants[id], name)
		}
		t.index = ix
	})
	return t.index
}

func reverseSpells(names SpellNames) map[Namespace]map[spellKey]string {
	rev := map[Namespace]map[spellKey]string{}
	for ns, entries := range names {
		rev[ns] = map[spellKey]string{}
		for name, e := range entries {
			rev[ns][spellKey{e.SpellID, e.Tag}] = name
		}
	}
	return rev
}

func reverseIDs(names map[string]int32) map[int32]string {
	rev := make(map[int32]string, len(names))
	for name, id := range names {
		rev[id] = name
	}
	return rev
}

// The spell-like entry name means in ns for a file of class: the class's own entry, else the shared one. A file
// with no spec line passes ClassUnknown and reads the shared entries only.
func (t *Names) Spell(class proto.Class, ns Namespace, name string) (*proto.ActionID, bool) {
	e, ok := t.spellEntry(class, ns, name)
	if !ok {
		return nil, false
	}
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: e.SpellID}, Tag: e.Tag, Rank: e.Rank}, true
}

// The name of a spell-like ActionID in ns for a file of class, by its spell id and tag whatever its rank: the class's
// own name, else the shared one. An ActionID that holds an item or other id has none.
func (t *Names) SpellName(class proto.Class, ns Namespace, id *proto.ActionID) (string, bool) {
	spellID := id.GetSpellId()
	if spellID == 0 {
		return "", false
	}
	ix := t.indexed()
	key := spellKey{spellID, id.GetTag()}
	if name, ok := ix.classes[class][ns][key]; ok {
		return name, true
	}
	name, ok := ix.shared[ns][key]
	return name, ok
}

func (t *Names) Item(name string) (int32, bool) {
	id, ok := t.Items[name]
	return id, ok
}

func (t *Names) ItemName(id int32) (string, bool) {
	name, ok := t.indexed().items[id]
	return name, ok
}

// The effect id an enchant name means.
func (t *Names) Enchant(name string) (int32, bool) {
	e, ok := t.Enchants[name]
	return e.EffectID, ok
}

// The name an enchant effect prints with on an item in slot: of the effect's names whose slots hold slot, the one
// with the most slots and the first in byte order of those, else the effect's first name in byte order. An effect a
// one-hand and a two-hand weapon row share thus prints with the one-hand row's name (main and off hand), which fits a
// two-hander too, where the two-hand row's (main hand) fits no one-hander.
func (t *Names) EnchantName(effectID int32, slot proto.ItemSlot) (string, bool) {
	names := t.indexed().enchants[effectID]
	best := ""
	for _, name := range names {
		if slots := t.Enchants[name].Slots; slices.Contains(slots, slot) && (best == "" || len(slots) > len(t.Enchants[best].Slots)) {
			best = name
		}
	}
	if best != "" {
		return best, true
	}
	if len(names) == 0 {
		return "", false
	}
	return names[0], true
}

func (t *Names) Suffix(name string) (int32, bool) {
	id, ok := t.Suffixes[name]
	return id, ok
}

func (t *Names) SuffixName(id int32) (string, bool) {
	name, ok := t.indexed().suffixes[id]
	return name, ok
}

// Every name of ns that a file of class reads, in byte order: the class's own and the shared ones for a spell-like
// namespace.
func (t *Names) Visible(class proto.Class, ns Namespace) []string {
	var keys []string
	switch ns {
	case NamespaceItem:
		keys = slices.Collect(maps.Keys(t.Items))
	case NamespaceEnchant:
		keys = slices.Collect(maps.Keys(t.Enchants))
	case NamespaceSuffix:
		keys = slices.Collect(maps.Keys(t.Suffixes))
	default:
		keys = slices.Collect(maps.Keys(t.Classes[class][ns]))
		for name := range t.Shared[ns] {
			if _, ok := t.Classes[class][ns][name]; !ok {
				keys = append(keys, name)
			}
		}
	}
	slices.Sort(keys)
	return keys
}

// The namespaces in which name means something for a file of class, in the order of AllNamespaces.
func (t *Names) Holding(class proto.Class, name string) []Namespace {
	var held []Namespace
	for _, ns := range AllNamespaces() {
		var ok bool
		switch ns {
		case NamespaceItem:
			_, ok = t.Items[name]
		case NamespaceEnchant:
			_, ok = t.Enchants[name]
		case NamespaceSuffix:
			_, ok = t.Suffixes[name]
		default:
			_, ok = t.Spell(class, ns, name)
		}
		if ok {
			held = append(held, ns)
		}
	}
	return held
}

// The short name of a class (SPEC.md §8.2), which keys its entries in names.json: warrior.
func ClassName(class proto.Class) string { return shortName(classEnum, int32(class)) }

// The class whose short name is short (SPEC.md §8.2); false for none and for ClassUnknown.
func ClassNamed(short string) (proto.Class, bool) {
	v, ok := shortValue(classEnum, short)
	if !ok || v.Number == 0 {
		return 0, false
	}
	return proto.Class(v.Number), true
}

// Whether the generator must not write word as a name: a keyword, head or unit word (SPEC.md §4 note 6).
func Reserved(word string) bool {
	ix, err := loadIndex()
	if err != nil {
		panic(err)
	}
	return ix.isReserved(word)
}

const classEnum = "proto.Class"

// The value of an enum the schema reaches, by its short name (SPEC.md §8.2).
func shortValue(enumName, short string) (*enumValue, bool) {
	ix, err := loadIndex()
	if err != nil {
		panic(err)
	}
	v := ix.enums[enumName].byName[short]
	return v, v != nil && v.Short == short
}

// The short name of an enum the schema reaches, by number; "" for a number that names no value.
func shortName(enumName string, number int32) string {
	ix, err := loadIndex()
	if err != nil {
		panic(err)
	}
	if v := ix.enums[enumName].byNumber[number]; v != nil {
		return v.Short
	}
	return ""
}

// Reads a name table in the form of SPEC.md §7.4, and refuses one that breaks a rule of what a loader checks. The JSON
// is read with keys exactly, numbers by value, so 1.0 is the whole number 1, and null as absent.
func ParseNames(data []byte) (*Names, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("names: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("names: text after the table")
	}
	if v == nil {
		return nil, fmt.Errorf("names: the table is null")
	}
	top, err := object("the table", v, "classes", "enchants", "icdOnlyItems", "items", "shared", "suffixes")
	if err != nil {
		return nil, err
	}
	t := &Names{
		Classes:      map[proto.Class]SpellNames{},
		Items:        map[string]int32{},
		Enchants:     map[string]EnchantName{},
		Suffixes:     map[string]int32{},
		ICDOnlyItems: map[string][]string{},
	}
	if t.Shared, err = spellNames("shared", top["shared"]); err != nil {
		return nil, err
	}
	classes, err := object("classes", top["classes"])
	if err != nil {
		return nil, err
	}
	for class, entries := range classes {
		number, ok := ClassNamed(class)
		if !ok {
			return nil, fmt.Errorf("names: %s is no class", class)
		}
		names, err := spellNames(class, entries)
		if err != nil {
			return nil, err
		}
		for ns, byName := range names {
			for name, e := range byName {
				if s, ok := t.Shared[ns][name]; ok && (s.SpellID != e.SpellID || s.Tag != e.Tag) {
					return nil, fmt.Errorf("names: %s means %s in %s.%s and %s in shared.%s", name, spellText(e), class, ns, spellText(s), ns)
				}
			}
		}
		t.Classes[proto.Class(number)] = names
	}
	if err := ids("items", top["items"], t.Items); err != nil {
		return nil, err
	}
	if err := ids("suffixes", top["suffixes"], t.Suffixes); err != nil {
		return nil, err
	}
	icdItems, err := object("icdOnlyItems", top["icdOnlyItems"])
	if err != nil {
		return nil, err
	}
	ix, err := loadIndex()
	if err != nil {
		return nil, err
	}
	for name, v := range icdItems {
		where := "icdOnlyItems." + name
		if _, ok := t.Items[name]; !ok {
			return nil, fmt.Errorf("names: %s names no item of items", where)
		}
		specs, err := stringList(where, v)
		if err != nil {
			return nil, err
		}
		if len(specs) == 0 {
			return nil, fmt.Errorf("names: %s lists no spec", where)
		}
		for _, spec := range specs {
			if _, ok := ix.specs[spec]; !ok {
				return nil, fmt.Errorf("names: %s lists %q, which is no spec", where, spec)
			}
		}
		t.ICDOnlyItems[name] = specs
	}
	enchants, err := object("enchants", top["enchants"])
	if err != nil {
		return nil, err
	}
	for name, v := range enchants {
		where := "enchants." + name
		e, err := object(where, v, "effectId", "slots")
		if err != nil {
			return nil, err
		}
		id, ok := idOf(e["effectId"])
		if !isName(name) || !ok {
			return nil, fmt.Errorf("names: %s is no name of a positive effect id", where)
		}
		if err := reserved(where, name); err != nil {
			return nil, err
		}
		slots, err := stringList(where+".slots", e["slots"])
		if err != nil {
			return nil, err
		}
		entry := EnchantName{EffectID: id}
		for _, s := range slots {
			v, ok := shortValue(slotEnum, s)
			if !ok {
				return nil, fmt.Errorf("names: %s: %s is no slot", where, s)
			}
			entry.Slots = append(entry.Slots, proto.ItemSlot(v.Number))
		}
		slices.Sort(entry.Slots)
		t.Enchants[name] = entry
	}
	return t, nil
}

func spellNames(scope string, v any) (SpellNames, error) {
	ix, err := loadIndex()
	if err != nil {
		return nil, err
	}
	spaces, err := object(scope, v)
	if err != nil {
		return nil, err
	}
	names := SpellNames{}
	for space, v := range spaces {
		ns := Namespace(space)
		if !ns.SpellLike() {
			return nil, fmt.Errorf("names: %s.%s is no spell-like namespace", scope, ns)
		}
		entries, err := object(scope+"."+space, v)
		if err != nil {
			return nil, err
		}
		names[ns] = map[string]SpellName{}
		named := map[spellKey]string{}
		for name, v := range entries {
			where := scope + "." + space + "." + name
			e, err := object(where, v, "spellId", "tag", "rank", "registeredBy", "icdOnlyBy", "uncastable")
			if err != nil {
				return nil, err
			}
			id, ok := idOf(e["spellId"])
			if !isName(name) || !ok {
				return nil, fmt.Errorf("names: %s is no name of a positive spell id", where)
			}
			if err := reserved(where, name); err != nil {
				return nil, err
			}
			tag, tagOK := int32Of(e["tag"])
			rank, rankOK := int32Of(e["rank"])
			if !tagOK || !rankOK {
				return nil, fmt.Errorf("names: %s has a tag or rank that is no int32", where)
			}
			specs, err := stringList(where+".registeredBy", e["registeredBy"])
			if err != nil {
				return nil, err
			}
			for _, spec := range specs {
				if _, ok := ix.specs[spec]; !ok {
					return nil, fmt.Errorf("names: %s is registered by %q, which is no spec", where, spec)
				}
			}
			icdOnly, err := stringList(where+".icdOnlyBy", e["icdOnlyBy"])
			if err != nil {
				return nil, err
			}
			for _, spec := range icdOnly {
				if ns != NamespaceAura || !slices.Contains(specs, spec) {
					return nil, fmt.Errorf("names: %s.icdOnlyBy lists %q, which is no spec of registeredBy on an aura", where, spec)
				}
			}
			uncastable, ok := e["uncastable"].(bool)
			if e["uncastable"] != nil && (!ok || ns != NamespaceSpell) {
				return nil, fmt.Errorf("names: %s.uncastable is true or left out, and only on a spell", where)
			}
			entry := SpellName{SpellID: id, Tag: tag, Rank: rank, RegisteredBy: specs, ICDOnlyBy: icdOnly, Uncastable: uncastable}
			key := spellKey{id, tag}
			if other, ok := named[key]; ok {
				return nil, fmt.Errorf("names: %s.%s names %s twice, as %s and %s", scope, ns, spellText(entry), min(name, other), max(name, other))
			}
			named[key] = name
			names[ns][name] = entry
		}
	}
	return names, nil
}

func ids(what string, v any, into map[string]int32) error {
	entries, err := object(what, v)
	if err != nil {
		return err
	}
	named := map[int32]string{}
	for name, x := range entries {
		id, ok := idOf(x)
		if !isName(name) || !ok {
			return fmt.Errorf("names: %s.%s is no name of a positive id", what, name)
		}
		if err := reserved(what+"."+name, name); err != nil {
			return err
		}
		if other, ok := named[id]; ok {
			return fmt.Errorf("names: %s names %d twice, as %s and %s", what, id, min(name, other), max(name, other))
		}
		named[id] = name
		into[name] = id
	}
	return nil
}

// The JSON object v, empty when v is null or absent; with keys, an object whose keys are among them.
func object(where string, v any, keys ...string) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("names: %s is no object", where)
	}
	for k := range m {
		if len(keys) > 0 && !slices.Contains(keys, k) {
			return nil, fmt.Errorf("names: %s has no key %s", where, k)
		}
	}
	return m, nil
}

// The JSON list of strings v, empty when v is null or absent.
func stringList(where string, v any) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	list, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("names: %s is no list", where)
	}
	out := make([]string, len(list))
	for i, x := range list {
		if out[i], ok = x.(string); !ok {
			return nil, fmt.Errorf("names: %s holds %v, which is no string", where, x)
		}
	}
	return out, nil
}

// The JSON number v as an int32 when it is a whole number in range; 0 when v is null or absent.
func int32Of(v any) (int32, bool) {
	if v == nil {
		return 0, true
	}
	x, ok := v.(float64)
	return int32(x), ok && x == math.Trunc(x) && x >= math.MinInt32 && x <= math.MaxInt32
}

// The JSON number v as a positive int32 id.
func idOf(v any) (int32, bool) {
	id, ok := int32Of(v)
	return id, ok && id > 0
}

func spellText(e SpellName) string {
	s := "spell " + strconv.Itoa(int(e.SpellID))
	if e.Tag != 0 {
		s += " tag " + strconv.Itoa(int(e.Tag))
	}
	return s
}

// Refuses a reserved word (SPEC.md §4 note 6), which a file never reads as a name.
func reserved(where, name string) error {
	ix, err := loadIndex()
	if err != nil {
		return err
	}
	if ix.isReserved(name) {
		return fmt.Errorf("names: %s is a reserved word", where)
	}
	return nil
}

// Whether s is a name the table may hold: an ASCII lower-case letter or _, then lower-case letters, digits and _.
func isName(s string) bool {
	if s == "" || isDigit(s[0]) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if c := s[i]; !isLower(c) && !isDigit(c) && c != '_' {
			return false
		}
	}
	return true
}

// The table in the form of SPEC.md §7.4: every object's keys in byte order, one entry to a line, and empty
// namespaces and classes left out.
func (t *Names) Encode() []byte {
	var b bytes.Buffer
	b.WriteString("{\n")
	type section struct {
		key   string
		write func(*bytes.Buffer)
	}
	var sections []section
	if classes := t.encodedClasses(); len(classes) > 0 {
		sections = append(sections, section{"classes", func(b *bytes.Buffer) {
			names := slices.Sorted(maps.Keys(classes))
			for i, name := range names {
				fmt.Fprintf(b, "    %s: {\n", quote(name))
				writeSpellNames(b, classes[name], "      ")
				b.WriteString("    }")
				b.WriteString(comma(i, len(names)))
			}
		}})
	}
	if len(t.Enchants) > 0 {
		sections = append(sections, section{"enchants", func(b *bytes.Buffer) {
			names := slices.Sorted(maps.Keys(t.Enchants))
			for i, name := range names {
				e := t.Enchants[name]
				slots := make([]string, len(e.Slots))
				for j, s := range e.Slots {
					slots[j] = quote(shortName(slotEnum, int32(s)))
				}
				fmt.Fprintf(b, "    %s: {\"effectId\": %d, \"slots\": [%s]}%s", quote(name), e.EffectID, strings.Join(slots, ", "), comma(i, len(names)))
			}
		}})
	}
	if len(t.ICDOnlyItems) > 0 {
		sections = append(sections, section{"icdOnlyItems", func(b *bytes.Buffer) {
			names := slices.Sorted(maps.Keys(t.ICDOnlyItems))
			for i, name := range names {
				fmt.Fprintf(b, "    %s: [%s]%s", quote(name), specList(t.ICDOnlyItems[name]), comma(i, len(names)))
			}
		}})
	}
	if len(t.Items) > 0 {
		sections = append(sections, section{"items", func(b *bytes.Buffer) { writeIDs(b, t.Items) }})
	}
	if hasSpellNames(t.Shared) {
		sections = append(sections, section{"shared", func(b *bytes.Buffer) { writeSpellNames(b, t.Shared, "    ") }})
	}
	if len(t.Suffixes) > 0 {
		sections = append(sections, section{"suffixes", func(b *bytes.Buffer) { writeIDs(b, t.Suffixes) }})
	}
	for i, s := range sections {
		fmt.Fprintf(&b, "  %s: {\n", quote(s.key))
		s.write(&b)
		b.WriteString("  }")
		b.WriteString(comma(i, len(sections)))
	}
	b.WriteString("}\n")
	return b.Bytes()
}

func (t *Names) encodedClasses() map[string]SpellNames {
	classes := map[string]SpellNames{}
	for class, names := range t.Classes {
		if hasSpellNames(names) {
			classes[ClassName(class)] = names
		}
	}
	return classes
}

func hasSpellNames(names SpellNames) bool {
	for _, entries := range names {
		if len(entries) > 0 {
			return true
		}
	}
	return false
}

func writeSpellNames(b *bytes.Buffer, names SpellNames, indent string) {
	var spaces []Namespace
	for ns, entries := range names {
		if len(entries) > 0 {
			spaces = append(spaces, ns)
		}
	}
	slices.Sort(spaces)
	for i, ns := range spaces {
		fmt.Fprintf(b, "%s%s: {\n", indent, quote(string(ns)))
		entries := names[ns]
		keys := slices.Sorted(maps.Keys(entries))
		for j, name := range keys {
			e := entries[name]
			fmt.Fprintf(b, "%s  %s: {\"spellId\": %d", indent, quote(name), e.SpellID)
			if e.Tag != 0 {
				fmt.Fprintf(b, ", \"tag\": %d", e.Tag)
			}
			if e.Rank != 0 {
				fmt.Fprintf(b, ", \"rank\": %d", e.Rank)
			}
			fmt.Fprintf(b, ", \"registeredBy\": [%s]", specList(e.RegisteredBy))
			if len(e.ICDOnlyBy) > 0 {
				fmt.Fprintf(b, ", \"icdOnlyBy\": [%s]", specList(e.ICDOnlyBy))
			}
			if e.Uncastable {
				fmt.Fprintf(b, ", \"uncastable\": true")
			}
			fmt.Fprintf(b, "}%s", comma(j, len(keys)))
		}
		fmt.Fprintf(b, "%s}%s", indent, comma(i, len(spaces)))
	}
}

// The specs, quoted and in byte order, as the members of a JSON list.
func specList(specs []string) string {
	quoted := slices.Sorted(slices.Values(specs))
	for i, s := range quoted {
		quoted[i] = quote(s)
	}
	return strings.Join(quoted, ", ")
}

func writeIDs(b *bytes.Buffer, names map[string]int32) {
	keys := slices.Sorted(maps.Keys(names))
	for i, name := range keys {
		fmt.Fprintf(b, "    %s: %d%s", quote(name), names[name], comma(i, len(keys)))
	}
}

func quote(s string) string {
	q, _ := json.Marshal(s)
	return string(q)
}

func comma(i, n int) string {
	if i < n-1 {
		return ",\n"
	}
	return "\n"
}
