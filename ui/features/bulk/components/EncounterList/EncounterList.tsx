import { BulkEncounterRef } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useMemo } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { bulkEncounterList, CURRENT_ENCOUNTER_REF, currentEncounterChecked, formatEncounterSummary, setBulkEncounterChecked } from '../../model/encounters';
import type { SavedRef } from '../../model/saved_refs';
import { BulkCheckList, type BulkCheckListProps, type BulkCheckRow } from '../BulkCheckList';

export interface EncounterListViewProps {
	current: NonNullable<BulkCheckListProps['current']>;
	presets: readonly BulkCheckRow[];
	saved: readonly BulkCheckRow[];
	onCheckedChange(ref: SavedRef, checked: boolean): void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.settings.encounters.${key}`, options);

export const EncounterListView = ({ current, presets, saved, onCheckedChange }: EncounterListViewProps) => (
	<BulkCheckList
		strings={{ presets: t('presets'), saved: t('saved'), empty: t('empty') }}
		current={current}
		presets={presets}
		saved={saved}
		onCheckedChange={onCheckedChange}
		testIds={{ list: 'bulk-encounters', row: 'bulk-encounter', current: 'bulk-encounter-current' }}
	/>
);

export const EncounterList = () => {
	const player = usePlayer();
	const { current, entries, checked } = useBulkEncounters();

	const rows = useMemo(() => {
		const list = bulkEncounterList(current, entries, checked).map((row): BulkCheckRow => ({
			ref: row.ref,
			name: row.name,
			summary: formatEncounterSummary(row.encounter),
			checked: row.checked,
			duplicate: false,
			disabledReason: row.duplicateOf === null ? undefined : t('duplicate', { encounter: row.duplicateOf }),
		}));
		return { presets: list.filter(row => row.ref.preset), saved: list.filter(row => !row.ref.preset) };
	}, [current, entries, checked]);

	const onCheckedChange = (ref: SavedRef, isChecked: boolean) =>
		setBulkEncounterChecked(
			player,
			BulkEncounterRef.create({ name: ref.name, preset: ref.preset }),
			isChecked,
			entries.map(entry => entry.ref),
		);

	return (
		<EncounterListView
			current={{
				label: t('current'),
				summary: formatEncounterSummary(current),
				checked: currentEncounterChecked(checked),
				onCheckedChange: isChecked => onCheckedChange(CURRENT_ENCOUNTER_REF, isChecked),
			}}
			presets={rows.presets}
			saved={rows.saved}
			onCheckedChange={onCheckedChange}
		/>
	);
};
