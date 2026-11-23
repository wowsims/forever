import { ResultReferenceDiff } from '@features/results/components/ResultMetricList';
import { ResultMetricValue } from '@features/results/components/ResultMetricValue';
import type { ReferenceDiff } from '@features/results/model/reference_diffs';
import i18n from '@i18n/config';
import { formatToNumber } from '@sim/utils/format';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId } from 'react';

import { formatBulkDps } from '../../model/results/format';

export interface BulkDpsCellProps {
	/** Sizes the cell in its row. */
	className?: string;
	testId?: string;
	/** A line above the DPS, such as "Current gear". */
	label?: string;
	avg: number;
	/** The 95% confidence interval of the row's own mean; zero or null leaves it out. */
	plusMinus: number | null;
	/** The difference to the row it is measured against; null shows none. */
	diff: ReferenceDiff | null;
	/** A row stopped before the final stage: a note under the DPS, whose tooltip takes the margin's place. */
	lowPrecision?: { note: string; tooltip: string };
}

// A result row's DPS with its margin and its difference, each with a tooltip.
export const BulkDpsCell = ({ className, testId, label, avg, plusMinus, diff, lowPrecision }: BulkDpsCellProps) => {
	const marginTooltipId = useId();
	const deltaTooltipId = useId();

	return (
		<div className={clsx('flex flex-col items-start gap-2 text-left font-bold', className)} data-testid={testId}>
			{label && <span className="text-sm font-normal">{label}</span>}
			<div>
				<ResultMetricValue
					average={formatBulkDps(avg)}
					unit={i18n.t('sidebar.results.metrics.dps.label')}
					margin={plusMinus !== null && plusMinus > 0 ? formatToNumber(plusMinus, { maximumFractionDigits: 0 }) : undefined}
					marginTooltipId={marginTooltipId}
				/>
			</div>
			{diff && <ResultReferenceDiff diff={diff} tooltipId={deltaTooltipId} hideLabel />}
			{lowPrecision && (
				<span className="text-xs font-normal text-warning" {...tooltipAnchorProps(marginTooltipId)}>
					{lowPrecision.note}
				</span>
			)}
			<Tooltip id={marginTooltipId} content={lowPrecision ? lowPrecision.tooltip : i18n.t('bulk_tab.results.margin_of_error')} />
			<Tooltip id={deltaTooltipId} />
		</div>
	);
};
