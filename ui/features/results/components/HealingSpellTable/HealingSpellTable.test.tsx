import { ActionId } from '@sim/proto/action_id';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { type HealingSpellRow, HealingSpellTableView } from './HealingSpellTable';

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', href: '', ready: true }) }));
vi.mock('@sim/proto/action_id/tooltip_data', () => ({ actionIdWowheadTooltipData: () => Promise.resolve('') }));

const spell = (name: string, rank: number, overrides: Partial<HealingSpellRow> = {}): HealingSpellRow => ({
	id: ActionId.empty(name),
	name,
	rank,
	periodic: false,
	absorb: false,
	critChance: 20,
	normal: { min: 100, avg: 120, max: 140 },
	critical: { min: 150, avg: 180, max: 210 },
	...overrides,
});
const names = () => screen.getAllByTestId('metrics-action-name').map(element => element.textContent);

describe('HealingSpellTableView', () => {
	it('groups ranks in the shared table and sorts spells alphabetically', async () => {
		render(<HealingSpellTableView spells={[spell('Renew', 2), spell('Flash Heal', 1), spell('Renew', 1)]} />);
		await act(async () => {});
		expect(names().map(name => name?.split(' (')[0])).toEqual(['Flash Heal', 'Renew', 'Renew', 'Renew']);
		const toggle = screen.getByRole('button', { name: 'Renew' });
		expect(toggle.getAttribute('aria-expanded')).toBe('true');
		fireEvent.click(toggle);
		expect(names()).toHaveLength(2);
		expect(screen.getByTestId('healing-spells-table').classList.contains('ui-metrics-table')).toBe(true);
	});

	it('allows a spec to put a spell ahead of the alphabetical order', async () => {
		render(<HealingSpellTableView spells={[spell('Flash Heal', 1), spell('Renew', 1)]} spellOrder={['Renew']} />);
		await act(async () => {});
		expect(names().map(name => name?.split(' (')[0])).toEqual(['Renew', 'Flash Heal']);
	});

	it('retains normal and critical tick averages, shield capacity and custom metric tooltips', async () => {
		render(
			<HealingSpellTableView
				spells={[
					spell('Renew', 1, { periodic: true }),
					spell('Shield', 1, { absorb: true, critical: undefined, customMetrics: { test_proc: { min: 10, avg: 20, max: 30 } } }),
				]}
			/>,
		);
		await act(async () => {});
		const table = screen.getByTestId('healing-spells-table');
		expect(within(table).getAllByText('120')).toHaveLength(2);
		expect(within(table).getByText('180')).toBeTruthy();
		expect(table.querySelector('[data-tooltip-id="healing-spell-custom-test_proc"]')?.textContent).toBeTruthy();
		expect(table.querySelectorAll('[data-tooltip-id="healing-spell-normal"]')).toHaveLength(2);
	});
});
