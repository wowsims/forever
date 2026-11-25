package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/stats"
)

func (result *SpellResult) applyResistances(sim *Simulation, spell *Spell, isPeriodic bool, attackTable *AttackTable) {
	resistanceMultiplier, outcome := spell.ResistanceMultiplier(sim, isPeriodic, attackTable)

	result.Damage *= resistanceMultiplier
	result.Outcome |= outcome

	result.ArmorAndResistanceMultiplier = resistanceMultiplier
}

// Modifies damage based on Armor or Magic resistances, depending on the damage type.
func (spell *Spell) ResistanceMultiplier(sim *Simulation, isPeriodic bool, attackTable *AttackTable) (float64, HitOutcome) {
	if spell.Flags.Matches(SpellFlagIgnoreResists) {
		return 1, OutcomeEmpty
	}

	if spell.SpellSchool.Matches(SpellSchoolPhysical) {
		// Periodic physical damage ignores armor unless flagged to meet all of it or a tenth of it.
		if isPeriodic && !spell.Flags.Matches(SpellFlagApplyArmorReduction) {
			if spell.Flags.Matches(SpellFlagTenthArmorTicks) {
				return attackTable.armorDamageModifier(0.1), OutcomeEmpty
			}
			return 1, OutcomeEmpty
		}

		// Physical resistance (armor).
		return attackTable.GetArmorDamageModifier(spell), OutcomeEmpty
	}

	// Magical resistance.
	if spell.Flags.Matches(SpellFlagNoPartialResist) {
		return 1, OutcomeEmpty
	}

	averageResist := attackTable.Defender.averageResist(spell, attackTable.Attacker, isPeriodic)
	// A binary spell's hit roll takes its whole resist, so one that lands loses no damage to a partial
	// resist, while Spell Piercing beyond the target's resistance still rolls its bonus on the hit, as on
	// any other spell (EV-0142).
	if averageResist == 0 || averageResist > 0 && spell.Flags.Matches(SpellFlagBinary) {
		return 1, OutcomeEmpty
	}

	chances := resistTenthsChances(math.Abs(averageResist))

	// An expected-damage call takes the roll's mean and draws nothing.
	if spell.expecting {
		share := 0.0
		for tenths, chance := range chances {
			share += chance * resistedShare(tenths)
		}
		return 1 - math.Copysign(share, averageResist), OutcomeEmpty
	}

	tenths := rollResistTenths(sim, chances)
	if tenths == 0 {
		return 1, OutcomeEmpty
	}
	share := resistedShare(tenths)
	if averageResist < 0 {
		return 1 + share, partialResistOutcome(-share)
	}
	return 1 - share, partialResistOutcome(share)
}

// A partial resist never takes more than 75% of a hit: the roll's 80, 90 and 100% steps land on 75%
// (EV-0142).
func resistedShare(tenths int) float64 {
	return min(0.1*float64(tenths), 0.75)
}

func (at *AttackTable) GetBinaryHitChance(spell *Spell) float64 {
	return at.Defender.binaryHitChance(spell, at.Attacker)
}

