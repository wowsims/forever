import i18n from '@i18n/config';
import type { BulkEncounterResults as BulkEncounterResultsEntry } from '@sim/bulk/types';
import { formatDeltaText } from '@sim/utils/format';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import type { ReactNode } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { encounterLabel, formatEncounterDuration } from '../../model/encounters';
import { formatBulkDps } from '../../model/format';

interface BulkEncounterResultRow {
	label: string;
	durationLabel: string;
	bestDps: string;
	/** The best set's gain over the current gear, signed, with its percentage; null when the best set is the current gear. */
	gain: string | null;
}

export interface BulkEncounterResultsViewProps {
	rows: readonly BulkEncounterResultRow[];
	selected: number;
	onSelect: (index: number) => void;
	/** Rendered under the rows, such as the CSV export. */
	footer?: ReactNode;
}

const t = (key: string) => i18n.t(`bulk_tab.results.${key}`);

export const BulkEncounterResultsView = ({ rows, selected, onSelect, footer }: BulkEncounterResultsViewProps) => (
	<TabPanelColumns.Right testId="bulk-encounter-results">
		<div className="sticky top-sim-header">
			<Box padding="md" className="grid gap-2 bg-background">
				<h6 className="mb-0">{t('encounters')}</h6>
				<p className="mb-0 text-xs">{t('encounters_caption')}</p>
				{rows.map((row, index) => (
					<Button
						key={index}
						size="sm"
						variant={index === selected ? 'secondary' : 'outline-secondary'}
						className="grid w-full grid-cols-1 text-left"
						aria-pressed={index === selected}
						onClick={() => onSelect(index)}>
						<span className="flex min-w-0 items-baseline gap-2">
							<span className="min-w-0 truncate" title={row.label}>
								{row.label}
							</span>{' '}
							<span className="shrink-0 text-xs opacity-75">{row.durationLabel}</span>
						</span>
						<span className="whitespace-nowrap tabular-nums">
							{row.bestDps}
							{row.gain !== null && (
								<>
									{' '}
									<span className="text-xs opacity-75">{row.gain}</span>
								</>
							)}
						</span>
					</Button>
				))}
				{footer}
			</Box>
		</div>
	</TabPanelColumns.Right>
);

const resultRow = (entry: BulkEncounterResultsEntry): BulkEncounterResultRow => {
	const { chains, originalGearResults } = entry.results;
	const best = chains[0]?.[0] ?? originalGearResults;
	const gain = best.dpsMetrics.avg - originalGearResults.dpsMetrics.avg;
	// Formatted as the result rows format a set's DPS and its gain, without the significance colour.
	return {
		label: encounterLabel(entry),
		durationLabel: formatEncounterDuration(entry),
		bestDps: formatBulkDps(best.dpsMetrics.avg),
		gain: gain === 0 ? null : formatDeltaText(originalGearResults.dpsMetrics.avg, best.dpsMetrics.avg, 2, undefined, true, true).text,
	};
};

export type BulkEncounterResultsProps = Pick<BulkEncounterResultsViewProps, 'selected' | 'onSelect' | 'footer'>;

// One row per finished encounter, the current one first; renders nothing before a run has finished one.
export const BulkEncounterResults = ({ selected, onSelect, footer }: BulkEncounterResultsProps) => {
	const results = useBulkState(slice => slice.results);
	if (!results) return null;
	return <BulkEncounterResultsView rows={results.map(resultRow)} selected={selected} onSelect={onSelect} footer={footer} />;
};
