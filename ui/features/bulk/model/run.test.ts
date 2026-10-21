import { saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions, BulkSettings } from '@generated/proto/api';
import { Encounter, ItemSlot, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter, UIEnchant } from '@generated/proto/ui';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { beforeEach, describe, expect, it, type MockInstance, vi } from 'vitest';

import { runCoreBulkSim } from './core_sim';
import { bulkProgress, cancelBulkBatch, refreshBulkCombinations, runBulkBatch } from './run';

vi.mock('./core_sim', () => ({ runCoreBulkSim: vi.fn() }));
vi.mock('../../../tracking/utils', () => ({ trackEvent: () => {} }));

const AGILITY = 10;

describe('bulk requests', () => {
	it('carry the enchant pool', async () => {
		const player = makeBulkPlayer({
			getGear: () => ({ getEquippedItem: () => null }),
			getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: AGILITY })] : []),
		});
		saveEnchantPreset(player, { name: 'Undead', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }] });
		setEnchantPresetChecked(player, 'Undead', true);
		const getBulkCombinationCount = vi.fn(async (_settings: BulkSettings) => ({ combinations: 2, iterations: 1 }));

		await refreshBulkCombinations({
			player,
			sim: { getBulkCombinationCount },
			handleCrash: (e: unknown) => Promise.reject(e),
		} as unknown as IndividualSimHost<any>);

		expect(getBulkCombinationCount.mock.calls[0][0].enchants).toEqual([
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }),
		]);
	});
});

const GEAR = { asSpec: () => ({ items: [] }) };
const CORE_RESULT = { referenceDpsMetrics: { avg: 1000, stdev: 1 }, topGearResults: [], iterations: 100, metrics: {} };
const SAVED_KEY = 'saved-encounters';

const CURRENT = Encounter.create({ duration: 180, durationVariation: 5, targets: [Target.create()] });
const PATCHWERK = Encounter.create({ duration: 300, durationVariation: 30, useHealth: true, targets: [Target.create()] });
const COUNCIL = Encounter.create({ duration: 90, durationVariation: 10, targets: [Target.create()] });

const ref = (name: string, preset = false) => BulkEncounterRef.create({ name, preset });

const batchHost = (checked: BulkEncounterRef[] = [ref('Patchwerk', true), ref('Same as current'), ref('Gone'), ref('Council')], powerTools = true) => {
	const player = makeBulkPlayer({ getGear: () => GEAR, setGearAsync: vi.fn(async () => {}) });
	patchSlice(player.sim.store, 'ui', { showPowerTools: powerTools });
	const saved = { 'Same as current': CURRENT, Council: COUNCIL };
	player.sim.env.storage.setItem(
		SAVED_KEY,
		JSON.stringify(
			Object.fromEntries(Object.entries(saved).map(([name, encounter]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter }))])),
		),
	);
	patchBulkState(player, { encounters: checked });
	return {
		player,
		sim: {
			store: player.sim.store,
			env: player.sim.env,
			encounter: { toProto: () => Encounter.clone(CURRENT) },
			waitForInit: async () => {},
			signalManager: { abortType: vi.fn(async () => {}) },
			getBulkCombinationCount: async () => ({ combinations: 1, iterations: 1 }),
			getShowPowerTools: () => player.sim.store.getState().ui.showPowerTools,
		},
		individualConfig: { presets: { encounters: [{ name: 'Patchwerk', encounter: PATCHWERK }] } },
		getSavedEncounterStorageKey: () => SAVED_KEY,
		handleCrash: (e: unknown) => Promise.reject(e),
	} as unknown as IndividualSimHost<any>;
};

const overrides = () => vi.mocked(runCoreBulkSim).mock.calls.map(call => call[3]);

const steps = (host: IndividualSimHost<any>) => bulkProgress(host.player)!.steps;

describe('runBulkBatch', () => {
	let consoleError: MockInstance;

	beforeEach(() => {
		vi.mocked(runCoreBulkSim).mockReset();
		consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
	});

	it('runs the current encounter, then each checked one, skipping duplicates and deleted refs', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(overrides()).toEqual([CURRENT, PATCHWERK, COUNCIL]);
		const results = bulkState(host.player).results!;
		expect(results.map(entry => [entry.name, entry.isCurrent])).toEqual([
			['', true],
			['Patchwerk', false],
			['Council', false],
		]);
		expect(results.map(entry => [entry.duration, entry.durationVariation])).toEqual([
			[180, 5],
			[null, null],
			[90, 10],
		]);
	});

	it('runs once when nothing is checked', async () => {
		const host = batchHost([]);
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(overrides()).toEqual([CURRENT]);
		expect(bulkState(host.player).results).toHaveLength(1);
	});

	it('runs only the current encounter with power tools off, keeping the checked ones for when they are back on', async () => {
		const checked = [ref('Patchwerk', true), ref('Council')];
		const host = batchHost(checked, false);
		vi.mocked(runCoreBulkSim).mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(overrides()).toEqual([CURRENT]);
		expect(steps(host).labels).toHaveLength(1);
		expect(bulkState(host.player).results!.map(entry => entry.isCurrent)).toEqual([true]);
		expect(bulkState(host.player).encounters).toEqual(checked);

		vi.mocked(runCoreBulkSim).mockClear();
		patchSlice(host.player.sim.store, 'ui', { showPowerTools: true });
		await runBulkBatch(host);

		expect(overrides()).toEqual([CURRENT, PATCHWERK, COUNCIL]);
	});

	it('keeps the finished encounters when cancelled', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockImplementationOnce(async () => {
				await cancelBulkBatch(host);
				throw new Error('Bulk Sim Aborted');
			});

		await runBulkBatch(host);

		expect(bulkState(host.player).results!.map(entry => entry.isCurrent)).toEqual([true]);
		expect(bulkState(host.player).isRunning).toBe(false);
	});

	it('keeps the finished encounters when a later one fails', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim)
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		expect(bulkState(host.player).results!.map(entry => entry.isCurrent)).toEqual([true]);
	});

	it('leaves no results when the first encounter fails', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim).mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		expect(bulkState(host.player).results).toBeNull();
	});

	it('lists every encounter as a step from the start and moves the running one along', async () => {
		const host = batchHost();
		const seen: Array<[number, number]> = [];
		const record = () => {
			const { labels, running } = steps(host);
			seen.push([labels.length, running]);
		};
		// The batch is marked running before it waits for the sim, so this is what the progress dialog finds as it mounts.
		host.sim.waitForInit = async () => record();
		vi.mocked(runCoreBulkSim).mockImplementation(async () => {
			record();
			return CORE_RESULT as never;
		});

		await runBulkBatch(host);

		expect(seen).toEqual([
			[3, 0],
			[3, 0],
			[3, 1],
			[3, 2],
		]);
		const { labels } = steps(host);
		expect(labels[1]).toBe('Patchwerk · bulk_tab.settings.encounters.health_summary · bulk_tab.settings.encounters.targets');
	});
});
