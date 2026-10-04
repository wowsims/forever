import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { useId } from 'react';

import { TooltipRow } from './TooltipRow';

export type MissInfo = ReturnType<Player<any>['getMissChanceInfo']>;

/** Tank-only: the target's chance to miss, broken down into base, defense skill and debuffs. */
export const MissRow = ({ info, pending }: { info: MissInfo; pending?: boolean }) => {
	const id = useId();
	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{i18n.t('sidebar.character_stats.tank_caps.miss_label')}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container">
							<TooltipAnchor tooltipId={id} testId="stat-value-link" className="text-white">
								{`${info.total.toFixed(2)}%`}
							</TooltipAnchor>
						</div>
						<Tooltip
							id={id}
							content={
								<>
									<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.base')} value={`${info.base.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.defense')} value={`${info.defense.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.debuffs')} value={`${info.debuffs.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.total')} value={`${info.total.toFixed(2)}%`} />
								</>
							}
						/>
					</>
				)}
			</TableCell>
		</TableRow>
	);
};
