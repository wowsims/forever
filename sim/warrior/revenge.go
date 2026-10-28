package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 5 is the top rank the trainer teaches; Rank 6 comes from Manual of Revenge VI.
var revengeTrainerRank = spellData.Revenge.ByID(11601)
var revengeBookRank = spellData.Revenge.ByID(25288)

// The "high amount of threat" the game files state with no number: Revenge's threat is 2.25 times
// its damage, measured in game on Rank 2.
// TODO: Ingame test needed -- confirm the multiplier on the top rank.
const revengeThreatMultiplier = 2.25

// The window a block, dodge or parry opens: the aura Defensive State (DND) fires on the hit taken.
// The rotation reads it by that aura's id, whichever rank it casts.
var revengeWindow = spellData.DefensiveStateTriggered.Highest()

func (warrior *Warrior) registerRevenge() {
	aura := warrior.RegisterAura(core.Aura{
		Label:    "Revenge",
		Duration: revengeWindow.Duration(),
		ActionID: core.ActionID{SpellID: revengeWindow.ID},
	})

	trigger := spelldata.ProcTrigger(&warrior.Character, spellData.DefensiveState.Highest(),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			aura.Activate(sim)
		})
	trigger.Outcome = core.OutcomeBlock | core.OutcomeDodge | core.OutcomeParry
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true
	warrior.MakeProcTriggerAura(trigger)

	warrior.registerRevengeRank(revengeTrainerRank, aura)
	warrior.registerRevengeRank(revengeBookRank, aura)
}

func (warrior *Warrior) registerRevengeRank(rank *spelldata.Spell, aura *core.Aura) {
	config := spelldata.SpellConfig(&warrior.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ThreatMultiplier = revengeThreatMultiplier

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return aura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := rank.DamageEffect().Roll(sim) + rank.DamageEffect().APCoeff()*spell.MeleeAttackPower(target)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		aura.Deactivate(sim)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	config.RelatedSelfBuff = aura

	warrior.RegisterSpell(config)
}
