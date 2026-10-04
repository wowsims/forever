import { makeEpColumns } from '@testing/storybook/fixtures/stat-weights';
import { fireEvent, render } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it, vi } from 'vitest';

import { EpRatiosRow } from './EpRatiosRow';

describe('EpRatiosRow', () => {
	it('gives each metric column a ratio input at its index, and reports an edit to it and the Update EP press', () => {
		const onEpRatioChange = vi.fn();
		const onComputeEp = vi.fn();
		const { getByTestId } = render(
			<Table>
				<thead>
					<EpRatiosRow
						columns={makeEpColumns({ damage: true, healing: true, threat: false })}
						epRatios={[1, 0.5, 0, 0, 0, 0]}
						onEpRatioChange={onEpRatioChange}
						onComputeEp={onComputeEp}
						showThreatMetrics={false}
					/>
				</thead>
			</Table>,
		);
		const inputs = [...getByTestId('ep-ratios').querySelectorAll('input')];
		expect(inputs.map(input => input.id)).toEqual(['ep-ratio-weight-0', 'ep-ratio-ep-0', 'ep-ratio-weight-1', 'ep-ratio-ep-1']);
		expect(inputs.map(input => input.value)).toEqual(['1.00', '1.00', '0.50', '0.50']);

		inputs[3].value = '2';
		fireEvent.change(inputs[3]);
		expect(onEpRatioChange).toHaveBeenCalledWith(1, 2);

		fireEvent.click(getByTestId('ep-ratios').querySelector('[data-testid="compute-ep"]')!);
		expect(onComputeEp).toHaveBeenCalledTimes(1);
	});
});
