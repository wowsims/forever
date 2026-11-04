import { useEffect, useState } from 'react';

import { resultKey, type SimResultData } from '../model/result_data';
import { useSimResult } from './useSimResult';

/**
 * The result a tab draws: the last one emitted while the tab was open, held while it is closed. A re-emit of the run
 * it holds under the same filter keeps the one held, so the tab does not rebuild what it already drew.
 */
export const useHeldSimResult = (active: boolean): SimResultData | null => {
	const live = useSimResult();
	const [shown, setShown] = useState<SimResultData | null>(null);
	useEffect(() => {
		if (!active || !live) return;
		setShown(prev => (resultKey(prev) === resultKey(live) ? prev : live));
	}, [active, live]);
	return shown;
};
