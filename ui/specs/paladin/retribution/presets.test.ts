// The Retribution builds load what their .siml files state: the Holy Shock build turns item swap on with the Manual
// Crowd Pummeler, the Unsealed Wrath build turns it off, and neither moves "in front of target".
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { isBuildActive } from '@app/preset_build_state';
import { applyBuild } from '@features/settings/model/apply_build';
import { ItemSlot, Spec } from '@generated/proto/common';
import { UIDatabase } from '@generated/proto/ui';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { DefaultBuild, HolyShockBuild } from './presets';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

// The real database, so the gear sets and the swap set find their items.
beforeAll(() => {
	const dbPath = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../../assets/database/db.json');
	serveDatabase(UIDatabase.fromJson(JSON.parse(readFileSync(dbPath, 'utf8'))));
});

const manualCrowdPummeler = 9449;

describe('a Retribution build', () => {
	it('turns item swap on with the Manual Crowd Pummeler for Holy Shock and off for Unsealed Wrath', async () => {
		const { player, host } = await makeIndividualSim({ spec: Spec.SpecRetributionPaladin });

		applyBuild(HolyShockBuild, host);
		expect(player.itemSwapSettings.getEnableItemSwap()).toBe(true);
		expect(player.itemSwapSettings.getItem(ItemSlot.ItemSlotMainHand)?.item.id).toBe(manualCrowdPummeler);

		applyBuild(DefaultBuild, host);
		expect(player.itemSwapSettings.getEnableItemSwap()).toBe(false);
	});

	it.each([DefaultBuild, HolyShockBuild])('leaves "in front of target" on and is active once $name loads', async build => {
		const { player, host } = await makeIndividualSim({ spec: Spec.SpecRetributionPaladin });
		player.setInFrontOfTarget(true);

		applyBuild(build, host);

		expect(player.getInFrontOfTarget()).toBe(true);
		expect(isBuildActive(build, host)).toBe(true);
	});
});
