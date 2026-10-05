import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore, patchKeyed, PLAYER_FIELDS, seedKeyed, type SimStore, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { fakeHost } from '@testing/sim/fake_host';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { useMemo } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { type OpenSelectorModal, type SelectorModalState, useSelectorModalState } from '../../hooks/useSelectorModal';
import { ALL_ITEM_SLOTS } from '../../model/gear_data';
import { type GearData, SelectorModalTabs } from '../../types';
import type { SelectorModalViewProps } from './SelectorModal';

vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions());

const tabs = vi.hoisted(() => ({ build: vi.fn(), eligibility: vi.fn() }));
vi.mock('./utils', () => ({ buildSelectorTabs: tabs.build, eligibilityFor: tabs.eligibility }));

const panes = vi.hoisted(() => ({ rendered: [] as Array<{ label: string; slot: number }> }));
vi.mock('./ItemList', () => ({
	ItemList: ({ tab, slot }: any) => {
		panes.rendered.push({ label: tab.label, slot });
		return <div data-pane={tab.label} />;
	},
}));

vi.mock('./SlotRail', () => ({ SlotRail: () => <div data-testid="gear-picker-modal-slots" /> }));

const { SelectorModal, SelectorModalView } = await import('./SelectorModal');

const tab = (label: SelectorModalTabs) => ({
	label,
	itemData: [],
	computeEP: () => 0,
	equippedToItem: () => null,
	onRemove: () => undefined,
});

const OPEN_SLOT = ALL_ITEM_SLOTS[3];

const item = (id: number): EquippedItem => {
	const stub = { id, withChallengeMode: () => stub, withDynamicStats: () => stub };
	return stub as unknown as EquippedItem;
};

// `GearTabBody` holds the state and the pickers open it; the harness stands in for that pair, and
// records what the modal's own rail asks for.
const openSpy = vi.fn<OpenSelectorModal>();
let openModal: OpenSelectorModal;
let setModalOpen: (open: boolean) => void;
const Harness = () => {
	const state = useSelectorModalState();
	const wrapped = useMemo<SelectorModalState>(
		() => ({
			...state,
			openTab: (slot, tab, gearData) => {
				openSpy(slot, tab, gearData);
				state.openTab(slot, tab, gearData);
			},
		}),
		[state],
	);
	openModal = wrapped.openTab;
	setModalOpen = wrapped.setOpen;
	return <SelectorModal state={wrapped} />;
};

/** Cleared first, so only what the rail opens after this point is counted. */
const railOpens = () => {
	openSpy.mockClear();
	return openSpy;
};

