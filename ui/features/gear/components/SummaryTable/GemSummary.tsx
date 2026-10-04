import type { UIGem as Gem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionIds } from '@sim/hooks/useActionIds';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { ActionId } from '@sim/proto/action_id';
import { useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { gemSummaryRows } from '../../model/summary_totals';
import { GemRow } from './GemRow';
import { SummaryTable } from './SummaryTable';

export interface GemSummaryViewRow {
	gem: Gem;
	count: number;
	iconUrl?: string;
}

export interface GemSummaryViewProps {
	rows: ReadonlyArray<GemSummaryViewRow>;
	onReset: () => void;
}

export const GemSummaryView = ({ rows, onReset }: GemSummaryViewProps) => (
	<SummaryTable
		title={i18n.t('gear_tab.gem_summary.title')}
		modifier="summary-table--gems"
		empty={!rows.length}
		reset={{ label: i18n.t('gear_tab.gem_summary.reset_gems'), onReset }}>
		{rows.map(row => (
			<GemRow key={row.gem.name} gem={row.gem} count={row.count} iconUrl={row.iconUrl} />
		))}
	</SummaryTable>
);

export const GemSummary = () => {
	const player = usePlayer();
	const gear = usePlayerStore('gear');
	const rows = useMemo(() => gemSummaryRows(gear.getAllGems()), [gear]);
	const icons = useActionIds(rows.map(row => ActionId.fromItemId(row.gem.id)));

	return (
		<GemSummaryView
			rows={rows.map((row, index) => ({ ...row, iconUrl: icons[index]?.iconUrl }))}
			onReset={() => {
				trackEvent({ action: 'click', category: 'gems', label: 'reset' });
				player.setGear(player.getGear().withoutGems());
			}}
		/>
	);
};
