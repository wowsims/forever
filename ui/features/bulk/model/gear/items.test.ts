import { Class, ItemSlot, ItemSpec, ItemType, Spec } from '@generated/proto/common';
import { BulkSimItemSlot } from '@sim/bulk/constants_auto_gen';
import type { BulkPickerEntry } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { bulkState, patchBulkState } from '@sim/settings/bulk_settings';
import { createSimStore, seedKeyed } from '@sim/state/sim_store';
import { beforeEach, describe, expect, it } from 'vitest';

import { loadEquippedBulkItems } from './items';

// `loadEquippedBulkItems` runs the real slot-eligibility helpers, so the fake item has to be a
// real enough `Item`: only `type`, `id` and `classAllowlist` are read on this path.
const STORE_KEY = 0;
const HELM_ID = 1000;
const helmSpec = ItemSpec.create({ id: HELM_ID });
const helm = {
	id: HELM_ID,
	item: { id: HELM_ID, type: ItemType.ItemTypeHead, classAllowlist: [] },
	_item: { id: HELM_ID, name: 'Helm' },
} as unknown as EquippedItem;

let store: ReturnType<typeof createSimStore>;
let equipped: Map<ItemSlot, EquippedItem | null>;
let player: Player<any>;

beforeEach(() => {
	store = createSimStore();
	seedKeyed(store, 'bulk', STORE_KEY, {
		items: [],
		pickerGroups: new Map<BulkSimItemSlot, readonly BulkPickerEntry[]>([[BulkSimItemSlot.ItemSlotHead, []]]),
		isRunning: false,
		v: { settings: 0, items: 0 },
	} as never);
	equipped = new Map();
	player = {
		storeKey: STORE_KEY,
		sim: { store, db: { lookupItemSpec: (spec: ItemSpec) => (spec.id === HELM_ID ? helm : null) } },
		getSpec: () => Spec.SpecDpsWarrior,
		getClass: () => Class.ClassWarrior,
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		getEquippedItems: () => equipped,
	} as unknown as Player<any>;
});

describe('loadEquippedBulkItems', () => {
	it('drops the batch spec of an item that has since been equipped, not just its picker entry', () => {
		patchBulkState(player, { items: [helmSpec], pickerGroups: new Map([[BulkSimItemSlot.ItemSlotHead, [{ index: 0, item: helm }]]]) });
		equipped.set(ItemSlot.ItemSlotHead, helm);

		loadEquippedBulkItems(player);

		expect(
			bulkState(player)
				.pickerGroups.get(BulkSimItemSlot.ItemSlotHead)
				?.map(entry => entry.index),
		).toEqual([-1]);
		// Left in the batch it would sim, count toward the combinations readout, and be re-persisted.
		expect(bulkState(player).items).toEqual([null]);
	});

	it('keeps a batch spec the equipped gear does not cover', () => {
		patchBulkState(player, { items: [helmSpec], pickerGroups: new Map([[BulkSimItemSlot.ItemSlotHead, [{ index: 0, item: helm }]]]) });

		loadEquippedBulkItems(player);

		expect(
			bulkState(player)
				.pickerGroups.get(BulkSimItemSlot.ItemSlotHead)
				?.map(entry => entry.index),
		).toEqual([0]);
		expect(bulkState(player).items).toEqual([helmSpec]);
	});
});
