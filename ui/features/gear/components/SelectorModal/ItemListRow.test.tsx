import { ItemQuality, type ItemSpec } from '@generated/proto/common';
import { UIEnchant as Enchant, type UIItem as Item } from '@generated/proto/ui';
import { renderWithHost } from '@testing/react';
import { fireEvent, render, screen } from '@testing-library/react';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { ItemData, ItemListType } from '../../types';
import { SelectorModalTabs } from '../../types';
import type { ItemListRowViewProps } from './ItemListRow';

// The batch seam is mocked rather than driven: the row only asks whether the item is in the batch
// and calls one of two writes, and both are plain functions over the player now.
const batch = vi.hoisted(() => ({
	hasItem: vi.fn((_spec?: unknown) => false),
	addItem: vi.fn((_spec?: unknown) => undefined),
	removeItem: vi.fn((_spec?: unknown) => undefined),
}));

const store = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	const state = {};
	return {
		subscribe: (callback: () => void) => {
			listeners.add(callback);
			return () => listeners.delete(callback);
		},
		getState: () => state,
		getInitialState: () => state,
		notify: () => listeners.forEach(listener => listener()),
	};
});

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon-url.png', name: 'item-name', href: 'https://example.com/item', ready: true }),
}));
vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(store.subscribe));
vi.mock('@features/bulk/model/gear/items', () => ({
	hasBulkItem: (_player: unknown, spec: ItemSpec) => batch.hasItem(spec),
	addBulkItem: (_player: unknown, spec: ItemSpec) => batch.addItem(spec),
	removeBulkItem: (_player: unknown, spec: ItemSpec) => batch.removeItem(spec),
}));
vi.mock('../../../../tracking/utils', () => ({ trackEvent: vi.fn() }));
vi.mock('./ItemSource', () => ({
	ItemSource: ({ item, faction }: { item: Pick<Item, 'id' | 'sources'>; faction?: number }) => (
		<div data-testid="item-source-marker" data-item-id={item.id} data-source-count={item.sources?.length} data-faction={faction} />
	),
}));
vi.mock('../GearPicker/ItemNoticeIcon', () => ({
	ItemNoticeIcon: ({ itemId, enchantId, infoNotice }: { itemId?: number; enchantId?: number; infoNotice?: ReactNode }) => (
		<div data-testid="item-notice-marker" data-item-id={itemId} data-enchant-id={enchantId}>
			{infoNotice}
		</div>
	),
}));

const { ItemListRow, ItemListRowView } = await import('./ItemListRow');
type ItemListRowProps = Parameters<typeof ItemListRow>[0];

const makeItemData = (overrides: Partial<ItemData<ItemListType, ReactNode>> = {}): ItemData<ItemListType, ReactNode> => ({
	item: { id: 1, ilvl: 0 } as Item,
	searchText: '',
	id: 1,
	actionId: {} as ItemData<ItemListType, ReactNode>['actionId'],
	quality: ItemQuality.ItemQualityEpic,
	phase: 1,
	ignoreEPFilter: false,
	nameDescription: '',
	onEquip: vi.fn(),
	name: 'Test Item',
	...overrides,
});

const renderRow = (
	overrides: Partial<ItemListRowProps> = {},
	batchOverrides: Partial<{ hasItem: ReturnType<typeof vi.fn>; addItem: ReturnType<typeof vi.fn>; removeItem: ReturnType<typeof vi.fn> }> = {},
) => {
	Object.assign(batch, { hasItem: vi.fn(() => false), addItem: vi.fn(), removeItem: vi.fn() }, batchOverrides);

	const props: ItemListRowProps = {
		itemData: makeItemData(),
		label: SelectorModalTabs.Items,
		active: false,
		ep: { ep: 0, equippedEP: null },
		favourited: false,
		favouriteTooltipId: 'fav-tooltip',
		compareTooltipId: 'compare-tooltip',
		onEquip: vi.fn(),
		onToggleFavourite: vi.fn(),
		...overrides,
	};

	const { container, rerender } = renderWithHost(<ItemListRow {...props} />, { sim: { store }, player: {} });

	return {
		container,
		batch,
		props,
		rerender: (next: Partial<ItemListRowProps>) => rerender(<ItemListRow {...props} {...next} />),
	};
};

