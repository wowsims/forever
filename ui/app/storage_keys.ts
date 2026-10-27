import { LOCAL_STORAGE_PREFIX } from '@sim/constants/other';

// The storage keys ui/app owns; the others live in @sim/state/storage_keys and @sim/state/persistence.
export const PRESET_FILTER_STORAGE_KEY = '__presetFilters__';
// Without the spec prefix: the tour opens by itself once per browser, on whichever spec page comes first.
export const ONBOARDING_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_onboarding`;
