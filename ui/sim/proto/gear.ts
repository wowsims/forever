import { ConsumesSpec, EquipmentSpec, HandType, ItemSlot, ItemSpec, Profession, WeaponType } from '@generated/proto/common';
import { ItemEffectRandPropPoints, SimDatabase, SimEnchant, SimItem } from '@generated/proto/db';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import { distinct, equalsOrBothNull, getEnumValues } from '../utils/collections';
import { sum } from '../utils/math';
import { Database } from './database';
import { EquippedItem } from './equipped_item';
import { isBluntWeaponType, isSharpWeaponType, validWeaponCombo } from './items';
import { adjustImbues, adjustSwapImbues } from './utils';

type InternalGear = Record<ItemSlot, EquippedItem | null>;

abstract class BaseGear {
	protected readonly gear: InternalGear;

	constructor(gear: Partial<InternalGear>) {
		this.getItemSlots().forEach(slot => {
			if (!gear[slot as ItemSlot]) gear[slot as ItemSlot] = null;
		});
		this.gear = gear as InternalGear;
	}

	abstract getItemSlots(): ItemSlot[];

	equals(other: BaseGear): boolean {
		const otherArray = other.asArray();
		return this.asArray().every((thisItem, slot) => equalsOrBothNull(thisItem, otherArray[slot], (a, b) => a.equals(b)));
	}

	getEquippedItem(slot: ItemSlot): EquippedItem | null {
		return this.gear[slot] || null;
	}

	getEquippedItems(): Array<EquippedItem | null> {
		return Object.values(this.gear);
	}

	asArray(): Array<EquippedItem | null> {
		return Object.values(this.gear);
	}

	asMap(): Partial<InternalGear> {
		const newInternalGear: Partial<InternalGear> = {};
		this.getItemSlots()
			.map(slot => Number(slot) as ItemSlot)
			.forEach(slot => {
				newInternalGear[slot] = this.getEquippedItem(slot);
			});
		return newInternalGear;
	}

	getTrinkets(): Array<EquippedItem | null> {
		return [this.getEquippedItem(ItemSlot.ItemSlotTrinket1), this.getEquippedItem(ItemSlot.ItemSlotTrinket2)];
	}

	hasTrinket(itemId: number): boolean {
		return this.getTrinkets()
			.map(t => t?.item.id)
			.includes(itemId);
	}

	hasTrinketFromOptions(itemIds: number[]): boolean {
		return this.getTrinkets()
			.filter((t): t is EquippedItem => !!t)
			.map(t => t.item.id)
			.some(id => itemIds.includes(id));
	}

	hasRelic(itemId: number): boolean {
		const relicItem = this.getEquippedItem(ItemSlot.ItemSlotRanged);

		if (!relicItem) {
			return false;
		}

		return relicItem!.item.id == itemId;
	}

	/**
	 * Returns a new Gear set with the item equipped.
	 *
	 * Checks for validity and removes/exchanges items as needed.
	 */
	protected withEquippedItemInternal(newSlot: ItemSlot, newItem: EquippedItem | null): Partial<InternalGear> {
		// Create a new identical set of gear
		const newInternalGear = this.asMap();

		if (newItem) {
			this.removeUniqueItems(newInternalGear, newItem);
		}

		// Actually assign the new item.
		newInternalGear[newSlot] = newItem;

		BaseGear.validateWeaponCombo(newInternalGear, newSlot);

		return newInternalGear;
	}

	private removeUniqueItems(gear: Partial<InternalGear>, newItem: EquippedItem) {
		if (newItem.item.unique || newItem.item.limitCategory) {
			this.getItemSlots()
				.map(slot => Number(slot) as ItemSlot)
				.forEach(slot => {
					if (
						(newItem.item.limitCategory && gear[slot]?.item.limitCategory == newItem.item.limitCategory) ||
						(newItem.item.unique && gear[slot]?.item.id == newItem.item.id)
					) {
						gear[slot] = null;
					}
				});
		}
	}

	private static validateWeaponCombo(gear: Partial<InternalGear>, newSlot: ItemSlot) {
		// Check for valid weapon combos.
		if (!validWeaponCombo(gear[ItemSlot.ItemSlotMainHand]?.item, gear[ItemSlot.ItemSlotOffHand]?.item)) {
			if (newSlot == ItemSlot.ItemSlotOffHand) {
				if (gear[ItemSlot.ItemSlotOffHand]?.item.handType == HandType.HandTypeTwoHand) {
					gear[ItemSlot.ItemSlotOffHand] = null;
				}

				gear[ItemSlot.ItemSlotMainHand] = null;
			} else {
				gear[ItemSlot.ItemSlotOffHand] = null;
			}
		}
	}

	toDatabase(db: Database): SimDatabase {
		const equippedItems = this.asArray().filter(ei => ei != null) as Array<EquippedItem>;
		const data = {
			items: distinct(equippedItems.map(ei => BaseGear.itemToDB(ei.item))),
			randomSuffixes: distinct(equippedItems.filter(ei => ei.randomSuffix).map(ei => ei.randomSuffix!)),
			itemEffectRandPropPoints: distinct(
				equippedItems.flatMap(ei => db.getItemEffectRandPropPoints(ei.ilvl)).filter((ieRpp): ieRpp is ItemEffectRandPropPoints => !!ieRpp),
			),
			enchants: distinct(
				equippedItems.flatMap(ei => {
					const out: ReturnType<typeof BaseGear.enchantToDB>[] = [];
					if (ei.enchant) out.push(BaseGear.enchantToDB(ei.enchant));
					return out;
				}),
			),
		};
		return SimDatabase.create(data);
	}

