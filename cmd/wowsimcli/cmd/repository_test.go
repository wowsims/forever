//go:build with_db

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/siml"
)

// Every file under ui/specs passes the check of sim/siml/SPEC.md §7.4.
func TestRepositoryFiles(t *testing.T) {
	files, err := simlFiles([]string{"../../../ui/specs"})
	if err != nil {
		t.Fatal(err)
	}
	if out, notes, err := runSiml(simlFmt, "-check", "../../../ui/specs"); err != nil {
		t.Errorf("%v\n%s%s", err, out, notes)
	}
	t.Logf("checked %d files under ui/specs", len(files))
}

// The check passes a hunter rotation of hunter names, and names what is wrong in the others.
func TestRepositoryCheck(t *testing.T) {
	committed, err := siml.DefaultNames()
	if err != nil {
		t.Fatal(err)
	}
	// A copy with an item the database does not hold, as a pinned name can be.
	names, err := siml.ParseNames(committed.Encode())
	if err != nil {
		t.Fatal(err)
	}
	names.Items["ghost_blade"] = 999999
	catalog, err := repositoryCatalog(".")
	if err != nil {
		t.Fatal(err)
	}
	const rotation = "siml 1\nspec hunter\n\npriority:\n  cast multi_shot if not target.aura(hunters_mark).is_active\n"
	const foreign = "siml 1\nspec hunter\n\npriority:\n  cast multi_shot if aura(cheat_death).is_active\n"
	cases := []struct {
		name, text string
		want       []string
	}{
		{"ui/specs/hunter/dps/apls/valid.apl.siml", rotation, nil},
		{"ui/specs/hunter/dps/gear_sets/valid.gear.siml", "siml 1\n\nhead       lieutenant_commanders_chain_helmet + falcons_call\nmain_hand  northern_shortsword of_the_monkey + weapon_crusader\n", nil},
		{"ui/specs/hunter/dps/apls/foreign.apl.siml", foreign, []string{
			"%s:5:27: cheat_death is an aura no hunter build registers",
		}},
		{"ui/specs/hunter/dps/gear_sets/wrong.gear.siml", "siml 1\n\nhead       lieutenant_commanders_chain_helmet + cloak_agility\nmain_hand  northern_shortsword of_agility_13019 + weapon_crusader\nranged     ghost_blade\n", []string{
			"%s:3:49: cloak_agility fits back, not head",
			"%s:4:32: Northern Shortsword rolls no of_agility_13019",
			"%s:5:12: the database has no item 999999",
		}},
		{"ui/specs/hunter/dps/gear_sets/one_hand.gear.siml", "siml 1\n\nmain_hand  thunderfury_blessed_blade_of_the_windseeker + enchant_2h_weapon_agility\n", []string{
			"%s:3:58: enchant_2h_weapon_agility does not fit Thunderfury, Blessed Blade of the Windseeker",
		}},
		{"ui/specs/hunter/dps/apls/arms.apl.siml", "siml 1\nspec dps_warrior\n\npriority:\n  cast heroic_strike\n", []string{
			"%s: spec dps_warrior is a warrior spec, and the file is not under ui/specs/warrior",
		}},
		{"ui/specs/default.apl.siml", rotation, []string{"%s: spec hunter is a hunter spec, and the file is not under ui/specs/hunter"}},
	}
	files := map[string]string{}
	for _, c := range cases {
		files[c.name] = c.text
	}
	dir := repositoryFiles(t, files)
	for _, c := range cases {
		path := filepath.Join(dir, c.name)
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		f, err := siml.Parse(path, src, names, os.ReadFile)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		want := make([]string, len(c.want))
		for i, w := range c.want {
			want[i] = fmt.Sprintf(w, path)
		}
		slices.Sort(want)
		got := repositoryProblems(siml.Analyze(path, src, names, os.ReadFile), catalog, path, f)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("%s gives the problems\n%s\nwant\n%s", c.name, strings.Join(got, "\n"), strings.Join(want, "\n"))
		}
	}

	// fmt -check runs it on the files under ui/specs, with the committed table.
	out, notes, err := runSiml(simlFmt, "-check", filepath.Join(dir, "ui/specs/hunter/dps/apls"))
	if err == nil || err.Error() != "2 of 3 .siml files are not in canonical form or fail the check" || out != "" {
		t.Errorf("-check: %v\n%s", err, out)
	}
	if want := filepath.Join(dir, "ui/specs/hunter/dps/apls/foreign.apl.siml") + ":5:27: cheat_death is an aura no hunter build registers"; !strings.Contains(notes, want) {
		t.Errorf("-check does not say %q:\n%s", want, notes)
	}
}
