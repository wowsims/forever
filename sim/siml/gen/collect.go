package gen

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	googleProto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/itemfit"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/siml"
)

// Builds every spec the sim can build and reads the database. root is the repository's root, which holds the
// talent trees (ui/sim/talents/trees). old is the committed table, whose spell ids get game names besides the
// registered ones.
func Collect(root string, old *siml.Names) (*Collection, error) {
	if !core.WITH_DB {
		return nil, errors.New("the generator needs the database: run it with --tags=with_db")
	}
	sim.RegisterAll()
	db := database.Load()

	c := &Collection{
		ICDItems:   map[int32][]string{},
		Registered: map[siml.Namespace]map[Identity]Registration{},
		Classes:    map[proto.Class]map[siml.Namespace]map[Identity]bool{},
		Shared:     map[siml.Namespace]map[Identity]bool{},
		Specs:      map[proto.Class][]string{},
		SpellNames: map[int32]string{},
	}
	specs, err := c.specBuilds(root)
	if err != nil {
		return nil, err
	}
	c.addShared(db, c.addClasses(specs))
	c.addSpellNames(db, old)
	c.Items, c.Enchants, c.Suffixes = databaseEntries(db)
	return c, nil
}

// The specs the sim can build, each with its talent ranks; the others are skipped.
func (c *Collection) specBuilds(root string) ([]*specBuild, error) {
	var specs []*specBuild
	for _, info := range siml.Specs() {
		c.Specs[info.Class] = append(c.Specs[info.Class], string(info.Field.Name()))
		probe := &proto.Player{}
		probe.ProtoReflect().Set(info.Field, probe.ProtoReflect().NewField(info.Field))
		if _, err := core.PlayerProtoToSpecSafe(probe); err != nil {
			c.Skipped = append(c.Skipped, fmt.Sprintf("%s: %v", info.Field.Name(), err))
			continue
		}
		talents, err := talentRanks(root, siml.ClassName(info.Class))
		if err != nil {
			return nil, err
		}
		specs = append(specs, &specBuild{info: info, talents: talents})
	}
	for class := range c.Specs {
		slices.Sort(c.Specs[class])
	}
	return specs, nil
}

// Builds each spec with no gear, every race with every talent at each rank and every option variant, records what
// they register as their class's, and gives the specs that built, each with its source configuration.
func (c *Collection) addClasses(specs []*specBuild) []*specBuild {
	var base []*build
	for _, s := range specs {
		for i, variant := range append([]optionVariant{nil}, optionVariants(s.info)...) {
			for _, race := range core.ClassRaceCapabilities[s.info.Class] {
				for _, talents := range s.talents {
					base = append(base, &build{spec: s, player: s.player(race, talents, variant), fallback: true, variant: i})
				}
			}
		}
	}
	run(base)
	var built []*specBuild
	for _, s := range specs {
		if s.pickSource(base) {
			built = append(built, s)
		}
	}
	for _, s := range specs {
		if !slices.ContainsFunc(base, func(b *build) bool { return b.spec == s && b.attempt >= 0 }) {
			i := slices.IndexFunc(base, func(b *build) bool { return b.spec == s })
			c.Skipped = append(c.Skipped, fmt.Sprintf("%s: %s", s.info.Field.Name(), base[i].errors[0]))
		}
	}
	for _, b := range base {
		c.add(b)
		for ns, ids := range b.found {
			for id := range ids {
				set(c.Classes, b.spec.info.Class, ns)[id] = true
			}
		}
	}
	c.Skipped = append(c.Skipped, fallbacks(base)...)
	return built
}

// Builds the specs with each source of gear, consumables and preset encounters, and records what a source adds
// beyond what its spec's class registers as shared.
func (c *Collection) addShared(db *proto.UIDatabase, specs []*specBuild) {
	builds := sourceBuilds(db, specs)
	run(builds)
	for _, b := range builds {
		c.add(b)
		for ns, ids := range b.found {
			for id := range ids {
				if !c.Classes[b.spec.info.Class][ns][id] {
					if c.Shared[ns] == nil {
						c.Shared[ns] = map[Identity]bool{}
					}
					c.Shared[ns][id] = true
				}
			}
		}
	}
	c.Skipped = append(c.Skipped, failures(builds)...)
	c.Skipped = append(c.Skipped, unusable(db, specs)...)
}

