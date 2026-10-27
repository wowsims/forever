package druid

import "github.com/wowsims/forever/sim/core"

func (druid *Druid) registerChallengingRoarSpell() {
	druid.ChallengingRoar = druid.registerBearUtility(spellData.ChallengingRoar.Highest(), func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
		}
	})
}
