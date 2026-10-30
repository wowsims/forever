package gen

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

// A name of the table: a class's or the shared spell-like entry, or an item, enchant or suffix entry.
type Pin struct {
	Class     proto.Class // ClassUnknown for a shared, item, enchant or suffix entry.
	Namespace siml.Namespace
	Name      string
}

// Reads <class>.<namespace>.<name>, shared.<namespace>.<name>, item.<name>, enchant.<name> or suffix.<name>.
func ParsePin(s string) (Pin, error) {
	parts := strings.Split(s, ".")
	switch {
	case len(parts) == 2 && !siml.Namespace(parts[0]).SpellLike() && slices.Contains(siml.AllNamespaces(), siml.Namespace(parts[0])):
		return Pin{Namespace: siml.Namespace(parts[0]), Name: parts[1]}, nil
	case len(parts) == 3 && siml.Namespace(parts[1]).SpellLike():
		if parts[0] == "shared" {
			return Pin{Namespace: siml.Namespace(parts[1]), Name: parts[2]}, nil
		}
		if class, ok := siml.ClassNamed(parts[0]); ok {
			return Pin{Class: class, Namespace: siml.Namespace(parts[1]), Name: parts[2]}, nil
		}
		return Pin{}, fmt.Errorf("%s: %s is no class", s, parts[0])
	}
	return Pin{}, fmt.Errorf("%s is not <class>.<namespace>.<name>, shared.<namespace>.<name>, item.<name>, enchant.<name> or suffix.<name>", s)
}

func (p Pin) String() string {
	if !p.Namespace.SpellLike() {
		return string(p.Namespace) + "." + p.Name
	}
	return scopeName(p.Class) + "." + string(p.Namespace) + "." + p.Name
}

// What a run changed and could not do.
type Report struct {
	Added        []string // One line per new entry, such as warrior.spell.heroic_strike = spell 25286 tag 1.
	Qualified    []string // New class entries whose plain name a shared entry holds, which the class's files mean by it.
	Repinned     []Repin  // The names a repin handed back to the new-name rules, by name.
	Dropped      []string // Names HEAD's table does not hold whose entries nothing gives now, such as hunter.spell.x = spell 1.
	Stale        []string // Pinned names no build registers while an id of the same game name is registered.
	Unregistered int      // Pinned spell-like entries no build registers.
	Skipped      []string
}

// A name a repin handed back to the new-name rules: what it meant before, and what it means now.
type Repin struct {
	Pin
	Was, Now string
}

// The table old grows into with c: every entry of old keeps its name and meaning, the registered-by lists are
// regenerated, and each spell, aura, item, enchant and random suffix of c that old does not name gets a new entry.
// With a pin, the entries of that name's namespace whose game names give the same name are dropped first, so the
// new-name rules decide what each of their names means again.
func Update(old *siml.Names, c *Collection, pin *Pin) (*siml.Names, *Report, error) {
	u := newUpdater(old, c)
	var before map[Pin]string
	if pin != nil {
		var err error
		if before, err = u.unpin(*pin); err != nil {
			return nil, nil, err
		}
	}
	u.registeredBy()
	// The classes first, so what a class casts takes its game name's name before what gear adds.
	for _, class := range slices.Sorted(maps.Keys(c.Classes)) {
		for _, ns := range spellNamespaces() {
			u.classSpells(class, ns)
		}
	}
	for _, ns := range spellNamespaces() {
		u.sharedSpells(ns)
	}
	u.items()
	u.icdItems()
	u.enchants()
	u.suffixes()
	for _, p := range slices.SortedFunc(maps.Keys(before), func(a, b Pin) int { return strings.Compare(a.Name, b.Name) }) {
		now, ok := u.meaning(p)
		if !ok {
			now = "nothing"
		}
		u.report.Repinned = append(u.report.Repinned, Repin{p, before[p], now})
	}
	u.stale()
	slices.Sort(u.report.Qualified)
	return u.t, u.report, nil
}

