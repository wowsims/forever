import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

describe('EpWeightsOptions', () => {
	it('shows the stats type and the show-all pick, and reports a change to either', async () => {
		const onStatsTypeChange = vi.fn();
		const onShowAllStatsChange = vi.fn();
		const { container } = render(
			<EpWeightsOptions
				statsType={StatsType.Weight}
				showAllStats={false}
				onStatsTypeChange={onStatsTypeChange}
				onShowAllStatsChange={onShowAllStatsChange}
			/>,
		);
		const select = container.querySelector<HTMLElement>('#ep-type-select')!;
		await userEvent.click(select);
		expect(screen.getAllByRole('option').map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'true']);
		await userEvent.click(screen.getAllByRole('option')[0]);
		expect(onStatsTypeChange).toHaveBeenCalledWith(StatsType.Ep);

		fireEvent.click(container.querySelector('#ep-show-all-stats')!);
		expect(onShowAllStatsChange).toHaveBeenCalledWith(true);
	});
});
