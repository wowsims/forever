import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { formatDeltaText, formatDurationSeconds, formatToNumber } from '@sim/utils/format';
import { Dialog } from '@ui-kit/Dialog';
import { useMemo } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import {
	type BulkComparison,
	bulkComparison,
	changedItemLabels,
	type ComparedBar,
	type ComparedEncounter,
	compareMatrix,
	encounterRanks,
	gainTint,
} from '../../model/compare';
import { formatBulkDps } from '../../model/format';
import { BulkCompareTable, type BulkCompareTableProps } from '../BulkCompareTable';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.compare.${key}`, options);

const currentLabel = () => i18n.t('bulk_tab.results.current_gear');

const shortLabel = (number: number) => t('short_set', { number });

const encounterLabel = ({ entry }: ComparedEncounter) => (entry.isCurrent ? i18n.t('bulk_tab.results.current') : entry.name);

const lengthLabel = (duration: number) => formatDurationSeconds(duration, { separatorStyle: 'colon', minimumUnit: 'minutes' });

// The rank and DPS with the fight-to-fight spread, the ± the sim results card shows, then the gain over the current gear.
// The current gear's own cells (rank null) carry neither rank nor gain.
const cellTooltip = (bar: ComparedBar, { entry }: ComparedEncounter, rank: number | null): string => {
	const base = entry.results.originalGearResults.dpsMetrics;
	const { avg, stdev } = bar.dpsMetrics;
	const dps = formatBulkDps(avg);
	const spread = t('readout_spread', { spread: formatToNumber(stdev, { maximumFractionDigits: 0 }) });
	if (rank === null) return `${t('readout_current', { dps })} ${spread}`;
	return `${t('readout_rank', { rank, dps })} ${spread} · ${formatDeltaText(base.avg, avg, 2, undefined, true, true).text}`;
};

const gainLabel = (gainPct: number) => `${formatToNumber(gainPct, { minimumFractionDigits: 2, maximumFractionDigits: 2, signDisplay: 'exceptZero' })}%`;

const tableProps = (comparison: BulkComparison): BulkCompareTableProps => {
	const { encounters, current } = comparison;
	const { rows, scale } = compareMatrix(comparison);
	const ranks = encounters.map(encounterRanks);
	return {
		columns: encounters.map(encounter => ({
			label: encounterLabel(encounter),
			length: encounter.entry.duration === null ? encounter.entry.durationLabel : lengthLabel(encounter.entry.duration),
		})),
		rows: rows.map(({ key, set, cells }) => ({
			key,
			label: set ? shortLabel(set.number) : currentLabel(),
			changes: set && current ? changedItemLabels(set.gear, current) : [],
			cells: cells.map(
				(cell, index) =>
					cell && {
						gain: gainLabel(cell.gainPct),
						dps: formatBulkDps(cell.bar.dpsMetrics.avg),
						tint: cell.tint,
						best: cell.best,
						tooltip: cellTooltip(cell.bar, encounters[index], set ? (ranks[index].get(key) ?? null) : null),
					},
			),
		})),
		legend: {
			low: gainLabel(scale.low),
			high: gainLabel(scale.high),
			lowTint: gainTint(scale.low, scale),
			highTint: gainTint(scale.high, scale),
		},
	};
};

// The dialog's popup mounts this on open, so the comparison is built only while the dialog shows.
const BulkCompareBody = ({ entries }: { entries: readonly BulkEncounterResults[] }) => {
	const table = useMemo(() => tableProps(bulkComparison(entries)), [entries]);
	return <BulkCompareTable {...table} />;
};

export interface BulkCompareDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const BulkCompareDialog = ({ open, onOpenChange }: BulkCompareDialogProps) => {
	const entries = useBulkState(slice => slice.results?.entries);

	return (
		<Dialog open={open} onOpenChange={onOpenChange} size="full" scrollContents title={t('title')}>
			{entries && <BulkCompareBody entries={entries} />}
		</Dialog>
	);
};
