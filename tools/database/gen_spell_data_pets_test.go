package database

// The warlock's demon abilities sit on the demons' own skill lines with no class bit, so the class
// file reaches them through the pet scaling spell those lines carry. The Felguard's line carries it
// too, but no ability on it is trained: it is a Season of Discovery rune. The summons are on the
// class's own Demonology line, each the one rank of its subtext "Summon". Reads SkillLineAbility,
// which the store's inputs do not capture, so it skips without tools/database/wowsims.db.

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func TestWarlockDemonAbilitiesAreGenerated(t *testing.T) {
	helper := openBuffTestDB(t)

	runeOnly, err := loadRuneOnlySpells(helper.db)
	if err != nil {
		t.Fatalf("%v", err)
	}
	trees, err := selectTraitTrees(helper)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var warlock dbc.DbcClass
	for _, candidate := range dbc.Classes {
		if candidate.ProtoClass == proto.Class_ClassWarlock {
			warlock = candidate
		}
	}
	ladders, _, _, err := discoverLadders(helper.db, warlock, trees[classMaskOf(warlock)], runeOnly)
	if err != nil {
		t.Fatalf("%v", err)
	}
	generated := map[string]int{}
	for _, l := range ladders {
		generated[l.Name] = len(l.Ranks)
	}

	for name, ranks := range map[string]int{
		"Firebolt":            7,
		"Torment":             6,
		"Lash of Pain":        6,
		"Spell Lock":          2,
		"Warlock Pet Scaling": 1,
		"Summon Imp":          1,
		"Summon Voidwalker":   1,
		"Summon Succubus":     1,
		"Summon Felhunter":    1,
	} {
		if generated[name] != ranks {
			t.Errorf("%s: %d ranks generated, want %d", name, generated[name], ranks)
		}
	}
	for _, name := range []string{"Avoidance", "Anguish", "Intercept"} {
		if _, ok := generated[name]; ok {
			t.Errorf("%s: the Felguard's ability is generated", name)
		}
	}
}
