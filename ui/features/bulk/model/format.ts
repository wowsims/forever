import { translateProtoStatName } from '@i18n/localization';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { formatToNumber } from '@sim/utils/format';

// Always two decimals, so "776.30" lines up with "797.41" in a column of results.
export const formatBulkDps = (dps: number): string => formatToNumber(dps, { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// "Item Suffix (Enchant)", as the result rows name it.
export const formatBulkItemLabel = (equipped: EquippedItem): string => {
	const suffix = equipped.randomSuffix ? ` ${translateProtoStatName(equipped.randomSuffix.name)}` : '';
	const enchant = equipped.enchant ? ` (${equipped.enchant.name})` : '';
	return `${equipped.item.name}${suffix}${enchant}`;
};
