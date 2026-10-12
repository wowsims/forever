import { useSimHost } from '@sim/context/SimHostContext';
import { parseStringArray, useLocalStorageState } from '@ui-kit/hooks/useLocalStorageState';

const STORAGE_SUFFIX = '__bulkSettingsGroups';

export const BULK_SETTINGS_GROUP = {
	durations: 'durations',
	enchants: 'enchants',
	freezes: 'freezes',
} as const;

export const useBulkSettingsGroups = () => useLocalStorageState<string[]>(useSimHost().getStorageKey(STORAGE_SUFFIX), parseStringArray, () => []);
