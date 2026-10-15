import type { BulkRunResults } from '@sim/bulk/types';
import { useState } from 'react';

import { useBulkState } from './useBulkState';

// The encounter shown in the Results pane, shared by the pane and the encounter sidebar. It is tied to the result list
// it was picked in, so a new run starts on its first encounter again.
export const useBulkResultSelection = (): [selected: number, select: (index: number) => void] => {
	const results = useBulkState(slice => slice.results);
	const [selection, setSelection] = useState<{ of: BulkRunResults | null; index: number }>({ of: null, index: 0 });
	const selected = results && selection.of === results ? selection.index : 0;
	return [selected, index => setSelection({ of: results, index })];
};
