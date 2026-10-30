//go:build with_db

package gen

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/siml"
)

const root = "../../.."

var committed = sync.OnceValues(func() ([]byte, error) { return os.ReadFile(filepath.Join(root, TablePath)) })

// One collection for every test here: building every spec takes a few seconds.
var collected = sync.OnceValues(func() (*Collection, error) {
	data, err := committed()
	if err != nil {
		return nil, err
	}
	old, err := siml.ParseNames(data)
	if err != nil {
		return nil, err
	}
	return Collect(root, old)
})

func committedTable(t *testing.T) (*siml.Names, []byte, *Collection) {
	t.Helper()
	data, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	old, err := siml.ParseNames(data)
	if err != nil {
		t.Fatalf("%s: %v", TablePath, err)
	}
	c, err := collected()
	if err != nil {
		t.Fatal(err)
	}
	return old, data, c
}

// The lines of b that a does not hold, each with prefix.
func missingLines(a, b []byte, prefix string) []string {
	have := map[string]bool{}
	for _, line := range strings.Split(string(a), "\n") {
		have[strings.TrimSuffix(strings.TrimSpace(line), ",")] = true
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		if key := strings.TrimSuffix(strings.TrimSpace(line), ","); !have[key] {
			out = append(out, prefix+key)
		}
	}
	return out
}

