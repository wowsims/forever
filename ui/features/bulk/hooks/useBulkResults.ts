import { bulkModeResults } from '../model/mode';
import { useBulkMode } from './useBulkMode';
import { useBulkState } from './useBulkState';

export const useBulkResults = () => {
	const { runMode } = useBulkMode();
	return useBulkState(slice => bulkModeResults(runMode, slice));
};
