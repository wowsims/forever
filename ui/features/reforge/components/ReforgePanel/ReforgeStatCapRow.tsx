import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { ReforgeOptimizerModel } from '@features/reforge/model/reforge_optimizer';
import { toDefaultUnitStatValue, toVisualUnitStatPercentage } from '@features/reforge/model/utils';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { subscribeReforgeChange } from '@sim/state/subscriptions';
import { TableCell, TableRow } from '@ui-kit/Table';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { type ReactNode, useMemo } from 'react';

import { useReforgeIdPrefix } from './context/ReforgeIdPrefixContext';

export interface ReforgeStatCapRowViewProps {
	statName: string;
	tooltip?: ReactNode;
	percentage: ReactNode;
	undershoot: ReactNode;
	/** The spec's preset caps for this stat, on a row of their own. */
	presets?: ReactNode;
}

export const ReforgeStatCapRowView = ({ statName, tooltip, percentage, undershoot, presets }: ReforgeStatCapRowViewProps) => {
	return (
		<>
			<TableRow>
				<TableCell className="w-0 whitespace-nowrap">
					{statName} {!!tooltip && <TooltipButton tooltip={tooltip} label={i18n.t('common.help')} testId="reforge-stat-cap-help" />}
				</TableCell>
				<TableCell className="w-18">{percentage}</TableCell>
				<TableCell className="w-14 text-right">{undershoot}</TableCell>
			</TableRow>
			{presets && (
				<TableRow>
					<TableCell />
					<TableCell className="w-18">{presets}</TableCell>
					<TableCell />
				</TableRow>
			)}
		</>
	);
};

export interface ReforgeStatCapRowProps {
	model: ReforgeOptimizerModel;
	player: Player<any>;
	unitStat: UnitStat;
	tooltip: ReactNode;
}

const CAP_FIELDS = ['reforge:useSoftCapBreakpoints', 'reforge:statCaps'] as const;

/** One stat's cap: the percentage, the undershoot flag, and — where the spec supplies them — a preset select on its own row. */
export const ReforgeStatCapRow = ({ model, player, unitStat, tooltip }: ReforgeStatCapRowProps) => {
	const settings = model.settings;
	const idPrefix = useReforgeIdPrefix();
	const statName = unitStat.getShortName(player.getClass());

	const getValue = () => toVisualUnitStatPercentage(settings.statCaps.getUnitStat(unitStat), unitStat);
	const setValue = (_player: Player<any>, newValue: number) =>
		settings.setStatCaps(settings.statCaps.withUnitStat(unitStat, toDefaultUnitStatValue(newValue, unitStat)));
	const enableWhen = () => model.isAllowedToOverrideStatCaps || !model.softCapsConfig.some(config => config.unitStat.equals(unitStat));

	const statPresets = model.statSelectionPresets?.find(entry => entry.unitStat.equals(unitStat))?.presets;
	const presetValues = useMemo(
		() =>
			statPresets
				? [
						{ name: i18n.t('sidebar.buttons.suggest_reforges.select_preset'), value: 0 },
						...[...statPresets.entries()].map(([key, percentValue]) => ({ name: `${key} - ${percentValue.toFixed(2)}%`, value: percentValue })),
					].sort((a, b) => a.value - b.value)
				: null,
		[statPresets],
	);

	return (
		<ReforgeStatCapRowView
			statName={statName}
			tooltip={tooltip}
			percentage={
				<ConnectedNumberField
					modObject={player}
					config={{
						id: `${idPrefix}-${statName}-percentage`,
						float: true,
						maxDecimalDigits: 5,
						showZeroes: false,
						positive: true,
						enableWhen,
						storeField: CAP_FIELDS,
						getValue,
						setValue,
					}}
					inputClassName="pr-2 pl-2 text-right"
				/>
			}
			undershoot={
				<ConnectedBooleanPicker
					modObject={player}
					config={{
						id: `${idPrefix}-${statName}-undershoot`,
						label: '',
						storeSubscribe: () => subscribeReforgeChange(settings),
						getValue: () => settings.undershootCaps.getUnitStat(unitStat) > 0,
						setValue: (_player, newValue) => {
							settings.undershootCaps = settings.undershootCaps.withUnitStat(unitStat, newValue ? 1 : 0);
						},
					}}
				/>
			}
			presets={
				presetValues && (
					<ConnectedChoiceField
						modObject={player}
						config={{
							id: `${idPrefix}-${statName}-presets`,
							label: '',
							values: presetValues,
							enableWhen,
							storeField: CAP_FIELDS,
							getValue,
							setValue,
						}}
					/>
				)
			}
		/>
	);
};
