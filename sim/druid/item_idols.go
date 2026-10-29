package druid

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The idols' equip spells reach their spells by class flags.
//
// TODO: Items without full game files data are not registered here:
//   - Idol of Shifting Tides 263411 (1270470, shapeshifting spells cost 40 less mana).
func init() {
	core.NewItemEffect(263435, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1270490)
		config := spelldata.AuraConfig(row)
		if character.HasItemEquipped(263435, []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}) {
			config.BuildPhase = core.CharacterBuildPhaseGear
		}
		aura := character.RegisterAura(config)
		spelldata.ParseEffects(character, aura, row)
		// Bonus item armor uses a separate stat from the base armor parsed above.
		bonus := 1 + row.FindEffect(dbcenums.E_APPLY_AURA, dbcenums.A_MOD_BONUS_ARMOR_PCT, 0).Percent()
		aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
			character.ApplyDynamicEquipScaling(sim, stats.BonusArmor, bonus)
		})
		aura.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			character.RemoveDynamicEquipScaling(sim, stats.BonusArmor, bonus)
		})
		character.ItemSwap.RegisterProc(263435, core.MakePermanent(aura))
	})

	core.NewItemEffect(279251, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		row := spelldata.MustFind(1306483)
		trigger := spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.PrimalBite.CD.Reset()
		})
		trigger.Outcome = core.OutcomeLanded
		// The first Lacerate hits for zero before it adds a stack, but still qualifies.
		trigger.RequireDamageDealt = false
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return druid.PrimalBite != nil && spellData.Lacerate.Contains(spell)
		}
		druid.ItemSwap.RegisterProc(279251, druid.MakeProcTriggerAura(trigger))
	})

	// Idol of Health
	// https://www.wowhead.com/forever/item=22399/idol-of-health
	//
	// Reduces the casting time of your Healing Touch spell by 0.15 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 27846},
		[]shared.ItemVariant{
			{ItemID: 22399, ItemName: "Idol of Health"},
		})

	// Howling Idol
	// https://www.wowhead.com/forever/item=272427/howling-idol
	//
	// Reduces the cooldown of your Shifting Power ability by 0.5 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291059},
		[]shared.ItemVariant{
			{ItemID: 272427, ItemName: "Howling Idol"},
		})

	// Enraged Idol
	// https://www.wowhead.com/forever/item=272428/enraged-idol
	//
	// Your Enrage ability generates an additional 10 Rage over its duration.
	//
	// TODO: 1291060 states the 10 as SPELLMOD_EFFECT1 on Enrage, which the parse reads as base
	// damage. The Enrage port has to add the rage off its effect 1.
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291060},
	// 	[]shared.ItemVariant{
	// 		{ItemID: 272428, ItemName: "Enraged Idol"},
	// 	})

	// Swarming Idol
	// https://www.wowhead.com/forever/item=272430/swarming-idol
	//
	// Increases the duration of your Insect Swarm ability by 2 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1291061},
		[]shared.ItemVariant{
			{ItemID: 272430, ItemName: "Swarming Idol"},
		})

	// Mystic Mushroom
	//
	// Increases your Spirit by 5%.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1248751},
		[]shared.ItemVariant{
			{ItemID: 249396, ItemName: "Mystic Mushroom"},
		})

	// Idol of Swiftness
	// https://www.wowhead.com/forever/item=279250/idol-of-swiftness
	//
	// Reduces the cooldown of your Swiftmend spell by 3 sec.
	shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 1306488},
		[]shared.ItemVariant{
			{ItemID: 279250, ItemName: "Idol of Swiftness"},
		})
}
