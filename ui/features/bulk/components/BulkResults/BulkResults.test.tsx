import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('./BulkResultRow', () => ({
	BulkResultRow: ({ iterations }: { iterations: number }) => <div data-testid="row" data-iterations={iterations} />,
}));

const { BulkResults } = await import('./BulkResults');

const STORE_KEY = 9;
const RUN_ITERATIONS = 30000;

const result = (avg: number) => ({ gear: {}, dpsMetrics: { avg, stdev: 1 } }) as unknown as TopGearResult;

const entry = (name: string, durationLabel: string, iterations: number, isCurrent = false): BulkEncounterResults => {
	const baseline = result(1000);
	return {
		name,
		isCurrent,
		durationLabel,
		results: { chains: [[result(1100)], [baseline]], originalGearResults: baseline, iterations },
	};
};

const mount = (plannedEncounters: number, entries: BulkEncounterResults[]) => {
	const store = createSimStore();
	const player = { sim: { store, getIterations: () => 100 }, storeKey: STORE_KEY } as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { started: true, results: { plannedEncounters, entries } });
	const host = { player, sim: player.sim } as never;
	return render(
		<SimHostProvider host={host}>
			<BulkResults onActivateTab={() => {}} />
		</SimHostProvider>,
	);
};

const rowIterations = () => screen.getAllByTestId('row').map(row => row.getAttribute('data-iterations'));

describe('BulkResults', () => {
	it('gives every row the iteration count the run was scored with, not the sidebar value', () => {
		mount(1, [entry('', '3:00 ±5s', RUN_ITERATIONS, true)]);
		expect(rowIterations()).toEqual([String(RUN_ITERATIONS), String(RUN_ITERATIONS)]);
	});

	it('shows no encounter tabs for a single-encounter run', () => {
		mount(1, [entry('', '3:00 ±5s', RUN_ITERATIONS, true)]);
		expect(screen.queryByTestId('bulk-encounter-tabs')).toBeNull();
	});

	it('names one tab per encounter, with its length, and switches the list on a click', () => {
		mount(3, [entry('', '3:00 ±5s', 100, true), entry('Patchwerk', '5:00 ±30s', 200), entry('Council', 'Health', 300)]);
		const group = screen.getByRole('group', { name: 'bulk_tab.results.encounters' });
		const tabs = within(group).getAllByRole('button');
		expect(tabs.map(tab => tab.textContent)).toEqual(['bulk_tab.results.current 3:00 ±5s', 'Patchwerk 5:00 ±30s', 'Council Health']);
		expect(tabs.map(tab => tab.getAttribute('aria-pressed'))).toEqual(['true', 'false', 'false']);
		expect(rowIterations()).toEqual(['100', '100']);

		fireEvent.click(tabs[1]);
		expect(rowIterations()).toEqual(['200', '200']);
		expect(tabs.map(tab => tab.getAttribute('aria-pressed'))).toEqual(['false', 'true', 'false']);
	});

	it('shows one tab for a cancelled multi-encounter run', () => {
		mount(3, [entry('', '3:00 ±5s', 100, true)]);
		expect(within(screen.getByTestId('bulk-encounter-tabs')).getAllByRole('button')).toHaveLength(1);
	});
});
