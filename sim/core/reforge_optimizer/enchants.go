package reforgeoptimizer

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Turns the request's enchant options into LP variables: one binary per (slot,
// enchant) and a per-slot row allowing at most one of them.

type enchantCandidate struct {
	slot      proto.ItemSlot
	enchant   core.Enchant
	tieBreak  float64 // The candidate's share of enchantTieBreak.
	weaponDps float64 // The enchant's weapon damage as bonus DPS on the slot's weapon.
}

// The objective bonus that breaks an EP tie toward giving a slot an enchant; the slot's equipped
// enchant gets twice as much, so a tie keeps it. A slot whose enchant stays in the baseline already
// has one, so a candidate there gets none. It sits far below any real EP difference and
// above the MIP gap the solve allows (mipRelGap, and HiGHS's default absolute gap of 1e-6); the
// reported score leaves it out.
const enchantTieBreak = 1e-4

func enchantVariableKey(slot proto.ItemSlot, effectID int32) string {
	return "Enchant_" + strconv.Itoa(int(slot)) + "_" + strconv.Itoa(int(effectID))
}

func enchantSlotKey(slot proto.ItemSlot) string {
	return "EnchantSlot_" + strconv.Itoa(int(slot))
}

// The static enchants that fit each listed, unfrozen, occupied slot, in slot then effect id order.
// The fitting row gives the stats, since an effect id can carry rows for different items. Makes the
// gear the baseline: a slot with candidates loses its enchant when that is a static row fitting the
// item. Any other enchant stays, so a slot with no pick keeps it.
func (o *reforgeOptimizer) enchantCandidates(gear *proto.EquipmentSpec) []enchantCandidate {
	options := slices.Clone(o.request.GetEnchantOptions())
	slices.SortFunc(options, func(a, b *proto.BulkEnchantSlotOptions) int { return cmp.Compare(a.GetSlot(), b.GetSlot()) })
	size := 0
	for _, option := range options {
		size += len(option.GetEffectIds())
	}
	candidates := make([]enchantCandidate, 0, size)
	for _, option := range options {
		slot := option.GetSlot()
		if o.frozenSlots[slot] || int(slot) >= len(gear.GetItems()) {
			continue
		}
		spec := gear.GetItems()[slot]
		item := core.GetItemByID(spec.GetId())
		if item == nil {
			continue
		}
		equipped := spec.GetEnchant()
		row, ok := core.FittingEnchantRow(equipped, *item)
		cleared := ok && row.IsStatic()
		stays := equipped != 0 && !cleared
		first := len(candidates)
		ids := slices.Clone(option.GetEffectIds())
		slices.Sort(ids)
		for _, id := range slices.Compact(ids) {
			row, ok := core.FittingEnchantRow(id, *item)
			if !ok || !row.IsStatic() {
				continue
			}
			tieBreak := enchantTieBreak
			switch {
			case id == equipped:
				tieBreak = 2 * enchantTieBreak
			case stays:
				tieBreak = 0
			}
			candidates = append(candidates, enchantCandidate{slot: slot, enchant: row, tieBreak: tieBreak, weaponDps: weaponDamageDps(row, item)})
		}
		if cleared && len(candidates) > first {
			spec.Enchant = 0
		}
	}
	return candidates
}

func (o *reforgeOptimizer) buildEnchantModel(weights core.UnitStats) (*lpVariables, *lpConstraints) {
	variables := newLPVariables()
	constraints := newLPConstraints()
	for _, candidate := range o.candidates {
		pseudoStats := make([]float64, stats.PseudoStatsLen)
		copy(pseudoStats, candidate.enchant.PseudoStats)
		if dpsStat, ok := weaponDpsPseudoStat[candidate.slot]; ok {
			pseudoStats[dpsStat] += candidate.weaponDps
		}

		capDelta := core.NewUnitStats()
		capDelta.Stats = candidate.enchant.Stats.Add(stats.FromPseudoStatsProto(pseudoStats))
		copy(capDelta.PseudoStats, pseudoStats)
		capCoeffs := o.resolveCapCoeffs(capDelta)
		capCoeffs[enchantSlotKey(candidate.slot)] = 1

		objCoeffs := map[string]float64{tieBreakCoeffKey: candidate.tieBreak}
		for statIdx := 0; statIdx < int(stats.ProtoStatsLen); statIdx++ {
			if value := candidate.enchant.Stats[statIdx]; value != 0 {
				o.applyReforgeStat(objCoeffs, stats.Stat(statIdx), value, weights)
			}
		}
		applyPseudoStats(objCoeffs, pseudoStats, weights, o.player.GetClass())

		key := enchantVariableKey(candidate.slot, candidate.enchant.EffectID)
		variables.set(key, capCoeffs)
		variables.setObj(key, objCoeffs)
		constraints.set(enchantSlotKey(candidate.slot), lessEq(1))
	}
	return variables, constraints
}