// Gives each spell id the builds register, and each the table names, its game name: the database's spell icon name,
// else the spell store's, else none.
func (c *Collection) addSpellNames(db *proto.UIDatabase, old *siml.Names) {
	icons := map[int32]string{}
	for _, icon := range db.SpellIcons {
		if _, ok := icons[icon.Id]; !ok && icon.Name != "" {
			icons[icon.Id] = icon.Name
		}
	}
	name := func(id int32) {
		if _, ok := c.SpellNames[id]; ok {
			return
		}
		if n, ok := icons[id]; ok {
			c.SpellNames[id] = n
		} else if spell := spelldata.Find(id); spell != nil {
			c.SpellNames[id] = spell.Name
		} else {
			c.SpellNames[id] = ""
		}
	}
	for _, regs := range c.Registered {
		for id := range regs {
			name(id.SpellID)
		}
	}
	for _, scope := range append(slices.Collect(maps.Values(old.Classes)), old.Shared) {
		for _, entries := range scope {
			for _, e := range entries {
				name(e.SpellID)
			}
		}
	}
}

func set(m map[proto.Class]map[siml.Namespace]map[Identity]bool, class proto.Class, ns siml.Namespace) map[Identity]bool {
	if m[class] == nil {
		m[class] = map[siml.Namespace]map[Identity]bool{}
	}
	if m[class][ns] == nil {
		m[class][ns] = map[Identity]bool{}
	}
	return m[class][ns]
}

// Merges what a build registered into the registrations, with the build's spec.
func (c *Collection) add(b *build) {
	spec := string(b.spec.info.Field.Name())
	for _, id := range b.icd {
		if !slices.Contains(c.ICDItems[id], spec) {
			c.ICDItems[id] = append(c.ICDItems[id], spec)
			slices.Sort(c.ICDItems[id])
		}
	}
	for ns, ids := range b.found {
		if c.Registered[ns] == nil {
			c.Registered[ns] = map[Identity]Registration{}
		}
		for id, r := range ids {
			reg := c.Registered[ns][id]
			if r.ICDOnly && !slices.Contains(reg.ICDSpecs, spec) {
				reg.ICDSpecs = append(reg.ICDSpecs, spec)
				slices.Sort(reg.ICDSpecs)
			}
			reg.Rank = max(reg.Rank, r.Rank)
			reg.Castable = reg.Castable || r.Castable
			if !slices.Contains(reg.Specs, spec) {
				reg.Specs = append(reg.Specs, spec)
				slices.Sort(reg.Specs)
			}
			c.Registered[ns][id] = reg
		}
	}
}

type specBuild struct {
	info    siml.SpecInfo
	talents []string
	source  *proto.Player // What the source builds start from: the first race, every talent at its highest rank, every buff.
}

type optionVariant func(m protoreflect.Message)

// A player of the spec with no gear, its options at their defaults but for variant.
func (s *specBuild) player(race proto.Race, talents string, variant optionVariant) *proto.Player {
	player := &proto.Player{Class: s.info.Class, Race: race, TalentsString: talents, Equipment: &proto.EquipmentSpec{Items: make([]*proto.ItemSpec, core.NumItemSlots)}}
	for i := range player.Equipment.Items {
		player.Equipment.Items[i] = &proto.ItemSpec{}
	}
	spec := player.ProtoReflect().NewField(s.info.Field)
	fillMessages(spec.Message(), 3)
	if variant != nil {
		variant(spec.Message())
	}
	player.ProtoReflect().Set(s.info.Field, spec)
	return player
}

// The talents strings with every talent at rank 1, then 2, up to the highest any talent has, each talent capped at
// its own highest: what a spell registers can depend on the rank (Improved Berserker Rage tags its aura with it).
func talentRanks(root, class string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "ui", "sim", "talents", "trees", class+".json"))
	if err != nil {
		return nil, err
	}
	var trees []struct {
		Talents []struct {
			MaxPoints int `json:"maxPoints"`
		} `json:"talents"`
	}
	if err := json.Unmarshal(data, &trees); err != nil {
		return nil, fmt.Errorf("%s talents: %w", class, err)
	}
	highest := 0
	for _, tree := range trees {
		for _, talent := range tree.Talents {
			highest = max(highest, talent.MaxPoints)
		}
	}
	strs := make([]string, 0, highest)
	for rank := 1; rank <= highest; rank++ {
		parts := make([]string, len(trees))
		for i, tree := range trees {
			var b strings.Builder
			for _, talent := range tree.Talents {
				fmt.Fprint(&b, min(rank, talent.MaxPoints))
			}
			parts[i] = b.String()
		}
		strs = append(strs, strings.Join(parts, "-"))
	}
	return strs, nil
}

