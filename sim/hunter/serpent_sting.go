package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var serpentStingRank = spellData.SerpentSting.Highest()

// TODO: In-game testing required. The row states no attack power share on the tick; the Forever
// notes say 10% of ranged attack power over the duration, 2% a tick, read at each tick.
const serpentStingTickAPShare = 0.02

func (hunter *Hunter) registerSerpentSting() {
	tick := serpentStingRank.PeriodicEffect()

	config := spelldata.SpellConfig(&hunter.Unit, serpentStingRank, spelldata.Flags(core.SpellFlagAPL))
	// A ranged ability, as the game files' ranged-ability procs (Rapid Recuperation) see it, so the
	// ticks crit at the ranged crit. The application deals no damage, so the on-hit listeners that
	// require damage skip it, and the ticks reach only periodic listeners.
	config.ProcMask = core.ProcMaskRangedSpecial
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.Dot = spelldata.DotConfig(serpentStingRank, tick)
	config.Dot.Aura.Tag = hunter.stingTag()
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		baseDamage := tick.Average() + serpentStingTickAPShare*dot.Spell.RangedAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, baseDamage, serpentStingRank.TickOutcome(dot))
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