describe('SelectorModal', () => {
	let equippedItems: Map<ItemSlot, EquippedItem | null>;
	let getEquippedItem: ReturnType<typeof vi.fn<() => EquippedItem | null>>;
	let host: IndividualSimHost<any>;
	let gearData: GearData;
	let realStore: SimStore;
	const STORE_KEY = 0;
	const makeGear = () => ({ getEquippedItem: (slot: ItemSlot) => equippedItems.get(slot) ?? null });
	const refreshGear = () => patchKeyed(realStore, 'players', STORE_KEY, { gear: makeGear() as never }, ['gear']);

	const setup = ({ tabSet = [tab(SelectorModalTabs.Items), tab(SelectorModalTabs.Enchants)], requestTab = SelectorModalTabs.Items } = {}) => {
		tabs.build.mockReturnValue(tabSet);
		const result = render(
			<SimHostProvider host={host}>
				<Harness />
			</SimHostProvider>,
		);
		act(() => openModal(OPEN_SLOT, requestTab, gearData));
		return result;
	};

	const tabButtons = () => Array.from(document.querySelectorAll<HTMLButtonElement>('[data-testid="selector-modal-tabs"] [role="tab"]'));
	const openPanes = () =>
		Array.from(document.querySelectorAll<HTMLElement>('[data-testid="selector-modal-tab-pane"][data-active] [data-pane]')).map(pane => pane.dataset.pane);
	const popup = () => document.querySelector('[data-testid="selector-modal"]')!;

	beforeEach(() => {
		openSpy.mockClear();
		panes.rendered.length = 0;
		equippedItems = new Map(ALL_ITEM_SLOTS.map(slot => [slot, item(slot)]));
		getEquippedItem = vi.fn(() => equippedItems.get(OPEN_SLOT) ?? null);
		gearData = { equipItem: () => undefined, getEquippedItem, subscribe: () => () => undefined };
		tabs.eligibility.mockReturnValue({ hasEnchants: true });

		const rootElem = document.createElement('div');
		rootElem.className = 'sim-ui';
		document.body.appendChild(rootElem);
		realStore = createSimStore();
		seedKeyed(realStore, 'players', STORE_KEY, { gear: makeGear(), challengeModeEnabled: false, v: zeroVersions(PLAYER_FIELDS) } as never);
		host = fakeHost({
			rootElem,
			player: {
				storeKey: STORE_KEY,
				sim: { store: realStore, encounter: { getAreaTypes: () => [] } },
				getGear: () => ({ getEquippedItem: (slot: ItemSlot) => equippedItems.get(slot) ?? null }),
				getChallengeModeEnabled: () => false,
				equipItem: () => undefined,
				getEquippedItem: (slot: ItemSlot) => equippedItems.get(slot) ?? null,
			},
		});
	});

	afterEach(() => {
		host.rootElem.remove();
	});

	it('renders one tab button and one pane per tab the data earned', () => {
		setup({ tabSet: [tab(SelectorModalTabs.Items), tab(SelectorModalTabs.Enchants), tab(SelectorModalTabs.RandomSuffixes)] });

		expect(tabButtons().map(button => button.dataset.label)).toEqual([
			SelectorModalTabs.Items,
			SelectorModalTabs.Enchants,
			SelectorModalTabs.RandomSuffixes,
		]);
		expect(panes.rendered.map(pane => pane.label)).toEqual([SelectorModalTabs.Items, SelectorModalTabs.Enchants, SelectorModalTabs.RandomSuffixes]);
	});

	it('opens on the requested tab, and on Items when the request is one the slot no longer offers', () => {
		tabs.eligibility.mockReturnValue({ hasEnchants: false });
		setup({ requestTab: SelectorModalTabs.Enchants });

		expect(tabButtons().find(button => button.getAttribute('aria-selected') === 'true')?.dataset.label).toBe(SelectorModalTabs.Items);
	});

	it('keeps the tab the user picks for the life of one request, and resets on the next', () => {
		setup();

		act(() => tabButtons()[1].click());
		expect(tabButtons().find(button => button.getAttribute('aria-selected') === 'true')?.dataset.label).toBe(SelectorModalTabs.Enchants);

		act(() => openModal(OPEN_SLOT, SelectorModalTabs.Items, gearData));
		expect(tabButtons().find(button => button.getAttribute('aria-selected') === 'true')?.dataset.label).toBe(SelectorModalTabs.Items);
	});

	it('falls back to the first tab when the one you were on is gone', () => {
		setup({ tabSet: [tab(SelectorModalTabs.Items), tab(SelectorModalTabs.Enchants), tab(SelectorModalTabs.RandomSuffixes)] });

		act(() => tabButtons()[2].click());
		expect(tabButtons().find(button => button.getAttribute('aria-selected') === 'true')?.dataset.label).toBe(SelectorModalTabs.RandomSuffixes);

		panes.rendered.length = 0;
		tabs.build.mockReturnValue([tab(SelectorModalTabs.Items), tab(SelectorModalTabs.Enchants)]);
		act(() => {
			equippedItems.set(OPEN_SLOT, item(999));
			refreshGear();
		});

		expect(tabButtons().find(button => button.getAttribute('aria-selected') === 'true')?.dataset.label).toBe(SelectorModalTabs.Items);
		expect(openPanes()).toEqual([SelectorModalTabs.Items]);
	});

	// Base UI stops keydown propagation at the popup, so the rail must listen on the popup rather
	// than the document.
	it('navigates the rail from the popup and not from the document', () => {
		setup();
		const openTab = railOpens();

		fireEvent.keyDown(document, { key: 'ArrowDown' });
		fireEvent.keyDown(document.body, { key: 'ArrowDown' });
		expect(openTab).not.toHaveBeenCalled();

		fireEvent.keyDown(popup(), { key: 'ArrowDown' });
		expect(openTab).toHaveBeenCalledTimes(1);
		expect(openTab.mock.calls[0][0]).toBe(ALL_ITEM_SLOTS[4]);
	});

	it('steps the rail by its own indices and wraps at both ends', () => {
		setup();
		const openTab = railOpens();

		fireEvent.keyDown(popup(), { key: 'ArrowUp' });
		expect(openTab.mock.calls[0][0]).toBe(ALL_ITEM_SLOTS[2]);

		act(() => openModal(ALL_ITEM_SLOTS[0], SelectorModalTabs.Items, gearData));
		fireEvent.keyDown(popup(), { key: 'ArrowUp' });
		expect(openTab.mock.calls.at(-1)![0]).toBe(ALL_ITEM_SLOTS.at(-1));

		act(() => openModal(ALL_ITEM_SLOTS.at(-1)!, SelectorModalTabs.Items, gearData));
		fireEvent.keyDown(popup(), { key: 'ArrowDown' });
		expect(openTab.mock.calls.at(-1)![0]).toBe(ALL_ITEM_SLOTS[0]);
	});

	it('carries the tab you are on into the slot the rail moves to', () => {
		setup();
		const openTab = railOpens();

		act(() => tabButtons()[1].click());
		fireEvent.keyDown(popup(), { key: 'ArrowDown' });
		expect(openTab.mock.calls[0][1]).toBe(SelectorModalTabs.Enchants);
	});

	it('leaves the rail keys alone once the modal is closed', () => {
		setup();
		const openTab = railOpens();
		const element = popup();

		act(() => setModalOpen(false));
		fireEvent.keyDown(element, { key: 'ArrowDown' });
		expect(openTab).not.toHaveBeenCalled();
	});

	// `Gear` copies the untouched slots' references forward, so the tab data is keyed on the open
	// slot's stored item rather than on the gear object.
	it('does not rebuild the tab data when a different slot changes', () => {
		setup();
		const before = getEquippedItem.mock.calls.length;

		act(() => {
			equippedItems.set(ALL_ITEM_SLOTS[0], item(42));
			refreshGear();
		});
		expect(getEquippedItem.mock.calls.length).toBe(before);

		act(() => {
			equippedItems.set(OPEN_SLOT, item(43));
			refreshGear();
		});
		expect(getEquippedItem.mock.calls.length).toBe(before + 1);
	});
});

