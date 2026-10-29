import i18n from '@i18n/config';

import { useBulkState } from '../../hooks/useBulkState';
import { BulkResultRow } from './BulkResultRow';
import { BulkResultTieGroup } from './BulkResultTieGroup';

export interface BulkResultsProps {
	onActivateTab: (id: string) => void;
	/** Index of the finished encounter whose ranking to show. */
	selected: number;
}

export const BulkResults = ({ onActivateTab, selected }: BulkResultsProps) => {
	const results = useBulkState(slice => slice.results);
	const started = useBulkState(slice => slice.started);

	if (!results) {
		// Starting a run empties the pane, so the invitation to run one does not come back.
		return started ? null : <div className="flex items-center justify-center p-6">{i18n.t('bulk_tab.results.run_simulation')}</div>;
	}

	const shown = results[selected].results;
	return (
		<div>
			{shown.chains.map((chain, chainIdx) =>
				chain.length > 1 ? (
					<BulkResultTieGroup key={chainIdx}>
						{chain.map((result, idx) => (
							<BulkResultRow key={idx} result={result} results={shown} onActivateTab={onActivateTab} />
						))}
					</BulkResultTieGroup>
				) : (
					<BulkResultRow
						key={chainIdx}
						result={chain[0]}
						results={shown}
						beforeGroup={(shown.chains[chainIdx + 1]?.length ?? 0) > 1}
						onActivateTab={onActivateTab}
					/>
				),
			)}
		</div>
	);
};
