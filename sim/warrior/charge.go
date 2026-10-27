package warrior

import (
	"time"

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

	aura := warrior.registerDashAura("Charge", actionID, config.Cast.CD.Duration, nil)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		aura.Duration = spell.CD.Duration
		aura.Activate(sim)
		warrior.AddRage(sim, chargeRage, metrics)
		warrior.MoveTo(spell.MinRange-3.5, sim) // movement aura is discretized in 1 yard intervals, so need to overshoot to guarantee melee range
	}

	warrior.RegisterSpell(config)
}

// TODO: Manual review needed -- the run speed and the callers' overshoot are the sim's movement model.
// The arrival callback runs only on reaching the target, never when the aura runs out or the iteration is torn down.
func (warrior *Warrior) registerDashAura(label string, actionID core.ActionID, duration time.Duration, onArrive func(sim *core.Simulation)) *core.Aura {
	aura := warrior.RegisterAura(core.Aura{
		Label:    label,
		ActionID: actionID,
		Duration: duration,
		OnGain: func(_ *core.Aura, sim *core.Simulation) {
			warrior.MultiplyMovementSpeed(sim, 3.0)
		},
		OnExpire: func(_ *core.Aura, sim *core.Simulation) {
			warrior.MultiplyMovementSpeed(sim, 1.0/3.0)
		},
	})

	warrior.RegisterMovementCallback(func(sim *core.Simulation, _ float64, kind core.MovementUpdateType) {
		if kind == core.MovementEnd && aura.IsActive() {
			aura.Deactivate(sim)
			if onArrive != nil {
				onArrive(sim)
			}
		}
	})

	return aura
}
