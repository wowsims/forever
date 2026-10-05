import type { ItemSlot } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { ReactNode } from 'react';

import { enchantsTabData, itemsTabData, randomSuffixesTabData } from '../../model/item_data';
import type { TabEligibility } from '../../model/tab_eligibility';
import type { GearData, ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';

export interface SelectorTab {
	label: SelectorModalTabs;
	itemData: ItemData<ItemListType, ReactNode>[];
	computeEP: (item: ItemListType) => number;
	equippedToItem: (equippedItem: EquippedItem | null) => ItemListType | null | undefined;
	onRemove?: () => void;
}

export interface SelectorTabsOptions {
	player: Player<any>;
	slot: ItemSlot;
	gearData: GearData;
	equippedItem: EquippedItem | null;
}

interface TabSpec<T extends ItemListType> {
	label: SelectorModalTabs;
	itemData: ItemData<T, ReactNode>[];
	computeEP: (item: T) => number;
	equippedToItem: (equippedItem: EquippedItem | null) => T | null | undefined;
	onRemove?: () => void;
}

// A tab with no rows is not built. The cast is necessary because every tab is uniform apart from
// the type its rows carry.
const describe = <T extends ItemListType>(spec: TabSpec<T>): SelectorTab | null => (spec.itemData.length ? (spec as unknown as SelectorTab) : null);

export const eligibilityFor = ({ player, slot }: Pick<SelectorTabsOptions, 'player' | 'slot'>): TabEligibility => ({
	hasEnchants: !!player.getEnchants(slot).length,
});

export const buildSelectorTabs = ({ player, slot, gearData, equippedItem }: SelectorTabsOptions): SelectorTab[] => {
	// `equippedItem.item` clones the proto on every read, and `computeEP` runs once per sort
	// comparison, so the suffix scale, the weapon's speed and what its DPS is worth are read once here.
	const randPropPoints = equippedItem?.item.randPropPoints ?? 0;
	const weaponSpeed = equippedItem?.item.weaponSpeed ?? 0;
	const epPerWeaponDps = weaponSpeed ? player.computeWeaponDpsEP(slot) : 0;

	const items = describe({
		label: SelectorModalTabs.Items,
		itemData: itemsTabData(gearData, player.getItems(slot)),
		computeEP: item => player.computeItemEP(item, slot),
		equippedToItem: item => item?.item,
		onRemove: gearData.addOnly ? undefined : () => gearData.equipItem(null),
	});
	if (gearData.addOnly) return items ? [items] : [];

	const tabs: Array<SelectorTab | null> = [
		items,
		describe({
			label: SelectorModalTabs.Enchants,
			itemData: enchantsTabData(gearData, player.getEnchants(slot)),
			computeEP: enchant => player.computeEnchantEP(enchant, weaponSpeed, epPerWeaponDps),
			equippedToItem: item => item?.enchant,
			onRemove: () => {
				const current = gearData.getEquippedItem();
				if (current) gearData.equipItem(current.withEnchant(null));
			},
		}),
		equippedItem?.item.randomSuffixOptions.length
			? describe({
					label: SelectorModalTabs.RandomSuffixes,
					itemData: randomSuffixesTabData(player, gearData, equippedItem),
					computeEP: randomSuffix => (player.computeRandomSuffixEP(randomSuffix) * randPropPoints) / 10000,
					equippedToItem: item => item?.randomSuffix,
					onRemove: () => {
						const current = gearData.getEquippedItem();
						if (current) gearData.equipItem(current.withItem(current.item).withRandomSuffix(null));
					},
				})
			: null,
	];

	return tabs.filter((tab): tab is SelectorTab => !!tab);
};

export const removeButtonLabel = (label: SelectorModalTabs, translate: (key: string) => string): string => {
	switch (label) {
		case SelectorModalTabs.Enchants:
			return translate('gear_tab.gear_picker.remove_buttons.remove_enchant');
		case SelectorModalTabs.RandomSuffixes:
			return translate('gear_tab.gear_picker.remove_buttons.remove_random_suffix');
		default:
			return translate('gear_tab.gear_picker.unequip_item');
	}
};

export const columnHeaderLabel = (label: SelectorModalTabs, translate: (tab: SelectorModalTabs) => string): string => {
	if ([SelectorModalTabs.Items, SelectorModalTabs.Enchants].includes(label)) return translate(label);
	return '';
};
