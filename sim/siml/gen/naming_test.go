package gen

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

const (
	warrior = proto.Class_ClassWarrior
	hunter  = proto.Class_ClassHunter
)

// A collection built by hand: registrations by class and namespace, and game names.
type fixture struct {
	c *Collection
}

func newFixture() *fixture {
	return &fixture{&Collection{
		Registered: map[siml.Namespace]map[Identity]Registration{},
		Classes:    map[proto.Class]map[siml.Namespace]map[Identity]bool{},
		Shared:     map[siml.Namespace]map[Identity]bool{},
		Specs:      map[proto.Class][]string{warrior: {"dps_warrior", "protection_warrior"}, hunter: {"hunter"}},
		SpellNames: map[int32]string{},
	}}
}

// Registers spell id with tag in ns for spec's class, or for the shared entries when class is ClassUnknown.
func (f *fixture) spell(class proto.Class, spec string, ns siml.Namespace, game string, id, tag, rank int32, castable bool) *fixture {
	f.c.SpellNames[id] = game
	key := Identity{id, tag}
	if f.c.Registered[ns] == nil {
		f.c.Registered[ns] = map[Identity]Registration{}
	}
	r := f.c.Registered[ns][key]
	r.Rank, r.Castable = max(r.Rank, rank), r.Castable || castable
	if !slices.Contains(r.Specs, spec) {
		r.Specs = append(r.Specs, spec)
		slices.Sort(r.Specs)
	}
	f.c.Registered[ns][key] = r
	if class == proto.Class_ClassUnknown {
		if f.c.Shared[ns] == nil {
			f.c.Shared[ns] = map[Identity]bool{}
		}
		f.c.Shared[ns][key] = true
	} else {
		set(f.c.Classes, class, ns)[key] = true
	}
	return f
}

func update(t *testing.T, old *siml.Names, c *Collection, pin *Pin) (*siml.Names, *Report) {
	t.Helper()
	if old == nil {
		old = &siml.Names{}
	}
	names, report, err := Update(old, c, pin)
	if err != nil {
		t.Fatal(err)
	}
	// The table must read back as itself, with every rule ParseNames checks.
	if _, err := siml.ParseNames(names.Encode()); err != nil {
		t.Fatalf("the updated table does not read back: %v\n%s", err, names.Encode())
	}
	return names, report
}

// The names of one namespace of a class, or of the shared entries, as "name=id/tag".
func spellNames(names *siml.Names, class proto.Class, ns siml.Namespace) []string {
	entries := names.Shared[ns]
	if class != proto.Class_ClassUnknown {
		entries = names.Classes[class][ns]
	}
	var out []string
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		e := entries[name]
		out = append(out, name+"="+itoa(e.SpellID)+"/"+itoa(e.Tag))
	}
	return out
}

func itoa(n int32) string { return strconv.Itoa(int(n)) }

func want(t *testing.T, what string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s:\n got %v\nwant %v", what, got, want)
	}
}

func TestNewSpellNames(t *testing.T) {
	f := newFixture().
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Heroic Strike", 25286, 1, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Heroic Strike", 25286, 0, 9, false).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Charge", 11578, 0, 3, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Charge", 100, 0, 1, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Overpower", 11585, 0, 4, true).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Overpower", 1282733, 0, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Battle Shout", 25289, 0, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Battle Shout", 25289, -1, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Windfury Totem", 10610, -1, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Windfury Totem", 10612, -1, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Windfury Totem", 10614, -1, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Seal", 7, -2, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Seal", 7, 3, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Seal", 8, 3, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "", 5, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "1st Aid", 6, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Wait", 9, 0, 0, true)
	names, report := update(t, nil, f.c, nil)

	want(t, "warrior spells", spellNames(names, warrior, siml.NamespaceSpell), []string{
		"charge=11578/0",        // The highest rank,
		"charge_rank1=100/0",    // and the rank tells the other apart.
		"heroic_strike=25286/1", // Castable first,
		"heroic_strike_tag0=25286/0",
		"overpower=11585/0",
		"spell_1st_aid=6/0", // A digit starts the game name,
		"spell_5=5/0",       // there is none,
		"wait_spell=9/0",    // or it is a reserved word.
	})
	want(t, "warrior auras", spellNames(names, warrior, siml.NamespaceAura), []string{
		"battle_shout=25289/0", // Tag 0 first,
		"battle_shout_external=25289/-1",
		"overpower=1282733/0", // Each namespace has its own names.
		"seal=7/-2",           // No tag 0 and no rank, so the lowest id, then the lowest tag.
		"seal_7_tag3=7/3",     // Neither tag nor rank tells it apart, and another has its id.
		"seal_8=8/3",
		"windfury_totem=10610/-1",
		"windfury_totem_10612=10612/-1", // No tag or rank tells them apart.
		"windfury_totem_10614=10614/-1",
	})
	if len(report.Added) != 17 {
		t.Errorf("the report lists %d added entries, want 17: %v", len(report.Added), report.Added)
	}
}

