package database

import (
	"cmp"
	"database/sql"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/buffmanifest"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// A partial update takes only some spell data groups from the game files database. Every other row
// stays as the committed inputs have it, so one class can move to a new game build while the others
// stay on the build their code was written for.
//
// A spell data group (a group, below) is one of the nine classes by package name, racials, items or
// shared. A row belongs to the group the first matching rule names:
//
//   - a hand-kept extra spell marked racials (the Eureka! rows) belongs to racials;
//   - a row that states a class family belongs to that class wherever it is reached from (a set
//     bonus, an enchant's class buff), because its flags are that class's;
//   - any other row belongs to the groups whose roots reach it, and moves when any of them is taken.
//
// Every run takes items. gen_db writes the item database whole, and the store's item rows have to
// match it. Every run takes shared as well unless -shared=false leaves it out.
const (
	groupItems   = "items"
	groupRacials = "racials"
	groupShared  = "shared"
)

func classGroup(class dbc.DbcClass) string {
	return strings.ToLower(dbc.ClassNameFromDBC(class))
}

func classGroups() []string {
	groups := make([]string, 0, len(dbc.Classes))
	for _, class := range dbc.Classes {
		groups = append(groups, classGroup(class))
	}
	return groups
}

func allSpellDataGroups() []string {
	return append(classGroups(), groupRacials, groupItems, groupShared)
}

// The groups a run takes from the game files database. Nil means every group: a full regeneration.
type SpellDataGroups map[string]bool

// The groups -classes, -racials and -shared name, plus items, which every run takes. classes is a
// comma-separated list of class package names, "all" or "none". With no flag set (shared stays true)
// it returns nil, the default selection.
func ParseSpellDataGroups(classes string, racials, shared bool) (SpellDataGroups, error) {
	classes = strings.TrimSpace(classes)
	if classes == "" && !racials && shared {
		return nil, nil
	}

	only := SpellDataGroups{groupItems: true, groupShared: shared, groupRacials: racials}
	var names []string
	switch classes {
	case "", "none":
	case "all":
		names = classGroups()
	default:
		names = strings.Split(classes, ",")
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if !slices.Contains(classGroups(), name) {
			return nil, fmt.Errorf("unknown class %q: one of %s, all or none; racials are taken with -racials and shared with -shared",
				name, strings.Join(slices.Sorted(slices.Values(classGroups())), ", "))
		}
		only[name] = true
	}
	return only, nil
}

func (only SpellDataGroups) Takes(group string) bool {
	return only == nil || only[group]
}

func (only SpellDataGroups) all() bool {
	return !slices.ContainsFunc(allSpellDataGroups(), func(p string) bool { return !only.Takes(p) })
}

// The game build the database was extracted from, which db2tool records in BuildInfo.
func buildVersion(helper *DBHelper) (string, error) {
	if !helper.tableExists("BuildInfo") {
		return "", fmt.Errorf("the game files database names no game build (no BuildInfo table): stamp the build it holds " +
			"(see docs/spell_data/partial-updates.md), or extract it again with `make db`")
	}
	var version string
	if err := helper.db.QueryRow(`SELECT Version FROM BuildInfo`).Scan(&version); err != nil {
		return "", fmt.Errorf("reading the game build from BuildInfo: %w", err)
	}
	return version, nil
}

// The highest hotfix push db2tool applied to the database, which it records in BuildInfo. A
// database whose BuildInfo has no HotfixPush column reads as 0, the same as one extracted with no
// hotfixes.
func hotfixPush(helper *DBHelper) (int32, error) {
	var columns int
	if err := helper.db.QueryRow(`SELECT count(*) FROM pragma_table_info('BuildInfo') WHERE name = 'HotfixPush'`).Scan(&columns); err != nil {
		return 0, fmt.Errorf("reading the columns of BuildInfo: %w", err)
	}
	if columns == 0 {
		return 0, nil
	}
	var push int32
	if err := helper.db.QueryRow(`SELECT HotfixPush FROM BuildInfo`).Scan(&push); err != nil {
		return 0, fmt.Errorf("reading the hotfix push from BuildInfo: %w", err)
	}
	return push, nil
}

// Refuses a database older in hotfixes than the committed inputs: taking a group from it would drop
// the hotfix rows its committed data holds. Only groups recorded on the database's game build are
// compared, and a group the inputs record no push for passes.
func checkHotfixes(builds map[string]string, hotfixes map[string]int32, only SpellDataGroups, version string, push int32) error {
	var newer []string
	for _, p := range allSpellDataGroups() {
		committed, ok := hotfixes[p]
		if ok && only.Takes(p) && builds[p] == version && committed > push {
			newer = append(newer, fmt.Sprintf("%s (push %d)", p, committed))
		}
	}
	if len(newer) == 0 {
		return nil
	}
	return fmt.Errorf("the game files database is %s with hotfixes up to push %d, and the committed inputs hold later hotfixes for %s: "+
		"log into the game to refresh DBCache.bin (or pass a newer cache with `make db DB2TOOL_FLAGS=\"--dbcache path/to/DBCache.bin\"`) "+
		"and run `make db` again", version, push, strings.Join(newer, ", "))
}

// Refuses the database when the committed inputs hold later hotfixes for a group only takes on the
// same game build.
func RequireCommittedHotfixes(helper *DBHelper, only SpellDataGroups) error {
	builds, hotfixes, err := readCommittedBuilds()
	if err != nil || hotfixes == nil {
		return err
	}
	version, err := buildVersion(helper)
	if err != nil {
		return err
	}
	push, err := hotfixPush(helper)
	if err != nil {
		return err
	}
	return checkHotfixes(builds, hotfixes, only, version, push)
}

// The groups whose committed build matches the database's. -check compares only these and reports a
// group on another build as not ported. Inputs that record no builds are checked whole.
func currentSpellDataGroups(builds map[string]string, version string) (only SpellDataGroups, behind []string) {
	if builds == nil {
		return nil, nil
	}
	only = SpellDataGroups{groupItems: true}
	for _, p := range allSpellDataGroups() {
		build, recorded := builds[p]
		if !recorded || build == version {
			only[p] = true
			continue
		}
		suffix := ""
		if only[p] {
			suffix = "; it is always taken, so it is compared with the database"
		}
		behind = append(behind, fmt.Sprintf("%s is on %s, the game files database is %s%s", p, describeBuild(build), version, suffix))
	}
	return only, behind
}

// The groups a run takes when it names none. While every group records the same game build, that is
// all of them: a full regeneration. Once a partial update has left groups on different builds, it is
// only the groups on the database's build, so a plain run never moves a class that has not been
// ported.
func defaultSpellDataGroups(helper *DBHelper) (SpellDataGroups, []string, error) {
	builds, _, err := readCommittedBuilds()
	if err != nil {
		return nil, nil, err
	}
	distinct := map[string]bool{}
	for _, build := range builds {
		distinct[build] = true
	}
	if len(distinct) <= 1 {
		return nil, nil, nil
	}
	version, err := buildVersion(helper)
	if err != nil {
		return nil, nil, err
	}
	only, behind := currentSpellDataGroups(builds, version)
	return only, behind, nil
}

// The groups a run takes: the ones -classes, -racials and -shared name, or the default groups when
// none is set, with a line for each group the default leaves on its committed build.
func ResolveSpellDataGroups(helper *DBHelper, classes string, racials, shared bool) (SpellDataGroups, []string, error) {
	only, err := ParseSpellDataGroups(classes, racials, shared)
	if err != nil || only != nil {
		return only, nil, err
	}
	return defaultSpellDataGroups(helper)
}

// The ids each group starts from: its captured roots, plus the hand-kept extra spells and buff
// manifest rows that belong to it. The extras and the manifest are not captured as roots (see
// withExtraIDs), so they are added here, the same way for every set of inputs.
func groupSeeds(roots map[string][]int32) map[string][]int32 {
	seeds := map[string][]int32{}
	for p, ids := range roots {
		seeds[p] = slices.Clone(ids)
	}
	for _, extra := range overrides.ExtraSpells {
		p := groupShared
		if extra.SpellDataGroup == groupRacials {
			p = groupRacials
		}
		seeds[p] = append(seeds[p], extra.SpellID)
	}
	for _, spec := range buffmanifest.All() {
		p := groupShared
		for _, class := range dbc.Classes {
			if class.ProtoClass == spec.Owner && spec.Owner != proto.Class_ClassUnknown {
				p = classGroup(class)
			}
		}
		seeds[p] = append(seeds[p], buffSpecIDs(spec)...)
	}
	return seeds
}

func racialExtraIDs() map[int32]bool {
	ids := map[int32]bool{}
	for _, extra := range overrides.ExtraSpells {
		if extra.SpellDataGroup == groupRacials {
			ids[extra.SpellID] = true
		}
	}
	return ids
}

func familyGroup(family int32) (string, bool) {
	for _, class := range dbc.Classes {
		if f, ok := core.ClassSpellFamilies[class.ProtoClass]; ok && f == family {
			return classGroup(class), true
		}
	}
	return "", false
}

// The group each row of a set of inputs belongs to.
type rowOwners struct {
	forced  map[int32]string   // a racials extra's own row, or a row stating a class family
	reached map[int32][]string // the groups whose seeds reach a row, in allSpellDataGroups order
}

// Whether the store the inputs render carries the id.
func (o rowOwners) has(id int32) bool {
	return len(o.reached[id]) > 0
}

// Computed once per set of inputs, since the merge and the change report both need it.
func ownersOf(in *storeInputs) rowOwners {
	if in.owners != nil {
		return *in.owners
	}
	t := &in.spellTables
	seeds := groupSeeds(in.SpellDataGroupRoots)
	o := rowOwners{forced: map[int32]string{}, reached: map[int32][]string{}}
	for _, p := range allSpellDataGroups() {
		for _, id := range reachableSpells(t, namedIDs(t, seeds[p])) {
			o.reached[id] = append(o.reached[id], p)
		}
	}

	racial := racialExtraIDs()
	for id := range o.reached {
		if racial[id] {
			o.forced[id] = groupRacials
		} else if p, ok := familyGroup(t.ClassOptions[id].Family); ok {
			o.forced[id] = p
		}
	}
	in.owners = &o
	return o
}

// Every group's roots in one sorted list, as storeInputs.Roots holds them.
func rootUnion(t *spellTables, roots map[string][]int32) []int32 {
	var lists [][]int32
	for _, p := range allSpellDataGroups() {
		lists = append(lists, roots[p])
	}
	return namedIDs(t, lists...)
}

// A committed root of a group that is not taken never moves, because that group's committed class
// file names the row and was written against it. Any other row moves with the group its racials tag or
// family names, or, when it has neither, as soon as a taken group reaches it.
func (o rowOwners) taken(id int32, only SpellDataGroups, kept map[int32]bool) bool {
	if kept[id] {
		return false
	}
	if p, ok := o.forced[id]; ok {
		return only.Takes(p)
	}
	return slices.ContainsFunc(o.reached[id], only.Takes)
}

// The inputs of a partial update. Rows, roots, talent nodes and builds come from fresh for the groups
// only takes and from old for the rest. The notes report what the rules alone would not show: inputs
// that record no groups being adopted, a spell taken because the new build adds it, a dropped name,
// and racial rows on a different build from their class.
func mergeStoreInputs(old, fresh *storeInputs, only SpellDataGroups, version string) (*storeInputs, []string, error) {
	var notes []string
	if old.SpellDataGroupRoots == nil {
		adopted, adoptNotes, err := adoptSpellDataGroups(old, fresh)
		if err != nil {
			return nil, nil, err
		}
		old = adopted
		notes = append(adoptNotes, fmt.Sprintf("%s records no spell data groups, so its roots and talent nodes are split the way %s splits them",
			spellStoreInputsPath, version))
	}

	kept := map[int32]bool{}
	for _, p := range allSpellDataGroups() {
		if !only.Takes(p) {
			for _, id := range old.SpellDataGroupRoots[p] {
				kept[id] = true
			}
		}
	}

	oldOwners, freshOwners := ownersOf(old), ownersOf(fresh)

	// The committed store's ids are the ones its inputs have rows for. Names lists every spell of the
	// build, so a hand-kept seed added since is named there without being carried. A taken row gets the
	// new build's version, except when the new build still names the spell but no longer reaches it. Its
	// rows were not captured then, so the committed row stays for the kept rows that still reach it.
	oldIDs := map[int32]bool{}
	moved := map[int32]bool{}
	for id := range oldOwners.reached {
		if !hasRows(&old.spellTables, id) {
			continue
		}
		oldIDs[id] = true
		_, named := fresh.Names[id]
		if oldOwners.taken(id, only, kept) && (freshOwners.has(id) || !named) {
			moved[id] = true
		}
	}
	for id := range freshOwners.reached {
		if freshOwners.taken(id, only, kept) {
			moved[id] = true
		}
	}

	roots := map[string][]int32{}
	source := func(p string) *storeInputs {
		if only.Takes(p) {
			return fresh
		}
		return old
	}
	for _, p := range allSpellDataGroups() {
		roots[p] = source(p).SpellDataGroupRoots[p]
	}

	// A spell the committed store does not carry is one the new build adds, when a taken row or root, a
	// hand-kept extra spell or a buff manifest row reaches it on the new build. No committed file names
	// it, so it comes from the new build whatever its group. The walk follows only rows taken from the
	// new build, because a kept row reaches what its committed version names.
	queue := slices.Sorted(maps.Keys(moved))
	for _, p := range allSpellDataGroups() {
		if only.Takes(p) {
			queue = append(queue, fresh.SpellDataGroupRoots[p]...)
		}
	}
	hand := groupSeeds(nil)
	for _, p := range allSpellDataGroups() {
		queue = append(queue, hand[p]...)
	}
	followed := map[int32]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if followed[id] || !freshOwners.has(id) {
			continue
		}
		followed[id] = true
		if !oldIDs[id] && !moved[id] {
			moved[id] = true
			notes = append(notes, fmt.Sprintf("row %d %s taken from %s: the new build adds it and a taken row reaches it",
				id, fresh.Names[id], version))
		}
		if moved[id] {
			queue = append(queue, spellEdges(&fresh.spellTables, id)...)
		}
	}

	t := cloneTables(&old.spellTables)
	for id := range moved {
		moveRow(t, &fresh.spellTables, id)
	}
	for id := range moved {
		grant, ok := fresh.EnchantGrants[id]
		if !ok || (oldIDs[grant] && !moved[grant]) {
			continue
		}
		if description, ok := fresh.Descriptions[grant]; ok {
			t.Descriptions[grant] = description
		}
	}
	union := rootUnion(t, roots)

	// A moved row can name a spell that the committed build names, the store does not carry, and the
	// new build no longer names. Its name is dropped, so the store has no row that is only a name.
	var ids []int32
	for {
		extended, err := withExtraIDs(t, union)
		if err != nil {
			return nil, nil, err
		}
		ids = reachableSpells(t, extended)
		changed := false
		for _, id := range ids {
			if oldIDs[id] || moved[id] {
				continue
			}
			changed = true
			delete(t.Names, id)
			notes = append(notes, fmt.Sprintf("spell %d is named by a row taken from %s and is no spell there, so the store drops it", id, version))
		}
		if !changed {
			break
		}
		t.refs = nil
	}

	var nodes []traitNode
	points := map[int32]map[int32]map[int32]float64{}
	for i, in := range []*storeInputs{fresh, old} {
		for _, node := range in.TraitNodes {
			if only.Takes(node.Class) != (i == 0) {
				continue
			}
			nodes = append(nodes, node)
			if curve, ok := in.TraitPoints[node.DefinitionID]; ok {
				points[node.DefinitionID] = curve
			}
		}
	}
	slices.SortStableFunc(nodes, func(a, b traitNode) int { return cmp.Compare(a.TreeID, b.TreeID) })

	merged := captureStoreInputs(t, union, ids, nodes, points)
	merged.SpellDataGroupRoots = roots
	merged.ItemRoots = source(groupItems).ItemRoots
	merged.Forms = source(groupShared).Forms
	merged.Builds = maps.Clone(old.Builds)
	if merged.Builds == nil {
		merged.Builds = map[string]string{}
	}
	merged.Hotfixes = maps.Clone(old.Hotfixes)
	if merged.Hotfixes == nil {
		merged.Hotfixes = map[string]int32{}
	}
	for _, p := range allSpellDataGroups() {
		if only.Takes(p) {
			merged.Builds[p] = version
			merged.Hotfixes[p] = fresh.Hotfixes[p]
		}
	}
	return merged, append(notes, racialBuildWarnings(merged)...), nil
}

