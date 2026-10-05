import { ItemQuality, ItemRandomSuffix } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { EquippedItem } from '@sim/proto/equipped_item';
import { StoreSubscribe } from '@sim/state/subscriptions';

export type ItemListType = Item | Enchant | ItemRandomSuffix;

// Everything a row carries except its rendered name, so the helpers that sort, search and favourite
// rows take this rather than a rendered row.
export interface ItemDataFields<T extends ItemListType> {
	item: T;
	searchText: string;
	id: number;
	actionId: ActionId;
	quality: ItemQuality;
	phase: number;
	ilvl?: number;
	ignoreEPFilter: boolean;
	nameDescription: string;
	onEquip: (item: T) => void;
}

export interface ItemData<T extends ItemListType, TName = string | HTMLElement> extends ItemDataFields<T> {
	name: TName;
}

export interface GearData {
	equipItem: (equippedItem: EquippedItem | null) => void;
	getEquippedItem: () => EquippedItem | null;
	// Fires when the equipped item for this slot changes.
	subscribe: StoreSubscribe;
	// Each pick adds an item rather than replacing one, so the picker offers items only, with nothing to remove.
	addOnly?: boolean;
}

export enum SelectorModalTabs {
	Items = 'Items',
	RandomSuffixes = 'Random Suffix',
	Enchants = 'Enchants',
}

// Helper function to get translated tab labels
export function getTranslatedTabLabel(tab: SelectorModalTabs): string {
	switch (tab) {
		case SelectorModalTabs.Items:
			return i18n.t('gear_tab.gear_picker.tabs.items');
		case SelectorModalTabs.RandomSuffixes:
			return i18n.t('gear_tab.gear_picker.tabs.random_suffix');
		case SelectorModalTabs.Enchants:
			return i18n.t('gear_tab.gear_picker.tabs.enchants');
		default:
			return tab;
	}
}
