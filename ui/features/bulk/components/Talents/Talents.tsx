import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useMemo } from 'react';

import { useBulkTalents } from '../../hooks/useBulkTalents';
import { bulkTalentList, formatTalentsSummary, setBulkTalentChecked } from '../../model/talents';
import { BulkCheckList, type BulkCheckRow } from '../BulkCheckList';

const t = (key: string) => i18n.t(`bulk_tab.settings.talents.${key}`);

export const Talents = () => {
	const player = usePlayer();
	const { playerSpec, current, entries, checked } = useBulkTalents();
	const rows = useMemo(() => {
		const list = bulkTalentList(playerSpec, current, entries, checked).map((item): BulkCheckRow => ({
			ref: item.ref,
			name: item.name,
			summary: formatTalentsSummary(item.talentsString),
			checked: item.checked,
			duplicate: item.duplicate,
		}));
		return { presets: list.filter(row => row.ref.preset), saved: list.filter(row => !row.ref.preset) };
	}, [playerSpec, current, entries, checked]);

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
			currentSummary={formatTalentsSummary(current)}
			presets={rows.presets}
			saved={rows.saved}
			onCheckedChange={(ref, value) =>
				setBulkTalentChecked(
					player,
					ref,
					value,
					entries.map(entry => entry.ref),
				)
			}
			testIds={{ list: 'bulk-talents', row: 'bulk-talent', current: 'bulk-talent-current' }}
		/>
	);
};