// Sets every message field down from m, so a spec's options and class options exist with their defaults.
func fillMessages(m protoreflect.Message, depth int) {
	if depth == 0 {
		return
	}
	fields := m.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() && fd.ContainingOneof() == nil {
			fillMessages(m.Mutable(fd).Message(), depth-1)
		}
	}
}

// Variants of the spec's options, each with every number of them at 1: one with nothing else changed, and one
// for each value a bool or enum field can take besides its default. Those are the options that register more, such
// as the warrior's use_battle_shout or the hunter's pet_type, which needs a pet_uptime above 0.
func optionVariants(info siml.SpecInfo) []optionVariant {
	var numbers, changes []optionVariant
	var walk func(path []protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor, depth int)
	walk = func(path []protoreflect.FieldDescriptor, md protoreflect.MessageDescriptor, depth int) {
		fields := md.Fields()
		for i := range fields.Len() {
			fd := fields.Get(i)
			if fd.IsList() || fd.IsMap() || fd.ContainingOneof() != nil {
				continue
			}
			at := append(slices.Clone(path), fd)
			switch fd.Kind() {
			case protoreflect.MessageKind:
				if depth > 1 {
					walk(at, fd.Message(), depth-1)
				}
			case protoreflect.BoolKind:
				changes = append(changes, setAt(at, protoreflect.ValueOfBool(true)))
			case protoreflect.EnumKind:
				values := fd.Enum().Values()
				for j := range values.Len() {
					if n := values.Get(j).Number(); n != fd.Default().Enum() {
						changes = append(changes, setAt(at, protoreflect.ValueOfEnum(n)))
					}
				}
			case protoreflect.Int32Kind:
				numbers = append(numbers, setAt(at, protoreflect.ValueOfInt32(1)))
			case protoreflect.DoubleKind:
				numbers = append(numbers, setAt(at, protoreflect.ValueOfFloat64(1)))
			case protoreflect.FloatKind:
				numbers = append(numbers, setAt(at, protoreflect.ValueOfFloat32(1)))
			}
		}
	}
	walk(nil, info.Field.Message(), 3)
	withNumbers := func(change optionVariant) optionVariant {
		return func(m protoreflect.Message) {
			for _, n := range numbers {
				n(m)
			}
			if change != nil {
				change(m)
			}
		}
	}
	variants := []optionVariant{withNumbers(nil)}
	for _, change := range changes {
		variants = append(variants, withNumbers(change))
	}
	return variants
}

func setAt(path []protoreflect.FieldDescriptor, v protoreflect.Value) optionVariant {
	return func(m protoreflect.Message) {
		for _, fd := range path[:len(path)-1] {
			m = m.Mutable(fd).Message()
		}
		m.Set(path[len(path)-1], v)
	}
}

// Turns on everything in a buffs message: every flag, the last value of every enum, one of every count. The auras
// external buffs and debuffs register then show.
func enableAll(m protoreflect.Message) {
	fields := m.Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if fd.IsList() || fd.IsMap() {
			continue
		}
		switch fd.Kind() {
		case protoreflect.BoolKind:
			m.Set(fd, protoreflect.ValueOfBool(true))
		case protoreflect.EnumKind:
			values := fd.Enum().Values()
			m.Set(fd, protoreflect.ValueOfEnum(values.Get(values.Len()-1).Number()))
		case protoreflect.Int32Kind:
			m.Set(fd, protoreflect.ValueOfInt32(1))
		case protoreflect.MessageKind:
			enableAll(m.Mutable(fd).Message())
		}
	}
}

// One ComputeStats run.
type build struct {
	spec      *specBuild
	player    *proto.Player
	encounter *proto.Encounter // The single default target when nil.
	source    string           // What the build adds to its spec's source build, for the report; "" for a build with no gear.
	fallback  bool             // Whether a failure with every buff is retried without the buffs, then without talents.
	variant   int              // Which option variant a build with no gear has; 0 for the defaults.

	found   map[siml.Namespace]map[Identity]Registration // What the build registered.
	icd     []int32                                      // The items whose effect the build lists only as an internal cooldown trigger.
	attempt int                                          // Which attempt succeeded: 0 with every buff, 1 without buffs, 2 without talents either; -1 when none did.
	errors  []string                                     // Why each failed attempt failed.
}

