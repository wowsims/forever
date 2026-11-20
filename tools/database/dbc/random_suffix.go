package dbc

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// A random suffix is one bonus list of an item's bonus tree; its ID is the bonus list's.
type RandomSuffix struct {
	ID               int
	Name             string
	Stats            []RandomSuffixStat
	EquipSpells      []int // The spells of the list's ITEM_BONUS_ITEM_EFFECT rows.
	ReducedCasterDps bool  // The list carries an ITEM_BONUS_REDUCED_CASTER_DPS row: a weapon with it deals a caster weapon's share of its damage table.
}

type RandomSuffixStat struct {
	ItemMod    int
	Allocation int // Per 10000 of the item's random property points.
}

func (raw RandomSuffix) ToProto() *proto.ItemRandomSuffix {
	suffix := &proto.ItemRandomSuffix{
		Name:             raw.Name,
		Id:               int32(raw.ID),
		Stats:            stats.Stats{}.ToProtoArray(),
		ReducedCasterDps: raw.ReducedCasterDps,
	}

	for _, stat := range raw.Stats {
		mapped, match := MapBonusStatIndexToStats(stat.ItemMod)
		if !match {
			continue
		}
		for _, s := range mapped {
			suffix.Stats[s] += float64(stat.Allocation)
		}
	}

	// The one equip spell suffixes carry, 13669 on "of Eluding", states a dodge percent and nothing else.
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	for _, spellID := range raw.EquipSpells {
		if addedStats, _ := AddEquipSpellStats(&stats.Stats{}, pseudoStats, spellID); addedStats {
			panic(fmt.Sprintf("random suffix %d: equip spell %d states flat stats", raw.ID, spellID))
		}
	}
	suffix.PseudoStats = NullFloat(pseudoStats)
	return suffix
}
