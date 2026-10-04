import { AVOIDANCE_INFO } from '@testing/storybook/fixtures/character-stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { AvoidanceRow, type AvoidanceRowProps } from './AvoidanceRow';

const openRow = async (props: Partial<AvoidanceRowProps> = {}) => {
	const view = render(
		<Table>
			<tbody>
				<AvoidanceRow info={AVOIDANCE_INFO} withHolyShield={false} hasParry hasBlock {...props} />
			</tbody>
		</Table>,
	);
	fireEvent.mouseEnter(view.container.querySelector('[data-testid="stat-value-link"]')!);
	await screen.findByText('sidebar.character_stats.tank_caps.crush');
	return screen.getAllByTestId('character-stats-tooltip-row');
};

describe('AvoidanceRow', () => {
	it('breaks the total into miss, dodge, parry and block, and marks the crush cap met and the shear cap missed', async () => {
		const rows = await openRow();
		expect(rows.map(row => row.firstElementChild!.textContent)).toEqual([
			'sidebar.character_stats.tank_caps.miss',
			'sidebar.character_stats.tank_caps.dodge',
			'sidebar.character_stats.tank_caps.parry',
			'sidebar.character_stats.tank_caps.block',
			'sidebar.character_stats.tank_caps.crush',
			'sidebar.character_stats.tank_caps.shear',
		]);
		expect(rows[4].dataset.cap).toBe('met');
		expect(rows[5].dataset.cap).toBe('missed');
	});

	it('drops parry and block for a class without them, and suffixes block-derived figures with Holy Shield', async () => {
		const rows = await openRow({ hasParry: false, hasBlock: false, withHolyShield: true });
		expect(rows).toHaveLength(4);
		expect(rows[2].lastElementChild!.textContent).toBe('103.00%sidebar.character_stats.tank_caps.with_holy_shield');
	});

	it('shows a skeleton and no value while pending', () => {
		const { container } = render(
			<Table>
				<tbody>
					<AvoidanceRow info={AVOIDANCE_INFO} withHolyShield={false} hasParry hasBlock pending />
				</tbody>
			</Table>,
		);
		expect(container.querySelector('[data-testid="stat-value-link"]')).toBeNull();
	});
});
