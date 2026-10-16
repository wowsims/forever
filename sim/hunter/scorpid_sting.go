package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var scorpidStingRank = spellData.ScorpidSting.Highest()

func (hunter *Hunter) registerScorpidSting() {
	auras := hunter.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// TODO: verify in game that Scorpid Sting and Insect Swarm share one slot.
		return buffs.NewSpellDataScorpidStingAuraFor(target, &hunter.Unit, hunter.stingTag())
	})

	config := spelldata.SpellConfig(&hunter.Unit, scorpidStingRank, spelldata.Flags(core.SpellFlagAPL))
	// No proc mask: a debuff with no damage, seen only by listeners that state no mask.
	config.ProcMask = core.ProcMaskEmpty

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)
		spell.DealCastResult(sim, result)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if result.Landed() {
				aura := auras.Get(target)
				dropOtherSting(sim, aura)
				aura.Activate(sim)
			}
			spell.DealOutcome(sim, result)
		})
	}

	config.RelatedAuraArrays = auras.ToMap()

	hunter.RegisterSpell(config)
}
