import { ConsumesSpec, ItemSlot } from '@generated/proto/common';
import { SavedSettings } from '@generated/proto/ui';
import { DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID } from '@sim/proto/sim_tables_auto_gen';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { MACE, swapSet, swapWeaponIds, SWORD, weaponsDatabase } from '@testing/sim/weapons';
import { makeEquipmentSpec } from '@testing/storybook/fixtures/protos';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { applySavedSettings } from './saved_settings';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase(weaponsDatabase()));

describe('applySavedSettings', () => {
	it("keeps the swap stones for the entry's swap weapons when the live swap set has none", async () => {
		const { player, host } = await makeIndividualSim();

		applySavedSettings(
			host,
			SavedSettings.create({
				enableItemSwap: true,
				itemSwap: swapSet({ mainHand: SWORD, offHand: MACE }),
				consumables: ConsumesSpec.create({ swapMhImbueId: DENSE_SHARPENING_STONE_ID, swapOhImbueId: DENSE_WEIGHTSTONE_ID }),
			}),
		);

		expect(player.itemSwapSettings.getEnableItemSwap()).toBe(true);
		expect(swapWeaponIds(player)).toEqual([SWORD, MACE]);
		expect([player.getConsumes().swapMhImbueId, player.getConsumes().swapOhImbueId]).toEqual([DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID]);
	});

	it("matches each stone to its weapon: the swap hand to the entry's swap weapon, the main hand to the gear", async () => {
		const { sim, player, host } = await makeIndividualSim();
		player.setGear(sim.db.lookupEquipmentSpec(makeEquipmentSpec({ [ItemSlot.ItemSlotMainHand]: { id: MACE } })));
		player.itemSwapSettings.setItemSwapSettings(true, sim.db.lookupItemSwap(swapSet({ mainHand: SWORD })));
		player.setConsumes(ConsumesSpec.create({ mhImbueId: DENSE_WEIGHTSTONE_ID, swapMhImbueId: DENSE_SHARPENING_STONE_ID }));

		applySavedSettings(
			host,
			SavedSettings.create({
				enableItemSwap: true,
				itemSwap: swapSet({ mainHand: MACE }),
				consumables: ConsumesSpec.create({ mhImbueId: DENSE_SHARPENING_STONE_ID, swapMhImbueId: DENSE_SHARPENING_STONE_ID }),
			}),
		);

		expect(swapWeaponIds(player)).toEqual([MACE, undefined]);
		expect([player.getConsumes().mhImbueId, player.getConsumes().swapMhImbueId]).toEqual([DENSE_WEIGHTSTONE_ID, DENSE_WEIGHTSTONE_ID]);
	});
});
