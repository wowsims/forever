package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var challengingShoutRank = spellData.ChallengingShout.Highest()

func (warrior *Warrior) registerChallengingShout() {
	config := spelldata.SpellConfig(&warrior.Unit, challengingShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		for _, aoeTarget := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealOutcome(sim, aoeTarget, spell.OutcomeMagicHit)
		}
	}

	warrior.ChallengingShout = warrior.RegisterSpell(config)
}
