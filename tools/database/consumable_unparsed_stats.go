package database

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The zones the Natural flasks' tooltip names for their bonus; the Barrow Deeps has no AreaTable row.
var naturalFlaskZones = []int32{616, 3606}

// Adds to the stats, pseudo stats, mob type stats and zone stats db.json lists for a consumable,
// which the settings offer and order it by, what its on-use grants that the stat parse leaves out:
// the stats against creature types as mob type stats under the mob types the aura's mask names
// (Potion of Demon Slaying 9224's 11406: attack power against Demons), or as plain stats where the
// mask names none, the hit and crit of one channel (Fire-toasted Bun 23327's 29332:
// A_MOD_HIT_CHANCE 2 as 2% melee and ranged hit), an all-channel crit (Elixir of the Mongoose
// 13452's 17538: A_MOD_CRIT_PCT 2 as 2% melee, ranged and spell crit), and a Natural flask's zone
// bonus as zone stats under naturalFlaskZones, at the rating one percent takes, attack and casting
// speed as haste.
func AddUnparsedStats(db *WowDatabase) error {
	for id, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		if err := addUnparsedStats(id, consumable, spelldata.Find(consumable.ItemEffects[0].BuffId)); err != nil {
			return err
		}
	}
	return nil
}

// One consumable's share of AddUnparsedStats, read off its on-use buff's row.
func addUnparsedStats(id int32, consumable *proto.Consumable, row *spelldata.Spell) error {
	var versus, percents stats.Stats
	var allCrit float64
	for i := range row.Effects {
		e := &row.Effects[i]
		if e.Type != dbcenums.E_APPLY_AURA {
			continue
		}
		if granted := spelldata.VersusStats(e.Aura); len(granted) > 0 {
			if mobTypes := spelldata.MobTypes(e.Misc); len(mobTypes) > 0 {
				bonus := map[int32]float64{}
				for _, stat := range granted {
					bonus[int32(stat)] = e.Average()
				}
				consumable.MobTypeStats = append(consumable.MobTypeStats, &proto.MobTypeStats{MobTypes: mobTypes, Stats: bonus})
			} else {
				for _, stat := range granted {
					versus[stat] += e.Average()
				}
			}
		}
		switch e.Aura {
		case dbcenums.A_MOD_HIT_CHANCE:
			percents[stats.PhysicalHitPercent] += e.Average()
		case dbcenums.A_MOD_SPELL_HIT_CHANCE:
			percents[stats.SpellHitPercent] += e.Average()
		case dbcenums.A_MOD_WEAPON_CRIT_PERCENT:
			percents[stats.PhysicalCritPercent] += e.Average()
		case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
			percents[stats.SpellCritPercent] += e.Average()
		case dbcenums.A_MOD_CRIT_PCT:
			allCrit += e.Average()
		}
	}
	if slices.Contains(forever.NaturalFlasks, id) {
		bonus, err := forever.NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average())
		if err != nil {
			return err
		}
		zoneGranted, zonePseudoStats := ratingsOf(bonus.Stats)
		zoneGranted[stats.HasteRating] += max(bonus.AttackHastePercent, bonus.CastHastePercent) * core.HasteRatingPerHastePercent
		if slices.ContainsFunc(zonePseudoStats, func(v float64) bool { return v != 0 }) {
			return fmt.Errorf("%d %s: the zone bonus reaches one channel alone, which zone stats cannot list", id, consumable.Name)
		}
		consumable.ZoneStats = append(consumable.ZoneStats, &proto.ZoneStats{ZoneIds: naturalFlaskZones, Stats: zoneGranted.ToProtoMap()})
	}
	granted, pseudoStats := ratingsOf(percents)
	granted = granted.Add(versus)
	if allCrit != 0 {
		pseudoStats[proto.PseudoStat_PseudoStatMeleeCritPercent] += allCrit
		pseudoStats[proto.PseudoStat_PseudoStatRangedCritPercent] += allCrit
		pseudoStats[proto.PseudoStat_PseudoStatSpellCritPercent] += allCrit
	}

	if granted != (stats.Stats{}) {
		consumable.Stats = stats.FromProtoArray(consumable.Stats).Add(granted).ToProtoArray()
	}
	if slices.ContainsFunc(pseudoStats, func(v float64) bool { return v != 0 }) {
		consumable.PseudoStats = pseudoStats
	}
	return nil
}

// The ratings and pseudo stats a consumable is listed by for these percents. One rating feeds the
// physical and the spell percent alike, so a consumable granting both is listed at the rating of the
// larger of the two, once. A consumable granting one of them is listed by that channel's pseudo
// stats: physical as melee and ranged, spell as spell. Expertise is listed as its rating.
func ratingsOf(percents stats.Stats) (stats.Stats, []float64) {
	var granted stats.Stats
	granted[stats.ExpertiseRating] = percents[stats.ExpertisePercent] * core.ExpertiseRatingPerExpertisePercent
	pseudoStats := make([]float64, stats.PseudoStatsLen)
	for _, r := range core.RatingChannelTable {
		if !r.HasPercentStats {
			continue
		}
		physical, spell := percents[r.PhysicalPercent], percents[r.SpellPercent]
		switch {
		case physical > 0 && spell > 0:
			granted[r.Rating] += max(physical, spell) * r.RatingPerPercent
		case physical > 0:
			pseudoStats[r.Melee] += physical
			pseudoStats[r.Ranged] += physical
		case spell > 0:
			pseudoStats[r.Spell] += spell
		}
	}
	return granted, pseudoStats
}
