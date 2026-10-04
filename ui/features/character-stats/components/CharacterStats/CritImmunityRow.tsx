import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { useId } from 'react';

import { TooltipRow } from './TooltipRow';
import { critImmunityClass } from './utils/stat_display';

export type CritImmunityInfo = ReturnType<Player<any>['getCritImmunityInfo']>;

/** Tank-only: how far the character is from the 5.6% reduced-crit-taken that makes it crit immune. */
export const CritImmunityRow = ({ info, text, pending }: { info: CritImmunityInfo; text: string; pending?: boolean }) => {
	const id = useId();
	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{i18n.t('sidebar.character_stats.tank_caps.crit_immunity_label')}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container">
							<TooltipAnchor tooltipId={id} testId="stat-value-link" className={critImmunityClass(info.delta)}>
								{`${text} `}
							</TooltipAnchor>
						</div>
						<Tooltip
							id={id}
							content={
								<>
									<TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.defense')} value={`${info.defense.toFixed(2)}%`} />
									{info.talents > 0 && (
										<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.talents')} value={`${info.talents.toFixed(2)}%`} />
									)}
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
