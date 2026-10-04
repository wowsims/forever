import type { ItemSlot } from '@generated/proto/common';
import { GemColor } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { gemMatchesSocket } from '@sim/proto/gems';
import { Stats } from '@sim/proto/stats';
import type { ReactNode } from 'react';

import { enchantsTabData, gemsTabData, itemsTabData, randomSuffixesTabData } from '../../model/item_data';
import type { TabEligibility } from '../../model/tab_eligibility';
import type { GearData, ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';

export interface SelectorTab {
	label: SelectorModalTabs;
	socketColor: GemColor;
	socketIdx?: number;
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
	socketColor?: GemColor;
	socketIdx?: number;
	itemData: ItemData<T, ReactNode>[];
	computeEP: (item: T) => number;
	equippedToItem: (equippedItem: EquippedItem | null) => T | null | undefined;
	onRemove?: () => void;
}

// A tab with no rows is not built. The cast is necessary because every tab is uniform apart from
// the type its rows carry.
const describe = <T extends ItemListType>(spec: TabSpec<T>): SelectorTab | null =>
	spec.itemData.length ? ({ socketColor: GemColor.GemColorUnknown, ...spec } as unknown as SelectorTab) : null;

export const eligibilityFor = ({ player, slot, equippedItem }: Omit<SelectorTabsOptions, 'gearData'>): TabEligibility => ({
	hasEnchants: !!player.getEnchants(slot).length,
	socketCount: equippedItem?.numSockets(),
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
		...gemTabs({ player, slot, gearData, equippedItem }),
	];

	return tabs.filter((tab): tab is SelectorTab => !!tab);
};

const gemTabs = ({ player, gearData, equippedItem }: SelectorTabsOptions): Array<SelectorTab | null> => {
	if (!equippedItem) return [];

	const socketBonusEP = player.computeStatsEP(new Stats(equippedItem.item.socketBonus)) / (equippedItem.item.gemSockets.length || 1);
	return equippedItem.curSocketColors().map((socketColor, socketIdx) =>
		describe({
			label: SelectorModalTabs[`Gem${socketIdx + 1}` as keyof typeof SelectorModalTabs],
			socketColor,
			socketIdx,
			itemData: gemsTabData(gearData, player.getGems(socketColor), socketIdx),
			computeEP: gem => player.computeGemEP(gem) + (gemMatchesSocket(gem, socketColor) ? socketBonusEP : 0),
			equippedToItem: item => item?.gems[socketIdx],
			onRemove: () => {
				const current = gearData.getEquippedItem();
				if (current) gearData.equipItem(current.withGem(null, socketIdx));
			},
		}),
	);
};

export const removeButtonLabel = (label: SelectorModalTabs, translate: (key: string) => string): string => {
	switch (label) {
		case SelectorModalTabs.Enchants:
			return translate('gear_tab.gear_picker.remove_buttons.remove_enchant');
		case SelectorModalTabs.RandomSuffixes:
			return translate('gear_tab.gear_picker.remove_buttons.remove_random_suffix');
		case SelectorModalTabs.Gem1:
		case SelectorModalTabs.Gem2:
		case SelectorModalTabs.Gem3:
			return translate('gear_tab.gear_picker.remove_buttons.remove_gem');
		default:
			return translate('gear_tab.gear_picker.unequip_item');
	}
};

export const columnHeaderLabel = (label: SelectorModalTabs, translate: (tab: SelectorModalTabs) => string): string => {
	if ([SelectorModalTabs.Gem1, SelectorModalTabs.Gem2, SelectorModalTabs.Gem3].includes(label)) return translate(SelectorModalTabs.Gem1);
	if ([SelectorModalTabs.Items, SelectorModalTabs.Enchants].includes(label)) return translate(label);
	return '';
};