// The average share of a spell's damage the target resists, R / (R + 8L + 19) for the caster's level L,
// with the caster's Spell Piercing taken off R unless the spell is flagged SpellFlagNoSpellPiercing.
// Piercing beyond the resistance makes the share negative, and the roll pays it out as bonus damage; a
// negative share under 1% pays nothing (EV-0142). The target's own resistance stops at 0, so a debuff
// that lowers it below 0 adds nothing. A target above the caster's level resists no more for its level.
func (unit *Unit) averageResist(spell *Spell, attacker *Unit, isPeriodic bool) float64 {
	if spell.SchoolIndex <= stats.SchoolIndexPhysical {
		return 0
	}

	resistance := 0.0
	if stat, ok := spell.SpellSchool.ResistanceStat(); ok {
		resistance = max(0, unit.GetStat(stat))
		if !spell.Flags.Matches(SpellFlagNoSpellPiercing) {
			resistance -= attacker.stats[stats.SpellPiercing] + attacker.PseudoStats.SchoolSpellPiercing[spell.SchoolIndex]
		}
	}
	// TODO: 8L + 19 is a fit to Spell Piercing logs and to mobs casting at players of known resistance
	// (EV-0091). ExpectedStat's ArmorConstant × 0.088 fits them nearly as well; a level 60 log tells the
	// two apart only with a large sample of direct hits.
	resistanceConstant := 8*float64(attacker.Level) + 19
	averageResist := resistance / (math.Abs(resistance) + resistanceConstant)

	// TODO: In-game testing required: whether a resistance share under 1% resists anything. A level 60
	// player with 5 resistance (0.99%) against one with 6 (1.19%), each taking a level 60 mob's plain
	// spell, settles it; the sim keeps the share.
	if averageResist < 0 && averageResist > -0.01 {
		averageResist = 0
	}

	// A periodic tick, a hybrid's or a channel's included, rolls its partial resist against a tenth of the
	// share, as do the hits of a spell flagged for it, and takes no Spell Piercing bonus. A tick at the
	// full share takes the bonus in full (EV-0092, EV-0142).
	// TODO: In-game testing required: whether piercing still lowers a positive resistance on a tick at a
	// tenth; the sim lets it.
	if isPeriodic && !spell.Flags.Matches(SpellFlagFullResistanceTicks) ||
		spell.Flags.Matches(SpellFlagTenthResistance) {
		averageResist = max(0, averageResist) / 10
	}

	return averageResist
}

// A binary spell is resisted in full with the target's average resist as the chance.
func (unit *Unit) binaryHitChance(spell *Spell, attacker *Unit) float64 {
	return 1 - max(0, unit.averageResist(spell, attacker, false))
}

// The chance of each step of the partial resist roll, in tenths of the damage, around the average
// share a. Up to a = 10% it gives 0, 1 or 2 tenths at 1 - 7.5a, 5a and 2.5a; from 90% the same mirrored
// from 10 tenths; in between each 10% step x within 20% of a at 0.5 - 2.5|x - a|.
func resistTenthsChances(a float64) (chances [11]float64) {
	switch {
	case a <= 0.1:
		chances[0], chances[1], chances[2] = 1-7.5*a, 5*a, 2.5*a
	case a >= 0.9:
		chances[10], chances[9], chances[8] = 1-7.5*(1-a), 5*(1-a), 2.5*(1-a)
	default:
		for tenths := range chances {
			chances[tenths] = max(0, 0.5-2.5*math.Abs(float64(tenths)/10-a))
		}
	}
	return chances
}

// One draw of the roll, taking the steps from 0 up.
func rollResistTenths(sim *Simulation, chances [11]float64) int {
	roll := sim.RandomFloat("Partial Resist")
	cumulative, last := 0.0, 0
	for tenths, chance := range chances {
		if chance <= 0 {
			continue
		}
		cumulative += chance
		last = tenths
		if roll < cumulative {
			return tenths
		}
	}
	return last
}

func (at *AttackTable) GetArmorDamageModifier(spell *Spell) float64 {
	return at.armorDamageModifier(1)
}

// The share of a hit left after armorShare of the defender's armor, taken of what the attacker's ignore
// factor and armor penetration leave.
func (at *AttackTable) armorDamageModifier(armorShare float64) float64 {
	if at.IgnoreArmor {
		return 1.0
	}

	ignoreArmorFactor := Clamp(at.ArmorIgnoreFactor, 0.0, 1.0)

	armorConstant := 400 + 85*float64(at.Attacker.Level)
	defenderArmor := at.Defender.Armor() - (at.Defender.Armor() * ignoreArmorFactor)
	// TODO: Forever possibly has different rules as there is Armor Piercing Flat and % values
	defenderArmor = max(defenderArmor-at.Attacker.stats[stats.ArmorPenetration], 0) * armorShare
	// Damage reduction from armor is capped at 75%.
	return max(1-defenderArmor/(defenderArmor+armorConstant), 0.25)
}
