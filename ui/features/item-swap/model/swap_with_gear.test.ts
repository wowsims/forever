import { ConsumesSpec, ItemSlot } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID } from '@sim/proto/sim_tables_auto_gen';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { MACE, swapSet, SWORD, weaponsDatabase } from '@testing/sim/weapons';
import { makeEquipmentSpec } from '@testing/storybook/fixtures/protos';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { swapWithGear } from './swap_with_gear';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

beforeAll(() => serveDatabase(weaponsDatabase()));

type Items = Partial<Record<ItemSlot, string>>;

/** A gear set reduced to the two calls the exchange makes, with item names standing in for equipped items. */
const gearOf = (items: Items) => ({
	items,
	getEquippedItem: (slot: ItemSlot) => items[slot] ?? null,
	withEquippedItem: (slot: ItemSlot, item: string | null) => gearOf({ ...items, [slot]: item ?? undefined }),
});
type FakeGear = ReturnType<typeof gearOf>;

const playerOf = (gear: Items, swapGear: Items, consumes: ConsumesSpec) => {
	const state = { gear: gearOf(gear), swapGear: gearOf(swapGear), consumes };
	const player = {
		getGear: () => state.gear,
		setGear: (next: FakeGear) => {
			state.gear = next;
		},
		getConsumes: () => ConsumesSpec.clone(state.consumes),
		setConsumes: (next: ConsumesSpec) => {
			state.consumes = ConsumesSpec.clone(next);
		},
		itemSwapSettings: {
			getGear: () => state.swapGear,
			setGear: (next: FakeGear) => {
				state.swapGear = next;
			},
		},
	} as unknown as Player<any>;
	return { player, state };
};

const CONSUMES = ConsumesSpec.create({
	mhImbueId: 1,
	ohImbueId: 2,
	mhWeaponImbueId: 3,
	ohWeaponImbueId: 4,
	swapMhImbueId: 5,
	swapOhImbueId: 6,
	swapMhWeaponImbueId: 7,
	swapOhWeaponImbueId: 8,
	foodId: 9,
});

const imbuesOf = (consumes: ConsumesSpec) => [
	consumes.mhImbueId,
	consumes.ohImbueId,
	consumes.mhWeaponImbueId,
	consumes.ohWeaponImbueId,
	consumes.swapMhImbueId,
	consumes.swapOhImbueId,
	consumes.swapMhWeaponImbueId,
	consumes.swapOhWeaponImbueId,
];

describe('swapWithGear', () => {
	it('exchanges the items and the imbues of both hands with the swap set', () => {
		const { player, state } = playerOf(
			{ [ItemSlot.ItemSlotMainHand]: 'sword', [ItemSlot.ItemSlotOffHand]: 'shield', [ItemSlot.ItemSlotTrinket1]: 'trinket' },
			{ [ItemSlot.ItemSlotMainHand]: 'mace', [ItemSlot.ItemSlotOffHand]: 'dagger' },
			CONSUMES,
		);

		swapWithGear(player, [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1]);

		expect(state.gear.items).toEqual({ [ItemSlot.ItemSlotMainHand]: 'mace', [ItemSlot.ItemSlotOffHand]: 'dagger' });
		expect(state.swapGear.items).toEqual({
			[ItemSlot.ItemSlotMainHand]: 'sword',
			[ItemSlot.ItemSlotOffHand]: 'shield',
			[ItemSlot.ItemSlotTrinket1]: 'trinket',
		});
		expect(imbuesOf(state.consumes)).toEqual([5, 6, 7, 8, 1, 2, 3, 4]);
		expect(state.consumes.foodId).toBe(9);
	});

	it('leaves the imbues of a hand it does not swap', () => {
		const { player, state } = playerOf({ [ItemSlot.ItemSlotMainHand]: 'sword' }, { [ItemSlot.ItemSlotMainHand]: 'mace' }, CONSUMES);

		swapWithGear(player, [ItemSlot.ItemSlotMainHand]);

		expect(imbuesOf(state.consumes)).toEqual([5, 2, 7, 4, 1, 6, 3, 8]);
	});

	it('keeps a stone that moves to a hand the other set leaves empty', async () => {
		const { sim, player } = await makeIndividualSim();
		player.setGear(sim.db.lookupEquipmentSpec(makeEquipmentSpec({ [ItemSlot.ItemSlotMainHand]: { id: SWORD }, [ItemSlot.ItemSlotOffHand]: { id: MACE } })));
		player.itemSwapSettings.setItemSwapSettings(true, sim.db.lookupItemSwap(swapSet({ mainHand: MACE })));
		player.setConsumes(ConsumesSpec.create({ mhImbueId: DENSE_SHARPENING_STONE_ID, ohImbueId: DENSE_WEIGHTSTONE_ID, swapMhImbueId: DENSE_WEIGHTSTONE_ID }));

		swapWithGear(player, [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand]);

		expect(player.itemSwapSettings.getItem(ItemSlot.ItemSlotOffHand)?.item.id).toBe(MACE);
		expect(player.getEquippedItem(ItemSlot.ItemSlotOffHand)).toBeNull();
		expect(imbuesOf(player.getConsumes())).toEqual([DENSE_WEIGHTSTONE_ID, 0, 0, 0, DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID, 0, 0]);
	});
});