func TestSharedAndClassNamesNeverClash(t *testing.T) {
	f := newFixture().
		spell(proto.Class_ClassUnknown, "dps_warrior", siml.NamespaceSpell, "Jom Gabbar", 29602, 0, 0, true).
		spell(proto.Class_ClassUnknown, "hunter", siml.NamespaceSpell, "Jom Gabbar", 29602, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Berserking", 20554, 0, 0, true).
		spell(hunter, "hunter", siml.NamespaceSpell, "Berserking", 26296, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Rend", 11574, 0, 7, true).
		spell(proto.Class_ClassUnknown, "hunter", siml.NamespaceSpell, "Rend", 13318, 0, 0, false)
	names, _ := update(t, nil, f.c, nil)
	// What a class casts takes the name before what gear adds, and a shared name is free in every class.
	want(t, "warrior", spellNames(names, warrior, siml.NamespaceSpell), []string{"berserking=20554/0", "rend=11574/0"})
	want(t, "hunter", spellNames(names, hunter, siml.NamespaceSpell), []string{"berserking=26296/0"})
	want(t, "shared", spellNames(names, proto.Class_ClassUnknown, siml.NamespaceSpell), []string{"jom_gabbar=29602/0", "rend_13318=13318/0"})
	if got := names.Shared[siml.NamespaceSpell]["jom_gabbar"].RegisteredBy; !slices.Equal(got, []string{"dps_warrior", "hunter"}) {
		t.Errorf("shared jom_gabbar is registered by %v, want every spec whose builds register it", got)
	}

	// A pinned shared name keeps its meaning for every class when a class gains a spell of the same game name,
	// and a new shared name stays clear of every class's names. The report says which name the class's files mean.
	before := names
	f.spell(warrior, "dps_warrior", siml.NamespaceSpell, "Jom Gabbar", 1, 0, 0, true).
		spell(proto.Class_ClassUnknown, "hunter", siml.NamespaceSpell, "Berserking", 30000, 0, 0, true)
	names, report := update(t, before, f.c, nil)
	want(t, "warrior after", spellNames(names, warrior, siml.NamespaceSpell), []string{"berserking=20554/0", "jom_gabbar_1=1/0", "rend=11574/0"})
	want(t, "qualified", report.Qualified, []string{
		"warrior.spell.jom_gabbar_1 = spell 1: a warrior file's jom_gabbar means shared.spell.jom_gabbar (spell 29602); -repin shared.spell.jom_gabbar, run on the table without this entry, gives the name to the class",
	})
	// The repin the report names, run on the table without the class's entry, gives the class the plain name.
	repinned, report := update(t, before, f.c, &Pin{Namespace: siml.NamespaceSpell, Name: "jom_gabbar"})
	want(t, "qualified after the repin", report.Qualified, nil)
	want(t, "warrior after the repin", spellNames(repinned, warrior, siml.NamespaceSpell), []string{"berserking=20554/0", "jom_gabbar=1/0", "rend=11574/0"})
	want(t, "shared after", spellNames(names, proto.Class_ClassUnknown, siml.NamespaceSpell), []string{"berserking_30000=30000/0", "jom_gabbar=29602/0", "rend_13318=13318/0"})
	for class, spell := range map[proto.Class]int32{warrior: 20554, hunter: 26296} {
		if id, _ := names.Spell(class, siml.NamespaceSpell, "berserking"); id.GetSpellId() != spell {
			t.Errorf("%s berserking means %v, want spell %d", class, id, spell)
		}
	}
}

func TestPinnedNamesKeepTheirMeaning(t *testing.T) {
	f := newFixture().
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Eye for an Eye", 9799, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Heroic Strike", 25286, 0, 9, false)
	old, _ := update(t, nil, f.c, nil)

	// The sim stops registering 9799 and registers 25997, of the same game name, and a castable Heroic Strike that
	// the rules would now prefer.
	g := newFixture().
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Eye for an Eye", 25997, 0, 0, true).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Heroic Strike", 25286, 0, 9, false).
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Heroic Strike", 25286, 1, 0, true)
	g.c.SpellNames[9799] = "Eye for an Eye"
	names, report := update(t, old, g.c, nil)
	want(t, "warrior spells", spellNames(names, warrior, siml.NamespaceSpell), []string{
		"eye_for_an_eye=9799/0",
		"eye_for_an_eye_25997=25997/0",
		"heroic_strike=25286/0",
		"heroic_strike_tag1=25286/1",
	})
	if got := names.Classes[warrior][siml.NamespaceSpell]["eye_for_an_eye"].RegisteredBy; len(got) != 0 {
		t.Errorf("the unregistered eye_for_an_eye is registered by %v, want none", got)
	}
	if got := names.Classes[warrior][siml.NamespaceSpell]["heroic_strike"].Rank; got != 9 {
		t.Errorf("heroic_strike has rank %d, want its pinned 9", got)
	}
	want(t, "stale", report.Stale, []string{"warrior.spell.eye_for_an_eye is pinned to spell 9799; the sim registers spell 25997"})
	if report.Unregistered != 1 {
		t.Errorf("the report counts %d unregistered names, want 1", report.Unregistered)
	}

	// Repinning hands the name back to the rules, which give it to the registered id.
	pin, err := ParsePin("warrior.spell.eye_for_an_eye")
	if err != nil {
		t.Fatal(err)
	}
	repinned, report := update(t, names, g.c, &pin)
	want(t, "repinned", spellNames(repinned, warrior, siml.NamespaceSpell), []string{
		"eye_for_an_eye=25997/0",
		"heroic_strike=25286/0",
		"heroic_strike_tag1=25286/1",
	})
	var changed []string
	for _, r := range report.Repinned {
		changed = append(changed, fmt.Sprintf("%s: was %s, now %s", r.Pin, r.Was, r.Now))
	}
	want(t, "repin report", changed, []string{
		"warrior.spell.eye_for_an_eye: was spell 9799, now spell 25997",
		"warrior.spell.eye_for_an_eye_25997: was spell 25997, now nothing",
	})
	if len(report.Repinned) != 2 {
		t.Errorf("the repin hands back %v, want both names of eye_for_an_eye's game name", report.Repinned)
	}
}