func spellNamespaces() []siml.Namespace {
	var spaces []siml.Namespace
	for _, ns := range siml.AllNamespaces() {
		if ns.SpellLike() {
			spaces = append(spaces, ns)
		}
	}
	slices.Sort(spaces)
	return spaces
}

type updater struct {
	c      *Collection
	t      *siml.Names
	report *Report

	// The database's entries by id or key, each the first the database lists.
	itemsByID    map[int32]Item
	rowsByKey    map[string]EnchantRow
	suffixesByID map[int32]Suffix
}

func newUpdater(old *siml.Names, c *Collection) *updater {
	t := &siml.Names{
		Classes:  map[proto.Class]siml.SpellNames{},
		Shared:   cloneSpells(old.Shared),
		Items:    map[string]int32{},
		Enchants: map[string]siml.EnchantName{},
		Suffixes: map[string]int32{},
	}
	for class, names := range old.Classes {
		t.Classes[class] = cloneSpells(names)
	}
	maps.Copy(t.Items, old.Items)
	maps.Copy(t.Suffixes, old.Suffixes)
	for name, e := range old.Enchants {
		t.Enchants[name] = siml.EnchantName{EffectID: e.EffectID, Slots: slices.Clone(e.Slots)}
	}
	u := &updater{c: c, t: t, report: &Report{Skipped: slices.Clone(c.Skipped)},
		itemsByID: map[int32]Item{}, rowsByKey: map[string]EnchantRow{}, suffixesByID: map[int32]Suffix{}}
	for _, it := range c.Items {
		if _, ok := u.itemsByID[it.ID]; !ok {
			u.itemsByID[it.ID] = it
		}
	}
	for _, row := range c.Enchants {
		if key := enchantKey(row.EffectID, row.Slots); u.rowsByKey[key].EffectID == 0 {
			u.rowsByKey[key] = row
		}
	}
	for _, s := range c.Suffixes {
		if _, ok := u.suffixesByID[s.ID]; !ok {
			u.suffixesByID[s.ID] = s
		}
	}
	return u
}

func cloneSpells(names siml.SpellNames) siml.SpellNames {
	c := siml.SpellNames{}
	for ns, entries := range names {
		c[ns] = map[string]siml.SpellName{}
		for name, e := range entries {
			e.RegisteredBy = slices.Clone(e.RegisteredBy)
			c[ns][name] = e
		}
	}
	return c
}

// Drops the entry p names and every other entry of its namespace and scope whose game name gives the same name,
// and returns what each dropped name meant.
func (u *updater) unpin(p Pin) (map[Pin]string, error) {
	if _, ok := u.meaning(p); !ok {
		return nil, fmt.Errorf("-repin %s: the table holds no such name", p)
	}
	base := u.base(p)
	before := map[Pin]string{}
	drop := func(names []string, del func(string)) {
		for _, name := range names {
			q := Pin{p.Class, p.Namespace, name}
			if name == p.Name || base != "" && u.base(q) == base {
				before[q], _ = u.meaning(q)
				del(name)
			}
		}
	}
	switch p.Namespace {
	case siml.NamespaceItem:
		drop(slices.Collect(maps.Keys(u.t.Items)), func(name string) { delete(u.t.Items, name) })
	case siml.NamespaceEnchant:
		drop(slices.Collect(maps.Keys(u.t.Enchants)), func(name string) { delete(u.t.Enchants, name) })
	case siml.NamespaceSuffix:
		drop(slices.Collect(maps.Keys(u.t.Suffixes)), func(name string) { delete(u.t.Suffixes, name) })
	default:
		entries := u.scope(p.Class)[p.Namespace]
		drop(slices.Collect(maps.Keys(entries)), func(name string) { delete(entries, name) })
	}
	return before, nil
}

