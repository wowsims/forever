package database

import (
	"strconv"
	"strings"
)

// The talents string with every talent at its maximum, in the form the UI parses
// (ui/sim/talents/talents_string.ts, serializeTalentsString): one digit per talent in tree
// order, trees joined by "-", trailing zeros and dashes dropped.
func MaxedTalentsString(tabs []TalentTabConfig) string {
	trees := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		var digits strings.Builder
		for _, talent := range tab.Talents {
			digits.WriteString(strconv.Itoa(talent.MaxPoints))
		}
		trees = append(trees, strings.TrimRight(digits.String(), "0"))
	}
	return strings.TrimRight(strings.Join(trees, "-"), "-")
}
