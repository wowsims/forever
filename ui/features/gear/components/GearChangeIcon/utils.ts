import type { EquippedItem } from '@sim/proto/equipped_item';

export const enchantChanged = (item: EquippedItem | undefined, previousItem: EquippedItem | undefined): boolean =>
	!!item?.enchant && !!previousItem && item.enchant.effectId !== previousItem.enchant?.effectId;
