import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { formatDeltaText, formatToNumber, formatToPercent } from '@sim/utils/format';
import { Dialog } from '@ui-kit/Dialog';
import { useMemo } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { changedItemLabels, compareMatrix, gainTint, type MatrixCell } from '../../model/compare';
import { encounterLabel, formatEncounterDuration } from '../../model/encounters';
import { formatBulkDps } from '../../model/format';
import { BulkCompareTable, type BulkCompareTableProps } from '../BulkCompareTable';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.compare.${key}`, options);

const currentLabel = () => i18n.t('bulk_tab.results.current_gear');

const shortLabel = (number: number) => t('short_set', { number });

const gainLabel = (gainPct: number) => formatToPercent(gainPct, { minimumFractionDigits: 2, maximumFractionDigits: 2, signDisplay: 'exceptZero' });

// The rank and DPS with the fight-to-fight spread, the ± the sim results card shows, then the gain over the current gear.
// The current gear's own cells (rank null) carry neither rank nor gain. The percentage is the cell's own, of the current
// gear's DPS.
const cellTooltip = ({ result, rank, gainPct }: MatrixCell, entry: BulkEncounterResults): string => {
	const base = entry.results.originalGearResults.dpsMetrics;
	const { avg, stdev } = result.dpsMetrics;
	const dps = formatBulkDps(avg);
	const spread = t('readout_spread', { spread: formatToNumber(stdev, { maximumFractionDigits: 0 }) });
	if (rank === null) return `${t('readout_current', { dps })} ${spread}`;
	return `${t('readout_rank', { rank, dps })} ${spread} · ${formatDeltaText(base.avg, avg, 2, undefined, true).text} (${gainLabel(gainPct)})`;
};

const tableProps = (entries: readonly BulkEncounterResults[], canDualWield: boolean): BulkCompareTableProps => {
	const { columns, rows, scale, current } = compareMatrix(entries);
	return {
		// The column gives the length alone; the variation is on the encounter list.
		columns: columns.map(entry => ({
			label: encounterLabel(entry),
			length: formatEncounterDuration({ duration: entry.duration, durationVariation: null }),
		})),
		rows: rows.map(({ key, set, cells }) => ({
			key,
			label: set ? shortLabel(set.number) : currentLabel(),
			changes: set && current ? changedItemLabels(set.gear, current, canDualWield) : [],
			cells: cells.map(
				(cell, index) =>
					cell && {
						gain: gainLabel(cell.gainPct),
						dps: formatBulkDps(cell.result.dpsMetrics.avg),
						tint: cell.tint,
						best: cell.best,
						tooltip: cellTooltip(cell, columns[index]),
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
	const canDualWield = getBulkPlayerCanDualWield(usePlayer());
	const table = useMemo(() => tableProps(entries, canDualWield), [entries, canDualWield]);
	return <BulkCompareTable {...table} />;
};

export interface BulkCompareDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const BulkCompareDialog = ({ open, onOpenChange }: BulkCompareDialogProps) => {
	const entries = useBulkState(slice => slice.results);

	return (
		<Dialog open={open} onOpenChange={onOpenChange} size="full" scrollContents title={t('title')}>
			{entries && <BulkCompareBody entries={entries} />}
		</Dialog>
	);
};
