import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { TableHeaderCell, TableRow } from '@ui-kit/Table';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import type { ReactNode } from 'react';

import type { EpColumn } from './types';
import { EP_TOOLTIP_ID } from './utils';

export interface EpWeightsHeaderProps {
	columns: EpColumn[];
	showThreatMetrics: boolean;
	showUpdateColumn?: boolean;
	statAction?: ReactNode;
}

// Every label wraps: with threat metrics on, fourteen metric columns share the dialog's width.
// The sticky cells draw the row's bottom rule themselves, in its transparent border's place: a
// collapsed row border stays with the scrolled table, not with the header.
export const EpWeightsHeader = ({ columns, showThreatMetrics, showUpdateColumn = true, statAction }: EpWeightsHeaderProps) => (
	<TableRow className="border-b-transparent">
		<TableHeaderCell wrap className={['ui-ep-weights-table-header-cell', showThreatMetrics && 'lg:max-xl:pr-0']}>
			<span className="inline-flex items-center gap-1">
				{i18n.t('sidebar.buttons.stat_weights.modal.column_headers.stat')}
				{statAction}
			</span>
		</TableHeaderCell>
		{showUpdateColumn && (
			<TableHeaderCell wrap className={['ui-ep-weights-table-header-cell', showThreatMetrics && 'lg:max-xl:pr-0']}>
				{i18n.t('sidebar.buttons.stat_weights.modal.column_headers.update')}
			</TableHeaderCell>
		)}
		{columns.map(column => {
			const isAction = column.type === 'action';
			return (
				<TableHeaderCell
					key={column.id}
					wrap
					data-column-type={isAction ? undefined : column.type}
					className={[
						'ui-ep-weights-table-header-cell',
						showThreatMetrics && 'lg:max-xl:pr-0',
						isAction
							? 'text-center'
							: [
									'text-right',
									column.type === 'weight' ? 'in-data-[stats-type=ep]:hidden' : 'in-data-[stats-type=weight]:hidden',
									showThreatMetrics && 'max-lg:text-center',
								],
					]}>
					<span {...tooltipAnchorProps(EP_TOOLTIP_ID, column.labelTooltip)}>{column.label}</span>
					<Button
						variant="unstyled"
						data-testid="col-action"
						className="ml-1"
						onClick={column.onCopy}
						{...tooltipAnchorProps(EP_TOOLTIP_ID, column.actionTooltip)}>
						<Icon name={isAction ? 'arrows-rotate' : 'copy'} />
					</Button>
				</TableHeaderCell>
			);
		})}
	</TableRow>
);
