import { applySavedSettings } from '@features/settings/model/saved_settings';
import { ItemSwap } from '@generated/proto/common';
import { DpsWarrior_Options } from '@generated/proto/warrior';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { settingsPresets } from './utils';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

describe('settingsPresets', () => {
	const rage = (startingRage: number) => DpsWarrior_Options.create({ classOptions: { startingRage } });

	it('keeps the spec options toggled after the item swap presets are built when one of them loads', async () => {
		const { player, host } = await makeIndividualSim();
		player.setSpecOptions(rage(0));
		const [swapPreset] = settingsPresets({ ...host.individualConfig.presets, itemSwaps: [{ name: 'Swap', itemSwap: ItemSwap.create() }] }, host);
		player.setSpecOptions(rage(50));

		applySavedSettings(host, swapPreset!.data);

		expect(player.getSpecOptions()).toEqual(rage(50));
	});
});
