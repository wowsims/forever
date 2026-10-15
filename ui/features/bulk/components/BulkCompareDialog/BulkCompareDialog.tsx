import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { formatDeltaText, formatToNumber } from '@sim/utils/format';
import { stDevToConf95 } from '@sim/utils/math';
import { Dialog } from '@ui-kit/Dialog';
import { useMemo, useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { type BulkComparison, bulkComparison, changedItemLabels, type ComparedBar, type ComparedEncounter } from '../../model/compare';
import { formatBulkDps } from '../../model/format';
import { BulkCompareChart, type BulkCompareGroup } from '../BulkCompareChart';
import { BulkCompareTable, type BulkCompareTableCell, type BulkCompareTableRow } from '../BulkCompareTable';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.compare.${key}`, options);

const setLabel = (bar: Pick<ComparedBar, 'isCurrent' | 'number'>) =>
	bar.isCurrent ? i18n.t('bulk_tab.results.current_gear') : t('set', { number: bar.number });

const encounterLabel = ({ entry }: ComparedEncounter) => (entry.isCurrent ? i18n.t('bulk_tab.results.current') : entry.name);

// As the result rows show a set's margin and its gain, without the significance colour.
const formatted = (bar: ComparedBar): BulkCompareTableCell => ({
	dps: formatBulkDps(bar.dpsMetrics.avg),
	margin: formatToNumber(stDevToConf95(bar.dpsMetrics.stdev, bar.iterations), { maximumFractionDigits: 0 }),
	gain: bar.isCurrent ? null : formatDeltaText(bar.dpsMetrics.avg - bar.gain, bar.dpsMetrics.avg, 2, undefined, true, true).text,
});

const chartGroups = ({ encounters }: BulkComparison): BulkCompareGroup[] =>
	encounters.map(encounter => ({
		label: encounterLabel(encounter),
		summary: encounter.entry.summaryLabel,
		bars: encounter.bars.map(bar => {
			const { dps, margin, gain } = formatted(bar);
			const set = setLabel(bar);
			return {
				key: bar.key,
				label: set,
				isCurrent: bar.isCurrent,
				avg: bar.dpsMetrics.avg,
				details: gain === null ? t('details', { set, dps, margin }) : t('details_gain', { set, dps, margin, gain }),
			};
		}),
	}));

// The current gear first, then every set by number, each with its result where an encounter ranked it.
const tableRows = ({ encounters, sets, current }: BulkComparison): BulkCompareTableRow[] => {
	const cellsOf = (key: string) =>
		encounters.map(encounter => {
			const bar = encounter.bars.find(candidate => candidate.key === key);
			return bar ? formatted(bar) : null;
		});
	const currentKey = encounters[0]?.bars.find(bar => bar.isCurrent)?.key ?? '';
	return [
		{ key: currentKey, label: setLabel({ isCurrent: true, number: 0 }), changes: [], cells: cellsOf(currentKey) },
		...sets.map(set => ({
			key: set.key,
			label: setLabel({ isCurrent: false, number: set.number }),
			changes: current ? changedItemLabels(set.gear, current) : [],
			cells: cellsOf(set.key),
		})),
	];
};

// The dialog's popup mounts this on open, so the comparison is built only while the dialog shows.
const BulkCompareBody = ({ entries }: { entries: readonly BulkEncounterResults[] }) => {
	const [highlighted, setHighlighted] = useState<string | null>(null);
	const { comparison, groups, rows } = useMemo(() => {
		const comparison = bulkComparison(entries);
		return { comparison, groups: chartGroups(comparison), rows: tableRows(comparison) };
	}, [entries]);

	return (
		<div className="grid grid-cols-1 gap-6">
			<BulkCompareChart groups={groups} highlighted={highlighted} onHighlight={setHighlighted} />
			<BulkCompareTable
				encounters={comparison.encounters.map(encounter => ({ label: encounterLabel(encounter), summary: encounter.entry.summaryLabel }))}
				rows={rows}
				highlighted={highlighted}
				onHighlight={setHighlighted}
			/>
		</div>
	);
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
