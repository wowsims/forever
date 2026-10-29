import { useMemo } from 'react';

import { bulkEnchantPickEntries } from '../model/enchant_picks';
import { useBulkState } from './useBulkState';

export const useBulkEnchantPicks = () => {
	const presets = useBulkState(slice => slice.enchantPresets);
	const checked = useBulkState(slice => slice.buildEnchantPresets);

	const entries = useMemo(() => bulkEnchantPickEntries(presets), [presets]);

	return { entries, checked };
};