	static itemToDB(item: Item): SimItem {
		return SimItem.fromJson(Item.toJson(item), { ignoreUnknownFields: true });
	}

	private static enchantToDB(enchant: Enchant): SimEnchant {
		return SimEnchant.fromJson(Enchant.toJson(enchant), { ignoreUnknownFields: true });
	}
}

/**
 * Represents a full gear set, including items/enchants for every slot.
 *
 * This is an immutable type.
 */
export class Gear extends BaseGear {
	constructor(gear: Partial<InternalGear>) {
		super(gear);
	}

	getItemSlots(): ItemSlot[] {
		return getEnumValues(ItemSlot);
	}

	withEquippedItem(newSlot: ItemSlot, newItem: EquippedItem | null): Gear {
		return new Gear(this.withEquippedItemInternal(newSlot, newItem));
	}

	asSpec(): EquipmentSpec {
		return EquipmentSpec.create({
			items: this.asArray().map(ei => (ei ? ei.asSpec() : ItemSpec.create())),
		});
	}

	// Removes enchanting profession bonus. The rings are set in place: re-equipping one drops every other item of its
	// limit category, and gear loaded from a spec can hold two rings of one.
	withoutEnchanting(): Gear {
		const rings = [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2].filter(slot => this.getEquippedItem(slot)?.enchant);
		if (!rings.length) return this;
		const slots = this.asMap();
		rings.forEach(slot => (slots[slot] = slots[slot]!.withEnchant(null)));
		return new Gear(slots);
	}
	hasMHWeapon(): boolean {
		return this.getEquippedItem(ItemSlot.ItemSlotMainHand) != null;
	}
	hasOHWeapon(): boolean {
		const weapon = this.getEquippedItem(ItemSlot.ItemSlotOffHand);
		return (
			weapon != null &&
			![HandType.HandTypeOffHand].includes(weapon.item.handType) &&
			![WeaponType.WeaponTypeOffHand, WeaponType.WeaponTypeShield].includes(weapon.item.weaponType)
		);
	}
	hasBluntMHWeapon(): boolean {
		const weapon = this.getEquippedItem(ItemSlot.ItemSlotMainHand);
		return weapon != null && isBluntWeaponType(weapon.item.weaponType);
	}
	hasSharpMHWeapon(): boolean {
		const weapon = this.getEquippedItem(ItemSlot.ItemSlotMainHand);
		return weapon != null && isSharpWeaponType(weapon.item.weaponType);
	}
	hasBluntOHWeapon(): boolean {
		const weapon = this.getEquippedItem(ItemSlot.ItemSlotOffHand);
		return weapon != null && isBluntWeaponType(weapon.item.weaponType);
	}
	hasSharpOHWeapon(): boolean {
		const weapon = this.getEquippedItem(ItemSlot.ItemSlotOffHand);
		return weapon != null && isSharpWeaponType(weapon.item.weaponType);
	}
	adjustImbues(consumes: ConsumesSpec): ConsumesSpec {
		return adjustImbues(consumes, this.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item, this.getEquippedItem(ItemSlot.ItemSlotOffHand)?.item);
	}

	getProfessionRequirements(): Array<Profession> {
		return distinct((this.asArray().filter(ei => ei != null) as Array<EquippedItem>).map(ei => ei.getProfessionRequirements()).flat());
	}
	getFailedProfessionRequirements(professions: Array<Profession>): Array<Item | Enchant> {
		return (this.asArray().filter(ei => ei != null) as Array<EquippedItem>).map(ei => ei.getFailedProfessionRequirements(professions)).flat();
	}
	getItemSetCount(name: string, alternativeName?: string) {
		if (!name) return 0;

		const validSetItemSlots = [ItemSlot.ItemSlotHead, ItemSlot.ItemSlotShoulder, ItemSlot.ItemSlotChest, ItemSlot.ItemSlotHands, ItemSlot.ItemSlotLegs];
		let setItemCount = 0;
		for (const slot of validSetItemSlots) {
			const item = this.getEquippedItem(slot);
			if (item?.item?.setName === name || (alternativeName && item?.item.setName === alternativeName)) setItemCount++;
		}

		return setItemCount;
	}

	getAverageItemLevel(): number {
		const items = this.getEquippedItems();
		let itemSlotsToCount = this.getItemSlots().length;
		let has2H = items?.[ItemSlot.ItemSlotMainHand]?.item?.handType === HandType.HandTypeTwoHand;

		// If user cannot dual wield 2H, then only count the main hand item as there is no off hand
		if (has2H) itemSlotsToCount -= 1;

		const totalIlvl = sum(items.filter((item): item is EquippedItem => item != null).map(item => item.ilvl));
		return totalIlvl / itemSlotsToCount;
	}
}

/**
 * Represents a item swap gear set, including items/enchants/bonusStats.
 *
 * This is an immutable type.
 */
export class ItemSwapGear extends BaseGear {
	constructor(gear: Partial<InternalGear>) {
		super(gear);
	}

	getItemSlots(): ItemSlot[] {
		return getEnumValues(ItemSlot);
	}

	withEquippedItem(newSlot: ItemSlot, newItem: EquippedItem | null): ItemSwapGear {
		return new ItemSwapGear(this.withEquippedItemInternal(newSlot, newItem));
	}

	adjustSwapImbues(consumes: ConsumesSpec): ConsumesSpec {
		return adjustSwapImbues(consumes, this.getEquippedItem(ItemSlot.ItemSlotMainHand)?.item, this.getEquippedItem(ItemSlot.ItemSlotOffHand)?.item);
	}
}
