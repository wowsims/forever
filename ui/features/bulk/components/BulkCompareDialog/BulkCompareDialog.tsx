import i18n from '@i18n/config';
import type { BulkEncounterResults } from '@sim/bulk/types';
import { getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Gear } from '@sim/proto/gear';
import { formatDeltaText, formatToNumber, formatToPercent } from '@sim/utils/format';
import { Dialog } from '@ui-kit/Dialog';
import { useMemo } from 'react';

import { useBulkResults } from '../../hooks/useBulkResults';
import { encounterLabel, formatEncounterDuration } from '../../model/builds/encounters';
import { gearSetupLabel } from '../../model/builds/gear_setups';
import { rotationLabel } from '../../model/builds/rotations';
import { settingsLabel } from '../../model/builds/settings_entries';
import { talentsLabel } from '../../model/builds/talents';
import { changedEnchantLabels, changedItemLabels, compareMatrix, type GainTint, type MatrixCell } from '../../model/results/compare';
import { formatBulkDps } from '../../model/results/format';
import { BulkCompareTable, type BulkCompareTableProps } from '../BulkCompareTable';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.compare.${key}`, options);

const currentLabel = () => i18n.t('bulk_tab.results.current');

const currentGearLabel = () => i18n.t('bulk_tab.results.current_gear');

const shortLabel = (number: number) => t('short_set', { number });

const gainLabel = (gainPct: number) => formatToPercent(gainPct, { minimumFractionDigits: 2, maximumFractionDigits: 2, signDisplay: 'exceptZero' });

// The rank and DPS with the fight-to-fight spread, the ± the sim results card shows, then the gain over the reference:
// the current gear, or the encounter's best set when the batch did not sim the current setup. The current gear's own
// cells (rank null) carry neither rank nor gain. The percentage is the cell's own, of the reference's DPS.
const cellTooltip = ({ result, rank, gainPct }: MatrixCell, entry: BulkEncounterResults): string => {
	const base = entry.results.originalGearResults.dpsMetrics;
	const { avg, stdev } = result.dpsMetrics;
	const dps = formatBulkDps(avg);
	const spread = t('readout_spread', { spread: formatToNumber(stdev, { maximumFractionDigits: 0 }) });
	if (rank === null) return `${t('readout_current', { dps })} ${spread}`;
	return `${t('readout_rank', { rank, dps })} ${spread} · ${formatDeltaText(base.avg, avg, 2, undefined, true).text} (${gainLabel(gainPct)})`;
};

// A set that changes no item (current gear under another talent setup) says so, as the result rows do.
const changesOf = (gear: Gear, current: Gear, canDualWield: boolean): string[] => {
	const changed = changedItemLabels(gear, current, canDualWield);
	return changed.length ? changed : [i18n.t('bulk_tab.results.gear_unchanged')];
};

const tableProps = (entries: readonly BulkEncounterResults[], canDualWield: boolean): BulkCompareTableProps => {
	const { columns, rows, current } = compareMatrix(entries);
	const tints = rows.flatMap(row => row.cells.flatMap(cell => cell?.tint ?? []));
	return {
		// The column gives the length alone; the variation is on the encounter list.
		columns: columns.map(entry => ({
			label: encounterLabel(entry),
			length: formatEncounterDuration({ duration: entry.duration, durationVariation: null }),
		})),
		// A Builds set, which carries its settings, is named by every part of its build; a Gear-mode set by its items.
		rows: rows.map(({ key, set, talents, rotation, settings, cells }) => ({
			key,
			label: set ? shortLabel(set.number) : currentLabel(),
			talents: settings ? talentsLabel(talents) : null,
			rotation: settings ? rotationLabel(rotation) : null,
			settings: settings ? settingsLabel(settings) : null,
			gear: !set ? currentGearLabel() : set.gearSetup ? gearSetupLabel(set.gearSetup) : null,
			changes: !set?.gearSetup && set && current ? changesOf(set.gear, current, canDualWield) : [],
			enchants: set ? changedEnchantLabels(set) : [],
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
		legend: { lowTint: Math.min(0, ...tints) as GainTint, highTint: Math.max(0, ...tints) as GainTint },
		builds: rows.some(row => row.settings !== null),
		againstBest: entries[0]?.results.referenceIsCurrent === false,
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
	const entries = useBulkResults();

	return (
		<Dialog open={open} onOpenChange={onOpenChange} size="full" scrollContents title={t('title')}>
			{entries && <BulkCompareBody entries={entries} />}
		</Dialog>
	);
};
