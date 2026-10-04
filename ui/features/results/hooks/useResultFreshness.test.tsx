import { Raid as RaidProto } from '@generated/proto/api';
import { Encounter as EncounterProto, EquipmentSpec, ItemSpec } from '@generated/proto/common';
import { SimRuns } from '@sim/sim_runs';
import { createSimStore, patchKeyed, patchSlice, PLAYER_FIELDS, seedKeyed, SimRunKind, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { SimResultsManager } from '../model/results_manager';
import { useResultFreshness } from './useResultFreshness';

const PLAYER = 1;

let store: SimStore;
let runs: SimRuns;
let results: SimResultsManager;

const Probe = () => <span>{String(useResultFreshness())}</span>;

const mount = () => renderWithHost(<Probe />, { sim: { store } });

// A run lands its result from inside the run, as the sim's result event does.
const simulate = (duringRun: () => void = () => {}) =>
	act(() =>
		runs.start(SimRunKind.IndividualSim, async () => {
			duringRun();
			results.setSimResult({ request: {} } as never);
		}),
	);

// Only `asSpec` is read: gear compares by the batch sim's gear identity key.
const gearWith = (itemId: number) => ({ asSpec: () => EquipmentSpec.create({ items: [ItemSpec.create({ id: itemId })] }) }) as never;
const equip = (itemId: number) => patchKeyed(store, 'players', PLAYER, { gear: gearWith(itemId) }, ['gear']);
const changePlayer = (field: 'gear' | 'talentsString') => (field === 'gear' ? equip(2) : patchKeyed(store, 'players', PLAYER, {}, [field]));

beforeEach(() => {
	store = createSimStore();
	seedKeyed(store, 'players', PLAYER, { gear: gearWith(1), v: zeroVersions(PLAYER_FIELDS) } as never);
	runs = new SimRuns(store, { abortType: async () => {} });
	results = new SimResultsManager({
		store,
		raid: { toProto: () => RaidProto.create() },
		encounter: { toProto: () => EncounterProto.create() },
	} as never);
});

describe('useResultFreshness', () => {
	it('is null before a result exists', () => {
		expect(mount().container.textContent).toBe('null');
	});

	it('is fresh right after a run', async () => {
		const { container } = mount();
		await simulate();
		expect(container.textContent).toBe('fresh');
	});

	it('turns stale after a gear change', async () => {
		const { container } = mount();
		await simulate();
		act(() => changePlayer('gear'));
		expect(container.textContent).toBe('stale');
	});

	it('reads fresh again once a gear change is undone', async () => {
		const { container } = mount();
		await simulate();
		act(() => equip(2));
		expect(container.textContent).toBe('stale');

		act(() => equip(1));
		expect(container.textContent).toBe('fresh');
	});

	it('turns stale after a talent change', async () => {
		const { container } = mount();
		await simulate();
		act(() => changePlayer('talentsString'));
		expect(container.textContent).toBe('stale');
	});

	it('lands stale when an input changed while the run was going', async () => {
		const { container } = mount();
		await simulate(() => changePlayer('gear'));
		expect(container.textContent).toBe('stale');
	});

	it('stays fresh after a UI setting and an iterations change', async () => {
		const { container } = mount();
		await simulate();
		act(() => patchSlice(store, 'ui', { showEPValues: true }));
		act(() => patchSlice(store, 'sim', { iterations: 1000 }));
		expect(container.textContent).toBe('fresh');
	});

	it('is fresh again after the next run', async () => {
		const { container } = mount();
		await simulate();
		act(() => changePlayer('gear'));
		await simulate();
		expect(container.textContent).toBe('fresh');
	});
});
