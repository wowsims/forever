import { ItemRandomSuffix } from '@generated/proto/common';
import { UIEnchant as Enchant, UIItem as Item } from '@generated/proto/ui';

import { ItemListType, SelectorModalTabs } from '../types';

export const getItemIdByItemType = (tab: SelectorModalTabs, item: ItemListType | null | undefined): number | null | undefined => {
	switch (tab) {
		case SelectorModalTabs.Enchants:
			return (item as Enchant)?.effectId;
		case SelectorModalTabs.Items:
		case SelectorModalTabs.RandomSuffixes:
			return (item as Item | ItemRandomSuffix)?.id;
		default:
			return null;
	}
};
