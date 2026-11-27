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
	cannotCrit := hurricaneTickRank.CannotCrit()

	tickConfig := spelldata.SpellConfig(&druid.Unit, hurricaneTickRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	if hurricaneTickRank.IsAProc() {
		tickConfig.Flags |= core.SpellFlagProc
	}
	// A tick takes no Spell Piercing bonus, though most binary hits do (EV-0142).
	tickConfig.Flags |= core.SpellFlagNoPartialResist
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		outcome := spell.OutcomeMagicHitAndCrit
		if cannotCrit {
			outcome = spell.OutcomeMagicHit
		}
		spell.CalcAndDealAoeDamage(sim, hurricaneTickDamage.Average(), outcome)
	}
	tickSpell := druid.RegisterSpell(tickConfig)

	// The channel ticks on an A_PERIODIC_DUMMY, which DotConfig does not read, so the dot is by hand.
	tickLength := hurricaneRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

	// The cast lands on every enemy in the area as the channel starts. Omen of Clarity rolls at the cast
	// itself, in registerOmenOfClarity (EV-0115, EV-0219).
	land := spelldata.RegisterAreaCastLanding(&druid.Unit, hurricaneRank)

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
		land(sim)
		spell.AOEDot().Apply(sim)
	}

	druid.Hurricane = druid.RegisterSpell(config)
	druid.Hurricane.RelatedDotSpell = tickSpell.Spell
}
