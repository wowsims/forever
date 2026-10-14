import { makeEpColumns } from '@testing/storybook/fixtures/stat-weights';
import { fireEvent, render } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it, vi } from 'vitest';

import { EpWeightsHeader } from './EpWeightsHeader';
import type { EpColumn } from './types';

const renderHeader = (columns: EpColumn[]) =>
	render(
		<Table>
			<thead>
				<EpWeightsHeader columns={columns} showThreatMetrics={false} />
			</thead>
		</Table>,
	);

describe('EpWeightsHeader', () => {
	it('tags each metric column with its type, and leaves Stat, Update and Current EP untagged', () => {
		const { container } = renderHeader(makeEpColumns());
		expect([...container.querySelectorAll('th')].map(th => th.getAttribute('data-column-type'))).toEqual([null, null, 'weight', 'ep', null]);
	});

	it('lets every column label wrap', () => {
		const { container } = renderHeader(makeEpColumns());
		expect([...container.querySelectorAll('th')].map(th => th.hasAttribute('data-wrap'))).toEqual([true, true, true, true, true]);
	});

	it('puts each column label on its tooltip anchor, and copies the column from its action button', () => {
		const onCopy = vi.fn();
		const columns = makeEpColumns().map((column, index) => (index === 1 ? { ...column, onCopy } : column));
		const { container } = renderHeader(columns);
		const dpsEp = container.querySelectorAll('th')[3];
		expect(dpsEp.querySelector('span')!.getAttribute('data-tooltip-content')).toContain('sidebar.buttons.stat_weights.modal.dps_ep.tooltip');

		fireEvent.click(dpsEp.querySelector('[data-testid="col-action"]')!);
		expect(onCopy).toHaveBeenCalledTimes(1);
	});

	it('puts the stat action in the Stat column header', () => {
		const { container } = render(
			<Table>
				<thead>
					<EpWeightsHeader columns={makeEpColumns()} showThreatMetrics={false} statAction={<button type="button">stat filter</button>} />
				</thead>
			</Table>,
		);
		expect(container.querySelectorAll('th')[0].querySelector('button')!.textContent).toBe('stat filter');
	});
});
