import type { ReforgeOptimizerModel } from '@features/reforge/model/reforge_optimizer';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { UnitStat } from '@sim/proto/stats';
import { Stats } from '@sim/proto/stats';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Table, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { type ReactNode, useId } from 'react';

import { ReforgeStatCapRow } from './ReforgeStatCapRow';
import { isIncludedStat } from './utils';

export interface ReforgeStatCapsViewProps {
	onReset: () => void;
	/** One `ReforgeStatCapRowView` per stat. */
	children: ReactNode;
}

export const ReforgeStatCapsView = ({ onReset, children }: ReforgeStatCapsViewProps) => {
	const resetTooltipId = useId();
	const resetLabel = i18n.t('sidebar.buttons.suggest_reforges.reset_to_defaults');

	return (
		<>
			<div className="flex pb-4">
				<h6 className="mr-1 mb-0">{i18n.t('sidebar.buttons.suggest_reforges.edit_stat_caps')}</h6>
				<TooltipButton
					tooltip={i18n.t('sidebar.buttons.suggest_reforges.stat_caps_tooltip')}
					label={i18n.t('common.help')}
					testId="reforge-stat-caps-help"
				/>
				<Button
					variant="unstyled"
					className="ml-auto inline"
					aria-label={resetLabel}
					data-testid="reforge-reset-stat-caps"
					{...tooltipAnchorProps(resetTooltipId)}
					onClick={onReset}>
					<Icon name="arrow-rotate-left" />
				</Button>
				<Tooltip id={resetTooltipId} content={resetLabel} />
			</div>
			<Table dense className="mb-2">
				<thead>
					<TableRow>
						<TableHeaderCell className="text-left">{i18n.t('sidebar.buttons.suggest_reforges.stat')}</TableHeaderCell>
						<TableHeaderCell className="text-right">%</TableHeaderCell>
						<TableHeaderCell className="text-left">{i18n.t('sidebar.buttons.suggest_reforges.max_column')}</TableHeaderCell>
					</TableRow>
				</thead>
				<tbody>{children}</tbody>
			</Table>
		</>
	);
};

export interface ReforgeStatCapsProps {
	model: ReforgeOptimizerModel;
	player: Player<any>;
	displayStats: UnitStat[];
	statTooltips: Partial<Record<number, ReactNode>>;
	useCustomEPValues: boolean;
}

/** The hard-cap table, rendered only while custom EP values are on. */
export const ReforgeStatCaps = ({ model, player, displayStats, statTooltips, useCustomEPValues }: ReforgeStatCapsProps) => {
	if (!useCustomEPValues) return null;

	return (
		<ReforgeStatCapsView onReset={() => model.settings.setStatCaps(model.defaults.statCaps || new Stats())}>
			{displayStats.filter(isIncludedStat).map(unitStat => (
				<ReforgeStatCapRow
					key={unitStat.getKey()}
					model={model}
					player={player}
					unitStat={unitStat}
					tooltip={unitStat.hasRootStat() ? statTooltips[unitStat.getRootStat()] : null}
				/>
			))}
		</ReforgeStatCapsView>
	);
};
