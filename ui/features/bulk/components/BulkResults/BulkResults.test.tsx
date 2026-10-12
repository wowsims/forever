import type { BulkDurationResults, TopGearResult } from '@sim/bulk/types';
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

const entry = (duration: number, iterations: number, isEncounter = false): BulkDurationResults => {
	const baseline = result(1000);
	return {
		duration,
		durationVariation: 5,
		isEncounter,
		results: { chains: [[result(1100)], [baseline]], originalGearResults: baseline, iterations },
	};
};

const mount = (plannedDurations: number, entries: BulkDurationResults[]) => {
	const store = createSimStore();
	const player = { sim: { store, getIterations: () => 100 }, storeKey: STORE_KEY } as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { started: true, results: { plannedDurations, entries } });
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
		mount(1, [entry(180, RUN_ITERATIONS, true)]);
		expect(rowIterations()).toEqual([String(RUN_ITERATIONS), String(RUN_ITERATIONS)]);
	});

	it('shows no pills for a single-duration run', () => {
		mount(1, [entry(180, RUN_ITERATIONS, true)]);
		expect(screen.queryByTestId('bulk-duration-pills')).toBeNull();
	});

	it('shows one pill per duration and switches the list on a click', () => {
		mount(3, [entry(180, 100, true), entry(90, 200), entry(300, 300)]);
		const pills = within(screen.getByTestId('bulk-duration-pills')).getAllByRole('button');
		expect(pills.map(pill => pill.textContent)).toEqual(['3:00 ±5sbulk_tab.results.encounter_duration', '1:30 ±5s', '5:00 ±5s']);
		expect(rowIterations()).toEqual(['100', '100']);

		fireEvent.click(pills[1]);
		expect(rowIterations()).toEqual(['200', '200']);
		expect(pills[1].getAttribute('aria-pressed')).toBe('true');
	});

	it('shows one pill for a cancelled multi-duration run', () => {
		mount(3, [entry(180, 100, true)]);
		expect(screen.getByTestId('bulk-duration-pills').querySelectorAll('button')).toHaveLength(1);
	});
});
