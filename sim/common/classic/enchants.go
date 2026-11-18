package classic

import (
	"fmt"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {

	// Enchant Gloves - Threat
	core.NewEnchantEffect(2613, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Increase Threat",
		})).AttachMultiplicativePseudoStatBuff(&character.PseudoStats.ThreatMultiplier, 1.02)

		character.ItemSwap.RegisterEnchantProc(2613, aura)
	})

	// Enchant Cloak - Subtlety
	core.NewEnchantEffect(2621, func(agent core.Agent) {
		character := agent.GetCharacter()
		aura := core.MakePermanent(character.RegisterAura(core.Aura{
			Label: "Decrease Threat",
		})).AttachMultiplicativePseudoStatBuff(&character.PseudoStats.ThreatMultiplier, 0.98)

		character.ItemSwap.RegisterEnchantProc(2621, aura)
	})

	// Crusader
	// EffectID: 1900, Proc SpellID: 20007
	// Permanently enchant a melee weapon so that often when attacking in melee
	// it heals for 75 to 125 and increases Strength by 100 for 15 sec.
	// The game files state no rate: 20007 has no proc chance and 458112's ProcChance 100 is the
	// sentinel beside "often". The enchant's effect entry carries it.
	core.NewEnchantEffect(1900, func(agent core.Agent) {
		character := agent.GetCharacter()
		duration := time.Second * 15
		actionID := core.ActionID{SpellID: 20007}
		healthMetrics := character.NewHealthMetrics(actionID)

		createCrusaderAuras := func(slot proto.ItemSlot) *core.StatBuffAura {
			labelSuffix := core.Ternary(slot == proto.ItemSlot_ItemSlotMainHand, "(MH)", "(OH)")
			aura := character.NewTemporaryStatsAura(
				fmt.Sprintf("Holy Strength %s", labelSuffix),
				actionID.WithHand(slot),
				stats.Stats{stats.Strength: 100},
				duration,
			)
			character.AddStatProcBuff(1900, aura, true, []proto.ItemSlot{slot})
			character.ItemSwap.RegisterWeaponEnchantBuff(aura.Aura, 1900)
			return aura
		}

		mhAura := createCrusaderAuras(proto.ItemSlot_ItemSlotMainHand)
		ohAura := createCrusaderAuras(proto.ItemSlot_ItemSlotOffHand)

		character.MakeProcTriggerAura(core.ProcTrigger{
			Name:         "Enchant Weapon - Crusader",
			ActionID:     actionID,
			IsWeaponProc: true,
			DPM:          character.NewDynamicLegacyProcForEnchant(1900, entryPPM(core.GetEnchantByEffectID(1900).EnchantEffects, 20007), 0),
			Outcome:      core.OutcomeLanded,
			Callback:     core.CallbackOnSpellHitDealt,
			Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				core.Ternary(spell.IsOH(), ohAura, mhAura).Activate(sim)
				character.GainHealth(sim, sim.Roll(75, 125), healthMetrics)
			},
		})
	})
}
