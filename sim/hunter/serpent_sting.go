package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 8 is the last rank the trainer teaches; rank 9 comes from the Guide: Serpent Sting IX book.
var serpentStingTrainerRank = spellData.SerpentSting.ByID(13555)
var serpentStingBookRank = spellData.SerpentSting.ByID(25295)

// The row states no attack power share on the tick. In game each tick adds 0.030 of ranged attack
// power, read at each tick, and spell power adds nothing.
const serpentStingTickAPShare = 0.030

func (hunter *Hunter) registerSerpentSting() {
	hunter.registerSerpentStingRank(serpentStingTrainerRank)
	hunter.registerSerpentStingRank(serpentStingBookRank)
}

func (hunter *Hunter) registerSerpentStingRank(rank *spelldata.Spell) {
	tick := rank.PeriodicEffect()

	config := spelldata.SpellConfig(&hunter.Unit, rank, spelldata.Flags(core.SpellFlagAPL))
	// A ranged ability, as the game files' ranged-ability procs (Rapid Recuperation) see it, so the
	// ticks crit at the ranged crit. The application deals no damage, so the on-hit listeners that
	// require damage skip it, and the ticks reach only periodic listeners.
	config.ProcMask = core.ProcMaskRangedSpecial
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Dot = spelldata.DotConfig(rank, tick)
	config.Dot.Aura.Tag = hunter.stingTag()
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		baseDamage := tick.Average() + serpentStingTickAPShare*dot.Spell.RangedAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, baseDamage, rank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)
		spell.DealCastResult(sim, result)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if result.Landed() {
				dot := spell.Dot(target)
				dropOtherSting(sim, dot.Aura)
				dot.Apply(sim)
			}
			spell.DealOutcome(sim, result)
		})
	}

	hunter.RegisterSpell(config)
}