// The name the game name of what p names gives it; "" when the database no longer holds it.
func (u *updater) base(p Pin) string {
	switch p.Namespace {
	case siml.NamespaceItem:
		if it, ok := u.itemsByID[u.t.Items[p.Name]]; ok {
			return baseName(it.Name, "item", it.ID)
		}
	case siml.NamespaceEnchant:
		e := u.t.Enchants[p.Name]
		if row, ok := u.rowsByKey[enchantKey(e.EffectID, e.Slots)]; ok {
			return baseName(row.Name, "enchant", row.EffectID)
		}
	case siml.NamespaceSuffix:
		if s, ok := u.suffixesByID[u.t.Suffixes[p.Name]]; ok {
			return baseName(s.Name, "suffix", s.ID)
		}
	default:
		e := u.scope(p.Class)[p.Namespace][p.Name]
		return baseName(u.c.SpellNames[e.SpellID], "spell", e.SpellID)
	}
	return ""
}

// What the entry p names means, as the report writes it.
func (u *updater) meaning(p Pin) (string, bool) {
	switch p.Namespace {
	case siml.NamespaceItem:
		id, ok := u.t.Items[p.Name]
		return "item " + strconv.Itoa(int(id)), ok
	case siml.NamespaceEnchant:
		e, ok := u.t.Enchants[p.Name]
		return enchantText(e), ok
	case siml.NamespaceSuffix:
		id, ok := u.t.Suffixes[p.Name]
		return "suffix " + strconv.Itoa(int(id)), ok
	}
	e, ok := u.scope(p.Class)[p.Namespace][p.Name]
	return entryText(e), ok
}

// The spell-like entries of a class, or the shared ones for ClassUnknown.
func (u *updater) scope(class proto.Class) siml.SpellNames {
	if class == proto.Class_ClassUnknown {
		if u.t.Shared == nil {
			u.t.Shared = siml.SpellNames{}
		}
		return u.t.Shared
	}
	if u.t.Classes[class] == nil {
		u.t.Classes[class] = siml.SpellNames{}
	}
	return u.t.Classes[class]
}

func scopeName(class proto.Class) string {
	if class == proto.Class_ClassUnknown {
		return "shared"
	}
	return siml.ClassName(class)
}

func entryText(e siml.SpellName) string {
	s := "spell " + strconv.Itoa(int(e.SpellID))
	if e.Tag != 0 {
		s += " tag " + strconv.Itoa(int(e.Tag))
	}
	if e.Rank != 0 {
		s += " rank " + strconv.Itoa(int(e.Rank))
	}
	return s
}

func enchantText(e siml.EnchantName) string {
	slots := make([]string, len(e.Slots))
	for i, s := range e.Slots {
		slots[i] = strings.TrimPrefix(s.String(), "ItemSlot")
	}
	return fmt.Sprintf("enchant %d (%s)", e.EffectID, strings.Join(slots, ", "))
}

// The specs whose builds register id in ns, those of class only for a class's entry.
func (u *updater) specs(class proto.Class, ns siml.Namespace, id Identity) []string {
	all := u.c.Registered[ns][id].Specs
	if class == proto.Class_ClassUnknown {
		return slices.Clone(all)
	}
	var specs []string
	for _, s := range all {
		if slices.Contains(u.c.Specs[class], s) {
			specs = append(specs, s)
		}
	}
	return specs
}

func (u *updater) registeredBy() {
	for class, names := range u.t.Classes {
		for ns, entries := range names {
			for name, e := range entries {
				entries[name] = u.withSpecs(class, ns, e)
			}
		}
	}
	for ns, entries := range u.t.Shared {
		for name, e := range entries {
			entries[name] = u.withSpecs(proto.Class_ClassUnknown, ns, e)
		}
	}
}

// e with its spec lists written from what the builds register: registeredBy, and of its specs those whose builds
// list the aura only as an internal cooldown trigger (icdOnlyBy); and for a spell no build registers castable,
// uncastable.
func (u *updater) withSpecs(class proto.Class, ns siml.Namespace, e siml.SpellName) siml.SpellName {
	id := Identity{e.SpellID, e.Tag}
	reg := u.c.Registered[ns][id]
	e.RegisteredBy, e.ICDOnlyBy = u.specs(class, ns, id), nil
	e.Uncastable = ns == siml.NamespaceSpell && len(e.RegisteredBy) > 0 && !reg.Castable
	for _, s := range e.RegisteredBy {
		if slices.Contains(reg.ICDSpecs, s) {
			e.ICDOnlyBy = append(e.ICDOnlyBy, s)
		}
	}
	return e
}

