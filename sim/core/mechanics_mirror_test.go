package core

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
)

// ui/sim/constants/mechanics.ts and base_stats_auto_gen.go are both written by
// tools/base_stats_parser.py from the same numbers, so they cannot disagree unless
// someone hand-edits one. This catches that, and catches a constant being added to
// one side only. The fix is `make basestats`, not an edit to the generated file.
func TestMechanicsConstantsMatchTheUI(t *testing.T) {
	const mechanicsPath = "../../ui/sim/constants/mechanics.ts"

	source, err := os.ReadFile(mechanicsPath)
	if err != nil {
		t.Fatalf("reading %s: %v", mechanicsPath, err)
	}

	goValues := map[string]float64{
		"CHARACTER_LEVEL":                                CharacterLevel,
		"BOSS_LEVEL":                                     CharacterLevel + 3,
		"EXPERTISE_RATING_PER_EXPERTISE_PERCENT":         ExpertiseRatingPerExpertisePercent,
		"HASTE_RATING_PER_HASTE_PERCENT":                 HasteRatingPerHastePercent,
		"CRIT_RATING_PER_CRIT_PERCENT":                   CritRatingPerCritPercent,
		"HIT_RATING_PER_HIT_PERCENT":                     HitRatingPerHitPercent,
		"DODGE_RATING_PER_DODGE_PERCENT":                 DodgeRatingPerDodgePercent,
		"PARRY_RATING_PER_PARRY_PERCENT":                 ParryRatingPerParryPercent,
		"DEFENSE_RATING_PER_DEFENSE_LEVEL":               DefenseRatingPerDefenseLevel,
		"BLOCK_RATING_PER_BLOCK_PERCENT":                 BlockRatingPerBlockPercent,
		"MISS_DODGE_PARRY_BLOCK_CRIT_CHANCE_PER_DEFENSE": MissDodgeParryBlockCritChancePerDefense,
	}

	// BOSS_LEVEL is an expression rather than a literal, so it is checked separately
	// below against the default raid boss level in target.go.
	declaration := regexp.MustCompile(`(?m)^export const ([A-Z_0-9]+) = ([0-9.]+);`)
	seen := map[string]bool{}
	for _, match := range declaration.FindAllStringSubmatch(string(source), -1) {
		name, literal := match[1], match[2]
		want, tracked := goValues[name]
		if !tracked {
			t.Errorf("%s declares %s, which no Go constant is pinned to -- add it to this test", mechanicsPath, name)
			continue
		}
		seen[name] = true
		got, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			t.Errorf("%s: %s = %q is not a number", mechanicsPath, name, literal)
			continue
		}
		if got != want {
			t.Errorf("%s: %s = %s, but Go says %s -- run `make basestats` rather than editing it", mechanicsPath, name, literal, strconv.FormatFloat(want, 'f', -1, 64))
		}
	}

	for name := range goValues {
		if name == "BOSS_LEVEL" {
			continue
		}
		if !seen[name] {
			t.Errorf("%s no longer declares %s", mechanicsPath, name)
		}
	}

	if bossLevel := regexp.MustCompile(`BOSS_LEVEL = CHARACTER_LEVEL \+ (\d+);`).FindStringSubmatch(string(source)); bossLevel == nil {
		t.Errorf("%s no longer derives BOSS_LEVEL from CHARACTER_LEVEL", mechanicsPath)
	} else if want := fmt.Sprint(int(goValues["BOSS_LEVEL"] - goValues["CHARACTER_LEVEL"])); bossLevel[1] != want {
		t.Errorf("%s: BOSS_LEVEL is CHARACTER_LEVEL + %s, but Go says + %s", mechanicsPath, bossLevel[1], want)
	}
}