func TestDeterministic(t *testing.T) {
	build := func() *Collection {
		f := newFixture()
		for i := range int32(40) {
			f.spell(warrior, "dps_warrior", siml.NamespaceAura, "Same", 1000+i, i%3-1, i%4, false)
			f.spell(proto.Class_ClassUnknown, "hunter", siml.NamespaceSpell, "Same", 2000+i, i%2, 0, i%5 == 0)
		}
		f.c.Items = []Item{{ID: 3, Name: "Ring", Level: 60}, {ID: 1, Name: "Ring", Level: 60}, {ID: 2, Name: "Ring", Level: 63}}
		return f.c
	}
	a, _ := update(t, nil, build(), nil)
	for range 5 {
		b, _ := update(t, nil, build(), nil)
		if !bytes.Equal(a.Encode(), b.Encode()) {
			t.Fatalf("two runs on the same input differ:\n%s\n%s", a.Encode(), b.Encode())
		}
	}
	again, _ := update(t, a, build(), nil)
	if !bytes.Equal(a.Encode(), again.Encode()) {
		t.Errorf("a run on its own output changed it:\n%s\n%s", a.Encode(), again.Encode())
	}
}

func TestItemEnchantAndSuffixNames(t *testing.T) {
	c := newFixture().c
	c.Items = []Item{
		{ID: 19510, Name: "Legionnaire's Band", Level: 63},
		{ID: 19511, Name: "Legionnaire's Band", Level: 53},
		{ID: 20429, Name: "Legionnaire's Band", Level: 23},
		{ID: 22630, Name: "Atiesh", Level: 90},
		{ID: 22589, Name: "Atiesh", Level: 90},
		{ID: 1, Name: "None"},
		{ID: 19510, Name: "Listed twice"}, // The first name of an id wins.
	}
	c.Enchants = []EnchantRow{
		{EffectID: 904, Name: "Enchant Bracer - Agility", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}},
		{EffectID: 904, Name: "Enchant Boots - Agility", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotFeet}},
		{EffectID: 7656, Name: "Enchant Bracer - Agility", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}},
		{EffectID: 7656, Name: "Enchant Bracer - Superior Agility", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}},
		{EffectID: 2646, Name: "Enchant 2H Weapon - Agility", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}},
	}
	c.Suffixes = []Suffix{{ID: 12727, Name: "of the Tiger"}, {ID: 12699, Name: "of the Tiger"}}
	names, _ := update(t, nil, c, nil)

	got := []string{}
	for _, name := range slices.Sorted(maps.Keys(names.Items)) {
		got = append(got, name+"="+itoa(names.Items[name]))
	}
	want(t, "items", got, []string{
		"atiesh=22589",
		"atiesh_22630=22630", // The same level, so the id.
		"legionnaires_band=19510",
		"legionnaires_band_ilvl23=20429", // The highest level is bare, and the level tells the others apart.
		"legionnaires_band_ilvl53=19511",
		"none_item=1",
	})
	got = nil
	for _, name := range slices.Sorted(maps.Keys(names.Enchants)) {
		e := names.Enchants[name]
		got = append(got, name+"="+itoa(e.EffectID))
	}
	want(t, "enchants", got, []string{
		"boots_agility=904",
		"bracer_agility=904",       // The lowest effect id is bare,
		"bracer_agility_7656=7656", // and 7656's second wrist row shares the entry of its first.
		"enchant_2h_weapon_agility=2646",
	})
	got = nil
	for _, name := range slices.Sorted(maps.Keys(names.Suffixes)) {
		got = append(got, name+"="+itoa(names.Suffixes[name]))
	}
	want(t, "suffixes", got, []string{"of_the_tiger=12699", "of_the_tiger_12727=12727"})

	// A renamed row of a named effect id and slots, and a renamed item, keep their names.
	c.Enchants[1].Name = "Enchant Boots - Nimbleness"
	c.Items[0].Name = "Band of the Legion"
	again, report := update(t, names, c, nil)
	if !bytes.Equal(names.Encode(), again.Encode()) || len(report.Added) != 0 {
		t.Errorf("renames changed the table: %v", report.Added)
	}
}