// Picks the source configuration of a spec from its builds with no gear: the one with default options, the first
// race and every talent at its highest rank, else the first that built with every buff. False when none did.
func (s *specBuild) pickSource(base []*build) bool {
	race, talents := core.ClassRaceCapabilities[s.info.Class][0], s.talents[len(s.talents)-1]
	var first *build
	for _, b := range base {
		if b.spec != s || b.attempt != 0 {
			continue
		}
		if b.variant == 0 && b.player.Race == race && b.player.TalentsString == talents {
			first = b
			break
		}
		if first == nil {
			first = b
		}
	}
	if first == nil {
		return false
	}
	s.source = first.player
	return true
}

// Runs the builds on every core.
func run(builds []*build) {
	work := make(chan *build)
	var wg sync.WaitGroup
	for range runtime.GOMAXPROCS(0) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range work {
				b.run()
			}
		}()
	}
	for _, b := range builds {
		work <- b
	}
	close(work)
	wg.Wait()
}

func (b *build) run() {
	attempts := []*proto.Player{b.player}
	if b.fallback {
		untalented := googleClone(b.player)
		untalented.TalentsString = ""
		attempts = append(attempts, b.player, untalented)
	}
	labels := []string{"with every buff", "without buffs", "without buffs or talents"}
	for i, player := range attempts {
		result, err := computeStats(player, i == 0, b.encounter)
		if err != nil {
			b.errors = append(b.errors, fmt.Sprintf("%s: %v", labels[i], err))
			continue
		}
		b.found, b.icd, b.attempt = registrations(result), icdItems(result), i
		return
	}
	b.attempt = -1
}

func computeStats(player *proto.Player, buffs bool, encounter *proto.Encounter) (result *proto.ComputeStatsResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	player = googleClone(player)
	partyBuffs, raidBuffs, debuffs := &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}
	if buffs {
		enableAll(partyBuffs.ProtoReflect())
		enableAll(raidBuffs.ProtoReflect())
		enableAll(debuffs.ProtoReflect())
		player.Buffs = &proto.IndividualBuffs{}
		enableAll(player.Buffs.ProtoReflect())
	}
	if encounter == nil {
		encounter = core.MakeSingleTargetEncounter(0)
	}
	result = core.ComputeStats(&proto.ComputeStatsRequest{
		Raid:      core.SinglePlayerRaidProto(player, partyBuffs, raidBuffs, debuffs),
		Encounter: encounter,
	})
	if result.ErrorResult != "" {
		return nil, errors.New(result.ErrorResult)
	}
	return result, nil
}

// The items whose effect the player lists only as an internal cooldown trigger, which only the ICD kinds find
// (aura(item.x).is_active does not).
func icdItems(result *proto.ComputeStatsResult) []int32 {
	metadata := result.RaidStats.Parties[0].Players[0].Metadata
	var out []int32
	for _, aura := range metadata.IcdAuras {
		id := aura.Id.GetItemId()
		if id != 0 && !slices.ContainsFunc(metadata.Auras, func(a *proto.AuraStats) bool { return a.Id.GetItemId() == id }) {
			out = append(out, id)
		}
	}
	return out
}