// Splits inputs that record no groups the way fresh splits the same ids. A talent node takes the class
// and tree of the fresh node with the same definition; failing that, of its spell's family; failing
// that, of the node before it, since a tree's nodes are captured together. A root goes to items if it is
// a gear root, otherwise to the first of: the class of the talent node granting it, the first group
// whose fresh roots hold it, the class its family names. A root none of these place goes to shared and
// is listed in the notes; only a spell the new build dropped that names no class ends up there. The
// adopted builds are unrecorded.
func adoptSpellDataGroups(old, fresh *storeInputs) (*storeInputs, []string, error) {
	adopted := *old
	adopted.owners = nil

	freshNodes := map[int32]traitNode{}
	classTree := map[string]int{}
	for _, node := range fresh.TraitNodes {
		freshNodes[node.DefinitionID] = node
		classTree[node.Class] = node.TreeID
	}
	adopted.TraitNodes = make([]traitNode, len(old.TraitNodes))
	nodeClass := map[int32]string{}
	var previous *traitNode
	for i, node := range old.TraitNodes {
		if f, ok := freshNodes[node.DefinitionID]; ok {
			node.Class, node.TreeID = f.Class, f.TreeID
		} else if class, ok := familyGroup(old.ClassOptions[node.SpellID].Family); ok {
			node.Class, node.TreeID = class, classTree[class]
		} else if previous != nil {
			node.Class, node.TreeID = previous.Class, previous.TreeID
		} else {
			return nil, nil, fmt.Errorf("talent definition %d (spell %d) names no class to adopt it by", node.DefinitionID, node.SpellID)
		}
		adopted.TraitNodes[i] = node
		previous = &adopted.TraitNodes[i]
		if _, ok := nodeClass[node.SpellID]; !ok {
			nodeClass[node.SpellID] = node.Class
		}
	}

	freshGroup := map[int32]string{}
	for _, p := range allSpellDataGroups() {
		for _, id := range fresh.SpellDataGroupRoots[p] {
			if _, ok := freshGroup[id]; !ok {
				freshGroup[id] = p
			}
		}
	}
	gear := map[int32]bool{}
	for _, id := range old.ItemRoots {
		gear[id] = true
	}
	adopted.SpellDataGroupRoots = map[string][]int32{}
	var notes []string
	for _, id := range old.Roots {
		var p string
		if gear[id] {
			p = groupItems
		} else if class, ok := nodeClass[id]; ok {
			p = class
		} else if group, ok := freshGroup[id]; ok {
			p = group
		} else if class, ok := familyGroup(old.ClassOptions[id].Family); ok {
			p = class
		} else {
			p = groupShared
			notes = append(notes, fmt.Sprintf("root %d %s names no spell data group and is adopted into %s", id, old.Names[id], p))
		}
		adopted.SpellDataGroupRoots[p] = append(adopted.SpellDataGroupRoots[p], id)
	}

	adopted.Builds = map[string]string{}
	for _, p := range allSpellDataGroups() {
		adopted.Builds[p] = ""
	}
	return &adopted, notes, nil
}