func TestQualifiedNameTaken(t *testing.T) {
	f := newFixture().spell(warrior, "dps_warrior", siml.NamespaceAura, "Rend 11574", 1, 0, 0, false)
	old, _ := update(t, nil, f.c, nil)
	f.spell(warrior, "dps_warrior", siml.NamespaceAura, "Rend", 11573, 0, 0, false).
		spell(warrior, "dps_warrior", siml.NamespaceAura, "Rend", 11574, 0, 0, false)
	names, _ := update(t, old, f.c, nil)
	want(t, "auras", spellNames(names, warrior, siml.NamespaceAura), []string{"rend=11573/0", "rend_11574=1/0", "rend_11574_2=11574/0"})
}

func TestParsePin(t *testing.T) {
	for _, c := range []struct {
		text string
		want Pin
	}{
		{"warrior.aura.overpower", Pin{warrior, siml.NamespaceAura, "overpower"}},
		{"shared.target_spell.cleave", Pin{proto.Class_ClassUnknown, siml.NamespaceTargetSpell, "cleave"}},
		{"item.goblin_sapper_charge", Pin{proto.Class_ClassUnknown, siml.NamespaceItem, "goblin_sapper_charge"}},
		{"enchant.falcons_call", Pin{proto.Class_ClassUnknown, siml.NamespaceEnchant, "falcons_call"}},
	} {
		got, err := ParsePin(c.text)
		if err != nil || got != c.want {
			t.Errorf("ParsePin(%s) = %v, %v; want %v", c.text, got, err, c.want)
		}
		if got.String() != c.text {
			t.Errorf("%v prints as %s, want %s", got, got, c.text)
		}
	}
	for _, text := range []string{"knight.spell.x", "warrior.item.x", "spell.x", "warrior.spell", "x"} {
		if _, err := ParsePin(text); err == nil {
			t.Errorf("ParsePin(%s) gave no error", text)
		}
	}
	if _, _, err := Update(&siml.Names{}, newFixture().c, &Pin{warrior, siml.NamespaceSpell, "missing"}); err == nil {
		t.Error("repinning a name the table does not hold gave no error")
	}
}

