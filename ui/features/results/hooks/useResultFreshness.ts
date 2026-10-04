import { useSim } from '@sim/context/SimHostContext';
import { useSyncExternalStore } from 'react';

import { type ResultFreshness, resultFreshnessTracker } from '../model/result_freshness';

/** Null before an individual sim result exists; 'stale' once an input it was simmed with has changed since its run started. */
export const useResultFreshness = (): ResultFreshness => {
	const tracker = resultFreshnessTracker(useSim().store);
	return useSyncExternalStore(tracker.subscribe, tracker.getSnapshot);
};
