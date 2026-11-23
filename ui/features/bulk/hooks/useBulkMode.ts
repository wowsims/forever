import { useUiStore } from '@sim/hooks/useUiStore';

import { bulkRunMode, bulkRunsBuilds, bulkRunsDrops } from '../model/mode';
import { useBulkState } from './useBulkState';

export const useBulkMode = () => {
	const powerTools = useUiStore('showPowerTools');
	const mode = useBulkState(slice => slice.mode);
	return { powerTools, mode, runMode: bulkRunMode(powerTools, mode), builds: bulkRunsBuilds(powerTools, mode), drops: bulkRunsDrops(powerTools, mode) };
};
