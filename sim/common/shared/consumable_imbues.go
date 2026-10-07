package shared

import (
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

// A temporary enchant consumable, applied to each hand core put its enchant on: the enchant's stats
// once per hand and each combat spell cast off that hand's hits at the chance its row states. Its flat
// weapon damage is core's, on every weapon built from that hand's item. The equip spells named are
// the enchant's that grant no stat of the enchant's own, such as Consecrated Sharpening Stone's 28893:
// their auras are parsed onto one permanent aura once per hand, since each weapon carries its own copy
// of the equip aura, and an aura bound to a matching weapon reaches that hand's attacks alone. An
// equip spell that procs, such as Scroll of Imbue Manablade's 1308913, is a listener built from its
// row instead. The piercing is a stat slot of the enchant that has no place in the stat array, added
// once per hand too.
func NewSpellDataTemporaryEnchant(itemID int32, piercing SchoolPiercing, equipSpellIDs ...int32) {
	core.NewSpellDataConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		consumable := core.GetConsumableByID(itemID)
		hands := applyTemporaryEnchant(character, consumable)
		if len(hands) == 0 {
			return
		}
		character.PseudoStats.SchoolSpellPiercing[piercing.School] += piercing.Amount * float64(len(hands))
		for _, spellID := range equipSpellIDs {
			row := spelldata.MustFind(spellID)
			if TemporaryEnchantEquipSpellProcs(row) {
				registerTemporaryEnchantProc(character, consumable, row, hands)
				continue
			}
			aura := character.GetOrRegisterAura(spelldata.AuraConfig(row, spelldata.Permanent()))
			for _, hand := range hands {
				spelldata.ParseEffects(character, aura, row, spelldata.OnHand(hand), spelldata.Count(1))
			}
		}
	})
}

// Returns the melee hands carrying the enchant, as proc masks.
func applyTemporaryEnchant(character *core.Character, consumable core.Consumable) []core.ProcMask {
	enchant := consumable.TemporaryEnchant
	if enchant == nil {
		return nil
	}

	var hands []core.ProcMask
	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		if !character.Equipment[slot].HasTempEnchant(enchant.EffectID) {
			continue
		}
		hands = append(hands, character.GetProcMaskForWeaponSlot(slot))
		character.AddStats(enchant.Stats.Add(stats.FromPseudoStatsProto(enchant.PseudoStats)))
	}
	if len(hands) == 0 {
		return nil
	}

	for _, effect := range enchant.EnchantEffects {
		registerTemporaryEnchantCombatSpell(character, consumable, effect)
	}
	return hands
}

// The combat spell rolls on the hits of the weapons carrying the enchant, at their own speeds.
func registerTemporaryEnchantCombatSpell(character *core.Character, consumable core.Consumable, effect *proto.ItemEffect) {
	damage := spelldata.MustFind(effect.BuffId)
	damageSpell := character.RegisterSpell(spellDataProcDamageSpell(character, damage, true))
	enchantID := consumable.TemporaryEnchant.EffectID
	chance := damage.StatedChance()

	config := spelldata.ProcTrigger(character, damage, nil, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(enchantID, 0, func(core.ProcMask) float64 { return chance })
	})
	config.Name = consumable.Name
	config.ActionID = core.ActionID{ItemID: consumable.Id}
	config.Handler = procDamageHandler(character, damageSpell, config.Callback)
	config.TriggerImmediately = true

	character.MakeProcTriggerAura(config)
}

func TemporaryEnchantEquipSpellProcs(row *spelldata.Spell) bool {
	return row.FirstAura(dbcenums.A_PROC_TRIGGER_SPELL) != spelldata.NilEffect
}

// An equip spell that procs, built from its row: the hits it hears, its chance and its internal
// cooldown. A listener on the wearer's own hits hears the hands carrying the enchant alone; one on
// the hits the wearer takes hears every one of them.
func registerTemporaryEnchantProc(character *core.Character, consumable core.Consumable, row *spelldata.Spell, hands []core.ProcMask) {
	actionID := core.ActionID{ItemID: consumable.Id}
	handler := FiredSpellHandler(character, actionID, row)
	if handler == nil {
		return
	}

	config := spelldata.ProcTrigger(character, row, handler)
	if config.Callback.Matches(core.CallbackOnSpellHitDealt) {
		var enchanted core.ProcMask
		for _, hand := range hands {
			enchanted |= hand
		}
		config.ProcMask &= enchanted
	}
	config.Name = consumable.Name
	config.ActionID = actionID

	character.MakeProcTriggerAura(config)
}