// The taken groups that the committed inputs record on a later game build than the database's. Taking
// them would move them back a build. Builds are compared by their trailing build number, which grows
// with every build; a build without one cannot be ordered and is skipped.
func groupsOnALaterBuild(committed *storeInputs, only SpellDataGroups, version string) []string {
	database, ok := buildNumber(version)
	if !ok {
		return nil
	}
	var later []string
	for _, p := range allSpellDataGroups() {
		if recorded, ok := buildNumber(committed.Builds[p]); ok && only.Takes(p) && recorded > database {
			later = append(later, fmt.Sprintf("%s (%s)", p, committed.Builds[p]))
		}
	}
	return later
}

func buildNumber(version string) (int, bool) {
	number, err := strconv.Atoi(version[strings.LastIndexByte(version, '.')+1:])
	return number, err == nil && version != ""
}

func describeBuild(build string) string {
	if build == "" {
		return "an unrecorded game build"
	}
	return build
}

// A racial row states the class flags of the class it serves, so it matches that class's spells only
// while both are on the same game build.
func racialBuildWarnings(in *storeInputs) []string {
	var warnings []string
	for _, id := range slices.Sorted(maps.Keys(racialExtraIDs())) {
		class, ok := familyGroup(in.ClassOptions[id].Family)
		if !ok || in.Builds[groupRacials] == in.Builds[class] {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s %d (racials, %s) states %s class flags, and %s is on %s",
			in.Names[id], id, describeBuild(in.Builds[groupRacials]), class, class, describeBuild(in.Builds[class])))
	}
	return warnings
}

