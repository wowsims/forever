import { Icon } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

export interface ResultMetricValueProps {
	/** Already formatted. */
	average: string;
	/** The metric's short name, set small after the number (e.g. "DPS"). */
	unit: string;
	/** Already formatted; left out when absent. */
	margin?: string;
	marginTooltipId?: string;
	size?: 'md' | 'sm';
}

// A metric's value line as the results sidebar shows it: the number large, its unit and "(±margin)" small beside it.
export const ResultMetricValue = ({ average, unit, margin, marginTooltipId, size = 'md' }: ResultMetricValueProps) => (
	<>
		<span data-testid="topline-result-avg" className={size === 'sm' ? 'text-md' : 'text-2xl leading-none'}>
			{average}
			<span className="text-ui font-normal"> {unit}</span>
		</span>
		{margin !== undefined && (
			<span data-testid="topline-result-stdev" className="text-ui font-normal" {...tooltipAnchorProps(marginTooltipId)}>
				(<Icon name="plus-minus" size="xs" />
				{margin})
			</span>
		)}
	</>
);
