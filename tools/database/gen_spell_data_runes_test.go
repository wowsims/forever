package database

// The class files carry what a Forever character can learn, so a family only a Season of Discovery
// rune grants is left out and a rune ability a trainer or a talent also teaches stays. Reads
// SkillLineAbility and SpellItemEnchantment, so it skips without tools/database/wowsims.db.

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestRuneOnlySpells(t *testing.T) {
	db := openBuffTestDB(t).db

	runeOnly, err := loadRuneOnlySpells(db)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, c := range []struct {
		name     string
		spellID  int32
		runeOnly bool
	}{
		{"Shadow Cleave", 403835, true},
		{"Metamorphosis : Demon Charge", 412783, true},
		{"Victory Rush", 402927, false},
		{"Incinerate", 412758, false},
		{"Primal Bite", 407995, false},
	} {
		if runeOnly[c.spellID] != c.runeOnly {
			t.Errorf("%s %d: rune only %v, want %v", c.name, c.spellID, runeOnly[c.spellID], c.runeOnly)
		}
	}
}

func TestRuneOnlyFamiliesAreNotGenerated(t *testing.T) {
	helper := openBuffTestDB(t)

	runeOnly, err := loadRuneOnlySpells(helper.db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	trees, err := selectTraitTrees(helper)
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, c := range []struct {
		class   proto.Class
		present []string
		absent  []string
	}{
		{proto.Class_ClassWarlock, []string{"Incinerate"}, []string{"Shadow Cleave", "Metamorphosis : Demon Charge"}},
		{proto.Class_ClassWarrior, []string{"Victory Rush"}, nil},
		{proto.Class_ClassDruid, []string{"Primal Bite"}, nil},
	} {
		var class dbc.DbcClass
		for _, candidate := range dbc.Classes {
			if candidate.ProtoClass == c.class {
				class = candidate
			}
		}

		ladders, _, _, err := discoverLadders(helper.db, class, trees[classMaskOf(class)], runeOnly)
		if err != nil {
			t.Fatalf("%s: %v", c.class, err)
		}
		generated := map[string]bool{}
		for _, l := range ladders {
			generated[l.Name] = true
		}

		for _, name := range c.present {
			if !generated[name] {
				t.Errorf("%s: %s is not generated", c.class, name)
			}
		}
		for _, name := range c.absent {
			if generated[name] {
				t.Errorf("%s: %s is generated", c.class, name)
			}
		}
	}
}
