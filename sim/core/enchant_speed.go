package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The haste pseudo stats of an item, of its enchant and of its temporary enchants multiply the
// wearer's melee, ranged and cast speed while the item is equipped. Each item and each enchant on an
// item applies its own, an enchant's aura up while the slot's item carries the enchant's id.
func (character *Character) registerEquipSpeedAuras() {
	registerItem, registerEnchant := character.ItemSwap.RegisterProcWithSlots, character.ItemSwap.RegisterEnchantProcWithSlots
	for i := range character.Equipment {
		slot := proto.ItemSlot(i)
		equipped := &character.Equipment[i]
		character.registerSpeedAura(slot, "Item", equipped.ID, equipped.PseudoStats, true, registerItem)
		character.registerSpeedAura(slot, "Enchant", equipped.Enchant.EffectID, equipped.Enchant.PseudoStats, true, registerEnchant)
		character.registerSpeedAura(slot, "Temporary Enchant", equipped.TempEnchant, equipped.TempEnchantPseudoStats, true, registerEnchant)
		character.registerSpeedAura(slot, "Weapon Imbue", equipped.WeaponImbue, equipped.WeaponImbuePseudoStats, true, registerEnchant)
		if !character.ItemSwap.IsEnabled() {
			continue
		}

		swapped := &character.ItemSwap.unEquippedItems[i]
		if swapped.ID != equipped.ID {
			character.registerSpeedAura(slot, "Item", swapped.ID, swapped.PseudoStats, false, registerItem)
		}
		if swapped.Enchant.EffectID != equipped.Enchant.EffectID {
			character.registerSpeedAura(slot, "Enchant", swapped.Enchant.EffectID, swapped.Enchant.PseudoStats, false, registerEnchant)
		}
		if swapped.TempEnchant != equipped.TempEnchant {
			character.registerSpeedAura(slot, "Temporary Enchant", swapped.TempEnchant, swapped.TempEnchantPseudoStats, false, registerEnchant)
		}
		if swapped.WeaponImbue != equipped.WeaponImbue {
			character.registerSpeedAura(slot, "Weapon Imbue", swapped.WeaponImbue, swapped.WeaponImbuePseudoStats, false, registerEnchant)
		}
	}
}

func (character *Character) registerSpeedAura(slot proto.ItemSlot, kind string, id int32, pseudoStats []float64, activeAtStart bool,
	register func(int32, *Aura, []proto.ItemSlot)) {
	if melee, ranged, cast := hastePercents(pseudoStats); melee == 0 && ranged == 0 && cast == 0 {
		return
	}

	aura := character.GetOrRegisterAura(Aura{
		Label:      fmt.Sprintf("%s %d Speed (%s)", kind, id, slot),
		BuildPhase: Ternary(activeAtStart, CharacterBuildPhaseGear, CharacterBuildPhaseNone),
		Duration:   NeverExpires,
	})
	if activeAtStart {
		aura = MakePermanent(aura)
	}
	register(id, aura.AttachHastePseudoStats(pseudoStats), []proto.ItemSlot{slot})
}

func hastePercents(pseudoStats []float64) (melee, ranged, cast float64) {
	return stats.PseudoStatValue(pseudoStats, proto.PseudoStat_PseudoStatMeleeHastePercent),
		stats.PseudoStatValue(pseudoStats, proto.PseudoStat_PseudoStatRangedHastePercent),
		stats.PseudoStatValue(pseudoStats, proto.PseudoStat_PseudoStatSpellHastePercent)
}

// Multiplies the unit's melee, ranged and cast speed by the haste percents pseudoStats states, indexed
// by proto.PseudoStat, while the aura is up.
func (aura *Aura) AttachHastePseudoStats(pseudoStats []float64) *Aura {
	melee, ranged, cast := hastePercents(pseudoStats)
	if melee != 0 {
		aura.AttachMultiplyMeleeSpeed(SymmetricMultiplierFromPercent(melee))
	}
	if ranged != 0 {
		aura.AttachMultiplyRangedSpeed(SymmetricMultiplierFromPercent(ranged))
	}
	if cast != 0 {
		aura.AttachMultiplyCastSpeed(SymmetricMultiplierFromPercent(cast))
	}
	return aura
}
