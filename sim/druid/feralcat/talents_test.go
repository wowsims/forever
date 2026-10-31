//go:build with_db

package feralcat

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

type talentLocation struct {
	RowIdx int `json:"rowIdx"`
	ColIdx int `json:"colIdx"`
}

// Talent ranks and prerequisites from the calculator's tree data.
type talentTree struct {
	Name    string `json:"name"`
	Talents []struct {
		FieldName      string          `json:"fieldName"`
		Location       talentLocation  `json:"location"`
		PrereqLocation *talentLocation `json:"prereqLocation"`
		MaxPoints      int             `json:"maxPoints"`
	} `json:"talents"`
}

func druidTrees(t *testing.T) []talentTree {
	t.Helper()
	raw, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []talentTree
	if err := json.Unmarshal(raw, &trees); err != nil {
		t.Fatal(err)
	}
	return trees
}

// Level 60 builds must respect rank limits, tiers, prerequisites, and the 51-point budget.
func talentsLegal(trees []talentTree, talents string) error {
	parts := strings.Split(talents, "-")
	if len(parts) != len(trees) {
		return fmt.Errorf("%d trees, want %d", len(parts), len(trees))
	}
	total := 0
	for i, tree := range trees {
		digits := parts[i]
		if len(digits) > len(tree.Talents) {
			return fmt.Errorf("%s: %d talents, the tree has %d", tree.Name, len(digits), len(tree.Talents))
		}
		perRow := map[int]int{}
		for j := range digits {
			if digits[j] < '0' || digits[j] > '9' {
				return fmt.Errorf("%s: invalid talent digit %q", tree.Name, digits[j])
			}
			points := int(digits[j] - '0')
			talent := tree.Talents[j]
			if points > talent.MaxPoints {
				return fmt.Errorf("%s: %s %d of %d", tree.Name, talent.FieldName, points, talent.MaxPoints)
			}
			perRow[talent.Location.RowIdx] += points
			total += points
		}
		for j := range digits {
			talent := tree.Talents[j]
			if digits[j] == '0' {
				continue
			}
			if talent.PrereqLocation != nil {
				found := false
				for k, parent := range tree.Talents {
					if parent.Location != *talent.PrereqLocation {
						continue
					}
					found = true
					points := 0
					if k < len(digits) {
						points = int(digits[k] - '0')
					}
					if points < parent.MaxPoints {
						return fmt.Errorf("%s: %s requires %s %d/%d, has %d", tree.Name, talent.FieldName, parent.FieldName, parent.MaxPoints, parent.MaxPoints, points)
					}
				}
				if !found {
					return fmt.Errorf("%s: %s has an unknown prerequisite", tree.Name, talent.FieldName)
				}
			}
			above := 0
			for row := 0; row < talent.Location.RowIdx; row++ {
				above += perRow[row]
			}
			if need := 5 * talent.Location.RowIdx; above < need {
				return fmt.Errorf("%s: %s is on row %d with %d points above, needs %d", tree.Name, talent.FieldName, talent.Location.RowIdx+1, above, need)
			}
		}
	}
	if total > 51 {
		return fmt.Errorf("%d points, a level 60 druid has 51", total)
	}
	return nil
}

// Every talent string the cat and bear UIs ship can be built in game.
func TestPresetTalentsAreLegal(t *testing.T) {
	trees := druidTrees(t)
	pattern := regexp.MustCompile(`talentsString: '([0-9-]+)'`)
	for _, file := range []string{
		"../../../ui/specs/druid/feralcat/presets.ts",
		"../../../ui/specs/druid/feralbear/presets.ts",
		"../../../ui/specs/druid/feralbear/spec.ts",
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
			if err := talentsLegal(trees, match[1]); err != nil {
				t.Errorf("%s: %s: %v", file, match[1], err)
			}
		}
	}
	for name, talents := range map[string]string{"catTalents": catTalents} {
		if err := talentsLegal(trees, talents); err != nil {
			t.Errorf("%s %s: %v", name, talents, err)
		}
	}
}

func TestTalentsLegalPrerequisites(t *testing.T) {
	trees := druidTrees(t)
	for _, talents := range []string{
		"050022-35200032122010125011-055",
		"050022-35200032121010125011-055",
	} {
		t.Run(talents, func(t *testing.T) {
			err := talentsLegal(trees, talents)
			if err == nil || !strings.Contains(err.Error(), "rendAndTear requires predatoryStrikes 3/3") {
				t.Fatalf("expected unmet Predatory Strikes prerequisite, got %v", err)
			}
		})
	}
}
