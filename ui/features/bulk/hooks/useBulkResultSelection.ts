import type { BulkEncounterResults } from '@sim/bulk/types';
import { useState } from 'react';

import { useBulkResults } from './useBulkResults';

// The encounter shown in the Results pane, shared by the pane and the encounter sidebar. It is tied to the result list
// it was picked in, so a new run, or a list it was not picked in, starts on its first encounter.
export const useBulkResultSelection = (): [selected: number, select: (index: number) => void] => {
	const results = useBulkResults();
	const [selection, setSelection] = useState<{ of: readonly BulkEncounterResults[] | null; index: number }>({ of: null, index: 0 });
	const selected = results && selection.of === results ? selection.index : 0;
	return [selected, index => setSelection({ of: results, index })];
};
