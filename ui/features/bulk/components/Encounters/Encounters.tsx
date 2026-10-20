import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useMemo } from 'react';

import { useBulkEncounters } from '../../hooks/useBulkEncounters';
import { bulkEncounterList, formatEncounterSummary, setBulkEncounterChecked } from '../../model/encounters';
import { BulkCheckList, type BulkCheckRow } from '../BulkCheckList';

const t = (key: string) => i18n.t(`bulk_tab.settings.encounters.${key}`);

export const Encounters = () => {
	const player = usePlayer();
	const { current, entries, checked } = useBulkEncounters();
	const rows = useMemo(() => {
		const list = bulkEncounterList(current, entries, checked).map((item): BulkCheckRow => ({
			ref: item.ref,
			name: item.name,
			summary: formatEncounterSummary(item.encounter),
			checked: item.checked,
			duplicate: item.duplicate,
		}));
		return { presets: list.filter(row => row.ref.preset), saved: list.filter(row => !row.ref.preset) };
	}, [current, entries, checked]);

	return (
		<BulkCheckList
			strings={{
				description: t('description'),
				current: t('current'),
				presets: t('presets'),
				saved: t('saved'),
				empty: t('empty'),
				duplicate: t('duplicate'),
			}}
			currentSummary={formatEncounterSummary(current)}
			presets={rows.presets}
			saved={rows.saved}
			onCheckedChange={(ref, value) =>
				setBulkEncounterChecked(
					player,
					ref,
					value,
					entries.map(entry => entry.ref),
				)
			}
			testIds={{ list: 'bulk-encounters', row: 'bulk-encounter', current: 'bulk-encounter-current' }}
		/>
	);
};
