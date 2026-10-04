package database

import (
	"slices"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Adds to the stats and pseudo stats db.json lists for a consumable, which the settings offer it by,
// what its on-use grants that the stat parse leaves out: the stats against creature types (Potion of
// Demon Slaying 9224's 11406: 180 attack power against Demons), the hit and crit of one channel
// (Fire-toasted Bun 23327's 29332: A_MOD_HIT_CHANCE 2 as 2% melee and ranged hit), an all-channel crit
// (Elixir of the Mongoose 13452's 17538: A_MOD_CRIT_PCT 2 as 2% melee, ranged and spell crit), and a
// Natural flask's zone bonus at the rating one percent takes, attack and casting speed as haste.
func AddUnparsedStats(db *WowDatabase) error {
	for id, consumable := range db.Consumables {
		if len(consumable.ItemEffects) == 0 {
			continue
		}
		row := spelldata.Find(consumable.ItemEffects[0].BuffId)

		var granted, percents stats.Stats
		var allCrit float64
		for i := range row.Effects {
			e := &row.Effects[i]
			if e.Type != dbcenums.E_APPLY_AURA {
				continue
			}
			if stat, ok := spelldata.VersusStat(e.Aura); ok {
				granted[stat] += e.Average()
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
			percents = percents.Add(bonus.Stats)
			granted[stats.HasteRating] += max(bonus.AttackHastePercent, bonus.CastHastePercent) * core.HasteRatingPerHastePercent
		}
		granted[stats.ExpertiseRating] += percents[stats.ExpertisePercent] * core.ExpertiseRatingPerExpertisePercent

		pseudoStats := make([]float64, stats.PseudoStatsLen)
		// One rating feeds the physical and the spell percent alike, so a consumable granting both is
		// listed at the rating of the larger of the two, once. A consumable granting one of them is listed
		// by that channel's pseudo stats: physical as melee and ranged, spell as spell.
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
	}
	return nil
}
