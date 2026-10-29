import i18n from '@i18n/config';
import type { BulkEncounterResults as BulkEncounterResultsEntry } from '@sim/bulk/types';
import { formatDeltaText } from '@sim/utils/format';
import { Box } from '@ui-kit/Box';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import type { ReactNode } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { encounterLabel, formatEncounterDuration } from '../../model/encounters';
import { formatBulkDps } from '../../model/format';

interface BulkEncounterResultRow {
	label: string;
	durationLabel: string;
	bestDps: string;
	/** The best set's gain over the current gear, signed, with its percentage; null when the best set is the current gear or was not compared with it. */
	gain: string | null;
}

export interface BulkEncounterResultsViewProps {
	rows: readonly BulkEncounterResultRow[];
	/** False when the batch did not sim the current setup, so the caption names no gain. */
	gains: boolean;
	selected: number;
	onSelect: (index: number) => void;
	/** Rendered under the rows, such as the CSV export. */
	footer?: ReactNode;
}

const t = (key: string) => i18n.t(`bulk_tab.results.${key}`);

export const BulkEncounterResultsView = ({ rows, gains, selected, onSelect, footer }: BulkEncounterResultsViewProps) => (
	<TabPanelColumns.Right testId="bulk-encounter-results">
		<div className="sticky top-sim-header">
			<Box padding="md" className="grid gap-2 bg-background">
				<h6 className="mb-0">{t('encounters')}</h6>
				<p className="mb-0 text-xs">{t(gains ? 'encounters_caption' : 'encounters_caption_best')}</p>
				<ToggleGroup
					tone="neutral"
					orientation="vertical"
					size="sm"
					aria-label={t('encounters')}
					value={String(selected)}
					onChange={next => onSelect(Number(next))}
					options={rows.map((row, index) => ({
						value: String(index),
						label: (
							<span className="grid w-full min-w-0 grid-cols-1">
								<span className="flex min-w-0 items-baseline gap-2">
									<span className="min-w-0 truncate" title={row.label}>
										{row.label}
									</span>{' '}
									<span className="ml-auto shrink-0 text-xs opacity-75">{row.durationLabel}</span>
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
							</span>
						),
					}))}
				/>
				{footer}
			</Box>
		</div>
	</TabPanelColumns.Right>
);

const resultRow = (entry: BulkEncounterResultsEntry): BulkEncounterResultRow => {
	const { chains, originalGearResults, referenceIsCurrent } = entry.results;
	const best = chains[0]?.[0] ?? originalGearResults;
	const gain = best.dpsMetrics.avg - originalGearResults.dpsMetrics.avg;
	// Formatted as the result rows format a set's DPS and its gain, without the significance colour.
	return {
		label: encounterLabel(entry),
		durationLabel: formatEncounterDuration(entry),
		bestDps: formatBulkDps(best.dpsMetrics.avg),
		gain:
			!referenceIsCurrent || gain === 0 ? null : formatDeltaText(originalGearResults.dpsMetrics.avg, best.dpsMetrics.avg, 2, undefined, true, true).text,
	};
};

export type BulkEncounterResultsProps = Pick<BulkEncounterResultsViewProps, 'selected' | 'onSelect' | 'footer'>;

// One row per finished encounter, the current one first; renders nothing before a run has finished one.
export const BulkEncounterResults = ({ selected, onSelect, footer }: BulkEncounterResultsProps) => {
	const results = useBulkState(slice => slice.results);
	if (!results) return null;
	return (
		<BulkEncounterResultsView
			rows={results.map(resultRow)}
			gains={results[0]?.results.referenceIsCurrent ?? true}
			selected={selected}
			onSelect={onSelect}
			footer={footer}
		/>
	);
};
