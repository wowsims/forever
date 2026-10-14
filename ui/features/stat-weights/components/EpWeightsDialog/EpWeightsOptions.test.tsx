import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

describe('EpWeightsOptions', () => {
	it('shows the stats type and reports a stats type change', async () => {
		const onStatsTypeChange = vi.fn();
		const { container } = render(<EpWeightsOptions statsType={StatsType.Weight} onStatsTypeChange={onStatsTypeChange} />);

		const select = container.querySelector<HTMLElement>('#ep-type-select')!;
		await userEvent.click(select);
		expect(screen.getAllByRole('option').map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'true']);
		await userEvent.click(screen.getAllByRole('option')[0]);
		expect(onStatsTypeChange).toHaveBeenCalledWith(StatsType.Ep);
	});
});
