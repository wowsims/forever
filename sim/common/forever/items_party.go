package forever

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Items whose effect reaches the wearer's whole party. The wearer is in its own party, so each puts
// the party buff's own copy on the wearer, and that copy joins the party buff's category: a party
// buff ticked in the settings and the wearer's own item are worth one buff, not two.
func init() {
	// Darkspear
	// https://www.wowhead.com/forever/item=12802/darkspear
	//
	// Chance on hit to grant nearby party members 4% increased critical strike chance.
	//
	// 1314143 states the 5% chance on melee swings and abilities and the 60 s between procs; the buff is
	// 16939, 4% crit for 20 s.
	core.NewItemEffect(12802, func(agent core.Agent) {
		character := agent.GetCharacter()
		buff := buffs.NewSpellDataDarkspearAura(&character.Unit, true, 0)

		trigger := spelldata.ProcTrigger(character, spelldata.MustFind(1314143), func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			buff.Activate(sim)
		})
		trigger.Name = "Darkspear"
		trigger.ActionID = core.ActionID{ItemID: 12802}
		character.ItemSwap.RegisterProc(12802, character.MakeProcTriggerAura(trigger))
	})

	// Admiral's Hat
	// https://www.wowhead.com/forever/item=10030/admiral-s-hat
	//
	// Gives 10 additional stamina to party members within 30 yards.
	//
	// The use is 12022, 10 Stamina for 10 min, on the item's 1 min cooldown.
	core.NewItemEffect(10030, func(agent core.Agent) {
		character := agent.GetCharacter()
		buff := buffs.NewSpellDataAdmiralsHatAura(&character.Unit, true, 0)
		row := spelldata.MustFind(12022)

		config := core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 10030},
			SpellSchool: row.SpellSchool(),
			ProcMask:    core.ProcMaskEmpty,
			Cast:        spelldata.Cast(row),
			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				buff.Activate(sim)
			},
		}
		for _, effect := range core.GetItemByID(10030).ItemEffects {
			if effect.GetOnUse() != nil {
				itemCast := shared.OnUseCast(character, effect)
				config.Cast.CD, config.Cast.SharedCD = itemCast.CD, itemCast.SharedCD
			}
		}

		character.AddMajorCooldown(core.MajorCooldown{
			Spell: character.RegisterSpell(config),
			Type:  core.CooldownTypeSurvival,
		})
	})
}