// The field indices of spellTables' tables: its exported maps, all keyed by spell id. Walking them by
// index means a table added later is cloned, moved and reported without another edit.
var tableFields = sync.OnceValue(func() []int {
	tables := reflect.TypeFor[spellTables]()
	var fields []int
	for i := range tables.NumField() {
		if tables.Field(i).IsExported() && tables.Field(i).Type.Kind() == reflect.Map {
			fields = append(fields, i)
		}
	}
	return fields
})

func cloneTables(t *spellTables) *spellTables {
	clone := newSpellTables()
	from, into := reflect.ValueOf(t).Elem(), reflect.ValueOf(clone).Elem()
	for _, i := range tableFields() {
		rows := from.Field(i).MapRange()
		for rows.Next() {
			into.Field(i).SetMapIndex(rows.Key(), rows.Value())
		}
	}
	return clone
}

// Each of into's rows for the id becomes from's row for it; where from has none, into's row goes too.
func moveRow(into, from *spellTables, id int32) {
	to, source, key := reflect.ValueOf(into).Elem(), reflect.ValueOf(from).Elem(), reflect.ValueOf(id)
	for _, i := range tableFields() {
		to.Field(i).SetMapIndex(key, source.Field(i).MapIndex(key))
	}
}

// Whether the tables carry any row of the id besides its name.
func hasRows(t *spellTables, id int32) bool {
	tables, key := reflect.ValueOf(t).Elem(), reflect.ValueOf(id)
	for _, i := range tableFields() {
		if tables.Type().Field(i).Name != "Names" && tables.Field(i).MapIndex(key).IsValid() {
			return true
		}
	}
	return false
}

// Each group's roots with the siblings of its reach: a rank that reads its number off a same-name
// sibling belongs to the rank's group.
func withSiblings(db *sql.DB, t *spellTables, roots map[string][]int32) (map[string][]int32, error) {
	seeds := groupSeeds(roots)
	reaches := map[string][]int32{}
	reached := map[int32]bool{}
	for _, p := range allSpellDataGroups() {
		reaches[p] = reachableSpells(t, namedIDs(t, seeds[p]))
		for _, id := range reaches[p] {
			reached[id] = true
		}
	}
	siblingsOf, err := siblingSpells(db, slices.Sorted(maps.Keys(reached)))
	if err != nil {
		return nil, err
	}

	out := map[string][]int32{}
	for _, p := range allSpellDataGroups() {
		var siblings []int32
		for _, id := range reaches[p] {
			siblings = append(siblings, siblingsOf[id]...)
		}
		out[p] = namedIDs(t, roots[p], siblings)
	}
	return out, nil
}
