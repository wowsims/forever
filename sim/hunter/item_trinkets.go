package hunter

import (
	"fmt"
	"slices"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var (
	arcaneInfused     = spelldata.MustFind(23721) // the on-use buff: 10 s, one charge
	arcaneDetonation  = spelldata.MustFind(23722) // the area damage it fires
	refocusedShotsFor = []spelldata.Ladder{spellData.AimedShot, spellData.MultiShot, spellData.Volley, spellData.ArcaneShot}
	ferrasCollar      = spelldata.MustFind(22854) // the equip spell: a flat 2 on every effect of Tamed Pet Passive
)

func init() {
	// Arcane Infused Gem
	// https://www.wowhead.com/forever/item=19336/arcane-infused-gem
	//
	// Infuses you with Arcane energy, causing your next Arcane Shot fired within 10s to detonate at
	// the target. The Arcane Detonation will deal 200 damage to enemies near the target.
	//
	// The buff's row fires the detonation on any ranged or spell hit and names no spell; the
	// tooltip's Arcane Shot is the condition. 23722 states 200 with a 0.15 variance on every enemy
	// within 8 yards.
	core.NewItemEffect(19336, func(agent core.Agent) {
		character := agent.GetCharacter()

		detonationConfig := spelldata.SpellConfig(&character.Unit, arcaneDetonation, spelldata.Magic(core.ProcMaskEmpty), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc))
		damage := arcaneDetonation.DamageEffect()
		detonationConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.CalcAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
				return damage.DirectAmount(sim) * sim.Encounter.AOECapMultiplier()
			})
			spell.DealBatchedAoeDamage(sim)
		}
		detonation := character.RegisterSpell(detonationConfig)

		var buff *core.Aura
		trigger := spelldata.ProcTrigger(character, arcaneInfused, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			detonation.Cast(sim, result.Target)
			buff.Deactivate(sim)
		})
		trigger.Name = "Arcane Infused - Trigger"
		trigger.TriggerImmediately = true
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spellData.ArcaneShot.Contains(spell)
		}
		buff = character.RegisterAura(spelldata.AuraConfig(arcaneInfused)).AttachProcTrigger(trigger)

		registerTrinketOnUse(character, 19336, buff, func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			buff.Activate(sim)
		}, nil)
	})

	// Renataki's Charm of Beasts
	// https://www.wowhead.com/forever/item=19953/renataki-s-charm-of-beasts
	//
	// Instantly clears the cooldowns of Aimed Shot, Multishot, Volley, and Arcane Shot.
	//
	// Refocus 24531 is a dummy the game scripts. Aimed Shot and Multi-Shot share one category's
	// cooldown. The charm is only worth using while one of the shots is cooling down.
	core.NewItemEffect(19953, func(agent core.Agent) {
		character := agent.GetCharacter()

		var shots []*core.Spell
		character.OnSpellRegistered(func(spell *core.Spell) {
			if slices.ContainsFunc(refocusedShotsFor, func(l spelldata.Ladder) bool { return l.Contains(spell) }) {
				shots = append(shots, spell)
			}
		})

		registerTrinketOnUse(character, 19953, nil, func(_ *core.Simulation, _ *core.Unit, _ *core.Spell) {
			for _, shot := range shots {
				if shot.CD.Timer != nil {
					shot.CD.Reset()
				}
			}
		}, func(sim *core.Simulation, _ *core.Character) bool {
			return slices.ContainsFunc(shots, func(shot *core.Spell) bool { return !shot.IsReady(sim) })
		})
	})

	// Ferra's Collar
	// https://www.wowhead.com/forever/item=18355/ferra-s-collar
	//
	// Increases the damage done by your pets by 4%.
	//
	// 22854's class mask names one spell, Tamed Pet Passive 8875, whose one effect raises the pet's
	// damage done in every school by its amount, 0 on its own. The flat 2 makes that 2%, not the
	// tooltip's 4%.
	core.NewItemEffect(18355, func(agent core.Agent) {
		character := agent.GetCharacter()
		slots := character.ItemSwap.EligibleSlotsForItem(18355)
		multiplier := core.MultiplierFromPercent(ferrasCollar.EffectN(1).Average())

		for _, pet := range character.Pets {
			if pet.IsGuardian() {
				continue
			}
			// Checked on reset rather than made permanent: a pet resets after its owner's item swap
			// has settled.
			config := spelldata.AuraConfig(ferrasCollar)
			config.Duration = core.NeverExpires
			config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
				if character.HasItemEquipped(18355, slots) {
					aura.Activate(sim)
				}
			}
			aura := pet.RegisterAura(config).AttachMultiplicativePseudoStatBuff(&pet.PseudoStats.DamageDealtMultiplier, multiplier)
			character.ItemSwap.RegisterProcWithSlots(18355, aura, slots)
		}
	})
}

// A DPS cooldown on the cooldown and the category cooldown the item's on-use effect states.
func registerTrinketOnUse(character *core.Character, itemID int32, buff *core.Aura, apply core.ApplySpellResults, shouldActivate core.CooldownActivationCondition) {
	effects := shared.ItemEffects(itemID)
	i := slices.IndexFunc(effects, func(effect *proto.ItemEffect) bool { return effect.GetOnUse() != nil })
	if i < 0 {
		panic(fmt.Sprintf("item %d has no on-use effect", itemID))
	}

	character.AddMajorCooldown(core.MajorCooldown{
		Spell: character.RegisterSpell(core.SpellConfig{
			ActionID:        core.ActionID{ItemID: itemID},
			Flags:           core.SpellFlagNoOnCastComplete,
			Cast:            shared.OnUseCast(character, effects[i]),
			ApplyEffects:    apply,
			RelatedSelfBuff: buff,
		}),
		Type:           core.CooldownTypeDPS,
		ShouldActivate: shouldActivate,
	})
}
