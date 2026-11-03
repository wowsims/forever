import { useUiStore } from '@sim/hooks/useUiStore';

import { bulkRunsBuilds } from '../model/mode';
import { useBulkState } from './useBulkState';

export const useBulkMode = () => {
	const powerTools = useUiStore('showPowerTools');
	const mode = useBulkState(slice => slice.mode);
	return { powerTools, mode, builds: bulkRunsBuilds(powerTools, mode) };
};
