import { useMemo } from 'react';

import { bulkEnchantPickEntries } from '../model/builds/enchant_picks';
import { useBulkState } from './useBulkState';

export const useDropsEnchantPicks = () => {
	const presets = useBulkState(slice => slice.enchantPresets);
	const checked = useBulkState(slice => slice.dropsEnchantPresets);

	const entries = useMemo(() => bulkEnchantPickEntries(presets), [presets]);

	return { entries, checked };
};
