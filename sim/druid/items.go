package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	core.NewItemEffect(220606, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(446212)
		aura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(row)))
		spelldata.ParseEffects(character, aura, row)
		character.ItemSwap.RegisterProc(220606, aura)
	})

	// Idol of the Moon. Its equip spell states the bonus as A_OVERRIDE_CLASS_SCRIPTS, which has no
	// parse row; the effect's class flags name Moonfire.
	core.NewItemEffect(23197, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(28854)
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label:    "Improved Moonfire",
			ActionID: core.ActionID{SpellID: row.ID},
		}).AttachSpellMod(core.SpellModConfig{
			ClassFlags: row.EffectN(1).ClassFlags,
			Kind:       core.SpellMod_BaseDamage_Flat,
			FloatValue: row.EffectN(1).BaseValue(),
		}))

		character.ItemSwap.RegisterProc(23197, aura)
	})

	// Talons of Wrath. The proc row names no spell; the tooltip names Wrath, whose class flags no other
	// druid spell shares.
	core.NewItemEffect(249441, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		row := spelldata.MustFind(1248996)
		energize := row.EffectN(1).Trigger()
		mana := energize.EnergizeEffect().Average()
		manaMetrics := druid.NewManaMetrics(core.ActionID{SpellID: energize.ID})

		trigger := spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			druid.AddMana(sim, mana, manaMetrics)
		})
		trigger.Name = "Talons of Wrath"
		trigger.ClassFlags = spellData.Wrath.Highest().ClassFlags

		druid.ItemSwap.RegisterProc(249441, druid.MakeProcTriggerAura(trigger))
	})

	// Wolfshead Helm (8345). Effect 1 raises effect 2 of Enrage, its instant rage, which the parser
	// would read as damage. Effect 2 is the energy Shifting Power adds; the row states it as a dummy.
	core.NewItemEffect(8345, func(agent core.Agent) {
		druid := agent.(DruidAgent).GetDruid()
		row := spelldata.MustFind(17768)
		energy := row.EffectN(2).BaseValue()
		enrage := row.EffectN(1).Percent()
		core.MakePermanent(druid.RegisterAura(core.Aura{
			Label:    "Wolfshead Helm",
			ActionID: core.ActionID{SpellID: row.ID},
			OnGain: func(_ *core.Aura, _ *core.Simulation) {
				druid.WolfsheadEnergyBonus += energy
				druid.WolfsheadEnrageBonus += enrage
			},
			OnExpire: func(_ *core.Aura, _ *core.Simulation) {
				druid.WolfsheadEnergyBonus -= energy
				druid.WolfsheadEnrageBonus -= enrage
			},
		}))
	})
}