// What a build registered, by namespace: the spell ids of the player's spells and auras, its pets' auras, and the
// targets' auras and spells. Item and other ids have no spell-like name.
func registrations(result *proto.ComputeStatsResult) map[siml.Namespace]map[Identity]Registration {
	found := map[siml.Namespace]map[Identity]Registration{}
	add := func(ns siml.Namespace, id *proto.ActionID, castable bool) {
		if id.GetSpellId() == 0 {
			return
		}
		if found[ns] == nil {
			found[ns] = map[Identity]Registration{}
		}
		key := Identity{id.GetSpellId(), id.GetTag()}
		r := found[ns][key]
		r.Rank = max(r.Rank, id.GetRank())
		r.Castable = r.Castable || castable
		found[ns][key] = r
	}
	player := result.RaidStats.Parties[0].Players[0]
	for _, spell := range player.Metadata.Spells {
		add(siml.NamespaceSpell, spell.Id, spell.IsCastable)
	}
	for _, aura := range player.Metadata.Auras {
		add(siml.NamespaceAura, aura.Id, false)
	}
	// icd_is_ready and trigger_icd read an aura name too, so the internal cooldown triggers they find are aura names,
	// marked as triggers when the build lists them as nothing else.
	for _, aura := range player.Metadata.IcdAuras {
		key := Identity{aura.Id.GetSpellId(), aura.Id.GetTag()}
		_, listed := found[siml.NamespaceAura][key]
		add(siml.NamespaceAura, aura.Id, false)
		if r, ok := found[siml.NamespaceAura][key]; ok && !listed {
			r.ICDOnly = true
			found[siml.NamespaceAura][key] = r
		}
	}
	for _, pet := range player.Pets {
		for _, aura := range pet.Metadata.Auras {
			add(siml.NamespacePetAura, aura.Id, false)
		}
	}
	for _, target := range result.EncounterStats.Targets {
		for _, aura := range target.Metadata.Auras {
			add(siml.NamespaceTargetAura, aura.Id, false)
		}
		for _, spell := range target.Metadata.Spells {
			add(siml.NamespaceTargetSpell, spell.Id, false)
		}
	}
	return found
}

// Why builds with no gear failed with every buff, for those that built without, once for each spec and reason.
func fallbacks(builds []*build) []string {
	seen := map[string]bool{}
	var lines []string
	for _, b := range builds {
		if b.attempt <= 0 {
			continue
		}
		line := fmt.Sprintf("%s: %s", b.spec.info.Field.Name(), b.errors[0])
		if !seen[line] {
			seen[line] = true
			lines = append(lines, line)
		}
	}
	return lines
}

// The source builds that failed, once for each source with the specs it failed for and the first spec's error.
func failures(builds []*build) []string {
	specs := map[string][]string{}
	first := map[string]string{}
	var order []string
	for _, b := range builds {
		if b.attempt >= 0 {
			continue
		}
		if specs[b.source] == nil {
			order = append(order, b.source)
			first[b.source] = b.errors[len(b.errors)-1]
		}
		specs[b.source] = append(specs[b.source], string(b.spec.info.Field.Name()))
	}
	lines := make([]string, len(order))
	for i, source := range order {
		lines[i] = fmt.Sprintf("%s for %s: %s", source, strings.Join(specs[source], ", "), first[source])
	}
	return lines
}

// The items with an effect that no spec the sim builds can use, so no build shows what they register.
func unusable(db *proto.UIDatabase, specs []*specBuild) []string {
	names := map[int32]string{}
	for _, it := range db.Items {
		if _, ok := names[it.Id]; !ok && core.HasItemEffect(it.Id) {
			names[it.Id] = it.Name
		}
	}
	var lines []string
	for _, id := range slices.Sorted(maps.Keys(names)) {
		item := core.GetItemByID(id)
		if !slices.ContainsFunc(specs, func(s *specBuild) bool { return wearable(item, s.info.Class) }) {
			lines = append(lines, fmt.Sprintf("item %d (%s): no spec the sim builds can use it", id, names[id]))
		}
	}
	return lines
}

// Whether a class can wear the item: its class allowlist, and for a relic, the class's relic type. A relic of another
// class's type fails to build.
func wearable(item *core.Item, class proto.Class) bool {
	switch item.RangedWeaponType {
	case proto.RangedWeaponType_RangedWeaponTypeIdol, proto.RangedWeaponType_RangedWeaponTypeLibram,
		proto.RangedWeaponType_RangedWeaponTypeTotem, proto.RangedWeaponType_RangedWeaponTypeSigil:
		if !slices.Contains(core.ClassRangedWeaponTypeCapabilities[class], item.RangedWeaponType) {
			return false
		}
	}
	return item.UsableBy(class)
}

// What the source builds draw on: the database's items, one per id and by id, the professions they need, the
// items of each item set, the effect ids of the enchants that have an effect, and the consumables by id.
type sources struct {
	items       []*proto.UIItem
	professions map[int32]proto.Profession
	sets        map[int32][]int32
	effects     []int32
	consumables []*proto.Consumable
}

