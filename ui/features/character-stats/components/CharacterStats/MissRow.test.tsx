import { MISS_INFO } from '@testing/storybook/fixtures/character-stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { MissRow } from './MissRow';

describe('MissRow', () => {
	it('shows the total and breaks it into base, defense and debuffs', async () => {
		const { container } = render(
			<Table>
				<tbody>
					<MissRow info={MISS_INFO} />
				</tbody>
			</Table>,
		);
		const link = container.querySelector('[data-testid="stat-value-link"]')!;
		expect(link.textContent).toBe('8.40%');

		fireEvent.mouseEnter(link);
		await screen.findByText('sidebar.character_stats.tank_caps.defense');
		expect(screen.getAllByTestId('character-stats-tooltip-row').map(row => [...row.children].map(cell => cell.textContent))).toEqual([
			['sidebar.character_stats.tooltip.base', '5.00%'],
			['sidebar.character_stats.tank_caps.defense', '2.40%'],
			['sidebar.character_stats.tooltip.debuffs', '1.00%'],
			['sidebar.character_stats.tooltip.total', '8.40%'],
		]);
	});
});
