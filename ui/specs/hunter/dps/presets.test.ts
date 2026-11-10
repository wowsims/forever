// The hunter builds set what they state and leave the rest as the page has it: a .siml build writes no bool it
// leaves off, and loading one leaves "in front of target" and item swap to the page.
import { isBuildActive } from '@app/preset_build_state';
import { applyBuild } from '@features/settings/model/apply_build';
import { Spec } from '@generated/proto/common';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { Builds } from './presets';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase());

// Every build goes through the same build() in presets.ts, so one stands for all.
describe('a hunter build', () => {
	const build = Builds[0];
	it('leaves "in front of target" on', async () => {
		const { player, host } = await makeIndividualSim({ spec: Spec.SpecHunter });
		player.setInFrontOfTarget(true);

		applyBuild(build, host);

		expect(player.getInFrontOfTarget()).toBe(true);
	});

	it('stays active once item swap is turned on after it loads', async () => {
		const { player, host } = await makeIndividualSim({ spec: Spec.SpecHunter });
		applyBuild(build, host);

		player.itemSwapSettings.setEnableItemSwap(true);

		expect(isBuildActive(build, host, [PresetConfigurationCategory.Settings])).toBe(true);
	});
});
