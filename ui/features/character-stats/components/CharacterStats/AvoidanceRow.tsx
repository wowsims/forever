import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { useId } from 'react';

import { TooltipRow } from './TooltipRow';

export type AvoidanceInfo = ReturnType<Player<any>['getAvoidanceInfo']>;

/** A crushing blow is off the table at 102.4% combined avoidance, a shear at 101.8% without miss. */
const CRUSH_CAP = 102.4;
const SHEAR_CAP = 101.8;

export interface AvoidanceRowProps {
	info: AvoidanceInfo;
	/** Protection Paladin's block is quoted with Holy Shield up, so every block-derived figure is suffixed. */
	withHolyShield: boolean;
	hasParry: boolean;
	hasBlock: boolean;
	pending?: boolean;
}

export const AvoidanceRow = ({ info, withHolyShield, hasParry, hasBlock, pending }: AvoidanceRowProps) => {
	const id = useId();
	const suffix = withHolyShield ? i18n.t('sidebar.character_stats.tank_caps.with_holy_shield') : '';
	const blockString = `${info.block.toFixed(2)}%${suffix}`;
	const totalString = `${info.total.toFixed(2)}%${suffix}`;
	const shearString = `${info.shear.toFixed(2)}%${suffix}`;

	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{i18n.t('sidebar.character_stats.tank_caps.avoidance_label')}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container">
							<TooltipAnchor tooltipId={id} testId="stat-value-link" className="text-white">
								{totalString}
							</TooltipAnchor>
						</div>
						<Tooltip
							id={id}
							content={
								<>
									<TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.miss')} value={`${info.miss.toFixed(2)}%`} />
									<TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.dodge')} value={`${info.dodge.toFixed(2)}%`} />
									{hasParry && <TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.parry')} value={`${info.parry.toFixed(2)}%`} />}
									{hasBlock && <TooltipRow label={i18n.t('sidebar.character_stats.tank_caps.block')} value={blockString} />}
									<TooltipRow
										className={info.total >= CRUSH_CAP ? 'text-success' : 'text-danger'}
										cap={info.total >= CRUSH_CAP ? 'met' : 'missed'}
										label={i18n.t('sidebar.character_stats.tank_caps.crush')}
										value={totalString}
									/>
									<TooltipRow
										className={info.shear >= SHEAR_CAP ? 'text-success' : 'text-danger'}
										cap={info.shear >= SHEAR_CAP ? 'met' : 'missed'}
										label={i18n.t('sidebar.character_stats.tank_caps.shear')}
										value={shearString}
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
