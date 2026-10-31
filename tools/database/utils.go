package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

func parseIntArrayField(jsonStr string, expectedLen int) ([]int, error) {
	var arr []int
	if jsonStr == "" {
		return arr, nil
	}
	if err := json.Unmarshal([]byte(jsonStr), &arr); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON array: %w", err)
	}
	if len(arr) != expectedLen {
		return nil, fmt.Errorf("invalid array length: expected %d, got %d", expectedLen, len(arr))
	}
	return arr, nil
}

func parseFloatArrayField(jsonStr string, expectedLen int) ([]float64, error) {
	var arr []float64
	if err := json.Unmarshal([]byte(jsonStr), &arr); err != nil {
		return nil, fmt.Errorf("unmarshaling JSON array: %w", err)
	}
	if len(arr) != expectedLen {
		return nil, fmt.Errorf("invalid array length: expected %d, got %d", expectedLen, len(arr))
	}
	return arr, nil
}

func ParseRandomSuffixOptions(optionsString sql.NullString) ([]int32, error) {
	if !optionsString.Valid || optionsString.String == "" {
		return []int32{}, nil
	}

	parts := strings.Split(optionsString.String, ",")
	var opts []int32
	var parseErrors []string

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		num, err := strconv.Atoi(part)
		if err != nil {
			parseErrors = append(parseErrors, fmt.Sprintf("part %d (%s): %v", i, part, err))
			continue
		}
		opts = append(opts, int32(num))
	}

	if len(parseErrors) > 0 {
		return opts, fmt.Errorf("some values couldn't be parsed: %s", strings.Join(parseErrors, "; "))
	}

	return opts, nil
}

// Formats the input string so that it does not use more than maxLength characters
// as soon a whole word exceeds the character limit a new line will be created
func formatStrings(maxLength int, input []string) []string {
	result := []string{}
	for _, line := range input {
		words := strings.Split(strings.Trim(line, "\n\r "), " ")
		currentLine := ""
		for _, word := range words {
			if len(currentLine) > maxLength {
				result = append(result, currentLine)
				currentLine = ""
			}

			if len(currentLine) > 0 {
				currentLine += " "
			}

			currentLine += word
		}

		if len(result) > 0 || len(currentLine) > 0 {
			result = append(result, currentLine)
		}
	}

	return result
}

func parseDungeonDifficulty(isRaid bool) proto.DungeonDifficulty {
	if isRaid {
		return proto.DungeonDifficulty_DifficultyRaid
	}
	return proto.DungeonDifficulty_DifficultyDungeon
}

func SpellHasDummyEffect(spellId int, instance *dbc.DBC) bool {
	return anySpellEffect(spellId, instance, isDummyAura)
}

func SpellHasTriggerEffect(spellId int, instance *dbc.DBC) bool {
	return anySpellEffect(spellId, instance, isProcTriggerAura)
}

func SpellUsesStacks(spellId int, instance *dbc.DBC) bool {
	if instance.Spells[spellId].MaxCumulativeStacks > 1 {
		return true
	}

	return anySpellEffect(spellId, instance, func(effect dbc.SpellEffect) bool {
		return isProcTriggerAura(effect) &&
			instance.Spells[effect.EffectTriggerSpell].MaxCumulativeStacks > 1
	})
}

// Reports whether any of the spell's effects satisfies pred. Effect order does not matter to a
// predicate, so this reads the map directly rather than paying for a sorted copy.
func anySpellEffect(spellId int, instance *dbc.DBC, pred func(dbc.SpellEffect) bool) bool {
	for _, effect := range instance.SpellEffects[spellId] {
		if pred(effect) {
			return true
		}
	}

	return false
}

func isDummyAura(effect dbc.SpellEffect) bool {
	return effect.IsDummy()
}

func isProcTriggerAura(effect dbc.SpellEffect) bool {
	return effect.IsProcTrigger()
}

func GetEffectStatString(itemEffect *proto.ItemEffect) string {
	if itemEffect == nil {
		return ""
	}

	stats := dbc.EffectStats(itemEffect)
	statsString := make([]string, 0, len(stats))
	for k := range stats {
		stat := proto.Stat(k)
		statsString = append(statsString, stat.String()[4:])
	}

	slices.Sort(statsString)

	return strings.Join(statsString, " / ")
}
