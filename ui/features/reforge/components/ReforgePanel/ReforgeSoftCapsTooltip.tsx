import { breakpointValueToDisplayPercentage } from '@features/reforge/model/utils';
import { StatCapType } from '@generated/proto/api';
import type { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import { statCapTypeNames } from '@sim/proto/names';
import type { StatCap } from '@sim/proto/stats';
import type { StatTooltipContent } from '@sim/spec_config';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';

export interface ReforgeSoftCapsTooltipProps {
	playerClass: Class;
	softCaps: StatCap[];
	additionalInformation: StatTooltipContent;
}

/** The breakpoint tables hanging off the optimize button, one per soft-capped stat. Built per open, so the limits they show are current. */
export const ReforgeSoftCapsTooltip = ({ playerClass, softCaps, additionalInformation }: ReforgeSoftCapsTooltipProps) => (
	<>
		<p>{i18n.t('sidebar.buttons.suggest_reforges.breakpoints_implemented')}</p>
		<div className="grid gap-4">
			{softCaps.map(({ unitStat, breakpoints, capType, postCapEPs }) => {
				const extra = unitStat.hasRootStat() ? additionalInformation[unitStat.getRootStat()]?.() : undefined;
				return (
					<Table key={unitStat.getKey()} dense>
						<thead>
							<TableRow>
								<TableHeaderCell className="text-left" colSpan={2}>
									{unitStat.getShortName(playerClass)}
								</TableHeaderCell>
								<TableHeaderCell className="text-right">{statCapTypeNames.get(capType)}</TableHeaderCell>
							</TableRow>
						</thead>
						<tbody>
							{extra !== undefined && (
								<TableRow>
									<TableCell colSpan={3}>{extra}</TableCell>
								</TableRow>
							)}
							<TableRow>
								<TableHeaderCell className="text-right">
									<em>%</em>
								</TableHeaderCell>
								<TableHeaderCell colSpan={2} className="text-right">
									<em>{i18n.t('sidebar.buttons.suggest_reforges.post_cap_ep')}</em>
								</TableHeaderCell>
							</TableRow>
							{breakpoints.map((breakpoint, breakpointIndex) => (
								<TableRow key={breakpoint}>
									<TableCell className="text-right">{breakpointValueToDisplayPercentage(breakpoint, unitStat)}</TableCell>
									<TableCell colSpan={2} className="text-right">
										{unitStat
											.convertEpToRatingScale(capType === StatCapType.TypeThreshold ? postCapEPs[0] : postCapEPs[breakpointIndex])
											.toFixed(2)}
									</TableCell>
								</TableRow>
							))}
						</tbody>
					</Table>
				);
			})}
		</div>
	</>
);
