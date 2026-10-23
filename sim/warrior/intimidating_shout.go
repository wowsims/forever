package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var intimidatingShoutRank = spellData.IntimidatingShout.Highest()

func (warrior *Warrior) registerIntimidatingShout() {
	config := spelldata.SpellConfig(&warrior.Unit, intimidatingShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	}

	warrior.IntimidatingShout = warrior.RegisterSpell(config)
}
