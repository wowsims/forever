package gen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

func readCommitted(t *testing.T) *siml.Names {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", TablePath))
	if err != nil {
		t.Fatal(err)
	}
	names, err := siml.ParseNames(data)
	if err != nil {
		t.Fatal(err)
	}
	return names
}

// Two names that swap their ids pass the up-to-date test, since it regenerates from the table it is given; the
// check against the base's table catches them.
func TestChangedCatchesSwappedNames(t *testing.T) {
	base := readCommitted(t)
	now := readCommitted(t)
	spells := now.Classes[hunter][siml.NamespaceSpell]
	arcane, multi := spells["arcane_shot"], spells["multi_shot"]
	spells["arcane_shot"], spells["multi_shot"] = multi, arcane
	// What no name's meaning holds: a registered-by list, and a new entry.
	e := spells["aimed_shot"]
	e.RegisteredBy = nil
	spells["aimed_shot"] = e
	spells["arcane_shot_new"] = siml.SpellName{SpellID: 1}
	for name := range now.Items {
		delete(now.Items, name)
		break
	}

	got := Changed(base, now, nil)
	if len(got) != 3 {
		t.Fatalf("got %d changed names, want the two swapped spells and the dropped item:\n%s", len(got), strings.Join(got, "\n"))
	}
	want(t, "the swapped spells", got[:2], []string{
		"hunter.spell.arcane_shot: was " + entryText(arcane) + ", now " + entryText(multi),
		"hunter.spell.multi_shot: was " + entryText(multi) + ", now " + entryText(arcane),
	})
	if !strings.HasPrefix(got[2], "item.") || !strings.HasSuffix(got[2], ", now nothing") {
		t.Errorf("the dropped item's line is %q, want item.<name>: was item <id>, now nothing", got[2])
	}

	repinned := RecordedNames([]byte(recordHeader+"hunter.spell.arcane_shot: was a, now b\n"), []byte(recordHeader))
	if got := Changed(base, now, repinned); len(got) != 2 || !strings.HasPrefix(got[0], "hunter.spell.multi_shot:") {
		t.Errorf("with arcane_shot's line added to the record, got %v, want the item and multi_shot only", got)
	}
	if got := Changed(base, readCommitted(t), nil); len(got) != 0 {
		t.Errorf("the committed table changed against itself: %v", got)
	}
}

func TestEnchantSlotsAndRanksAreMeanings(t *testing.T) {
	base := &siml.Names{
		Classes:  map[proto.Class]siml.SpellNames{warrior: {siml.NamespaceAura: {"rend": {SpellID: 11574, Rank: 7}}}},
		Enchants: map[string]siml.EnchantName{"crusader": {EffectID: 1900, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}}},
	}
	now := &siml.Names{
		Classes:  map[proto.Class]siml.SpellNames{warrior: {siml.NamespaceAura: {"rend": {SpellID: 11574, Rank: 6}}}},
		Enchants: map[string]siml.EnchantName{"crusader": {EffectID: 1900, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}}},
	}
	want(t, "changed", Changed(base, now, nil), []string{
		"enchant.crusader: was enchant 1900 (MainHand), now enchant 1900 (OffHand)",
		"warrior.aura.rend: was spell 11574 rank 7, now spell 11574 rank 6",
	})
}

func TestRecord(t *testing.T) {
	p := Pin{warrior, siml.NamespaceSpell, "overpower"}
	repins := []Repin{
		{p, "spell 11585", "spell 11584"},
		{Pin{warrior, siml.NamespaceSpell, "overpower_rank4"}, "spell 11584", "spell 11584"}, // Its meaning held.
	}
	record := AppendRecord(nil, repins)
	want(t, "the record", strings.Split(string(record), "\n"), append(strings.Split(recordHeader, "\n")[:3],
		"warrior.spell.overpower: was spell 11585, now spell 11584", ""))
	if again := AppendRecord(record, repins); string(again) != string(record) {
		t.Errorf("a line the record holds was added again:\n%s", again)
	}
	committed, err := os.ReadFile("repinned.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(committed), recordHeader) {
		t.Errorf("%s does not open with the header the generator writes", RecordPath)
	}

	later := AppendRecord(record, []Repin{{Pin{Namespace: siml.NamespaceItem, Name: "atiesh"}, "item 22589", "item 22630"}})
	got := RecordedNames(later, record)
	if len(got) != 1 || !got["item.atiesh"] {
		t.Errorf("the names the record adds are %v, want item.atiesh alone", got)
	}
	if got := RecordedNames(later, nil); len(got) != 2 || !got["warrior.spell.overpower"] {
		t.Errorf("the names of the whole record are %v, want overpower and atiesh", got)
	}
}

