// What a player may equip: item and enchant eligibility, weapon-type rules and
// the gear identity keys the reforge cache is keyed on.
import {
	EnchantProcKind,
	EnchantType,
	EquipmentSpec,
	HandType,
	ItemSlot,
	ItemType,
	Profession,
	RangedWeaponType,
	Spec,
	WeaponType,
} from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import { PlayerSpec } from '../player/player_spec';
import { PlayerSpecs } from '../player/specs';
import { intersection, swap } from '../utils/collections';
import { BLUNT_WEAPON_TYPES, ITEM_TYPE_TO_SLOTS, SHARP_WEAPON_TYPES } from './sim_tables_auto_gen';

export function isSharpWeaponType(weaponType: WeaponType): boolean {
	return SHARP_WEAPON_TYPES.includes(weaponType);
}

export function isBluntWeaponType(weaponType: WeaponType): boolean {
	return BLUNT_WEAPON_TYPES.includes(weaponType);
}

// Returns true if this item may be equipped in at least 1 slot for the given Spec.
export function canEquipItem<SpecType extends Spec>(item: Item, playerSpec: PlayerSpec<SpecType>, slot: ItemSlot | undefined): boolean {
	const playerClass = PlayerSpecs.getPlayerClass(playerSpec);
	if (item.classAllowlist.length > 0 && !item.classAllowlist.includes(playerClass.classID)) {
		return false;
	}

	if ([ItemType.ItemTypeFinger, ItemType.ItemTypeTrinket].includes(item.type)) {
		return true;
	}

	if (item.type == ItemType.ItemTypeWeapon) {
		const eligibleWeaponType = playerClass.weaponTypes.find(wt => wt.weaponType == item.weaponType);
		if (!eligibleWeaponType) {
			return false;
		}

		if (
			(item.handType == HandType.HandTypeOffHand || (item.handType == HandType.HandTypeOneHand && slot == ItemSlot.ItemSlotOffHand)) &&
			![WeaponType.WeaponTypeShield, WeaponType.WeaponTypeOffHand].includes(item.weaponType) &&
			!playerSpec.canDualWield
		) {
			return false;
		}

		if (item.handType == HandType.HandTypeTwoHand && !eligibleWeaponType.canUseTwoHand) {
			return false;
		}
		if (item.handType == HandType.HandTypeTwoHand && slot == ItemSlot.ItemSlotOffHand) {
			return false;
		}

		return true;
	}

	if (item.type == ItemType.ItemTypeRanged) {
		return playerClass.rangedWeaponTypes.includes(item.rangedWeaponType);
	}

	// At this point, we know the item is an armor piece (feet, chest, legs, etc).
	return playerClass.armorTypes[0] >= item.armorType;
}

// Honor rank and battleground item name prefixes. "Premier " items are Forever's reissued PvP items and all count.
const PVP_NAME_PREFIXES = [
	'Premier ',
	"Private's ",
	"Corporal's ",
	"Sergeant's ",
	"Master Sergeant's ",
	"Sergeant Major's ",
	"Knight-Lieutenant's ",
	"Knight-Captain's ",
	"Knight-Champion's ",
	"Lieutenant Commander's ",
	"Marshal's ",
	"Field Marshal's ",
	"Grand Marshal's ",
	"Scout's ",
	"Grunt's ",
	"Senior Sergeant's ",
	"First Sergeant's ",
	"Stone Guard's ",
	"Blood Guard's ",
	"Legionnaire's ",
	"Centurion's ",
	"Champion's ",
	"Lieutenant General's ",
	"General's ",
	"Warlord's ",
	"High Warlord's ",
	"Highlander's ",
	"Defiler's ",
	"Sentinel's ",
	"Outrider's ",
	'Stormpike ',
	'Frostwolf ',
	'Insignia of the ',
];

// PvE items whose names start with a PvP prefix.
const PVE_ITEM_NAMES = new Set([
	"Legionnaire's Leggings",
	"Sentinel's Guard",
	"Sentinel's Belt",
	"Sentinel's Gauntlet",
	"Sentinel's Plate",
	"Scout's Dagger",
	"Champion's Legplates",
]);

export const isPVPItem = (item: Item) => {
	const name = item?.name;
	if (!name || PVE_ITEM_NAMES.has(name)) return false;
	return PVP_NAME_PREFIXES.some(prefix => name.startsWith(prefix));
};

export function getEligibleItemSlots(item: Item): Array<ItemSlot> {
	if (ITEM_TYPE_TO_SLOTS[item.type]) {
		return ITEM_TYPE_TO_SLOTS[item.type]!;
	}

	if (item.type == ItemType.ItemTypeWeapon) {
		if (item.handType == HandType.HandTypeMainHand) {
			return [ItemSlot.ItemSlotMainHand];
		} else if (item.handType == HandType.HandTypeOffHand) {
			return [ItemSlot.ItemSlotOffHand];
		} else {
			return [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];
		}
	}

	// Should never reach here
	throw new Error('Could not find item slots for item: ' + Item.toJsonString(item));
}

export const isSecondaryItemSlot = (slot: ItemSlot) => slot === ItemSlot.ItemSlotFinger2 || slot === ItemSlot.ItemSlotTrinket2;

// Returns whether the given main-hand and off-hand items can be worn at the
// same time.
export function validWeaponCombo(mainHand: Item | null | undefined, offHand: Item | null | undefined): boolean {
	return mainHand?.handType != HandType.HandTypeTwoHand && offHand?.handType != HandType.HandTypeTwoHand;
}

