package forever

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Dragonbreath Chili
// https://www.wowhead.com/forever/item=12217/dragonbreath-chili
//
// 15852 applies an A_PROC_TRIGGER_SPELL the parser does not attach, so the item routes to no shape and
// ships through ConsumableAllowList. Its row states the roll, the melee hits that make it and the
// internal cooldown; 15851 deals the fire damage to a cone in front of the eater, which the sim takes as
// every target. The 10 minute buff is up for the whole fight.
func init() {
	core.NewConsumableEffect(12217, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(15851)
		damage := row.DamageEffect()

		config := spelldata.SpellConfig(&character.Unit, row, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc())
		if row.IsAProc() {
			config.Flags |= core.SpellFlagProc
		}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			for _, target := range sim.Encounter.ActiveTargetUnits {
				spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
			}
		}
		breath := character.RegisterSpell(config)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(15852), func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
			breath.Cast(sim, character.CurrentTarget)
		})
		trigger.Name = "Dragonbreath Chili"
		character.MakeProcTriggerAura(trigger)
	})
}
