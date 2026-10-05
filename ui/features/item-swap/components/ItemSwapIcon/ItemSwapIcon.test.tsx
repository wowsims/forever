import { OpenSelectorModalContext } from '@features/gear/hooks/useSelectorModal';
import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import type { GearData } from '@features/gear/types';
import { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { ItemSwapGear } from '@sim/proto/gear';
import { createSimStore, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { fakeHost } from '@testing/sim/fake_host';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ItemSwapIconViewProps } from './ItemSwapIcon';

const trackReleases = (store: SimStore) => {
	const releases: Array<ReturnType<typeof vi.fn>> = [];
	const real = store.subscribe.bind(store) as (...args: Array<unknown>) => () => void;
	vi.spyOn(store, 'subscribe').mockImplementation(((...args: Array<unknown>) => {
		const release = vi.fn(real(...args));
		releases.push(release);
		return release;
	}) as unknown as SimStore['subscribe']);
	return releases;
};

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: (actionId?: { itemId: number }) =>
		actionId
			? { iconUrl: `icon-${actionId.itemId}.jpg`, name: `Item ${actionId.itemId}`, href: `https://wowhead.test/item=${actionId.itemId}`, ready: true }
			: { iconUrl: '', name: '', href: '', ready: true },
}));

const tooltip = { settles: [] as Array<(url: string) => void> };
vi.mock('@sim/proto/action_id/tooltip_data', () => ({
	equippedItemWowheadTooltipData: () => new Promise<string>(resolve => tooltip.settles.push(resolve)),
}));

const { ItemSwapIcon, ItemSwapIconView } = await import('./ItemSwapIcon');

const SLOTS = [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand, ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2];

const equippedItem = (itemId: number) =>
	({
		asActionId: () => ({ itemId }),
		// The selector modal reads a swap item through the same decorators the gear picker uses, so the
		// double has to answer them or `createItemSwapGearData` throws before it returns anything.
		withDynamicStats() {
			return this;
		},
	}) as unknown as EquippedItem;

const setup = (swap: Map<ItemSlot, EquippedItem> = new Map(), slots: ItemSlot[] = SLOTS) => {
	const equipItem = vi.fn();
	const openTab = vi.fn();
	const store = createSimStore();
	const releases = trackReleases(store);
	const storeKey = 0;
	seedKeyed(store, 'players', storeKey, {
		itemSwapGear: new ItemSwapGear(Object.fromEntries(swap)),
		v: zeroVersions(PLAYER_FIELDS),
	} as never);
	const player = {
		storeKey,
		sim: { store, encounter: { getAreaTypes: () => [] } },
		itemSwapSettings: { getItem: (slot: ItemSlot) => swap.get(slot) ?? null, equipItem },
	} as unknown as Player<any>;
	const host = fakeHost({ player });
	const view = render(
		<SimHostProvider host={host}>
			<OpenSelectorModalContext value={openTab}>
				{slots.map(slot => (
					<ItemSwapIcon key={slot} slot={slot} />
				))}
			</OpenSelectorModalContext>
		</SimHostProvider>,
	);
	return { view, equipItem, openTab, releases };
};

const icons = (container: Element) => [
	...container.querySelectorAll<HTMLAnchorElement>('[data-testid="icon-picker-root"] > a[data-testid="icon-picker-button"]'),
];

beforeEach(() => {
	tooltip.settles.length = 0;
});

describe('ItemSwapIcon', () => {
	it('marks a filled slot active and links it to the item, which is what the swap probe reads', () => {
		const { view } = setup(new Map([[ItemSlot.ItemSlotMainHand, equippedItem(1234)]]));

		const [mainHand] = icons(view.container);
		expect(mainHand.hasAttribute('data-active')).toBe(true);
		expect(mainHand.getAttribute('href')).toBe('https://wowhead.test/item=1234');
		expect(mainHand.style.backgroundImage).toContain('icon-1234.jpg');
	});

	it('writes the resolved tooltip data onto a filled slot, and marks it as carrying its own icon', async () => {
		const { view } = setup(new Map([[ItemSlot.ItemSlotMainHand, equippedItem(1234)]]));

		expect(icons(view.container).map(icon => icon.dataset.whtticon)).toEqual(['false', undefined, undefined, undefined]);
		expect(tooltip.settles).toHaveLength(1);

		tooltip.settles[0]('item=1234');
		await waitFor(() => expect(icons(view.container)[0].dataset.wowhead).toBe('item=1234'));
	});

	it('opens one shared modal on the slot the icon was clicked for', () => {
		const { view, openTab } = setup();

		icons(view.container).forEach(icon => icon.click());

		expect(openTab).toHaveBeenCalledTimes(SLOTS.length);
		expect(openTab.mock.calls.map(([slot]) => slot)).toEqual(SLOTS);
		expect(openTab.mock.calls.map(([, tab]) => tab)).toEqual(['Items', 'Items', 'Items', 'Items']);
	});

	it('hands the modal gear data that reads and writes the swap set, not the equipped one', () => {
		const item = equippedItem(1234);
		const { view, openTab, equipItem } = setup(new Map([[ItemSlot.ItemSlotOffHand, item]]));

		icons(view.container)[1].click();

		const gearData = openTab.mock.calls[0][2] as GearData;
		expect(gearData.getEquippedItem()).toBe(item);
		gearData.equipItem(null);
		expect(equipItem).toHaveBeenCalledWith(ItemSlot.ItemSlotOffHand, null);
	});

	it('holds a live store listener per mounted icon, and releases every one of them on unmount', () => {
		const { view, releases } = setup();

		expect(releases.length).toBeGreaterThan(0);
		expect(releases.filter(release => release.mock.calls.length > 0)).toHaveLength(0);

		view.unmount();

		expect(releases.filter(release => release.mock.calls.length === 0)).toHaveLength(0);
	});
});

const renderIcon = (props: Partial<ItemSwapIconViewProps> = {}) => {
	const onOpen = vi.fn();
	render(<ItemSwapIconView slot={ItemSlot.ItemSlotMainHand} item={null} iconUrl="" name="" href="" onOpen={onOpen} {...props} />);
	return { onOpen };
};

const icon = () => screen.getByTestId('icon-picker-button');

describe('ItemSwapIconView', () => {
	it('draws an empty slot as its placeholder, unfilled and named for the slot', () => {
		renderIcon();

		expect(icon().hasAttribute('data-active')).toBe(false);
		expect(icon().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotMainHand));
		expect(icon().getAttribute('aria-label')).toBe(translateSlotName(ItemSlot.ItemSlotMainHand));
		expect(icon().hasAttribute('href')).toBe(false);
	});

	it('draws a filled slot active, with its icon, name, link and wowhead data', () => {
		renderIcon({
			item: makeEquippedItem(),
			iconUrl: 'sword.jpg',
			name: 'Thunderfury',
			href: 'https://wowhead.test/item=19019',
			wowhead: { 'data-wowhead': 'item=19019' },
		});

		expect(icon().hasAttribute('data-active')).toBe(true);
		expect(icon().style.backgroundImage).toContain('sword.jpg');
		expect(icon().getAttribute('aria-label')).toBe('Thunderfury');
		expect(icon().getAttribute('href')).toBe('https://wowhead.test/item=19019');
		expect(icon().dataset.whtticon).toBe('false');
		expect(icon().dataset.wowhead).toBe('item=19019');
	});

	it('opens the slot when the icon is clicked', () => {
		const { onOpen } = renderIcon();

		fireEvent.click(icon());
		expect(onOpen).toHaveBeenCalledOnce();
	});
});