const TABS: SelectorModalViewProps['tabs'] = [
	{ label: SelectorModalTabs.Items, pane: <p>items pane</p> },
	{ label: SelectorModalTabs.Enchants, pane: <p>enchants pane</p> },
	{ label: SelectorModalTabs.RandomSuffixes, pane: <p>random suffix pane</p> },
];

const renderView = (props: Partial<SelectorModalViewProps> = {}) => {
	const all: SelectorModalViewProps = {
		id: 'modal',
		open: true,
		onOpenChange: vi.fn(),
		title: 'Main Hand',
		tabs: TABS,
		activeTab: SelectorModalTabs.Items,
		onSelectTab: vi.fn(),
		sequence: 1,
		...props,
	};
	renderWithHost(<SelectorModalView {...all} />);
	return all;
};

const viewTabButtons = () => screen.getAllByRole('tab');
const activePanes = () => Array.from(document.querySelectorAll('[data-testid="selector-modal-tab-pane"][data-active]')).map(pane => pane.textContent);

describe('SelectorModalView', () => {
	it('titles the dialog with the slot and puts the rail in its header', () => {
		renderView({ rail: <div data-testid="rail-marker" /> });
		expect(screen.getByTestId('selector-modal-title').textContent).toBe('Main Hand');
		expect(screen.getByTestId('rail-marker')).toBeTruthy();
	});

	it('wires each tab button to its pane under the given id', () => {
		renderView();
		expect(viewTabButtons().map(button => button.dataset.label)).toEqual([
			SelectorModalTabs.Items,
			SelectorModalTabs.Enchants,
			SelectorModalTabs.RandomSuffixes,
		]);
		for (const button of viewTabButtons()) {
			const pane = document.getElementById(button.getAttribute('aria-controls')!)!;
			expect(pane.id.startsWith('modal-')).toBe(true);
			expect(pane.getAttribute('aria-labelledby')).toBe(button.id);
		}
	});

	it('labels each tab with its name', () => {
		renderView();
		expect(viewTabButtons()[0].textContent).toBe(i18n.t('gear_tab.gear_picker.tabs.items'));
		expect(viewTabButtons()[2].textContent).toBe(i18n.t('gear_tab.gear_picker.tabs.random_suffix'));
	});

	it('shows the active pane and reports a tab pick', () => {
		const props = renderView({ activeTab: SelectorModalTabs.Enchants });
		expect(activePanes()).toEqual(['enchants pane']);

		act(() => viewTabButtons()[2].click());
		expect(props.onSelectTab).toHaveBeenCalledWith(SelectorModalTabs.RandomSuffixes);
	});

	it('renders no panes without a request', () => {
		renderView({ sequence: undefined });
		expect(document.querySelector('[data-testid="selector-modal-tab-pane"]')).toBeNull();
		expect(viewTabButtons()).toHaveLength(3);
	});

	it('hands key presses on the popup to its handler', () => {
		const onKeyDown = vi.fn();
		renderView({ onKeyDown });
		fireEvent.keyDown(screen.getByTestId('selector-modal'), { key: 'ArrowDown' });
		expect(onKeyDown).toHaveBeenCalledOnce();
	});
});