// The committed file is in the generator's form, and a regeneration gives every entry of it back as it is: the
// name, its meaning and its spec lists, and the items listed as internal cooldown triggers. What the sim and the
// database give beyond the table is only reported, since a database update or a class change adds names without
// changing what is committed until make siml-names regenerates the table.
func TestTableKeepsItsEntries(t *testing.T) {
	old, data, c := committedTable(t)
	if got := old.Encode(); !bytes.Equal(got, data) {
		t.Errorf("%s is not in the generator's form; run %s and commit the table:\n  %s", TablePath, MakeCommand,
			strings.Join(append(missingLines(data, got, "+ "), missingLines(got, data, "- ")...), "\n  "))
	}
	names, report, err := Update(old, c, nil)
	if err != nil {
		t.Fatal(err)
	}
	var changed []string
	spellLike := func(scope string, old, now siml.SpellNames) {
		for _, ns := range slices.Sorted(maps.Keys(old)) {
			for _, name := range slices.Sorted(maps.Keys(old[ns])) {
				if e, g := old[ns][name], now[ns][name]; e.SpellID != g.SpellID || e.Tag != g.Tag || e.Rank != g.Rank || !slices.Equal(e.RegisteredBy, g.RegisteredBy) ||
					!slices.Equal(e.ICDOnlyBy, g.ICDOnlyBy) || e.Uncastable != g.Uncastable {
					changed = append(changed, fmt.Sprintf("%s.%s.%s is %+v, want %+v", scope, ns, name, g, e))
				}
			}
		}
	}
	for _, class := range slices.Sorted(maps.Keys(old.Classes)) {
		spellLike(scopeName(class), old.Classes[class], names.Classes[class])
	}
	spellLike("shared", old.Shared, names.Shared)
	for _, name := range slices.Sorted(maps.Keys(old.Items)) {
		if e, g := old.Items[name], names.Items[name]; e != g {
			changed = append(changed, fmt.Sprintf("item.%s is item %d, want item %d", name, g, e))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(old.Enchants)) {
		if e, g := old.Enchants[name], names.Enchants[name]; !reflect.DeepEqual(e, g) {
			changed = append(changed, fmt.Sprintf("enchant.%s is %s, want %s", name, enchantText(g), enchantText(e)))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(old.Suffixes)) {
		if e, g := old.Suffixes[name], names.Suffixes[name]; e != g {
			changed = append(changed, fmt.Sprintf("suffix.%s is suffix %d, want suffix %d", name, g, e))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(old.ICDOnlyItems)) {
		if e, g := old.ICDOnlyItems[name], names.ICDOnlyItems[name]; !slices.Equal(e, g) {
			changed = append(changed, fmt.Sprintf("icdOnlyItems.%s is %v, want %v", name, g, e))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(names.ICDOnlyItems)) {
		if _, listed := old.ICDOnlyItems[name]; !listed && old.Items[name] != 0 {
			changed = append(changed, fmt.Sprintf("icdOnlyItems.%s is %v, want none", name, names.ICDOnlyItems[name]))
		}
	}
	if len(changed) > 0 {
		t.Errorf("%s holds entries the sim and the database do not give now; run %s and commit the table:\n  %s",
			TablePath, MakeCommand, strings.Join(changed[:min(len(changed), 60)], "\n  "))
	}
	if len(report.Added) > 0 {
		t.Logf("%s lacks %d names the sim and the database give; %s adds them:\n  %s",
			TablePath, len(report.Added), MakeCommand, strings.Join(report.Added[:min(len(report.Added), 60)], "\n  "))
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	old, _, _ := committedTable(t)
	var runs [][]byte
	for range 2 {
		c, err := Collect(root, old)
		if err != nil {
			t.Fatal(err)
		}
		// From an empty table every name is new, so the whole run is compared, not only what the table lacks.
		names, _, err := Update(&siml.Names{}, c, nil)
		if err != nil {
			t.Fatal(err)
		}
		runs = append(runs, names.Encode())
	}
	if !bytes.Equal(runs[0], runs[1]) {
		t.Errorf("two runs differ:\n%s", strings.Join(append(missingLines(runs[0], runs[1], "+ "), missingLines(runs[1], runs[0], "- ")...), "\n"))
	}
}

// A pinned name keeps its meaning when the sim stops registering it, and the spell it would name now takes another
// name.
func TestPinnedNamesSurviveTheSim(t *testing.T) {
	old, _, c := committedTable(t)
	cleave := old.Classes[proto.Class_ClassWarrior][siml.NamespaceSpell]["cleave"]
	if cleave.Tag == 0 || !slices.Contains(cleave.RegisteredBy, "dps_warrior") {
		t.Fatalf("warrior.spell.cleave is %+v, want a tagged spell dps_warrior registers", cleave)
	}
	const pinnedTag = 7
	pinned := &siml.Names{Classes: maps.Clone(old.Classes), Shared: old.Shared, Items: old.Items, Enchants: old.Enchants, Suffixes: old.Suffixes}
	pinned.Classes[proto.Class_ClassWarrior] = maps.Clone(old.Classes[proto.Class_ClassWarrior])
	spells := maps.Clone(old.Classes[proto.Class_ClassWarrior][siml.NamespaceSpell])
	spells["cleave"] = siml.SpellName{SpellID: cleave.SpellID, Tag: pinnedTag}
	pinned.Classes[proto.Class_ClassWarrior][siml.NamespaceSpell] = spells

	names, report, err := Update(pinned, c, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := names.Classes[proto.Class_ClassWarrior][siml.NamespaceSpell]
	if e := got["cleave"]; e.SpellID != cleave.SpellID || e.Tag != pinnedTag || len(e.RegisteredBy) != 0 {
		t.Errorf("cleave is %+v, want the pinned spell %d tag %d with no registered-by", e, cleave.SpellID, pinnedTag)
	}
	tagged := fmt.Sprintf("cleave_tag%d", cleave.Tag)
	if e := got[tagged]; e.SpellID != cleave.SpellID || e.Tag != cleave.Tag || !slices.Contains(e.RegisteredBy, "dps_warrior") {
		t.Errorf("%s is %+v, want the registered spell %d tag %d", tagged, e, cleave.SpellID, cleave.Tag)
	}
	stale := fmt.Sprintf("warrior.spell.cleave is pinned to spell %[1]d tag %[2]d; the sim registers spell %[1]d, spell %[1]d tag %[3]d", cleave.SpellID, pinnedTag, cleave.Tag)
	if !slices.Contains(report.Stale, stale) {
		t.Errorf("the report does not name the stale cleave: %v", report.Stale)
	}
	for name, e := range spells {
		if g, ok := got[name]; !ok || g.SpellID != e.SpellID || g.Tag != e.Tag || g.Rank != e.Rank {
			t.Errorf("warrior.spell.%s is %+v, want its pinned %+v", name, g, e)
		}
	}
	added := slices.DeleteFunc(slices.Clone(report.Added), func(line string) bool { return !strings.HasPrefix(line, "warrior.spell.") })
	if want := []string{fmt.Sprintf("warrior.spell.%s = spell %d tag %d", tagged, cleave.SpellID, cleave.Tag)}; !reflect.DeepEqual(added, want) {
		t.Errorf("the run added %v to warrior.spell, want %s alone", added, tagged)
	}
}
