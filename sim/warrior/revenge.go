package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// TODO: Manual review needed -- spell 25288 states "a high amount of threat" with no number;
// none is modelled until measured in game.
var revengeRank = spellData.Revenge.Highest()

// The window a block, dodge or parry opens: the aura Defensive State (DND) fires on the hit taken.
var revengeWindow = spellData.DefensiveStateTriggered.Highest()

func (warrior *Warrior) registerRevenge() {
	actionID := core.ActionID{SpellID: revengeRank.ID}

	aura := warrior.RegisterAura(core.Aura{
		Label:    "Revenge",
		Duration: revengeWindow.Duration(),
		ActionID: actionID,
	})

	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.DefensiveState.Highest(),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		})
	trigger.Outcome = core.OutcomeBlock | core.OutcomeDodge | core.OutcomeParry
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true
	warrior.MakeProcTriggerAura(trigger)

	config := spelldata.SpellConfig(&warrior.Unit, revengeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return aura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := revengeRank.DamageEffect().Roll(sim) + revengeRank.DamageEffect().APCoeff()*spell.MeleeAttackPower(target)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		aura.Deactivate(sim)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	config.RelatedSelfBuff = aura

	warrior.RegisterSpell(config)
}
