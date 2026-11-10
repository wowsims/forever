import type { MeleeCritCapInfo } from '@generated/proto/api';
import i18n from '@i18n/config';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { useId } from 'react';

import { TooltipRow } from './TooltipRow';
import { critCapClass } from './utils/stat_display';

export interface CritCapRowProps {
	info: MeleeCritCapInfo;
	text: string;
	pending?: boolean;
}

export const CritCapRow = ({ info, text, pending }: CritCapRowProps) => {
	const id = useId();
	const tone = info.delta === 0 ? 'neutral' : info.delta > 0 ? 'negative' : 'positive';
	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{i18n.t('sidebar.character_stats.melee_crit_cap')}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container" data-tone={tone}>
							<TooltipAnchor tooltipId={id} testId="stat-value-link" className={critCapClass(info.delta)}>
								{`${text} `}
							</TooltipAnchor>
						</div>
						<Tooltip
							id={id}
							content={
								<>
									<TooltipRow label={i18n.t('sidebar.character_stats.attack_table.glancing')} value={`${info.glancing.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.attack_table.suppression')} value={`${info.suppression.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.attack_table.to_hit_cap')} value={`${info.remainingHit.toFixed(2)}%`} />
									<TooltipRow
										label={i18n.t('sidebar.character_stats.attack_table.to_exp_cap')}
										value={`${info.remainingExpertise.toFixed(2)}%`}
									/>
									{info.block > 0 && (
										<TooltipRow label={i18n.t('sidebar.character_stats.attack_table.block')} value={`${info.block.toFixed(2)}%`} />
									)}
									{info.specOffset !== 0 && (
										<TooltipRow
											label={i18n.t('sidebar.character_stats.attack_table.spec_offsets')}
											value={`${info.specOffset.toFixed(2)}%`}
										/>
									)}
									<TooltipRow label={i18n.t('sidebar.character_stats.attack_table.final_crit_cap')} value={`${info.critCap.toFixed(2)}%`} />
									<hr />
									<TooltipRow
										label={i18n.t('sidebar.character_stats.attack_table.can_raise_by')}
										value={`${(info.remainingExpertise + info.remainingHit).toFixed(2)}%`}
									/>
								</>
							}
						/>
					</>
				)}
			</TableCell>
		</TableRow>
	);
};
