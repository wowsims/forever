package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerGrowlSpell() {
	druid.Growl = druid.registerBearUtility(spellData.Growl.Highest(), func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	})
}

// A taunt castable from an APL. applyEffects rolls its hit; the taunt itself is not modelled.
func (druid *Druid) registerBearUtility(rank *spelldata.Spell, applyEffects core.ApplySpellResults) *DruidSpell {
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ProcMask = core.ProcMaskEmpty // A landed taunt rolls no proc, Omen of Clarity included.
	config.ThreatMultiplier = 1
	config.ApplyEffects = applyEffects
	return druid.RegisterSpell(config)
}
