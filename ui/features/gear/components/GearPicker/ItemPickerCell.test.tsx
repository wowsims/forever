import { ItemQuality, ItemSlot, Spec } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { createSimStore, patchKeyed, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const tooltip = vi.hoisted(() => ({ settles: [] as Array<(url: string) => void> }));

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: '', name: '', href: 'https://www.wowhead.com/item=1', ready: true }) }));
vi.mock('@sim/proto/action_id/tooltip_data', () => ({
	equippedItemWowheadTooltipData: () => new Promise<string>(resolve => tooltip.settles.push(resolve)),
	actionIdWowheadTooltipData: () => new Promise<string>(() => {}),
}));

const { ItemPickerCell } = await import('./ItemPickerCell');

const equippedItem = (id: number) =>
	({
		item: { id, name: `Item ${id}`, quality: ItemQuality.ItemQualityEpic, nameDescription: '' },
		ilvl: 120,
		randomSuffix: undefined,
		enchant: undefined,
		asActionId: () => ({ itemId: id, spellId: 0 }),
		hasRandomSuffixOptions: () => false,
	}) as any;

describe('ItemPickerCell', () => {
	const setup = () => {
		const store = createSimStore();
		const storeKey = 0;
		const gear = (id: number) => ({ getEquippedItem: () => equippedItem(id), asArray: () => [equippedItem(id)] });
		seedKeyed(store, 'players', storeKey, { gear: gear(1), v: zeroVersions(PLAYER_FIELDS) } as never);
		const player = {
			storeKey,
			sim: { store },
			getGear: () => store.getState().players[storeKey].gear,
			getSpec: () => Spec.SpecUnknown,
		} as unknown as Player<any>;
		tooltip.settles.length = 0;
		const { container } = renderWithHost(<ItemPickerCell slot={ItemSlot.ItemSlotHead} ready={false} />, { player, sim: { store } });
		return {
			container,
			equip: (id: number) => act(() => patchKeyed(store, 'players', storeKey, { gear: gear(id) } as never, ['gear'])),
		};
	};

	const anchors = (container: HTMLElement) => [
		container.querySelector<HTMLElement>('[data-testid="item-picker-icon"]')!,
		container.querySelector<HTMLElement>('[data-testid="item-picker-name-container"]')!,
	];

	it('shows the tooltip the equipped item resolved to, on both the icon and the name', async () => {
		const { container } = setup();

		tooltip.settles[0]('item=1');
		await waitFor(() => expect(anchors(container).map(element => element.dataset.wowhead)).toEqual(['item=1', 'item=1']));
		expect(anchors(container).map(element => element.dataset.whtticon)).toEqual(['false', 'false']);
	});

	// The defect: a tooltip request for the item that was in the slot a moment ago must not land
	// on the cell after the slot moved on.
	it('drops a tooltip that resolved after the equipped item changed', async () => {
		const { container, equip } = setup();

		equip(2);
		expect(anchors(container).some(element => element.hasAttribute('data-wowhead'))).toBe(false);
		expect(tooltip.settles).toHaveLength(2);

		tooltip.settles[1]('item=2');
		await waitFor(() => expect(anchors(container).map(element => element.dataset.wowhead)).toEqual(['item=2', 'item=2']));

		await act(async () => tooltip.settles[0]('item=1'));
		expect(anchors(container).map(element => element.dataset.wowhead)).toEqual(['item=2', 'item=2']);
	});
});
