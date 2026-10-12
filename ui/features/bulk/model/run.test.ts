import { saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions, BulkSettings } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { BulkDuration, UIEnchant } from '@generated/proto/ui';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { describe, expect, it, vi } from 'vitest';

import { runCoreBulkSim } from './core_sim';
import { cancelBulkBatch, refreshBulkCombinations, runBulkBatch } from './run';

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

const batchHost = () => {
	const player = makeBulkPlayer({ getGear: () => GEAR, setGearAsync: vi.fn(async () => {}) });
	patchSlice(player.sim.store, 'encounter', { duration: 180, durationVariation: 5, useHealth: false });
	patchBulkState(player, {
		durations: [BulkDuration.create({ duration: 90, durationVariation: 10 }), BulkDuration.create({ duration: 300, durationVariation: 30 })],
	});
	return {
		player,
		sim: {
			store: player.sim.store,
			waitForInit: async () => {},
			signalManager: { abortType: vi.fn(async () => {}) },
			getBulkCombinationCount: async () => ({ combinations: 1, iterations: 1 }),
		},
		handleCrash: (e: unknown) => Promise.reject(e),
	} as unknown as IndividualSimHost<any>;
};

describe('runBulkBatch', () => {
	it('runs the encounter, then each extra duration', async () => {
		const host = batchHost();
		vi.mocked(runCoreBulkSim)
			.mockReset()
			.mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(vi.mocked(runCoreBulkSim).mock.calls.map(call => call[3]?.encounterOverride)).toEqual([
			{ duration: 180, durationVariation: 5 },
			{ duration: 90, durationVariation: 10 },
			{ duration: 300, durationVariation: 30 },
		]);
		const results = bulkState(host.player).results!;
		expect(results.plannedDurations).toBe(3);
		expect(results.entries.map(entry => [entry.duration, entry.isEncounter])).toEqual([
			[180, true],
			[90, false],
			[300, false],
		]);
	});

	it('keeps the finished durations when cancelled', async () => {
		const host = batchHost();
		const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
		vi.mocked(runCoreBulkSim)
			.mockReset()
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockImplementationOnce(async () => {
				await cancelBulkBatch(host);
				throw new Error('Bulk Sim Aborted');
			});

		await runBulkBatch(host);
		consoleError.mockRestore();

		const results = bulkState(host.player).results!;
		expect(results.plannedDurations).toBe(3);
		expect(results.entries.map(entry => entry.duration)).toEqual([180]);
		expect(bulkState(host.player).isRunning).toBe(false);
	});

	it('keeps the finished durations when a later one fails', async () => {
		const host = batchHost();
		const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
		vi.mocked(runCoreBulkSim)
			.mockReset()
			.mockResolvedValueOnce(CORE_RESULT as never)
			.mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		consoleError.mockRestore();

		expect(bulkState(host.player).results!.entries.map(entry => entry.duration)).toEqual([180]);
	});

	it('leaves no results when the first duration fails', async () => {
		const host = batchHost();
		const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
		vi.mocked(runCoreBulkSim).mockReset().mockRejectedValueOnce(new Error('worker crashed'));

		await runBulkBatch(host);

		expect(consoleError).toHaveBeenCalled();
		consoleError.mockRestore();

		expect(bulkState(host.player).results).toBeNull();
	});

	it('runs once for a health fight', async () => {
		const host = batchHost();
		patchSlice(host.player.sim.store, 'encounter', { useHealth: true });
		vi.mocked(runCoreBulkSim)
			.mockReset()
			.mockResolvedValue(CORE_RESULT as never);

		await runBulkBatch(host);

		expect(vi.mocked(runCoreBulkSim)).toHaveBeenCalledTimes(1);
		expect(bulkState(host.player).results!.plannedDurations).toBe(1);
	});
});
