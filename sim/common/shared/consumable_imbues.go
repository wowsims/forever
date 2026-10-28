package shared

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Target resistance a temporary enchant ignores on one school's spells.
type SchoolPiercing struct {
	School stats.SchoolIndex
	Amount float64
}

// A temporary enchant consumable, applied to each hand core put its enchant on, in the main set or
// the item swap set. Its stats, speed percents and flat weapon damage are core's, on the item
// carrying the enchant. Each combat spell is cast off the hits of the weapons carrying the enchant at
// the chance its row states. The equip spells named are the enchant's that grant no stat of the
// enchant's own, such as Consecrated Sharpening Stone's 28893: their auras are parsed onto an aura
// per hand, since each weapon carries its own copy of the equip aura, and an aura bound to a matching
// weapon reaches that hand's attacks alone. An equip spell that procs, such as Scroll of Imbue
// Manablade's 1308913, is a listener built from its row instead. The piercing is a stat slot of the
// enchant that has no place in the stat array, held by the same per-hand aura.
func NewSpellDataTemporaryEnchant(itemID int32, piercing SchoolPiercing, equipSpellIDs ...int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		slots := applyTemporaryEnchant(character, consumable)
		if len(slots) == 0 {
			return
		}
		var equipSpells []*spelldata.Spell
		for _, spellID := range equipSpellIDs {
			row := spelldata.MustFind(spellID)
			if TemporaryEnchantEquipSpellProcs(row) {
				registerTemporaryEnchantProc(character, consumable, row, slots)
				continue
			}
			equipSpells = append(equipSpells, row)
		}
		if piercing.Amount == 0 && len(equipSpells) == 0 {
			return
		}
		for _, slot := range slots {
			registerTemporaryEnchantHand(character, consumable, slot, piercing, equipSpells)
		}
	})
}

// Returns the melee hands carrying the enchant in the main set or the item swap set.
func applyTemporaryEnchant(character *core.Character, consumable core.Consumable) []proto.ItemSlot {
	enchant := consumable.TemporaryEnchant
	if enchant == nil {
		return nil
	}

	var slots []proto.ItemSlot
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		swapped := character.ItemSwap.GetUnequippedItemBySlot(slot).HasTempEnchant(enchant.EffectID)
		if character.Equipment[slot].HasTempEnchant(enchant.EffectID) || swapped {
			slots = append(slots, slot)
		}
	}
	if len(slots) == 0 {
		return nil
	}

	for _, effect := range enchant.EnchantEffects {
		registerTemporaryEnchantCombatSpell(character, consumable, effect)
	}
	return slots
}

// The combat spell rolls on the hits of the weapons carrying the enchant, at their own speeds.
func registerTemporaryEnchantCombatSpell(character *core.Character, consumable core.Consumable, effect *proto.ItemEffect) {
	damage := spelldata.MustFind(effect.BuffId)
	enchantID := consumable.TemporaryEnchant.EffectID
	chance := damage.StatedChance()

	config := spelldata.ProcTrigger(character, damage, nil, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(enchantID, 0, func(core.ProcMask) float64 { return chance })
	})
	config.Name = consumable.Name
	config.ActionID = core.ActionID{ItemID: consumable.Id}
	attachProcDamage(character, &config, spellDataProcDamageSpell(character, damage, true))

	character.MakeProcTriggerAura(config)
}

// The enchant's copy on one hand, up while the weapon in that hand carries the enchant.
func registerTemporaryEnchantHand(character *core.Character, consumable core.Consumable, slot proto.ItemSlot, piercing SchoolPiercing, equipSpells []*spelldata.Spell) {
	enchantID := consumable.TemporaryEnchant.EffectID
	actionID := core.ActionID{ItemID: consumable.Id}
	if len(equipSpells) == 1 {
		actionID = core.ActionID{SpellID: equipSpells[0].ID}
	}
	equipped := character.Equipment[slot].HasTempEnchant(enchantID)
	aura := character.RegisterAura(core.Aura{
		Label:    fmt.Sprintf("%s (%s)", consumable.Name, core.Ternary(slot == proto.ItemSlot_ItemSlotMainHand, "Main Hand", "Off Hand")),
		ActionID: actionID.WithTag(int32(slot)),
		Duration: core.NeverExpires,
		// On the hand from the start, the enchant's equip spells show among the sheet's consumables.
		BuildPhase: core.Ternary(equipped, core.CharacterBuildPhaseConsumes, core.CharacterBuildPhaseNone),
	})
	if piercing.Amount != 0 {
		aura.AttachAdditivePseudoStatBuff(&character.PseudoStats.SchoolSpellPiercing[piercing.School], piercing.Amount)
	}
	hand := character.GetProcMaskForWeaponSlot(slot)
	for _, row := range equipSpells {
		spelldata.ParseEffects(character, aura, row, spelldata.OnHand(hand), spelldata.Count(1))
	}

	if equipped {
		core.MakePermanent(aura)
	}
	character.ItemSwap.RegisterEnchantProcWithSlots(enchantID, aura, []proto.ItemSlot{slot})
}

func TemporaryEnchantEquipSpellProcs(row *spelldata.Spell) bool {
	return row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL) != spelldata.NilEffect
}

// An equip spell that procs, built from its row: the hits it hears, its chance and its internal
// cooldown, listening while a weapon carrying the enchant is equipped. A listener on the wearer's own
// hits hears the hands carrying the enchant alone; one on the hits the wearer takes hears every one
// of them.
func registerTemporaryEnchantProc(character *core.Character, consumable core.Consumable, row *spelldata.Spell, slots []proto.ItemSlot) {
	actionID := core.ActionID{ItemID: consumable.Id}
	handler := FiredSpellHandler(character, actionID, row)
	if handler == nil {
		return
	}
	enchantID := consumable.TemporaryEnchant.EffectID

	config := spelldata.ProcTrigger(character, row, handler)
	if config.Callback.Matches(core.CallbackOnSpellHitDealt) {
		var hands core.ProcMask
		for _, slot := range slots {
			hands |= character.GetProcMaskForWeaponSlot(slot)
		}
		config.ProcMask &= hands

		if character.ItemSwap.IsEnabled() {
			enchanted := character.GetDynamicProcMaskForWeaponTempEnchant(enchantID)
			extraCondition := config.ExtraCondition
			config.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
				return (spell == nil || spell.ProcMask.Matches(*enchanted)) && (extraCondition == nil || extraCondition(sim, spell, result))
			}
		}
	}
	config.Name = consumable.Name
	config.ActionID = actionID

	character.ItemSwap.RegisterEnchantProcWithSlots(enchantID, character.MakeProcTriggerAura(config), slots)
}
