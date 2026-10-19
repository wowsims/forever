import type { StatWeightValues } from '@generated/proto/api';
import type { UnitStats } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { UnitStat } from '@sim/proto/stats';
import { TableCell } from '@ui-kit/Table';
import { toneTextClass } from '@ui-kit/utils/css';

import { StatWeightValue } from './StatWeightValue';

export interface StatWeightCellsProps {
	stat: UnitStat;
	statWeights?: StatWeightValues;
	iterations: number;
	epRatio: number;
	/** In the stored unit, like the weights; the cells show both divided by ratingScale. */
	epDelta: number;
	/** The rating points one stored weight unit covers; 1 shows the stored unit. */
	ratingScale?: number;
	cellClassName: string;
}

const NotApplicable = () => (
	<span data-testid="results-avg" data-notapplicable className="pr-6 font-bold">
		{i18n.t('sidebar.buttons.stat_weights.modal.not_applicable')}
	</span>
);

export const StatWeightCells = ({ stat, statWeights, iterations, epRatio, epDelta, ratingScale = 1, cellClassName }: StatWeightCellsProps) => {
	const shown = (values: UnitStats) => stat.getProtoValue(values) / ratingScale;
	const shownDelta = epDelta / ratingScale;
	const unused = !!statWeights && epRatio === 0;
	const rounded = shownDelta.toFixed(2);
	const delta = !statWeights || unused || rounded === '0.00' ? undefined : shownDelta > 0 ? 'positive' : 'negative';

	return (
		<>
			<TableCell
				testId="stdev-cell"
				data-column-type="weight"
				className={['text-right in-data-[stats-type=ep]:hidden', cellClassName, unused && 'text-white/70']}
				data-unused={unused ? '' : undefined}>
				{statWeights ? (
					<StatWeightValue value={shown(statWeights.weights!)} stdev={shown(statWeights.weightsStdev!)} iterations={iterations} />
				) : (
					<NotApplicable />
				)}
			</TableCell>
			<TableCell
				testId="stdev-cell"
				data-column-type="ep"
				className={['text-right in-data-[stats-type=weight]:hidden', cellClassName, unused && 'text-white/70']}
				data-unused={unused ? '' : undefined}>
				{statWeights ? (
					<StatWeightValue
						value={shown(statWeights.epValues!)}
						stdev={shown(statWeights.epValuesStdev!)}
						iterations={iterations}
						className={toneTextClass(delta)}
						sign={delta}
					/>
				) : (
					<NotApplicable />
				)}
			</TableCell>
		</>
	);
};
