package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var chargeRank = spellData.Charge.ByID(11578)

var vanguardChargeRank = chargeRank.OverriddenBy(spellData.Vanguard.Highest())

func (warrior *Warrior) registerCharge() {
	rank := core.Ternary(warrior.Talents.Vanguard, vanguardChargeRank, chargeRank)
	actionID := core.ActionID{SpellID: chargeRank.ID}
	metrics := warrior.NewRageMetrics(actionID)
	metrics.NoThreat = rank.GeneratesNoThreat() // The rage makes no threat where the row carries ATTR_EX_1_NO_THREAT.

	chargeRage := rank.EnergizeEffect().Tenths() + spellData.ImprovedCharge.TenthsAt(warrior.Talents.ImprovedCharge)

	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagMovesToTarget))
	// The APL names Charge by the spell the talent replaces.
	config.ActionID = actionID

	aura := core.MakeDashAura(&warrior.Unit, "Charge", actionID, config.Cast.CD.Duration, nil)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		aura.Duration = spell.CD.Duration
		aura.Activate(sim)
		warrior.AddRage(sim, chargeRage, metrics)
		warrior.MoveTo(spell.MinRange-3.5, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
	}

	warrior.RegisterSpell(config)
}