// Adds pseudo-stats to an objective coefficient map the way Stats.computeEP in
// ui/sim/proto/stats.ts values them:
//   - The part of a rating's melee, ranged and spell percents that all three channels share counts
//     once under the rating when the rating carries EP, otherwise under each channel.
//   - A channel's percent beyond that shared part, and any other percent with a root rating, counts
//     under its own key when that key carries EP. Otherwise it falls back to EP through the rating:
//     for a channel, what the rating is worth beyond the channels' weights; for any other percent,
//     the rating's full worth.
//   - One rating's fallbacks count once, as the largest, under the rating's key.
//   - Ranged Hit % has no fallback for a class in core.ClassesWithoutRangedHitFallback.
//   - Any other pseudo-stat counts under its own key.
func applyPseudoStats(coeffs map[string]float64, pseudoStats []float64, weights core.UnitStats, class proto.Class) {
	fallbacks := map[stats.Stat]float64{}
	addPercent := func(rating stats.Stat, pseudoStat proto.PseudoStat, value float64, fallbackPerPercent float64) {
		if value == 0 {
			return
		}
		if getUnitStat(weights, stats.UnitStatFromPseudoStat(pseudoStat)) != 0 {
			coeffs[pseudoStatCoeffKey(pseudoStat)] += value
			return
		}
		ep := value * fallbackPerPercent
		if counted, ok := fallbacks[rating]; !ok || ep > counted {
			fallbacks[rating] = ep
		}
	}

	for _, channels := range core.RatingChannelTable {
		channelStats := channels.Channels()
		values := make([]float64, len(channelStats))
		weighted := 0.0
		for i, pseudoStat := range channelStats {
			values[i] = stats.PseudoStatValue(pseudoStats, pseudoStat)
			weighted += getUnitStat(weights, stats.UnitStatFromPseudoStat(pseudoStat))
		}
		perPercent := channels.RatingPerPercent * weights.Stats[channels.Rating]
		common := max(0, slices.Min(values))
		if common != 0 && perPercent != 0 {
			coeffs[statCoeffKey(proto.Stat(channels.Rating))] += common * channels.RatingPerPercent
		} else if common != 0 {
			for _, pseudoStat := range channelStats {
				coeffs[pseudoStatCoeffKey(pseudoStat)] += common
			}
		}
		residual := max(0, perPercent-weighted)
		for i, pseudoStat := range channelStats {
			fallback := 0.0
			if core.ChannelHasEPFallback(pseudoStat, class) {
				fallback = residual
			}
			addPercent(channels.Rating, pseudoStat, values[i]-common, fallback)
		}
	}

	for idx, value := range pseudoStats {
		pseudoStat := proto.PseudoStat(idx)
		if value == 0 || isChannelPseudoStat(pseudoStat) {
			continue
		}
		if entry, ok := percentRootRating(pseudoStat); ok {
			addPercent(entry.rating, pseudoStat, value, entry.ratingPerPercent*weights.Stats[entry.rating])
		} else {
			coeffs[pseudoStatCoeffKey(pseudoStat)] += value
		}
	}

	// A fallback is EP at these weights, kept under the rating's key in rating units so it moves with
	// the rating's weight.
	for rating, ep := range fallbacks {
		if ep != 0 && weights.Stats[rating] != 0 {
			coeffs[statCoeffKey(proto.Stat(rating))] += ep / weights.Stats[rating]
		}
	}
}

func isChannelPseudoStat(pseudoStat proto.PseudoStat) bool {
	return slices.ContainsFunc(core.RatingChannelTable, func(channels core.RatingChannels) bool {
		return slices.Contains(channels.Channels(), pseudoStat)
	})
}

// The conversion from the rating a percent outside the rating channels belongs to, as
// UnitStat.getRootStat gives it: hit and crit rating for the off hand's percents, and the expertise,
// dodge, parry and block ratings for theirs.
func percentRootRating(pseudoStat proto.PseudoStat) (ratingPercent, bool) {
	for _, entry := range ratingPercents {
		if entry.root && entry.percent == pseudoStat {
			return entry, true
		}
	}
	return ratingPercent{}, false
}

var weaponDpsPseudoStat = map[proto.ItemSlot]proto.PseudoStat{
	proto.ItemSlot_ItemSlotMainHand: proto.PseudoStat_PseudoStatMainHandDps,
	proto.ItemSlot_ItemSlotOffHand:  proto.PseudoStat_PseudoStatOffHandDps,
	proto.ItemSlot_ItemSlotRanged:   proto.PseudoStat_PseudoStatRangedDps,
}

// The enchant's weapon damage as bonus DPS on the weapon it sits on. A swing deals
// weapon damage plus bonus DPS times the weapon's speed, so the two convert by the speed.
func weaponDamageDps(enchant core.Enchant, item *core.Item) float64 {
	if enchant.WeaponDamage == 0 || item.SwingSpeed == 0 {
		return 0
	}
	return enchant.WeaponDamage / item.SwingSpeed
}
