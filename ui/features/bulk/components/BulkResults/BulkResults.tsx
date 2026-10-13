import i18n from '@i18n/config';
import type { BulkEncounterResults, BulkRunResults } from '@sim/bulk/types';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { BulkResultRow } from './BulkResultRow';
import { BulkResultTieGroup } from './BulkResultTieGroup';

export interface BulkEncounterPillsProps {
	entries: ReadonlyArray<Pick<BulkEncounterResults, 'name' | 'isCurrent' | 'durationLabel'>>;
	selected: number;
	onSelect: (index: number) => void;
}

export const BulkEncounterPills = ({ entries, selected, onSelect }: BulkEncounterPillsProps) => (
	<div className="mb-4 flex flex-wrap gap-2" role="group" aria-label={i18n.t('bulk_tab.results.encounters')} data-testid="bulk-encounter-pills">
		{entries.map((entry, index) => (
			<Button
				key={index}
				size="sm"
				className="rounded-full"
				variant={index === selected ? 'primary' : 'outline-primary'}
				aria-pressed={index === selected}
				onClick={() => onSelect(index)}>
				{entry.isCurrent ? i18n.t('bulk_tab.results.current') : entry.name} <span className="ml-1 text-xs opacity-75">{entry.durationLabel}</span>
			</Button>
		))}
	</div>
);

export interface BulkResultsProps {
	onActivateTab: (id: string) => void;
}

export const BulkResults = ({ onActivateTab }: BulkResultsProps) => {
	const results = useBulkState(slice => slice.results);
	const started = useBulkState(slice => slice.started);
	// Tied to the result list it was picked in, so a new run starts on its first encounter again.
	const [selection, setSelection] = useState<{ of: BulkRunResults | null; index: number }>({ of: null, index: 0 });

	if (!results) {
		// Starting a run empties the pane, so the invitation to run one does not come back.
		return started ? null : <div className="flex items-center justify-center p-6">{i18n.t('bulk_tab.results.run_simulation')}</div>;
	}

	const index = selection.of === results ? selection.index : 0;
	const { results: shown } = results.entries[index];
	const iterations = shown.iterations;
	return (
		<>
			{results.plannedEncounters > 1 && (
				<BulkEncounterPills entries={results.entries} selected={index} onSelect={next => setSelection({ of: results, index: next })} />
			)}
			{shown.chains.map((chain, chainIdx) =>
				chain.length > 1 ? (
					<BulkResultTieGroup key={chainIdx}>
						{chain.map((result, idx) => (
							<BulkResultRow
								key={idx}
								result={result}
								baseResult={shown.originalGearResults}
								iterations={iterations}
								onActivateTab={onActivateTab}
							/>
						))}
					</BulkResultTieGroup>
				) : (
					<BulkResultRow
						key={chainIdx}
						result={chain[0]}
						baseResult={shown.originalGearResults}
						iterations={iterations}
						onActivateTab={onActivateTab}
					/>
				),
			)}
		</>
	);
};
