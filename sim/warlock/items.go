package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const blackBookItemID = 19337

func init() {
	// The Black Book: Blessing of the Black Book doubles the damage and the armor of the demon out. Every
	// effect of the row lands on the pet, so the item cannot be used with no demon out.
	core.NewItemEffect(blackBookItemID, func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		itemEffect := core.GetItemByID(blackBookItemID).ItemEffects[0]
		onUse := itemEffect.GetOnUse()
		row := spelldata.MustFind(itemEffect.BuffId)

		petAuras := make(map[*WarlockPet]*core.Aura)
		for _, pet := range warlock.demons() {
			aura := pet.RegisterAura(spelldata.AuraConfig(row))
			spelldata.ParseEffects(&pet.Character, aura, row)
			petAuras[pet] = aura
		}

		cast := spelldata.Cast(row)
		cast.CD = core.Cooldown{
			Timer:    warlock.NewTimer(),
			Duration: time.Duration(onUse.CooldownMs) * time.Millisecond,
		}
		cast.SharedCD = core.Cooldown{
			Timer:    warlock.GetOrInitSpellCategoryTimer(onUse.CategoryId),
			Duration: time.Duration(onUse.CategoryCooldownMs) * time.Millisecond,
		}

		spell := warlock.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: blackBookItemID},
			SpellSchool: row.SpellSchool(),
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete,
			Cast:        cast,

			ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
				return warlock.ActivePet != nil && warlock.ActivePet.IsEnabled()
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
				petAuras[warlock.ActivePet].Activate(sim)
			},
		})

		warlock.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
		})
	})
}
