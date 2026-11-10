import { CRIT_CAP_INFO } from '@testing/storybook/fixtures/character-stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { CritCapRow } from './CritCapRow';

const renderRow = (info = CRIT_CAP_INFO) =>
	render(
		<Table>
			<tbody>
				<CritCapRow info={info} text="8.40% under" />
			</tbody>
		</Table>,
	);

const openRows = async (container: HTMLElement) => {
	fireEvent.mouseEnter(container.querySelector('[data-testid="stat-value-link"]')!);
	await screen.findByText('sidebar.character_stats.attack_table.can_raise_by');
	return screen.getAllByTestId('character-stats-tooltip-row').map(row => [...row.children].map(cell => cell.textContent));
};

describe('CritCapRow', () => {
	it('shows the text, and a breakdown ending in how far hit and expertise can raise the cap', async () => {
		const { container } = renderRow();
		const link = container.querySelector('[data-testid="stat-value-link"]')!;
		expect(link.textContent).toBe('8.40% under ');

		const rows = await openRows(container);
		const labels = rows.map(([label]) => label);
		expect(labels).not.toContain('sidebar.character_stats.attack_table.spec_offsets');
		expect(labels).not.toContain('sidebar.character_stats.attack_table.block');
		expect(rows).toContainEqual(['sidebar.character_stats.attack_table.final_crit_cap', '61.10%']);
		expect(rows.at(-1)).toEqual(['sidebar.character_stats.attack_table.can_raise_by', '3.70%']);
	});

	it('adds the spec offset row when there is one', async () => {
		const { container } = renderRow({ ...CRIT_CAP_INFO, specOffset: 1.5 });
		expect(await openRows(container)).toContainEqual(['sidebar.character_stats.attack_table.spec_offsets', '1.50%']);
	});

	it('adds the block row when the target blocks, which hit and expertise cannot raise', async () => {
		const { container } = renderRow({ ...CRIT_CAP_INFO, block: 5 });
		const rows = await openRows(container);
		expect(rows).toContainEqual(['sidebar.character_stats.attack_table.block', '5.00%']);
		expect(rows.at(-1)).toEqual(['sidebar.character_stats.attack_table.can_raise_by', '3.70%']);
	});

	it('colours the value by whether the player is under or over the crit cap', () => {
		const under = renderRow(CRIT_CAP_INFO);
		expect(under.container.querySelector('[data-tone]')!.getAttribute('data-tone')).toBe('positive');

		const over = renderRow({ ...CRIT_CAP_INFO, delta: 8.4 });
		expect(over.container.querySelector('[data-tone]')!.getAttribute('data-tone')).toBe('negative');
	});
});
