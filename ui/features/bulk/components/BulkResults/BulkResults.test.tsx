import type { BulkEncounterResults, TopGearResult } from '@sim/bulk/types';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { patchBulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('./BulkResultRow', () => ({
	BulkResultRow: ({ iterations }: { iterations: number }) => <div data-testid="row" data-iterations={iterations} />,
}));

const { BulkResults } = await import('./BulkResults');

const STORE_KEY = 9;
const RUN_ITERATIONS = 30000;

const result = (avg: number) => ({ gear: {}, dpsMetrics: { avg, stdev: 1 } }) as unknown as TopGearResult;

const entry = (iterations: number): BulkEncounterResults => {
	const baseline = result(1000);
	return {
		name: '',
		isCurrent: true,
		durationLabel: '3:00 ±5s',
		summaryLabel: '3:00 ±5s · 1 target',
		duration: 180,
		durationVariation: 5,
		results: { chains: [[result(1100)], [baseline]], originalGearResults: baseline, iterations },
	};
};

const mount = (entries: BulkEncounterResults[]) => {
	const store = createSimStore();
	const player = { sim: { store, getIterations: () => 100 }, storeKey: STORE_KEY } as unknown as Player<any>;
	seedBulkSettings(player);
	patchBulkState(player, { started: true, results: { entries } });
	const host = { player, sim: player.sim } as never;
	return render(
		<SimHostProvider host={host}>
			<BulkResults onActivateTab={() => {}} selected={0} />
		</SimHostProvider>,
	);
};

const rowIterations = () => screen.getAllByTestId('row').map(row => row.getAttribute('data-iterations'));

describe('BulkResults', () => {
	it('gives every row the iteration count the run was scored with, not the sidebar value', () => {
		mount([entry(RUN_ITERATIONS)]);
		expect(rowIterations()).toEqual([String(RUN_ITERATIONS), String(RUN_ITERATIONS)]);
	});
});
