import { type AreaType, ItemRandomSuffix, ItemSlot, ItemSpec, Profession, PseudoStat, ScalingItemProperties } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import { distinct } from '../utils/collections';
import { ActionId } from './action_id';
import { enchantAppliesToItem } from './items';
import { Stats } from './stats';

export const getWeaponDPS = (item: Item, upgradeStep: 0): number => {
	const { weaponDamageMin, weaponDamageMax } = item.scalingOptions[upgradeStep];
	return (weaponDamageMin + weaponDamageMax) / 2 / (item.weaponSpeed || 1);
};

export const getWeaponDpsStatsBySlot = (weaponDps: number, slot: ItemSlot): Stats => {
	const stats = new Stats();
	if (slot === ItemSlot.ItemSlotMainHand) {
		return stats.withPseudoStat(PseudoStat.PseudoStatMainHandDps, weaponDps);
	} else if (slot === ItemSlot.ItemSlotOffHand) {
		return stats.withPseudoStat(PseudoStat.PseudoStatOffHandDps, weaponDps);
	} else if (slot === ItemSlot.ItemSlotRanged) {
		return stats.withPseudoStat(PseudoStat.PseudoStatRangedDps, weaponDps);
	}
	return stats;
};

export const getWeaponStatsBySlot = (item: Item, slot: ItemSlot, upgradeStep: 0) => {
	if (item.weaponSpeed > 0) {
		return getWeaponDpsStatsBySlot(getWeaponDPS(item, upgradeStep), slot);
	}
	return new Stats();
};

type EquippedItemOptions = {
	item: Item;
	enchant?: Enchant | null;
	randomSuffix?: ItemRandomSuffix | null;
};

/**
 * Represents an equipped item along with the enchant attached to it.
 *
 * This is an immutable type.
 */
export class EquippedItem {
	readonly _item: Item;
	readonly _randomSuffix: ItemRandomSuffix | null;
	readonly _enchant: Enchant | null;

	constructor({ item, enchant, randomSuffix }: EquippedItemOptions) {
		this._item = item;
		this._enchant = enchant || null;
		this._randomSuffix = randomSuffix || null;
	}

	get item(): Item {
		// Make a defensive copy
		return Item.clone(this._item);
	}

	get id(): number {
		return this._item.id;
	}

	get randomSuffix(): ItemRandomSuffix | null {
		return this._randomSuffix ? ItemRandomSuffix.clone(this._randomSuffix) : null;
	}
	get enchant(): Enchant | null {
		// Make a defensive copy
		return this._enchant ? Enchant.clone(this._enchant) : null;
	}
	get randomPropPoints(): number {
		return this.item.scalingOptions[0].randPropPoints || this.item.randPropPoints;
	}
	get ilvl(): number {
		return this.item.scalingOptions[0].ilvl;
	}

	getBaseScalingItemProperties(): ScalingItemProperties {
		return this._item.scalingOptions[0];
	}

	equals(other: EquippedItem, ignoreEnchants?: boolean) {
		if (this.id != other.id) return false;
		if (!Item.equals(this._item, other.item)) return false;

		if ((this._randomSuffix == null) != (other.randomSuffix == null)) return false;

		if (this._randomSuffix && other.randomSuffix && !ItemRandomSuffix.equals(this._randomSuffix, other.randomSuffix)) return false;

		if ((this._enchant == null) != (other.enchant == null) && !ignoreEnchants) return false;

		if (this._enchant && other.enchant && !Enchant.equals(this._enchant, other.enchant) && !ignoreEnchants) return false;

		return true;
	}

	/**
	 * Replaces the item and tries to keep the existing enchant if possible.
	 */
	withItem(item: Item): EquippedItem {
		let newEnchant = null;
		if (this._enchant && enchantAppliesToItem(this._enchant, item)) newEnchant = this._enchant;

		return new EquippedItem({
			item,
			enchant: newEnchant,
		});
	}

	/**
	 * Returns a new EquippedItem with the given enchant applied.
	 */
	withEnchant(enchant: Enchant | null): EquippedItem {
		return new EquippedItem({
			item: this._item,
			enchant,
			randomSuffix: this._randomSuffix,
		});
	}

	withRandomSuffix(randomSuffix: ItemRandomSuffix | null): EquippedItem {
		return new EquippedItem({
			item: this._item,
			enchant: this._enchant,
			randomSuffix,
		});
	}

	// Folds in the stats of the given areas, as the sim does.
	withDynamicStats(areaTypes: readonly AreaType[] = []) {
		const item = this.item;
		const scalingOptions = item.scalingOptions[0];

		item.stats = new Stats().asProtoArray().map((_, index) => scalingOptions.stats[index] || 0);
		item.weaponDamageMin = scalingOptions.weaponDamageMin;
		item.weaponDamageMax = scalingOptions.weaponDamageMax;
		item.randPropPoints = scalingOptions.randPropPoints;

		if (this._randomSuffix) {
			item.stats = item.stats.map((stat, index) =>
				this._randomSuffix!.stats[index] > 0 ? Math.floor((this._randomSuffix!.stats[index] * item.randPropPoints) / 10000) : stat,
			);
		}

		for (const areaStats of scalingOptions.areaStats.filter(bonus => areaTypes.includes(bonus.areaType))) {
			item.stats = item.stats.map((stat, index) => stat + (areaStats.stats[index] || 0));
		}

		return new EquippedItem({
			item,
			enchant: this._enchant,
			randomSuffix: this._randomSuffix,
		});
	}

	// Returns calculated stats
	// if slot is provided it will include slot specific stats like weapon DPS
	calcStats(slot?: ItemSlot): Stats {
		const item = this.item;
		let stats = new Stats(item.stats, item.pseudoStats);
		if (typeof slot === 'number') stats = stats.add(getWeaponStatsBySlot(item, slot, 0));
		return stats;
	}

	asActionId(): ActionId {
		return ActionId.fromItemId(this._item.id, undefined, this._randomSuffix?.id || undefined);
	}

	asSpec(): ItemSpec {
		return ItemSpec.create({
			id: this._item.id,
			randomSuffix: this._randomSuffix?.id,
			enchant: this._enchant?.effectId,
		});
	}

	hasRandomSuffixOptions(): boolean {
		return !!this._item.randomSuffixOptions.length;
	}

	getProfessionRequirements(): Array<Profession> {
		const profs: Array<Profession> = [];
		if (this._item.requiredProfession != Profession.ProfessionUnknown) {
			profs.push(this._item.requiredProfession);
		}
		if (this._enchant != null && this._enchant.requiredProfession != Profession.ProfessionUnknown) {
			profs.push(this._enchant.requiredProfession);
		}
		return distinct(profs);
	}
	getFailedProfessionRequirements(professions: Array<Profession>): Array<Item | Enchant> {
		const failed: Array<Item | Enchant> = [];
		if (this._item.requiredProfession != Profession.ProfessionUnknown && !professions.includes(this._item.requiredProfession)) {
			failed.push(this._item);
		}
		if (
			this._enchant != null &&
			this._enchant.requiredProfession != Profession.ProfessionUnknown &&
			!professions.includes(this._enchant.requiredProfession)
		) {
			failed.push(this._enchant);
		}
		return failed;
	}
}