func newSources(db *proto.UIDatabase) *sources {
	src := &sources{professions: map[int32]proto.Profession{}, sets: map[int32][]int32{}}
	for _, it := range db.Items {
		if _, ok := src.professions[it.Id]; !ok {
			src.professions[it.Id] = it.RequiredProfession
			src.items = append(src.items, it)
		}
	}
	slices.SortFunc(src.items, func(a, b *proto.UIItem) int { return cmp.Compare(a.Id, b.Id) })
	for _, it := range src.items {
		if it.SetId != 0 {
			src.sets[it.SetId] = append(src.sets[it.SetId], it.Id)
		}
	}
	for _, e := range db.Enchants {
		if core.HasEnchantEffect(e.EffectId) && !slices.Contains(src.effects, e.EffectId) {
			src.effects = append(src.effects, e.EffectId)
		}
	}
	slices.Sort(src.effects)
	src.consumables = slices.Clone(db.Consumables)
	slices.SortFunc(src.consumables, func(a, b *proto.Consumable) int { return cmp.Compare(a.Id, b.Id) })
	return src
}

// The builds of each spec that built, each with one source added to its source configuration: an item with an
// effect, a whole item set, an enchant with an effect, a consumable, or a preset encounter.
func sourceBuilds(db *proto.UIDatabase, specs []*specBuild) []*build {
	src := newSources(db)
	var builds []*build
	for _, s := range specs {
		if s.source != nil {
			builds = slices.Concat(builds, s.gearBuilds(src), s.consumableBuilds(src), s.encounterBuilds())
		}
	}
	return builds
}

// The item the spec's class can wear in some slot.
func (s *specBuild) usable(id int32) (*core.Item, bool) {
	item := core.GetItemByID(id)
	return item, item != nil && wearable(item, s.info.Class) && len(core.EligibleSlotsForItem(item)) > 0
}

// The source configuration with the items on, each in its first free slot, with the profession it needs; nil when
// the class cannot wear one of them or no slot is free.
func (s *specBuild) withGear(src *sources, source string, ids ...int32) *build {
	player := googleClone(s.source)
	taken := map[proto.ItemSlot]bool{}
	for _, id := range ids {
		item, ok := s.usable(id)
		if !ok {
			return nil
		}
		slot := slices.IndexFunc(core.EligibleSlotsForItem(item), func(slot proto.ItemSlot) bool { return !taken[slot] })
		if slot < 0 {
			return nil
		}
		at := core.EligibleSlotsForItem(item)[slot]
		taken[at] = true
		player.Equipment.Items[at].Id = id
		if p := src.professions[id]; p != proto.Profession_ProfessionUnknown {
			player.Profession1 = p
		}
	}
	return &build{spec: s, player: player, source: source}
}

// A build for each item with an effect, each item set, and each enchant with an effect on the first item by id that
// it fits, that has no item effect and that the class can wear.
func (s *specBuild) gearBuilds(src *sources) []*build {
	var builds []*build
	add := func(b *build) {
		if b != nil {
			builds = append(builds, b)
		}
	}
	for _, it := range src.items {
		if core.HasItemEffect(it.Id) {
			add(s.withGear(src, fmt.Sprintf("item %d (%s)", it.Id, it.Name), it.Id))
		}
	}
	for _, id := range slices.Sorted(maps.Keys(src.sets)) {
		if len(src.sets[id]) > 1 {
			add(s.withGear(src, fmt.Sprintf("item set %d", id), src.sets[id]...))
		}
	}
	for _, effect := range src.effects {
		for _, it := range src.items {
			item, ok := s.usable(it.Id)
			if !ok || core.HasItemEffect(it.Id) || !core.EnchantAppliesToItem(effect, *item) {
				continue
			}
			if b := s.withGear(src, fmt.Sprintf("enchant %d", effect), it.Id); b != nil {
				b.player.Equipment.Items[core.EligibleSlotsForItem(item)[0]].Enchant = effect
				builds = append(builds, b)
			}
			break
		}
	}
	return builds
}

