import { makeEpColumns } from '@testing/storybook/fixtures/stat-weights';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { EpWeightsTableView, type EpWeightsTableViewProps } from './EpWeightsTable';
import { StatsType } from './types';

const renderTable = (props: Partial<EpWeightsTableViewProps> = {}) =>
	render(
		<EpWeightsTableView
			columns={makeEpColumns()}
			statsType={StatsType.Ep}
			epRatios={[1, 0, 0, 0, 0, 0]}
			onEpRatioChange={() => {}}
			onComputeEp={() => {}}
			showThreatMetrics={false}
			showEpRatios
			{...props}>
			<tr data-testid="stat-row" />
		</EpWeightsTableView>,
	);

describe('EpWeightsTableView', () => {
	it('tags the table with its stats type and puts the rows in the body', () => {
		const { getByTestId } = renderTable({ statsType: StatsType.Weight });
		const table = getByTestId('results-ep-table');
		expect(table.getAttribute('data-stats-type')).toBe('weight');
		expect(table.querySelector('tbody [data-testid="stat-row"]')).not.toBeNull();
	});

	it('heads the table with Stat, Update, the metric columns and Current EP', () => {
		const { getByTestId } = renderTable();
		const headers = [...getByTestId('results-ep-table').querySelectorAll('thead tr:first-child th')].map(th => th.textContent);
		expect(headers).toEqual([
			'sidebar.buttons.stat_weights.modal.column_headers.stat',
			'sidebar.buttons.stat_weights.modal.column_headers.update',
			'sidebar.buttons.stat_weights.modal.dps_weight.label',
			'sidebar.buttons.stat_weights.modal.dps_ep.label',
			'sidebar.buttons.stat_weights.modal.current_ep.label',
		]);
	});

	it('drops the Update column for a gear planner, and the ratio row when it is not shown', () => {
		const { getByTestId } = renderTable({ showUpdateColumn: false, showEpRatios: false });
		const table = getByTestId('results-ep-table');
		expect(table.querySelectorAll('thead tr')).toHaveLength(1);
		expect(table.querySelectorAll('thead th')).toHaveLength(4);
	});
});
