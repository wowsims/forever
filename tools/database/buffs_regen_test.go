package database

// Re-renders the buff files from the game files rows committed in
// assets/db_inputs/spell_store_inputs.json and asserts the committed ones agree, plus the resolver
// invariants the rendered code cannot show. No game files database: the buffs render from the same
// capture the store does, so this runs in CI.

import (
	"fmt"
	"maps"
	"testing"

	"github.com/wowsims/forever/tools/database/buffmanifest"
)

func resolveCommittedBuffs(t *testing.T) []ResolvedBuff {
	t.Helper()
	inRepositoryRoot(t)

	rows, err := resolveBuffManifest(committedInputs(t))
	if err != nil {
		t.Fatalf("resolving the manifest: %v", err)
	}
	return rows
}

func TestBuffFilesRegenerateFromTheCommittedInputs(t *testing.T) {
	inRepositoryRoot(t)

	files, err := renderBuffOutputs(committedInputs(t))
	if err != nil {
		t.Fatalf("rendering the buff files: %v", err)
	}
	for name, rendered := range files {
		assertRendersCommitted(t, name, rendered)
	}
}

func TestResolvedBuffInvariants(t *testing.T) {
	rows := resolveCommittedBuffs(t)
	if len(rows) != len(buffmanifest.All()) {
		t.Fatalf("resolved %d rows for a manifest of %d", len(rows), len(buffmanifest.All()))
	}

	for _, row := range rows {
		if row.Kind == buffmanifest.KindFlag {
			if row.SpellID != 0 {
				t.Errorf("%s: %s resolved to spell %d", row.Field, row.Kind, row.SpellID)
			}
			continue
		}
		if row.Proto == buffmanifest.ProtoTristate && row.TalentRanks == 0 && row.ImpAction == nil {
			t.Errorf("%s: declared ProtoTristate but neither a talent nor an ImpAction prices it", row.Field)
		}
		if want, pinned := pinnedTalentRanks[row.Field]; pinned && row.TalentRanks != want {
			t.Errorf("%s: the talent takes %d points, want %d", row.Field, row.TalentRanks, want)
		}
		if want, pinned := pinnedCategories[row.Field]; pinned && row.Category != want {
			t.Errorf("%s: the aura competes under %q, want %q", row.Field, row.Category, want)
		}
		if want, pinned := pinnedAmounts[row.Field]; pinned {
			got := map[string]float64{}
			for _, applied := range row.Applied {
				got[applied.Kind] = applied.Value
			}
			if !maps.Equal(got, want) {
				t.Errorf("%s: grants %v, want %v", row.Field, got, want)
			}
		}
	}
}

// Every raid and individual row names the group it reaches, so each is checked two ways: that the
// game files agree with the scope, and that they state a group at all. A party or debuff row the
// game files state nothing for is exempt, because its scope is then the sim's grouping rather than
// a game fact: the buff Windfury's proc lands on its wielder, a debuff the game files reach by the
// area around its caster.
func TestScopeMatchesTheClientTargeting(t *testing.T) {
	for _, row := range resolveCommittedBuffs(t) {
		if row.SpellID == 0 {
			continue
		}
		complaint := scopeComplaint(row)
		if reason, exempt := scopeExemptions[row.Field]; exempt {
			if complaint == "" {
				t.Errorf("%s: spell %d now states %s, so the exemption %q is stale",
					row.Field, row.SpellID, row.Scope, reason)
			}
			continue
		}
		if complaint != "" {
			t.Errorf("%s: %s", row.Field, complaint)
		}
	}
}

// The rows whose spell the game files target somewhere else than the scope reaches,
// each with the reason the scope is still what the sim wants.
var scopeExemptions = map[string]string{}

// What the game files say against the scope the manifest files the row under, and
// "" when the two agree or the row is one the game files may state nothing for.
func scopeComplaint(row ResolvedBuff) string {
	scope, stated := clientScope(row)
	switch {
	case stated && scope == row.Scope:
		return ""
	case stated:
		return fmt.Sprintf("the manifest says %s, spell %d states %s", row.Scope, row.SpellID, scope)
	case row.Scope == buffmanifest.ScopeRaid || row.Scope == buffmanifest.ScopeIndividual:
		return fmt.Sprintf("the manifest says %s, spell %d states no group it reaches", row.Scope, row.SpellID)
	}
	return ""
}

// What the parse attaches for a row, by the kind it attaches as. The two auras
// the game files state as a bare A_MOD_CRIT_PCT carry no school: what they are
// worth is the game files' 3, and it lands on every kind of crit.
//
// Expose Armor states nothing on the effect itself: its -450 armor is per combo
// point, and the raid config's debuff is the five-point finisher.
//
// The paladin auras state a healing-taken row of 0 beside the aura, which the
// parse leaves out.
var pinnedAmounts = map[string]map[string]float64{
	"leader_of_the_pack": {"stat PhysicalCritPercent": 3, "stat SpellCritPercent": 3},
	"moonkin_aura":       {"stat PhysicalCritPercent": 3, "stat SpellCritPercent": 3},
	"expose_armor":       {"stat Armor": -2250},
	"devotion_aura":      {"stat Armor": 735},
	"concentration_aura": {"pushback": -0.35},
	"retribution_aura":   {},
}

// The talents that price an improved state: Restorative Totems on Mana Spring, five points,
// Guardian Totems on Stoneskin Totem, two points, and Vindication's talent on its debuff, three.
var pinnedTalentRanks = map[string]int32{
	"mana_spring_totem": 5,
	"stoneskin_totem":   2,
	"vindication":       3,
}

// What a pinned row competes under, as (stats, own aura). A resistance source
// that has no exclusivity beyond the school itself keeps no category of its own,
// which is how every totem, Aspect of the Wild and Shadow Protection read; a
// paladin aura also holds its own slot. Armor is not a school, so Devotion Aura
// has only the slot. Furious Howl holds a category of its own, and the attack
// power reductions bid in Demoralizing Shout's.
var pinnedCategories = map[string]string{
	"frost_resistance_totem":      "",
	"aspect_of_the_wild":          "",
	"prayer_of_shadow_protection": "",
	"frost_resistance_aura":       "FrostResistanceAura",
	"shadow_resistance_aura":      "ShadowResistanceAura",
	"devotion_aura":               "DevotionAura",
	"furious_howl":                "BattleShout",
	"hex_of_weakness":             "Demoralizing",
	"demoralizing_screech":        "Demoralizing",
	"vindication":                 "Demoralizing",
}
