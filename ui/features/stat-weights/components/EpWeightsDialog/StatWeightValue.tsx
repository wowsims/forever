import { stDevToConf90 } from '@sim/utils/math';
import { Icon } from '@ui-kit/Icon';

export interface StatWeightValueProps {
	value: number;
	stdev: number;
	iterations: number;
	className?: string;
	sign?: 'positive' | 'negative';
}

export const StatWeightValue = ({ value, stdev, iterations, className, sign }: StatWeightValueProps) => (
	<>
		<span data-testid="results-avg" className={className} data-sign={sign}>
			{value.toFixed(2)}
		</span>{' '}
		<span className="whitespace-nowrap">
			{'('}
			<Icon name="plus-minus" />
			{stDevToConf90(stdev, iterations).toFixed(2)}
			{')'}
		</span>
	</>
);
