import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { breakpointValueToDisplayPercentage } from '@features/reforge/model/utils';
import { StatCapType } from '@generated/proto/api';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { StatCap } from '@sim/proto/stats';
import type { ReforgeSettings } from '@sim/settings/reforge_settings';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { TooltipButton } from '@ui-kit/TooltipButton';
import type { ReactNode } from 'react';

import { useReforgeIdPrefix } from './context/ReforgeIdPrefixContext';
import { isIncludedStat } from './utils';

export interface ReforgeBreakpointLimitRow {
	key: string;
	statName: string;
	picker: ReactNode;
}

export interface ReforgeBreakpointLimitsViewProps {
	rows: ReadonlyArray<ReforgeBreakpointLimitRow>;
}

export const ReforgeBreakpointLimitsView = ({ rows }: ReforgeBreakpointLimitsViewProps) => {
	return (
		<>
			<div className="flex pb-4">
				<h6 className="mr-1 mb-0">{i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit')}</h6>
				<TooltipButton
					tooltip={i18n.t('sidebar.buttons.suggest_reforges.breakpoint_limit_tooltip')}
					label={i18n.t('common.help')}
					testId="reforge-breakpoint-limit-help"
				/>
			</div>
			<Table dense className="mb-2">
				<tbody>
					{rows.map(row => (
						<TableRow key={row.key}>
							<TableCell className="w-0 whitespace-nowrap">{row.statName}</TableCell>
							<TableCell className="w-18">{row.picker}</TableCell>
						</TableRow>
					))}
				</tbody>
			</Table>
		</>
	);
};

export interface ReforgeBreakpointLimitsProps {
	settings: ReforgeSettings;
	softCapsConfig: StatCap[];
	player: Player<any>;
	useSoftCapBreakpoints: boolean;
}

/** Caps the highest breakpoint the optimizer will chase, one select per soft-capped stat. */
export const ReforgeBreakpointLimits = ({ settings, softCapsConfig, player, useSoftCapBreakpoints }: ReforgeBreakpointLimitsProps) => {
	const idPrefix = useReforgeIdPrefix();

	if (!useSoftCapBreakpoints) return null;

	const rows = softCapsConfig
		.filter(
			config =>
				(config.capType === StatCapType.TypeThreshold || config.capType === StatCapType.TypeSoftCap) &&
				config.breakpoints.length > 0 &&
				isIncludedStat(config.unitStat),
		)
		.map(({ breakpoints, unitStat }) => {
			const statName = unitStat.getShortName(player.getClass());
			return {
				key: unitStat.getKey(),
				statName,
				picker: (
					<ConnectedChoiceField
						modObject={player}
						config={{
							id: `${idPrefix}-${statName}-breakpoint-limit`,
							label: '',
							values: [
								{ name: i18n.t('sidebar.buttons.suggest_reforges.no_limit_set'), value: 0 },
								...breakpoints.map(breakpoint => ({
									name: `${breakpointValueToDisplayPercentage(breakpoint, unitStat)}%`,
									value: breakpoint,
								})),
							].sort((a, b) => a.value - b.value),
							storeField: 'reforge:useSoftCapBreakpoints',
							getValue: () => {
								const limit = settings.breakpointLimits.getUnitStat(unitStat);
								return breakpoints.some(breakpoint => breakpoint == limit) ? limit : 0;
							},
							setValue: (_player, newValue) => settings.setBreakpointLimits(settings.breakpointLimits.withUnitStat(unitStat, newValue)),
						}}
					/>
				),
			};
		});

	return <ReforgeBreakpointLimitsView rows={rows} />;
};
