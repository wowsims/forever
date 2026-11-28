import i18n from '@i18n/config';

import { useBulkMode } from '../../hooks/useBulkMode';
import { useBulkResults } from '../../hooks/useBulkResults';
import { useBulkState } from '../../hooks/useBulkState';
import { bulkModeStarted } from '../../model/mode';
import { BulkRankingList } from '../BulkRankingList';
import { BulkResultRow } from './BulkResultRow';

export interface BulkResultsProps {
	onActivateTab: (id: string) => void;
	/** Index of the finished encounter whose ranking to show. */
	selected: number;
}

export const BulkResults = ({ onActivateTab, selected }: BulkResultsProps) => {
	const { runMode } = useBulkMode();
	const results = useBulkResults();
	const started = useBulkState(slice => bulkModeStarted(runMode, slice));

	if (!results) {
		// Starting a run empties the pane, so the invitation to run one does not come back.
		return started ? null : <div className="flex items-center justify-center p-6">{i18n.t('bulk_tab.results.run_simulation')}</div>;
	}

	const shown = results[selected].results;
	return (
		<BulkRankingList
			key={`${runMode}:${selected}`}
			chains={shown.chains}
			renderRow={result => <BulkResultRow result={result} results={shown} onActivateTab={onActivateTab} />}
		/>
	);
};
