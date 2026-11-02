package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var riposteRank = spellData.Riposte.Highest()

// The window a parry opens. The game files do not state which Defensive State aura Riposte reads, so it is
// the one the rogue's own passive (16092) fires.
var riposteWindow = spellData.DefensiveStateTriggered.Highest()

func (rogue *Rogue) registerRiposte() {
	if !rogue.Talents.Riposte {
		return
	}

	rogue.RiposteAura = rogue.RegisterAura(spelldata.AuraConfig(riposteWindow, spelldata.Label("Riposte Aura")))

	trigger := spelldata.ProcTrigger(&rogue.Character, spellData.DefensiveState.Highest(),
		func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			rogue.RiposteAura.Activate(sim)
		})
	trigger.Name = "Riposte - Trigger"
	trigger.Outcome = core.OutcomeParry
	trigger.RequireDamageDealt = false
	trigger.TriggerImmediately = true
	rogue.MakeProcTriggerAura(trigger)

	config := spelldata.SpellConfig(&rogue.Unit, riposteRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return rogue.RiposteAura.IsActive()
	}

	// The disarm (A_MOD_DISARM) is not modelled: the target's swings are not cancelled.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		baseDamage := riposteRank.WeaponStrike(sim, spell, spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		rogue.RiposteAura.Deactivate(sim)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	rogue.Riposte = rogue.RegisterSpell(config)
}
