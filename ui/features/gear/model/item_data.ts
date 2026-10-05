import { ItemRandomSuffix } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import { translateProtoStatName } from '@i18n/localization';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { EquippedItem } from '@sim/proto/equipped_item';

import { GearData, ItemData } from '../types';

export const itemsTabData = (gearData: GearData, items: Item[]): ItemData<Item, string>[] =>
	items.map(item => {
		const equippedItem = new EquippedItem({ item });
		return {
			item: item,
			id: item.id,
			actionId: equippedItem.asActionId(),
			ilvl: item.scalingOptions?.[0].ilvl || item.ilvl,
			name: item.name,
			searchText: item.name,
			quality: item.quality,
			nameDescription: item.nameDescription,
			phase: item.phase,
			ignoreEPFilter: false,
			onEquip: item => {
				const equippedItem = gearData.getEquippedItem();
				if (equippedItem) {
					gearData.equipItem(equippedItem.withItem(item));
				} else {
					gearData.equipItem(new EquippedItem({ item }));
				}
			},
		};
	});

export const byLvlDescending = (itemA: Enchant, itemB: Enchant) => itemB.lvl - itemA.lvl || itemB.effectId - itemA.effectId;

export const enchantActionId = (enchant: Enchant): ActionId => (enchant.itemId ? ActionId.fromItemId(enchant.itemId) : ActionId.fromSpellId(enchant.spellId));

const enchantRow = (
	enchant: Enchant,
	equip: (equippedItem: EquippedItem, enchant: Enchant) => EquippedItem,
	gearData: GearData,
): ItemData<Enchant, string> => ({
	item: enchant,
	id: enchant.effectId,
	actionId: enchantActionId(enchant),
	ilvl: enchant.lvl,
	name: enchant.name,
	searchText: enchant.name,
	quality: enchant.quality,
	phase: enchant.phase || 1,
	ignoreEPFilter: true,
	nameDescription: '',
	onEquip: enchant => {
		const equippedItem = gearData.getEquippedItem();
		if (equippedItem) gearData.equipItem(equip(equippedItem, enchant));
	},
});

export const enchantsTabData = (gearData: GearData, enchants: Enchant[]): ItemData<Enchant, string>[] =>
	enchants.sort(byLvlDescending).map(enchant => enchantRow(enchant, (equippedItem, value) => equippedItem.withEnchant(value), gearData));

export const randomSuffixesTabData = (player: Player<any>, gearData: GearData, equippedItem: EquippedItem): ItemData<ItemRandomSuffix, string>[] => {
	const itemProto = equippedItem.item;
	return player.getRandomSuffixes(itemProto).map((randomSuffix: ItemRandomSuffix) => {
		const label = translateProtoStatName(randomSuffix.name);

		return {
			item: randomSuffix,
			id: randomSuffix.id,
			actionId: ActionId.fromRandomSuffix(itemProto, randomSuffix),
			name: label,
			searchText: label,
			quality: itemProto.quality,
			phase: itemProto.phase,
			nameDescription: '',
			ignoreEPFilter: true,
			onEquip: randomSuffix => {
				const equippedItem = gearData.getEquippedItem();
				if (equippedItem) {
					gearData.equipItem(equippedItem.withItem(equippedItem.item).withRandomSuffix(randomSuffix));
				}
			},
		};
	});
};
