import { bulkModeResults } from '../model/mode';
import { useBulkMode } from './useBulkMode';
import { useBulkState } from './useBulkState';

export const useBulkResults = () => {
	const { builds } = useBulkMode();
	return useBulkState(slice => bulkModeResults(builds, slice));
};
