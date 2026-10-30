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
	rank := spellData.Judgement.Highest()

	config := spelldata.SpellConfig(&paladin.Unit, rank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagPassiveSpell|core.SpellFlagNoOnCastComplete))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskJudgement
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

	paladin.Judgement = paladin.RegisterSpell(config)

	// Every melee strike that lands refreshes the judgement debuffs on its target.
	paladin.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Judgement Refresh" + paladin.Label,
		Callback:           core.CallbackOnSpellHitDealt,
		ProcMask:           core.ProcMaskMelee,
		Outcome:            core.OutcomeLanded,
		TriggerImmediately: true,
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			paladin.refreshJudgements(sim, result.Target)
		},
	})
}