// The names a new entry must not take: those that mean something else where the entry is looked up.
type conflicts map[string][]string // Name → the keys of what it names.

func (c conflicts) add(name, key string) {
	if !slices.Contains(c[name], key) {
		c[name] = append(c[name], key)
	}
}

func (c conflicts) free(name, key string) bool {
	for _, k := range c[name] {
		if k != key {
			return false
		}
	}
	return !siml.Reserved(name)
}

// A thing to name, or one a name already holds.
type candidate struct {
	key   string   // What the name means, unique among what one namespace can name.
	base  string   // The name its game name gives it (baseName).
	id    int32    // The spell id, item id, effect id or suffix id, which the last qualifier writes.
	quals []string // The qualifiers that can tell it apart, in order: tag then rank, or the item level; "" where it has none.
	again string   // What follows _<id> when another thing of the base has the same id: a spell-like entry's tag qualifier.
}

func spellKey(id Identity) string { return fmt.Sprintf("spell %d %d", id.SpellID, id.Tag) }

func tagText(tag int32) string {
	switch {
	case tag == -1:
		return "_external"
	case tag < 0:
		return "_tag_m" + strconv.Itoa(int(-tag))
	}
	return "_tag" + strconv.Itoa(int(tag))
}

// The rank qualifier; none for rank 0, which is no rank.
func rankText(rank int32) string {
	switch {
	case rank < 0:
		return "_rank_m" + strconv.Itoa(int(-rank))
	case rank == 0:
		return ""
	}
	return "_rank" + strconv.Itoa(int(rank))
}

// The name a game name gives a thing of kind: identifier(game name), with a leading enchant_ dropped from an
// enchant's when a letter follows; <kind>_<id> when that is empty, <kind>_<base> when it starts with a digit, and
// <base>_<kind> when it is a reserved word.
func baseName(game, kind string, id int32) string {
	base := siml.Identifier(game)
	if rest, ok := strings.CutPrefix(base, "enchant_"); ok && kind == "enchant" && rest != "" && rest[0] >= 'a' && rest[0] <= 'z' {
		base = rest
	}
	switch {
	case base == "":
		base = kind + "_" + strconv.Itoa(int(id))
	case base[0] >= '0' && base[0] <= '9':
		base = kind + "_" + base
	}
	if siml.Reserved(base) {
		base += "_" + kind
	}
	return base
}

// Names the new candidates, each slice in preference order within its base. holders are the things that already
// hold names where the candidates are looked up. The first candidate of a base takes it when it is free; every other
// one takes the base with the first of its qualifiers that no other candidate or holder of the base shares, else
// with _<id> (followed by the tag qualifier when another of them has the same spell id); and _2, _3 … follow while
// that name is taken.
func assign(fresh []candidate, holders []candidate, taken conflicts) map[string]string {
	byBase := map[string][]candidate{}
	for _, c := range fresh {
		byBase[c.base] = append(byBase[c.base], c)
	}
	held := map[string][]candidate{}
	for _, h := range holders {
		held[h.base] = append(held[h.base], h)
	}
	names := map[string]string{}
	for _, base := range slices.Sorted(maps.Keys(byBase)) {
		group := byBase[base]
		everyone := append(slices.Clone(group), held[base]...)
		for i, c := range group {
			name := base
			if i > 0 || !taken.free(base, c.key) {
				q := base + qualifier(c, everyone)
				name = q
				for k := 2; !taken.free(name, c.key); k++ {
					name = q + "_" + strconv.Itoa(k)
				}
			}
			taken.add(name, c.key)
			names[c.key] = name
		}
	}
	return names
}

