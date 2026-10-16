package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var hurricaneRank = spellData.Hurricane.Highest()

// The damage each second is its own row, which the channel casts on every tick.
var hurricaneTickRank = spellData.HurricaneTriggered.Highest()
var hurricaneTickDamage = hurricaneTickRank.DamageEffect()

func (druid *Druid) registerHurricaneSpell() {
	tickConfig := spelldata.SpellConfig(&druid.Unit, hurricaneTickRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, hurricaneTickDamage.Average(), spell.OutcomeMagicHit)
	}
	tickSpell := druid.RegisterSpell(tickConfig)

	// The channel ticks on an A_PERIODIC_DUMMY, which DotConfig does not read, so the dot is by hand.
	tickLength := hurricaneRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

	config := spelldata.SpellConfig(&druid.Unit, hurricaneRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.Dot = core.DotConfig{
		IsAOE:               true,
		Aura:                spelldata.AuraConfig(hurricaneRank),
		NumberOfTicks:       int32(hurricaneRank.Duration() / tickLength),
		TickLength:          tickLength,
		AffectedByCastSpeed: true,
		OnTick: func(sim *core.Simulation, target *core.Unit, _ *core.Dot) {
			tickSpell.Cast(sim, target)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}

	druid.Hurricane = druid.RegisterSpell(config)
	druid.Hurricane.RelatedDotSpell = tickSpell.Spell
}
