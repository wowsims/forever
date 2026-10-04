import { makeStatWeightsResult, STAT_WEIGHT_STATS } from '@testing/storybook/fixtures/stat-weights';
import { render } from '@testing-library/react';
import { Table, TableRow } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { StatWeightCells, type StatWeightCellsProps } from './StatWeightCells';

const renderCells = (props: Partial<StatWeightCellsProps>) =>
	render(
		<Table>
			<tbody>
				<TableRow>
					<StatWeightCells stat={STAT_WEIGHT_STATS[1]} iterations={1000} epRatio={1} epDelta={0} cellClassName="" {...props} />
				</TableRow>
			</tbody>
		</Table>,
	);

const epValue = (container: HTMLElement) => container.querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!;

describe('StatWeightCells', () => {
	it('shows the weight and the EP, signing the EP against the current weight', () => {
		const statWeights = makeStatWeightsResult().dps;
		const { container, rerender } = renderCells({ statWeights, epDelta: 0.3 });
		expect(container.querySelector('[data-column-type="weight"] [data-testid="results-avg"]')!.textContent).toBe('1.90');
		expect(epValue(container).getAttribute('data-sign')).toBe('positive');

		rerender(
			<Table>
				<tbody>
					<TableRow>
						<StatWeightCells stat={STAT_WEIGHT_STATS[1]} statWeights={statWeights} iterations={1000} epRatio={1} epDelta={0.001} cellClassName="" />
					</TableRow>
				</tbody>
			</Table>,
		);
		expect(epValue(container).hasAttribute('data-sign')).toBe(false);
	});

	it('marks both cells unused when the metric has no EP ratio, and leaves them unsigned', () => {
		const { container } = renderCells({ statWeights: makeStatWeightsResult().dps, epRatio: 0, epDelta: 1 });
		const cells = [...container.querySelectorAll('[data-testid="stdev-cell"]')];
		expect(cells.map(cell => cell.hasAttribute('data-unused'))).toEqual([true, true]);
		expect(epValue(container).hasAttribute('data-sign')).toBe(false);
	});

	it('reads N/A in both cells without a result', () => {
		const { container } = renderCells({});
		expect(container.querySelectorAll('[data-notapplicable]')).toHaveLength(2);
	});
});