// The first qualifier of c that no other of everyone shares, else _<id>, with c.again after it when another has
// the same id.
func qualifier(c candidate, everyone []candidate) string {
	for i, q := range c.quals {
		if q == "" {
			continue
		}
		if !slices.ContainsFunc(everyone, func(o candidate) bool { return o.key != c.key && i < len(o.quals) && o.quals[i] == q }) {
			return q
		}
	}
	id := "_" + strconv.Itoa(int(c.id))
	if slices.ContainsFunc(everyone, func(o candidate) bool { return o.key != c.key && o.id == c.id }) {
		id += c.again
	}
	return id
}

func (u *updater) spellCandidate(id Identity, rank int32) candidate {
	return candidate{
		key:   spellKey(id),
		base:  baseName(u.c.SpellNames[id.SpellID], "spell", id.SpellID),
		id:    id.SpellID,
		quals: []string{tagText(id.Tag), rankText(rank)},
		again: tagText(id.Tag),
	}
}

// Preference among new spell-like entries for one name: castable first in the spell namespace, then tag 0, the
// highest rank and the lowest spell id; the lowest tag last, so the order is total.
func (u *updater) preferSpells(ns siml.Namespace, ids []Identity) {
	reg := u.c.Registered[ns]
	slices.SortFunc(ids, func(a, b Identity) int {
		ra, rb := reg[a], reg[b]
		if ns == siml.NamespaceSpell && ra.Castable != rb.Castable {
			if ra.Castable {
				return -1
			}
			return 1
		}
		if (a.Tag == 0) != (b.Tag == 0) {
			if a.Tag == 0 {
				return -1
			}
			return 1
		}
		return cmp.Or(cmp.Compare(rb.Rank, ra.Rank), cmp.Compare(a.SpellID, b.SpellID), cmp.Compare(a.Tag, b.Tag))
	})
}

// Adds the spell-like entries the shared scope lacks. A shared name must be free in the shared entries and in
// every class's, so adding it changes what no name means anywhere.
func (u *updater) sharedSpells(ns siml.Namespace) {
	scopes := []siml.SpellNames{u.t.Shared}
	for _, class := range slices.Sorted(maps.Keys(u.t.Classes)) {
		scopes = append(scopes, u.t.Classes[class])
	}
	named := identities(u.t.Shared[ns])
	var need []Identity
	for id := range u.c.Shared[ns] {
		if !named[id] {
			need = append(need, id)
		}
	}
	u.addSpells(proto.Class_ClassUnknown, ns, need, scopes)
}

// Adds the spell-like entries a class lacks: what its builds register that neither its own nor the shared entries
// name. Its new names must be free in both.
func (u *updater) classSpells(class proto.Class, ns siml.Namespace) {
	own, shared := identities(u.t.Classes[class][ns]), identities(u.t.Shared[ns])
	var need []Identity
	for id := range u.c.Classes[class][ns] {
		if !own[id] && !shared[id] {
			need = append(need, id)
		}
	}
	u.addSpells(class, ns, need, []siml.SpellNames{u.t.Classes[class], u.t.Shared})
}

func identities(entries map[string]siml.SpellName) map[Identity]bool {
	ids := map[Identity]bool{}
	for _, e := range entries {
		ids[Identity{e.SpellID, e.Tag}] = true
	}
	return ids
}

func (u *updater) addSpells(class proto.Class, ns siml.Namespace, need []Identity, scopes []siml.SpellNames) {
	if len(need) == 0 {
		return
	}
	taken := conflicts{}
	var holders []candidate
	for _, scope := range scopes {
		for _, name := range slices.Sorted(maps.Keys(scope[ns])) {
			e := scope[ns][name]
			id := Identity{e.SpellID, e.Tag}
			taken.add(name, spellKey(id))
			holders = append(holders, u.spellCandidate(id, e.Rank))
		}
	}
	u.preferSpells(ns, need)
	entries := u.scope(class)
	if entries[ns] == nil {
		entries[ns] = map[string]siml.SpellName{}
	}
	rank := func(id Identity) int32 { return u.c.Registered[ns][id].Rank }
	candidateOf := func(id Identity) candidate { return u.spellCandidate(id, rank(id)) }
	addNames(u, need, candidateOf, holders, taken, func(id Identity, name string) string {
		e := u.withSpecs(class, ns, siml.SpellName{SpellID: id.SpellID, Tag: id.Tag, Rank: rank(id)})
		entries[ns][name] = e
		line := fmt.Sprintf("%s.%s.%s = %s", scopeName(class), ns, name, entryText(e))
		if base := candidateOf(id).base; class != proto.Class_ClassUnknown && name != base {
			if shared, ok := u.t.Shared[ns][base]; ok {
				u.report.Qualified = append(u.report.Qualified, fmt.Sprintf("%s: a %s file's %s means shared.%s.%s (%s); -repin shared.%s.%s, run on the table without this entry, gives the name to the class",
					line, scopeName(class), base, ns, base, entryText(shared), ns, base))
			}
		}
		return line
	})
}

