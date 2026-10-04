import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { TableCell, TableRow } from '@ui-kit/Table';
import { NumberField } from '@ui-kit/TextField';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';

import type { EpColumn } from './types';
import { EP_TOOLTIP_ID } from './utils';

export interface EpRatiosRowProps {
	columns: EpColumn[];
	epRatios: number[];
	onEpRatioChange: (ratioIndex: number, value: number) => void;
	onComputeEp: () => void;
	showThreatMetrics: boolean;
}

export const EpRatiosRow = ({ columns, epRatios, onEpRatioChange, onComputeEp, showThreatMetrics }: EpRatiosRowProps) => {
	const cellClassName = clsx('ui-ep-weights-table-cell', showThreatMetrics && 'max-lg:pl-0');

	return (
		<TableRow testId="ep-ratios" className="bg-(--table-row-even-bg)">
			<TableCell className={cellClassName}>{i18n.t('sidebar.buttons.stat_weights.modal.column_headers.ep_ratio')}</TableCell>
			<TableCell className={cellClassName} />
			{columns
				.filter(column => column.type !== 'action')
				.map(column => (
					<TableCell
						key={column.id}
						data-column-type={column.type}
						className={[
							cellClassName,
							'text-right',
							column.type === 'weight' ? 'in-data-[stats-type=ep]:hidden' : 'in-data-[stats-type=weight]:hidden',
						]}>
						<NumberField
							id={`ep-ratio-${column.type}-${column.ratioIndex}`}
							float
							value={epRatios[column.ratioIndex!]}
							onChange={value => onEpRatioChange(column.ratioIndex!, value)}
							inputClassName={clsx('ui-ep-weights-input-align-right ml-auto max-w-24', showThreatMetrics && 'ui-ep-weights-compact-input')}
						/>
					</TableCell>
				))}
			<TableCell className={[cellClassName, 'text-center align-middle']}>
				<Button
					data-testid="compute-ep"
					className="whitespace-nowrap"
					onClick={onComputeEp}
					{...tooltipAnchorProps(EP_TOOLTIP_ID, i18n.t('sidebar.buttons.stat_weights.modal.tooltips.compute_weighted_ep'))}>
					<Icon name="calculator" className="inline align-middle" />
					<span data-testid="not-tiny" className={clsx(showThreatMetrics && 'max-lg:hidden')}>
						{i18n.t('sidebar.buttons.stat_weights.modal.column_headers.update_ep_button')}
					</span>
				</Button>
			</TableCell>
		</TableRow>
	);
};