// A build for each consumable with an effect that the class may use: in extra_ids, or as an imbue on the main hand
// of the first weapon that takes it, with the race and profession it needs.
func (s *specBuild) consumableBuilds(src *sources) []*build {
	class := s.info.Class
	var builds []*build
	for _, cons := range src.consumables {
		if !core.HasConsumableEffect(cons.Id) || (len(cons.ClassAllowlist) > 0 && !slices.Contains(cons.ClassAllowlist, class)) {
			continue
		}
		player := googleClone(s.source)
		if len(cons.RaceAllowlist) > 0 && !slices.Contains(cons.RaceAllowlist, player.Race) {
			i := slices.IndexFunc(core.ClassRaceCapabilities[class], func(r proto.Race) bool { return slices.Contains(cons.RaceAllowlist, r) })
			if i < 0 {
				continue
			}
			player.Race = core.ClassRaceCapabilities[class][i]
		}
		if cons.RequiredProfession != proto.Profession_ProfessionUnknown {
			player.Profession1 = cons.RequiredProfession
			player.Prof1Rank = max(cons.ReqProfRank, 300)
		}
		player.Consumables = &proto.ConsumesSpec{}
		switch cons.Type {
		case proto.ConsumableType_ConsumableTypeImbue, proto.ConsumableType_ConsumableTypeWeaponImbue:
			weapon := slices.IndexFunc(src.items, func(it *proto.UIItem) bool {
				item, ok := s.usable(it.Id)
				return ok && !core.HasItemEffect(it.Id) && core.EligibleSlotsForItem(item)[0] == proto.ItemSlot_ItemSlotMainHand &&
					core.GetConsumableByID(cons.Id).TakesItem(proto.ItemSlot_ItemSlotMainHand, *item)
			})
			if weapon < 0 {
				continue
			}
			player.Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id = src.items[weapon].Id
			if cons.Type == proto.ConsumableType_ConsumableTypeImbue {
				player.Consumables.MhImbueId = cons.Id
			} else {
				player.Consumables.MhWeaponImbueId = cons.Id
			}
		default:
			player.Consumables.ExtraIds = []int32{cons.Id}
		}
		builds = append(builds, &build{spec: s, player: player, source: fmt.Sprintf("consumable %d (%s)", cons.Id, cons.Name)})
	}
	return builds
}

// A build for each preset encounter.
func (s *specBuild) encounterBuilds() []*build {
	var builds []*build
	for _, preset := range core.PresetEncounters {
		encounter := core.MakeSingleTargetEncounter(0)
		encounter.Targets = nil
		for _, t := range preset.Targets {
			encounter.Targets = append(encounter.Targets, googleClone(t.Target))
		}
		builds = append(builds, &build{spec: s, player: googleClone(s.source), encounter: encounter, source: "encounter " + preset.Path})
	}
	return builds
}

func googleClone[M googleProto.Message](m M) M { return googleProto.Clone(m).(M) }

// The database's items, enchant rows and random suffixes. Items come from its items, then its consumables and item
// icons, so an id the items name keeps that name and level.
func databaseEntries(db *proto.UIDatabase) ([]Item, []EnchantRow, []Suffix) {
	var items []Item
	for _, it := range db.Items {
		items = append(items, Item{ID: it.Id, Name: it.Name, Level: it.ScalingOptions[0].GetIlvl()})
	}
	for _, cons := range db.Consumables {
		items = append(items, Item{ID: cons.Id, Name: cons.Name, Level: cons.Lvl})
	}
	for _, icon := range db.ItemIcons {
		items = append(items, Item{ID: icon.Id, Name: icon.Name})
	}
	var enchants []EnchantRow
	for _, e := range db.Enchants {
		enchants = append(enchants, EnchantRow{EffectID: e.EffectId, Name: e.Name, Slots: enchantSlots(e)})
	}
	var suffixes []Suffix
	for _, s := range db.RandomSuffixes {
		suffixes = append(suffixes, Suffix{ID: s.Id, Name: s.Name})
	}
	return items, enchants, suffixes
}

// The slots of the item types an enchant row fits: those of its type and extra types, a weapon's main and off
// hand, narrowed to the main hand for a two-hander or staff enchant and to the off hand for a shield or off-hand one.
func enchantSlots(e *proto.UIEnchant) []proto.ItemSlot {
	slots := itemfit.EnchantSlots(itemfit.Enchant{Type: e.Type, ExtraTypes: e.ExtraTypes})
	switch e.EnchantType {
	case proto.EnchantType_EnchantTypeTwoHand, proto.EnchantType_EnchantTypeStaff:
		slots = slices.DeleteFunc(slots, func(s proto.ItemSlot) bool { return s == proto.ItemSlot_ItemSlotOffHand })
	case proto.EnchantType_EnchantTypeShield, proto.EnchantType_EnchantTypeOffHand:
		slots = slices.DeleteFunc(slots, func(s proto.ItemSlot) bool { return s == proto.ItemSlot_ItemSlotMainHand })
	}
	slices.Sort(slots)
	return slices.Compact(slots)
}