// Names need, in preference order, by the candidates candidateOf gives: holders are what already holds a name where
// they are read, and taken the names they must not take. add writes each entry and gives its report line.
func addNames[T any](u *updater, need []T, candidateOf func(T) candidate, holders []candidate, taken conflicts, add func(T, string) string) {
	fresh := make([]candidate, len(need))
	for i, x := range need {
		fresh[i] = candidateOf(x)
	}
	names := assign(fresh, holders, taken)
	added := make([]string, len(need))
	for i, x := range need {
		added[i] = add(x, names[fresh[i].key])
	}
	slices.Sort(added)
	u.report.Added = append(u.report.Added, added...)
}

// What the entries of an item, enchant or suffix namespace hold: the names they take, the candidates of those the
// database still gives, which the qualifiers of new names tell apart, and the keys they name.
func held[V any](entries map[string]V, key func(V) string, current func(V) (candidate, bool)) (conflicts, []candidate, map[string]bool) {
	taken := conflicts{}
	var holders []candidate
	named := map[string]bool{}
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		v := entries[name]
		named[key(v)] = true
		taken.add(name, key(v))
		if c, ok := current(v); ok {
			holders = append(holders, c)
		}
	}
	return taken, holders, named
}

func itemKey(id int32) string { return "item " + strconv.Itoa(int(id)) }

func (u *updater) itemCandidate(it Item) candidate {
	level := ""
	if it.Level > 0 {
		level = "_ilvl" + strconv.Itoa(int(it.Level))
	}
	return candidate{key: itemKey(it.ID), base: baseName(it.Name, "item", it.ID), id: it.ID, quals: []string{level}}
}

// Adds an entry for every item of the database the table does not name. Preference: the highest item level, then
// the lowest id; the item level is the readable qualifier.
func (u *updater) items() {
	taken, holders, named := held(u.t.Items, itemKey, func(id int32) (candidate, bool) {
		it, ok := u.itemsByID[id]
		return u.itemCandidate(it), ok
	})
	var need []Item
	for _, it := range u.itemsByID {
		if !named[itemKey(it.ID)] {
			need = append(need, it)
		}
	}
	slices.SortFunc(need, func(a, b Item) int { return cmp.Or(cmp.Compare(b.Level, a.Level), cmp.Compare(a.ID, b.ID)) })
	addNames(u, need, u.itemCandidate, holders, taken, func(it Item, name string) string {
		u.t.Items[name] = it.ID
		return fmt.Sprintf("item.%s = item %d", name, it.ID)
	})
}

func enchantKey(effectID int32, slots []proto.ItemSlot) string {
	return fmt.Sprintf("enchant %d %v", effectID, slots)
}

func enchantCandidate(row EnchantRow) candidate {
	return candidate{key: enchantKey(row.EffectID, row.Slots), base: baseName(row.Name, "enchant", row.EffectID), id: row.EffectID}
}

// Adds an entry for every effect id and slots of the database's enchant rows that the table does not name. Rows
// of one effect id that fit the same slots share one entry, named by the first of them; preference among new
// entries is the lowest effect id, and no qualifier is readable.
// The items whose effect a spec knows only as an internal cooldown trigger, written again from what the builds list.
func (u *updater) icdItems() {
	u.t.ICDOnlyItems = map[string][]string{}
	for name, id := range u.t.Items {
		if specs := u.c.ICDItems[id]; len(specs) > 0 {
			u.t.ICDOnlyItems[name] = slices.Clone(specs)
		}
	}
}

