import { ENCHANT_PRESET_SLOTS, enchantPresetPoolSlot } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { Class, EnchantProcKind, ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { MISSING_ENCHANT_EFFECTS } from '@sim/constants/missing_effects_auto_gen';
import type { Player } from '@sim/player/player';
import { enchantEP, type EquippedItem, getWeaponDpsStatsBySlot } from '@sim/proto/equipped_item';
import type { Gear } from '@sim/proto/gear';
import { enchantAppliesToItem, isStaticEnchant } from '@sim/proto/items';
import type { Stats } from '@sim/proto/stats';

export type SkippedEnchantSlot = { slot: ItemSlot; reason: 'proc' | 'unlisted' };

type FittingEnchants = { slot: ItemSlot; equipped: EquippedItem; fitting: Enchant[] };

export type EnchantSlot = SkippedEnchantSlot | FittingEnchants;

const ENCHANTABLE_SLOTS = [...ENCHANT_PRESET_SLOTS, ItemSlot.ItemSlotFinger2].sort((a, b) => a - b);

// A slot whose item can take a proc enchant the sim implements is left to the sim. With a preset pool, a slot draws
// only on the enchants the checked presets list for it.
export const classifyEnchantSlots = (
	player: Player<any>,
	gear: Gear,
	frozenSlots: readonly ItemSlot[],
	presetPool?: readonly BulkEnchantSlotOptions[],
): EnchantSlot[] => {
	const survivalBlocks = !!presetPool || player.playerSpec.isTankSpec;
	return ENCHANTABLE_SLOTS.flatMap((slot): EnchantSlot[] => {
		const equipped = gear.getEquippedItem(slot);
		if (!equipped || frozenSlots.includes(slot)) return [];
		const item = equipped.item;
		let fitting = player.getEnchants(slot).filter(enchant => enchantAppliesToItem(enchant, item));
		if (presetPool && fitting.length) {
			const listed = presetPool.find(pooled => pooled.slot === enchantPresetPoolSlot(slot))?.effectIds ?? [];
			fitting = fitting.filter(enchant => listed.includes(enchant.effectId));
			if (!fitting.length) return [{ slot, reason: 'unlisted' }];
		}
		const blocked = fitting.some(
			enchant =>
				(enchant.procKind === EnchantProcKind.EnchantProcKindCombat ||
					(survivalBlocks && enchant.procKind === EnchantProcKind.EnchantProcKindSurvival)) &&
				!MISSING_ENCHANT_EFFECTS.has(enchant.effectId),
		);
		return [blocked ? { slot, reason: 'proc' } : { slot, equipped, fitting }];
	});
};

export const skippedEnchantSlots = (slots: readonly EnchantSlot[]): SkippedEnchantSlot[] =>
	slots.filter((entry): entry is SkippedEnchantSlot => 'reason' in entry);

// The static enchants the gear optimizer may pick per slot: those worth more than 0 EP, or from a preset pool, every
// listed one whatever its EP.
export const buildEnchantOptions = (slots: readonly EnchantSlot[], weights: Stats, playerClass: Class, fromPresetPool: boolean): BulkEnchantSlotOptions[] =>
	slots.flatMap(entry => {
		if ('reason' in entry) return [];
		const { slot, equipped, fitting } = entry;
		const weaponSpeed = equipped.item.weaponSpeed;
		const epPerWeaponDps = weaponSpeed ? getWeaponDpsStatsBySlot(1, slot).computeEP(weights, playerClass) : 0;
		const effectIds = fitting
			.filter(isStaticEnchant)
			.filter(enchant => fromPresetPool || enchantEP(enchant, weights, playerClass, weaponSpeed, epPerWeaponDps) > 0)
			.map(enchant => enchant.effectId);
		return effectIds.length ? [BulkEnchantSlotOptions.create({ slot, effectIds: [...new Set(effectIds)].sort((a, b) => a - b) })] : [];
	});