func TestPruneDropsWorkInProgress(t *testing.T) {
	f := newFixture().
		spell(warrior, "dps_warrior", siml.NamespaceSpell, "Overpower", 11585, 0, 0, true).
		spell(proto.Class_ClassUnknown, "hunter", siml.NamespaceAura, "Haste", 13533, 0, 0, false)
	f.c.Items = []Item{{ID: 19019, Name: "Thunderfury"}}
	f.c.Enchants = []EnchantRow{{EffectID: 1900, Name: "Crusader", Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}}}
	f.c.Suffixes = []Suffix{{ID: 12699, Name: "of the Tiger"}}
	head := &siml.Names{
		Classes: map[proto.Class]siml.SpellNames{warrior: {siml.NamespaceSpell: {"heroic_strike": {SpellID: 25286}}}},
		Items:   map[string]int32{"old_item": 1},
	}
	old := &siml.Names{
		Classes: map[proto.Class]siml.SpellNames{warrior: {siml.NamespaceSpell: {
			"heroic_strike": {SpellID: 25286}, // HEAD's, unregistered: pinned.
			"overpower":     {SpellID: 11585}, // New and registered: kept.
			"typo":          {SpellID: 11858}, // New and unregistered: dropped.
			"other_class":   {SpellID: 13533}, // Registered by a hunter build only: dropped from the warrior.
		}}},
		Shared: siml.SpellNames{siml.NamespaceAura: {
			"haste":   {SpellID: 13533},
			"retired": {SpellID: 2},
		}},
		Items:    map[string]int32{"old_item": 1, "thunderfury": 19019, "gone": 3},
		Enchants: map[string]siml.EnchantName{"crusader": {EffectID: 1900, Slots: []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}}, "gone": {EffectID: 4}},
		Suffixes: map[string]int32{"of_the_tiger": 12699, "gone": 5},
	}
	dropped := prune(old, head, f.c)
	want(t, "dropped", dropped, []string{
		"enchant.gone = enchant 4 ()",
		"item.gone = item 3",
		"shared.aura.retired = spell 2",
		"suffix.gone = suffix 5",
		"warrior.spell.other_class = spell 13533",
		"warrior.spell.typo = spell 11858",
	})
	want(t, "warrior spells", spellNames(old, warrior, siml.NamespaceSpell), []string{"heroic_strike=25286/0", "overpower=11585/0"})
	want(t, "shared auras", spellNames(old, proto.Class_ClassUnknown, siml.NamespaceAura), []string{"haste=13533/0"})
	if _, ok := old.Items["old_item"]; !ok || len(old.Items) != 2 || len(old.Enchants) != 1 || len(old.Suffixes) != 1 {
		t.Errorf("items %v, enchants %v, suffixes %v: want HEAD's and the database's kept", old.Items, old.Enchants, old.Suffixes)
	}
	if again := prune(old, head, f.c); len(again) != 0 {
		t.Errorf("a second prune dropped %v", again)
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestHeadTable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	plain := t.TempDir()
	if IsCheckout(plain) {
		t.Fatal("a directory outside any checkout reads as one")
	}
	if head, err := headTable(plain); head != nil || err != nil {
		t.Errorf("outside a checkout, HEAD's table is %v, %v; want none", head, err)
	}

	root := t.TempDir()
	git(t, root, "init", "-q")
	if head, err := headTable(root); head != nil || err != nil {
		t.Errorf("before the first commit, HEAD's table is %v, %v; want none", head, err)
	}
	table := &siml.Names{Items: map[string]int32{"thunderfury": 19019}}
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(TablePath)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, TablePath), table.Encode(), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, root, "add", ".")
	git(t, root, "commit", "-q", "-m", "table")
	if err := os.WriteFile(filepath.Join(root, TablePath), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	head, err := headTable(root)
	if err != nil || head == nil || head.Items["thunderfury"] != 19019 {
		t.Fatalf("HEAD's table is %v, %v; want the committed one, whatever the working tree holds", head, err)
	}
	if _, ok, err := GitFile(root, "HEAD", RecordPath); ok || err != nil {
		t.Errorf("a file HEAD does not hold reads as %v, %v; want not ok", ok, err)
	}
	if _, _, err := GitFile(root, "no-such-branch", TablePath); err == nil {
		t.Error("a commit that does not exist gave no error")
	}
	if IsCheckout(filepath.Join(root, "sim")) {
		t.Error("a directory inside a checkout reads as its top")
	}
}
