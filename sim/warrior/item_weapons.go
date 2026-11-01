package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	// Old Guard Retaliator
	// https://www.wowhead.com/forever/item=208222/old-guard-retaliator
	//
	// On landing a killing blow that grants experience or honor, your next attack will critically
	// strike.
	//
	// 418509 hears the killing blow, and its dummy effect's 100 is the crit chance the next attack
	// gets. The game files state no buff for it, so the buff lasts until a melee or ranged attack
	// spends it, landed or not.
	core.NewItemEffect(208222, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(418509)

		buff := character.RegisterAura(core.Aura{
			Label:    "Primal Bloodthirst",
			ActionID: core.ActionID{SpellID: row.ID},
			Duration: core.NeverExpires,
		}).AttachStatBuff(stats.PhysicalCritPercent, row.EffectN(1).BaseValue())
		buff.AttachProcTrigger(core.ProcTrigger{
			Name:               "Primal Bloodthirst - Spend",
			Callback:           core.CallbackOnSpellHitDealt,
			ProcMask:           core.ProcMaskMeleeOrRanged,
			TriggerImmediately: true,
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				buff.Deactivate(sim)
			},
		})

		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		})
		trigger.Name = "Old Guard Retaliator"
		trigger.ActionID = core.ActionID{ItemID: 208222}
		character.ItemSwap.RegisterProc(208222, character.MakeProcTriggerAura(trigger))
	})
}
