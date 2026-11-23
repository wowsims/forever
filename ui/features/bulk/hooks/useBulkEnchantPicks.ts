import { useMemo } from 'react';

import { bulkEnchantPickEntries, type BulkEnchantPickField } from '../model/builds/enchant_picks';
import { useBulkState } from './useBulkState';

export const useBulkEnchantPicks = (field: BulkEnchantPickField) => {
	const presets = useBulkState(slice => slice.enchantPresets);
	const checked = useBulkState(slice => slice[field]);

	const entries = useMemo(() => bulkEnchantPickEntries(presets), [presets]);

	return { entries, checked };
};
