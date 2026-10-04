import i18n from '@i18n/config';

import { useBulkState } from '../../hooks/useBulkState';
import { BulkResultRow } from './BulkResultRow';
import { BulkResultTieGroup } from './BulkResultTieGroup';

export interface BulkResultsProps {
	onActivateTab: (id: string) => void;
}

export const BulkResults = ({ onActivateTab }: BulkResultsProps) => {
	const results = useBulkState(slice => slice.results);
	const started = useBulkState(slice => slice.started);

	if (!results) {
		// Starting a run empties the pane, so the invitation to run one does not come back.
		return started ? null : <div className="flex items-center justify-center p-6">{i18n.t('bulk_tab.results.run_simulation')}</div>;
	}

	const iterations = results.iterations;
	return (
		<>
			{results.chains.map((chain, chainIdx) =>
				chain.length > 1 ? (
					<BulkResultTieGroup key={chainIdx}>
						{chain.map((result, idx) => (
							<BulkResultRow
								key={idx}
								result={result}
								baseResult={results.originalGearResults}
								iterations={iterations}
								onActivateTab={onActivateTab}
							/>
						))}
					</BulkResultTieGroup>
				) : (
					<BulkResultRow
						key={chainIdx}
						result={chain[0]}
						baseResult={results.originalGearResults}
						iterations={iterations}
						onActivateTab={onActivateTab}
					/>
				),
			)}
		</>
	);
};
