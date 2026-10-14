import i18n from '@i18n/config';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { useMemo } from 'react';

import { useSimResult } from '../../hooks/useSimResult';
import { targetOptions } from './utils';

export interface ResultsFilterProps {
	/** `ALL_UNITS`, or the encounter index of one target. */
	target: number;
	onTargetChange: (target: number) => void;
}

/**
 * The detailed-results toolbar's target filter. It owns the option list, which only exists once a
 * run has produced targets; the selection itself belongs to the pane, because the pane is what
 * re-emits the result every table reads through it.
 */
export const ResultsFilter = ({ target, onTargetChange }: ResultsFilterProps) => {
	const resultData = useSimResult();
	const options = useMemo(() => (resultData ? targetOptions(resultData.result) : []), [resultData]);

	return (
		<div data-testid="results-filter-root" className="flex items-center text-lg">
			<ChoiceField
				id="results-filter-target-filter"
				className={['py-4', !resultData && 'hidden']}
				options={options}
				value={target}
				onChange={onTargetChange}
				disabled={!resultData}
				variant="link"
				placeholder={i18n.t('results_tab.details.unit')}
			/>
		</div>
	);
};
