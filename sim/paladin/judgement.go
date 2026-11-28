package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Judgement
// https://www.wowhead.com/forever/spell=20271
//
// Unleash the energy of a Seal spell upon an enemy. Does not consume the Seal. Refer to individual
// Seals for Judgement effect.
func (paladin *Paladin) registerJudgement() {
	paladin.Judgement = paladin.RegisterSpell(paladin.judgementConfig(spellData.Judgement.Highest(), SpellMaskJudgement))

	// Debuff refreshes in EV-0117. Holy Strike refreshes the judgement debuffs only through Sacred
	// Arbiter's own trigger.
	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Judgement Refresh" + paladin.Label,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           swingMask,
		Outcome:            core.OutcomeLanded,
		ExtraCondition:     hearsSwing,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			paladin.refreshJudgements(sim, result.Target)
		},
	})
}

// A cast that unleashes the active seal's judgement, the one Judgement and Swift Judgement share. A
// cast listener hears this cast and the judgement it unleashes. Cast rolls in EV-0144.
func (paladin *Paladin) judgementConfig(rank *spelldata.Spell, classMask int64) core.SpellConfig {
	config := spelldata.SpellConfig(&paladin.Unit, rank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagPassiveSpell))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = classMask
	config.DamageMultiplier = 1
	config.ThreatMultiplier = 1

	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return paladin.activeSeal() != nil
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		paladin.activeSeal().judgement.Cast(sim, target)

		// Let the rotation react when the cooldown ends, since nothing on the GCD marks it.
		pa := sim.GetConsumedPendingActionFromPool()
		pa.NextActionAt = sim.CurrentTime + spell.TimeToReady(sim) + core.SpellBatchWindow
		pa.OnAction = func(sim *core.Simulation) {
			paladin.ReactToEvent(sim, false, false)
		}
		sim.AddPendingAction(pa)
	}

	return config
}
