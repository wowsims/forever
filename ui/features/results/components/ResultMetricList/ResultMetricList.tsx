import { Icon } from '@ui-kit/Icon';
import { Table, TableCell, TableHeaderCell, TableRow } from '@ui-kit/Table';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId } from 'react';

import type { ReferenceDiffs } from '../../model/reference_diffs';
import { resultMetricCategories, type ResultMetrics } from '../../model/sim_results';
import type { ResultMetric } from '../../model/topline_metrics';
import { ResultMetricValue } from '../ResultMetricValue';
import { ResultReferenceDiff } from './ResultReferenceDiff';
import { formatAverage, formatStdev, hasMetricTooltip, type ResultMetricLayout, resultMetricTooltip } from './utils';

export interface ResultMetricListProps {
	metrics: Array<ResultMetric>;
	layout: ResultMetricLayout;
	referenceDiffs?: ReferenceDiffs;
	size?: 'md' | 'sm';
}

export const ResultMetricList = ({ metrics, layout, referenceDiffs, size = 'md' }: ResultMetricListProps) => {
	const tooltipId = useId();
	const diffTooltipId = useId();
	const anchorFor = (metric: keyof ResultMetrics) => tooltipAnchorProps(hasMetricTooltip(metric, layout) ? tooltipId : undefined);

	return (
		<>
			{layout === 'row' ? (
				<Table striped hover className="ui-metrics-table" testId="metrics-table">
					<thead data-testid="metrics-table-header">
						<TableRow testId="metrics-table-header-row">
							{metrics.map(({ metric, name, extraClass }) => (
								<TableHeaderCell
									key={metric}
									testId="metrics-table-header-cell"
									className={['ui-metrics-header-cell font-bold', extraClass]}
									data-metric={metric}
									data-metric-category={resultMetricCategories[metric]}
									{...anchorFor(metric)}>
									{name}
								</TableHeaderCell>
							))}
						</TableRow>
					</thead>
					<tbody data-testid="metrics-table-body">
						<TableRow>
							{metrics.map(column => (
								<TableCell key={column.metric} className={['ui-metrics-cell text-center align-top font-bold', column.extraClass]}>
									<div data-testid="topline-result-avg" className="text-2xl">
										{formatAverage(column, layout)}
									</div>
									{column.stdev ? (
										<div data-testid="topline-result-stdev" className="text-ui font-normal">
											<Icon name="plus-minus" size="xs" /> {formatStdev(column, layout)}
										</div>
									) : undefined}
									<ResultReferenceDiff diff={referenceDiffs?.[column.metric]} tooltipId={diffTooltipId} />
								</TableCell>
							))}
						</TableRow>
					</tbody>
				</Table>
			) : (
				metrics.map(column => (
					<div
						key={column.metric}
						data-testid="results-metric"
						className={clsx('text-left font-bold', column.extraClass)}
						data-metric={column.metric}
						data-metric-category={resultMetricCategories[column.metric]}
						data-size={size}
						{...anchorFor(column.metric)}>
						<ResultMetricValue
							average={formatAverage(column, layout)}
							unit={column.name}
							margin={column.stdev ? formatStdev(column, layout) : undefined}
							size={size}
						/>
						<ResultReferenceDiff diff={referenceDiffs?.[column.metric]} tooltipId={diffTooltipId} />
					</div>
				))
			)}
			<Tooltip
				id={tooltipId}
				place={layout === 'list' ? 'right' : 'top'}
				render={({ activeAnchor }) => resultMetricTooltip(activeAnchor?.getAttribute('data-metric') as keyof ResultMetrics | null, layout)}
			/>
			{referenceDiffs && <Tooltip id={diffTooltipId} />}
		</>
	);
};
