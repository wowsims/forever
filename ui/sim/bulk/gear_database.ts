// Building the trimmed SimDatabase a bulk run ships to the workers.
import { BulkSettings } from '@generated/proto/api';
import { ItemRandomSuffix } from '@generated/proto/common';
import { ItemEffectRandPropPoints, SimDatabase, SimEnchant, SimItem } from '@generated/proto/db';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import { Database } from '../proto/database';
import { EquippedItem } from '../proto/equipped_item';
import { Gear } from '../proto/gear';

const enchantRowKey = (enchant: Enchant) => `${enchant.effectId}-${enchant.itemId}-${enchant.spellId}`;

// An effect id can carry several enchant rows of different types, and the sim judges fit by any of them, so every row of a shipped effect id ships.
export const makeBulkGearDatabase = (db: Database, gearSets: Gear[], extraItems: EquippedItem[] = [], extraEnchantIds: readonly number[] = []): SimDatabase => {
	const items = new Map<number, Item>();
	const randomSuffixes = new Map<number, ItemRandomSuffix>();
	const itemEffectRandPropPoints = new Map<number, ItemEffectRandPropPoints>();
	const enchants = new Map<string, Enchant>();
	const enchantIds = new Set(extraEnchantIds);

	const addEquippedItem = (equippedItem: EquippedItem) => {
		const item = equippedItem.item;
		items.set(item.id, item);

		const randomSuffix = equippedItem.randomSuffix;
		if (randomSuffix) randomSuffixes.set(randomSuffix.id, randomSuffix);

		const scalingIlvls = new Set([equippedItem.ilvl]);
		Object.values(item.scalingOptions ?? {}).forEach(opt => {
			if (opt?.ilvl) scalingIlvls.add(opt.ilvl);
		});
		scalingIlvls.forEach(ilvl => {
			const rpp = db.getItemEffectRandPropPoints(ilvl);
			if (rpp) itemEffectRandPropPoints.set(rpp.ilvl, rpp);
		});

		const enchant = equippedItem.enchant;
		if (enchant) {
			enchants.set(enchantRowKey(enchant), enchant);
			enchantIds.add(enchant.effectId);
		}
	};

	for (const gearSet of gearSets) {
		for (const equippedItem of gearSet.asArray()) {
			if (equippedItem) addEquippedItem(equippedItem);
		}
	}
	for (const equippedItem of extraItems) {
		addEquippedItem(equippedItem);
	}
	for (const effectId of enchantIds) {
		for (const enchant of db.getEnchantsByEffectId(effectId)) enchants.set(enchantRowKey(enchant), enchant);
	}

	return SimDatabase.create({
		items: Array.from(items.values()).map(item => SimItem.fromJson(Item.toJson(item), { ignoreUnknownFields: true })),
		randomSuffixes: Array.from(randomSuffixes.values()),
		itemEffectRandPropPoints: Array.from(itemEffectRandPropPoints.values()),
		enchants: Array.from(enchants.values()).map(enchant => SimEnchant.fromJson(Enchant.toJson(enchant), { ignoreUnknownFields: true })),
	});
};

export const makeBulkItemDatabaseFromSpecs = (db: Database, baselineGear: Gear, settings: BulkSettings): SimDatabase => {
	const extraItems = settings.items.map(itemSpec => (itemSpec ? db.lookupItemSpec(itemSpec) : null)).filter((item): item is EquippedItem => item != null);
	return makeBulkGearDatabase(
		db,
		[baselineGear],
		extraItems,
		settings.enchants.flatMap(options => options.effectIds),
	);
};
