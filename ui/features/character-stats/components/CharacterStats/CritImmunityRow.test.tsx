import { CRIT_IMMUNITY_INFO } from '@testing/storybook/fixtures/character-stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { CritImmunityRow } from './CritImmunityRow';

const openRows = async (info = CRIT_IMMUNITY_INFO) => {
	const { container } = render(
		<Table>
			<tbody>
				<CritImmunityRow info={info} text="Immune" />
			</tbody>
		</Table>,
	);
	fireEvent.mouseEnter(container.querySelector('[data-testid="stat-value-link"]')!);
	await screen.findByText('sidebar.character_stats.tank_caps.defense');
	return screen.getAllByTestId('character-stats-tooltip-row').map(row => row.firstElementChild!.textContent);
};

describe('CritImmunityRow', () => {
	it('breaks the total into defense and talents', async () => {
		expect(await openRows()).toEqual([
			'sidebar.character_stats.tank_caps.defense',
			'sidebar.character_stats.tooltip.talents',
			'sidebar.character_stats.tooltip.total',
		]);
	});

	it('leaves out the talents row without a talent contribution', async () => {
		expect(await openRows({ ...CRIT_IMMUNITY_INFO, talents: 0 })).toEqual([
			'sidebar.character_stats.tank_caps.defense',
			'sidebar.character_stats.tooltip.total',
		]);
	});
});