func TestFilesUsing(t *testing.T) {
	root := t.TempDir()
	table := &siml.Names{
		Classes: map[proto.Class]siml.SpellNames{
			warrior: {siml.NamespaceSpell: {
				"overpower":       {SpellID: 11585},
				"overpower_rank4": {SpellID: 11584},
				"battle_shout":    {SpellID: 25289},
			}},
			hunter: {siml.NamespaceSpell: {"arcane_shot": {SpellID: 14287}}},
		},
		Shared:   siml.SpellNames{siml.NamespaceSpell: {"goblin_sapper": {SpellID: 13241}, "battle_shout": {SpellID: 25289}}},
		Items:    map[string]int32{"arcanite_reaper": 12784},
		Enchants: map[string]siml.EnchantName{},
		Suffixes: map[string]int32{},
	}
	files := map[string]string{
		TablePath: string(table.Encode()),
		"ui/specs/warrior/dps/apls/arms.apl.siml":     "siml 1\nspec dps_warrior\n\npriority:\n  cast overpower\n  cast battle_shout\n",
		"ui/specs/warrior/dps/apls/fury.apl.siml":     "siml 1\nspec dps_warrior\n\npriority:\n  cast overpower_rank4\n  cast goblin_sapper\n",
		"ui/specs/warrior/dps/apls/notes.apl.siml":    "siml 1\nspec dps_warrior\n\nvar $overpower = 1\n",
		"ui/specs/warrior/dps/gear_sets/p1.gear.json": "overpower",
		"ui/specs/warrior/dps/gear_sets/p1.gear.siml": "siml 1\n\nmain_hand arcanite_reaper\n",
		"ui/specs/hunter/dps/defaults.siml":           "siml 1\nspec hunter\n\nparty_buffs: battle_shout\n",
		"ui/specs/hunter/dps/apls/default.apl.siml":   "siml 1\nspec hunter\n\npriority:\n  cast arcane_shot\n  cast goblin_sapper\n",
		"ui/specs/hunter/dps/apls/shout.apl.siml":     "siml 1\nspec hunter\n\npriority:\n  cast battle_shout\n",
	}
	for path, text := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		pin  Pin
		want []string
	}{
		{Pin{warrior, siml.NamespaceSpell, "overpower"}, []string{"ui/specs/warrior/dps/apls/arms.apl.siml"}},
		// The hunter defaults write battle_shout as a party buff, which is no name.
		{Pin{warrior, siml.NamespaceSpell, "battle_shout"}, []string{"ui/specs/warrior/dps/apls/arms.apl.siml"}},
		{Pin{Namespace: siml.NamespaceSpell, Name: "goblin_sapper"}, []string{
			"ui/specs/hunter/dps/apls/default.apl.siml",
			"ui/specs/warrior/dps/apls/fury.apl.siml",
		}},
		{Pin{Namespace: siml.NamespaceItem, Name: "arcanite_reaper"}, []string{"ui/specs/warrior/dps/gear_sets/p1.gear.siml"}},
		// A warrior file means the warrior's own battle_shout, which a repin of the shared one leaves alone.
		{Pin{Namespace: siml.NamespaceSpell, Name: "battle_shout"}, []string{"ui/specs/hunter/dps/apls/shout.apl.siml"}},
	} {
		got, err := FilesUsing(root, c.pin)
		if err != nil {
			t.Fatal(err)
		}
		want(t, c.pin.String(), got, c.want)
	}
}