describe('ItemListRow', () => {
	it('omits the level cell outside the items and enchants tabs', () => {
		const { container } = renderRow({ label: SelectorModalTabs.RandomSuffixes });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ilvl-container"]')).toBeNull();
	});

	it.each([
		['an enchant’s lvl', 47, '47'],
		['an empty cell for an enchant with no lvl', 0, ''],
	])('shows %s on the enchants tab', (_name, lvl, shown) => {
		const { container } = renderRow({
			label: SelectorModalTabs.Enchants,
			itemData: makeItemData({ ilvl: lvl, item: Enchant.create({ effectId: 1894, lvl }) as unknown as ItemListType }),
		});
		expect(container.querySelector('[data-testid="selector-modal-list-item-ilvl-container"]')?.textContent).toBe(shown);
	});

	it.each([
		['prefers the itemData ilvl over the item ilvl', 500, 400, '500'],
		['falls back to the item ilvl when the itemData ilvl is unset', 0, 450, '450'],
	])('%s', (_name, ilvl, itemIlvl, shown) => {
		const { container } = renderRow({ itemData: makeItemData({ ilvl, item: { id: 1, ilvl: itemIlvl } as Item }) });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ilvl-container"]')?.textContent).toBe(shown);
	});

	it('renders the source cell on the items and enchants tabs alone', () => {
		const sourceCell = '[data-testid="selector-modal-list-item-source-container"]';
		expect(renderRow({ label: SelectorModalTabs.Items }).container.querySelector(sourceCell)).not.toBeNull();
		expect(renderRow({ label: SelectorModalTabs.Enchants }).container.querySelector(sourceCell)).not.toBeNull();
		expect(renderRow({ label: SelectorModalTabs.RandomSuffixes }).container.querySelector(sourceCell)).toBeNull();
	});

	it.each([
		['items', SelectorModalTabs.Items],
		['enchants', SelectorModalTabs.Enchants],
	])("hands the %s tab's source cell the faction it is given", (_tab, label) => {
		const { container } = renderRow({ label, faction: 2, itemData: makeItemData({ item: Enchant.create({ effectId: 1 }) as unknown as ItemListType }) });
		expect(container.querySelector('[data-testid="item-source-marker"]')?.getAttribute('data-faction')).toBe('2');
	});

	it('shows an enchant row the sources of the enchant, under its own item id', () => {
		const enchant = Enchant.create({
			effectId: 1900,
			itemId: 16252,
			sources: [{ source: { oneofKind: 'drop', drop: { zoneId: 28, npcId: 4494, difficulty: 0, otherName: '', category: '' } } }],
		});
		const { container } = renderRow({ label: SelectorModalTabs.Enchants, itemData: makeItemData({ id: 1900, item: enchant as unknown as ItemListType }) });
		const marker = container.querySelector('[data-testid="item-source-marker"]');

		expect(marker?.getAttribute('data-item-id')).toBe('16252');
		expect(marker?.getAttribute('data-source-count')).toBe('1');
	});

	it('looks the notice up by enchant id on the enchants tab and by item id elsewhere', () => {
		const enchants = renderRow({ label: SelectorModalTabs.Enchants, itemData: makeItemData({ id: 1894 }) });
		const enchantNotice = enchants.container.querySelector('[data-testid="item-notice-marker"]')!;
		expect(enchantNotice.getAttribute('data-enchant-id')).toBe('1894');
		expect(enchantNotice.hasAttribute('data-item-id')).toBe(false);

		const items = renderRow({ label: SelectorModalTabs.Items, itemData: makeItemData({ id: 17182 }) });
		const itemNotice = items.container.querySelector('[data-testid="item-notice-marker"]')!;
		expect(itemNotice.getAttribute('data-item-id')).toBe('17182');
		expect(itemNotice.hasAttribute('data-enchant-id')).toBe(false);
	});

	describe('the first-seen notice', () => {
		const NOW_MS = Date.UTC(2026, 10, 8, 12);
		const DAY_SECONDS = 24 * 60 * 60;
		const notice = (container: HTMLElement) => container.querySelector('[data-testid="item-notice-marker"]')!.textContent;

		beforeEach(() => {
			vi.spyOn(Date, 'now').mockReturnValue(NOW_MS);
		});

		afterEach(() => {
			vi.restoreAllMocks();
		});

		it('shows on an item first seen inside the last seven days', () => {
			const { container } = renderRow({ firstSeen: NOW_MS / 1000 - 3 * DAY_SECONDS });
			expect(notice(container)).toBe('gear_tab.item_notices.discovered_on');
		});

		it.each([
			['an item first seen before the last seven days', SelectorModalTabs.Items, NOW_MS / 1000 - 8 * DAY_SECONDS],
			['an item with no first-seen date', SelectorModalTabs.Items, undefined],
			['an enchant whose row has a recent date', SelectorModalTabs.Enchants, NOW_MS / 1000 - 3 * DAY_SECONDS],
		])('stays off %s', (_name, label, firstSeen) => {
			const { container } = renderRow({ label, firstSeen });
			expect(notice(container)).toBe('');
		});
	});

	it('hands the ep and equipped ep to the ep cell', () => {
		const { container } = renderRow({ ep: { ep: 80, equippedEP: 50 } });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ep-value"]')?.textContent).toBe('80');
		expect(container.querySelector('[data-testid="selector-modal-list-item-ep-delta"]')?.getAttribute('data-sign')).toBe('positive');
	});

	it('draws no ep cell when the list has no EP column', () => {
		const { container } = renderRow({ ep: undefined });
		expect(container.querySelector('[data-testid="selector-modal-list-item-ep"]')).toBeNull();
	});

	it('wires the icon and anchor to useActionId', () => {
		const { container } = renderRow();
		expect(container.querySelector('[data-testid="selector-modal-list-item-icon"]')?.getAttribute('src')).toBe('icon-url.png');
		expect(container.querySelector('[data-testid="selector-modal-list-item-link"]')?.getAttribute('href')).toBe('https://example.com/item');
	});

	it('reflects whether the item is already in the bulk batch', () => {
		const inBatch = renderRow({}, { hasItem: vi.fn(() => true) });
		expect(inBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')?.getAttribute('data-in-batch')).toBe('true');

		const notInBatch = renderRow({}, { hasItem: vi.fn(() => false) });
		expect(notInBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')?.getAttribute('data-in-batch')).toBe('false');
	});

	it('adds the item to the batch when absent, and removes it when present', () => {
		const notInBatch = renderRow({ itemData: makeItemData({ id: 7 }) }, { hasItem: vi.fn(() => false) });
		fireEvent.click(notInBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);
		expect(notInBatch.batch.addItem).toHaveBeenCalledTimes(1);
		expect(notInBatch.batch.removeItem).not.toHaveBeenCalled();

		const inBatch = renderRow({ itemData: makeItemData({ id: 7 }) }, { hasItem: vi.fn(() => true) });
		fireEvent.click(inBatch.container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);
		expect(inBatch.batch.removeItem).toHaveBeenCalledTimes(1);
		expect(inBatch.batch.addItem).not.toHaveBeenCalled();
	});

	it('re-reads the batch flag when a recycled row is handed a different item', () => {
		const { container, rerender } = renderRow({ itemData: makeItemData({ id: 1 }) }, { hasItem: vi.fn((spec: any) => spec.id === 2) });
		const compare = () => container.querySelector('[data-testid="selector-modal-list-item-compare"]')!;
		expect(compare().getAttribute('data-in-batch')).toBe('false');

		rerender({ itemData: makeItemData({ id: 2 }) });

		expect(compare().getAttribute('data-in-batch')).toBe('true');
	});

	it('acts on the item the recycled row now shows, not the one it showed before', () => {
		const { container, batch, rerender } = renderRow({ itemData: makeItemData({ id: 1 }) }, { hasItem: vi.fn((spec: any) => spec.id === 1) });

		rerender({ itemData: makeItemData({ id: 2 }) });
		fireEvent.click(container.querySelector('[data-testid="selector-modal-list-item-compare"]')!);

		expect(batch.addItem).toHaveBeenCalledTimes(1);
		expect(batch.removeItem).not.toHaveBeenCalled();
	});

	it('renders the compare container on the items tab alone', () => {
		const items = renderRow({ label: SelectorModalTabs.Items });
		expect(items.container.querySelector('[data-testid="selector-modal-list-item-compare"]')).not.toBeNull();

		const enchants = renderRow({ label: SelectorModalTabs.Enchants });
		expect(enchants.container.querySelector('[data-testid="selector-modal-list-item-compare"]')).toBeNull();
	});
});

const renderRowView = (props: Partial<ItemListRowViewProps> = {}) => {
	const all: ItemListRowViewProps = {
		name: 'Thunderfury',
		quality: ItemQuality.ItemQualityLegendary,
		nameDescription: '',
		iconUrl: 'icon.jpg',
		href: 'https://example.com/item',
		active: false,
		onEquip: vi.fn(),
		favourited: false,
		favouriteTooltipId: 'favourite',
		onToggleFavourite: vi.fn(),
		...props,
	};
	const { container } = render(
		<div data-testid="row">
			<ItemListRowView {...all} />
		</div>,
	);
	const cells = () => Array.from(screen.getByTestId('row').children).map(child => child.getAttribute('data-testid'));
	return { container, props: all, cells };
};

describe('ItemListRowView', () => {
	it('renders every column in order when each is given', () => {
		const { cells } = renderRowView({
			ilvl: 80,
			source: 'Molten Core',
			ep: { ep: 12, equippedEP: 10 },
			compare: { inBatch: false, tooltipId: 'compare', onToggle: vi.fn() },
		});
		expect(cells()).toEqual([
			'selector-modal-list-item-ilvl-container',
			'selector-modal-list-label-cell',
			'selector-modal-list-item-source-container',
			'selector-modal-list-item-ep',
			'selector-modal-list-item-favorite',
			'selector-modal-list-item-compare',
			'selector-modal-list-item-row-link',
		]);
		expect(screen.getByTestId('selector-modal-list-item-ilvl-container').textContent).toBe('80');
		expect(screen.getByTestId('selector-modal-list-item-source-container').textContent).toBe('Molten Core');
	});

	it('keeps the label and favourite columns alone when nothing else is given', () => {
		expect(renderRowView().cells()).toEqual(['selector-modal-list-label-cell', 'selector-modal-list-item-favorite', 'selector-modal-list-item-row-link']);
	});

	it('leaves the description label off when the item has none', () => {
		renderRowView({ nameDescription: '' });
		expect(screen.getByTestId('selector-modal-list-item-name').textContent).toBe('Thunderfury');
	});

	it('shows the icon, the quality-coloured name, its description and the notice beside them', () => {
		renderRowView({ nameDescription: 'Heroic', notice: <span data-testid="notice" />, active: true });
		expect(screen.getByTestId('selector-modal-list-item-icon').getAttribute('src')).toBe('icon.jpg');
		expect(screen.getByTestId('selector-modal-list-item-link').getAttribute('href')).toBe('https://example.com/item');
		expect(screen.getByTestId('selector-modal-list-item-name').dataset.quality).toBe(String(ItemQuality.ItemQualityLegendary));
		expect(screen.getByTestId('heroic-label').textContent).toBe('(Heroic)');
		expect(screen.getByTestId('selector-modal-list-label-cell').hasAttribute('data-active')).toBe(true);
		expect(screen.getByTestId('notice')).toBeTruthy();
	});

	it('equips from the row link', () => {
		const { props } = renderRowView();
		fireEvent.click(screen.getByTestId('selector-modal-list-item-link'));
		expect(props.onEquip).toHaveBeenCalledOnce();
	});

	it('equips from a click on the row outside its controls', () => {
		const { props } = renderRowView({ source: 'Molten Core' });
		fireEvent.click(screen.getByTestId('selector-modal-list-item-row-link'));
		expect(props.onEquip).toHaveBeenCalledOnce();
	});

	it('leaves a source link and the favourite star to themselves', () => {
		const onSource = vi.fn();
		const { props } = renderRowView({
			source: (
				<WowheadAnchor href="https://example.com/npc" testId="source-link" onClick={onSource}>
					Ragnaros
				</WowheadAnchor>
			),
		});

		fireEvent.click(screen.getByTestId('source-link'));
		fireEvent.click(screen.getByTestId('selector-modal-list-item-favorite'));

		expect(onSource).toHaveBeenCalledOnce();
		expect(props.onToggleFavourite).toHaveBeenCalledOnce();
		expect(props.onEquip).not.toHaveBeenCalled();
	});
});