export function getEligibleEnchantSlots(enchant: Enchant): Array<ItemSlot> {
	return [enchant.type]
		.concat(enchant.extraTypes || [])
		.map(type => {
			if (ITEM_TYPE_TO_SLOTS[type]) {
				return ITEM_TYPE_TO_SLOTS[type]!;
			}

			if (type == ItemType.ItemTypeWeapon) {
				return [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand];
			}

			// Should never reach here
			throw new Error('Could not find item slots for enchant: ' + Enchant.toJsonString(enchant));
		})
		.flat();
}

export function isOffHandOnlyEnchant(enchant: Enchant): boolean {
	return enchant.enchantType === EnchantType.EnchantTypeShield || enchant.enchantType === EnchantType.EnchantTypeOffHand;
}

export function enchantAppliesToItem(enchant: Enchant, item: Item): boolean {
	const sharedSlots = intersection(getEligibleEnchantSlots(enchant), getEligibleItemSlots(item));
	if (!sharedSlots.length) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeTwoHand && item.handType !== HandType.HandTypeTwoHand) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeStaff && item.weaponType !== WeaponType.WeaponTypeStaff) return false;

	if (enchant.enchantType === EnchantType.EnchantTypeShield && item.weaponType !== WeaponType.WeaponTypeShield) return false;

	if ((enchant.enchantType === EnchantType.EnchantTypeOffHand) !== (item.weaponType === WeaponType.WeaponTypeOffHand)) return false;

	if (item.weaponType === WeaponType.WeaponTypeShield && enchant.enchantType !== EnchantType.EnchantTypeShield) return false;

	if (enchant.type == ItemType.ItemTypeRanged) {
		if (
			![RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeCrossbow, RangedWeaponType.RangedWeaponTypeGun].includes(
				item.rangedWeaponType,
			)
		)
			return false;
	}

	if (item.rangedWeaponType != RangedWeaponType.RangedWeaponTypeWand && item.rangedWeaponType > 0 && enchant.type != ItemType.ItemTypeRanged) {
		return false;
	}

	return true;
}

export function canEquipEnchant<SpecType extends Spec>(enchant: Enchant, playerSpec: PlayerSpec<SpecType>, hasEnchanting: boolean): boolean {
	if (enchant.classAllowlist.length > 0 && !enchant.classAllowlist.includes(playerSpec.classID)) {
		return false;
	}

	// This enchant requires the Enchanting profession.
	if (enchant.requiredProfession == Profession.Enchanting && !hasEnchanting) {
		return false;
	}

	return true;
}

// An enchant with no proc that grants stats, pseudo-stats or weapon damage: its EP is the whole of its worth.
export const isStaticEnchant = (enchant: Enchant): boolean =>
	enchant.procKind === EnchantProcKind.EnchantProcKindNone &&
	(enchant.stats.some(value => value !== 0) || enchant.pseudoStats.some(value => value !== 0) || enchant.weaponDamage !== 0);

/**
 * Fingerprint for comparing or deduplicating gear sets: item, random suffix and enchant.
 *
 * NOT a cache key. Use getReforgeCacheGearKey for anything that keys an optimizer result.
 */
export function getGearIdentityKey(spec: EquipmentSpec): string {
	return buildGearKey(spec);
}

/**
 * Cache key for an optimizer result: the identity fingerprint plus each frozen slot's
 * frozen marker.
 */
export function getReforgeCacheGearKey(spec: EquipmentSpec, frozenItemSlots?: readonly ItemSlot[]): string {
	return buildGearKey(spec, frozenItemSlots);
}

function buildGearKey(spec: EquipmentSpec, frozenItemSlots?: readonly ItemSlot[]): string {
	const items = spec.items;
	const frozenSlots = frozenItemSlots ?? [];
	const frozenSlotMask = frozenSlots.length ? new Uint8Array(items.length) : undefined;
	if (frozenSlotMask) {
		for (let i = 0; i < frozenSlots.length; i++) {
			const slot = frozenSlots[i];
			if (slot >= 0 && slot < items.length) {
				frozenSlotMask[slot] = 1;
			}
		}
	}
	const itemKeys = new Array<string>(items.length);
	for (let slotIdx = 0; slotIdx < items.length; slotIdx++) {
		const item = items[slotIdx];
		if (!item?.id) {
			itemKeys[slotIdx] = '';
			continue;
		}

		const itemSlot = slotIdx as ItemSlot;
		const isFrozen = !!frozenSlotMask?.[itemSlot];
		const reforgeFingerprint = 0;
		itemKeys[slotIdx] = [item.id, item.randomSuffix ?? 0, item.enchant ?? 0, reforgeFingerprint].join(':');
		// Frozen-ness has to travel with the item through the paired-slot normalization
		// below, or swapping two rings with one of them frozen collapses to a single key
		// while the optimizer (which freezes by slot index) must leave a different ring
		// alone in each case.
		if (isFrozen) {
			itemKeys[slotIdx] += ':frozen';
		}
	}

	const reorderPairedSlots = (firstSlot: ItemSlot, secondSlot: ItemSlot): void => {
		if (itemKeys[firstSlot] > itemKeys[secondSlot]) {
			swap(itemKeys, firstSlot, secondSlot);
		}
	};

	// Normalize interchangeable slots so equivalent gear layouts share a cache key.
	reorderPairedSlots(ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2);
	reorderPairedSlots(ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2);

	return itemKeys.join('|');
}