func (u *updater) enchants() {
	key := func(e siml.EnchantName) string { return enchantKey(e.EffectID, e.Slots) }
	taken, holders, named := held(u.t.Enchants, key, func(e siml.EnchantName) (candidate, bool) {
		row, ok := u.rowsByKey[key(e)]
		return enchantCandidate(row), ok
	})
	var need []EnchantRow
	for k, row := range u.rowsByKey {
		if !named[k] {
			need = append(need, row)
		}
	}
	slices.SortFunc(need, func(a, b EnchantRow) int {
		return cmp.Or(cmp.Compare(a.EffectID, b.EffectID), slices.Compare(a.Slots, b.Slots))
	})
	addNames(u, need, enchantCandidate, holders, taken, func(row EnchantRow, name string) string {
		e := siml.EnchantName{EffectID: row.EffectID, Slots: slices.Clone(row.Slots)}
		u.t.Enchants[name] = e
		return fmt.Sprintf("enchant.%s = %s", name, enchantText(e))
	})
}

func suffixKey(id int32) string { return "suffix " + strconv.Itoa(int(id)) }

func suffixCandidate(s Suffix) candidate {
	return candidate{key: suffixKey(s.ID), base: baseName(s.Name, "suffix", s.ID), id: s.ID}
}

// Adds an entry for every random suffix the table does not name. Preference: the lowest id; no qualifier is
// readable.
func (u *updater) suffixes() {
	taken, holders, named := held(u.t.Suffixes, suffixKey, func(id int32) (candidate, bool) {
		s, ok := u.suffixesByID[id]
		return suffixCandidate(s), ok
	})
	var need []Suffix
	for _, s := range u.suffixesByID {
		if !named[suffixKey(s.ID)] {
			need = append(need, s)
		}
	}
	slices.SortFunc(need, func(a, b Suffix) int { return cmp.Compare(a.ID, b.ID) })
	addNames(u, need, suffixCandidate, holders, taken, func(s Suffix, name string) string {
		u.t.Suffixes[name] = s.ID
		return fmt.Sprintf("suffix.%s = suffix %d", name, s.ID)
	})
}

// Counts the pinned spell-like names no build registers, and reports those whose game name gives the name of an
// id the namespace's builds do register.
func (u *updater) stale() {
	check := func(class proto.Class, ns siml.Namespace, entries map[string]siml.SpellName, registered map[Identity]bool) {
		for _, name := range slices.Sorted(maps.Keys(entries)) {
			e := entries[name]
			if len(e.RegisteredBy) > 0 {
				continue
			}
			u.report.Unregistered++
			base := baseName(u.c.SpellNames[e.SpellID], "spell", e.SpellID)
			var now []string
			for id := range registered {
				if id != (Identity{e.SpellID, e.Tag}) && baseName(u.c.SpellNames[id.SpellID], "spell", id.SpellID) == base {
					now = append(now, entryText(siml.SpellName{SpellID: id.SpellID, Tag: id.Tag}))
				}
			}
			if len(now) > 0 {
				slices.Sort(now)
				u.report.Stale = append(u.report.Stale, fmt.Sprintf("%s.%s.%s is pinned to %s; the sim registers %s",
					scopeName(class), ns, name, entryText(siml.SpellName{SpellID: e.SpellID, Tag: e.Tag}), strings.Join(now, ", ")))
			}
		}
	}
	for _, ns := range spellNamespaces() {
		check(proto.Class_ClassUnknown, ns, u.t.Shared[ns], u.c.Shared[ns])
	}
	for _, class := range slices.Sorted(maps.Keys(u.t.Classes)) {
		for _, ns := range spellNamespaces() {
			check(class, ns, u.t.Classes[class][ns], u.c.Classes[class][ns])
		}
	}
}
