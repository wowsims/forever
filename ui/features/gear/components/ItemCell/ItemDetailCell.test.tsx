import { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { makeEquippedItem } from '@testing/storybook/fixtures/gear';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { SelectorModalTabs } from '../../types';
import { ItemDetailCellView, type ItemDetailCellViewProps } from './ItemDetailCell';

const renderCell = (props: Partial<ItemDetailCellViewProps> = {}) =>
	render(<ItemDetailCellView slot={ItemSlot.ItemSlotMainHand} item={makeEquippedItem()} iconUrl="sword.jpg" href="https://example.com/item" {...props} />);

const icon = () => screen.getByTestId('item-picker-icon');
const name = () => screen.getByTestId('item-picker-name-container');

describe('ItemDetailCellView', () => {
	it('shows the item level, name, icon and link of the equipped item', () => {
		renderCell();
		expect(screen.getByTestId('item-picker-ilvl').textContent).toBe('80');
		expect(screen.getByTestId('item-picker-name').textContent).toBe('Thunderfury, Blessed Blade of the Windseeker');
		expect(icon().style.backgroundImage).toContain('sword.jpg');
		expect([icon(), name()].map(anchor => anchor.getAttribute('href'))).toEqual(['https://example.com/item', 'https://example.com/item']);
	});

	it('shows the slot name and its empty icon when nothing is equipped', () => {
		renderCell({ item: null, iconUrl: '' });
		expect(name().textContent).toBe(translateSlotName(ItemSlot.ItemSlotMainHand));
		expect(icon().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotMainHand));
	});

	it('puts the wowhead tooltip on both the icon and the name', () => {
		renderCell({ wowhead: { 'data-wowhead': 'item=19019' } });
		expect([icon(), name()].map(anchor => anchor.dataset.wowhead)).toEqual(['item=19019', 'item=19019']);
	});

	it('renders the notice beside the name and the extra labels after the enchant', () => {
		renderCell({ notice: <span data-testid="notice-marker" />, extraLabels: <span data-testid="extra-marker" /> });
		expect(screen.getByTestId('item-picker-name-row').querySelector('[data-testid="notice-marker"]')).not.toBeNull();
		expect(screen.getByTestId('item-picker-labels-container').lastElementChild!.getAttribute('data-testid')).toBe('extra-marker');
	});

	it('opens the items tab from the icon and the name', () => {
		const onOpen = vi.fn();
		renderCell({ onOpen, href: '' });

		fireEvent.click(icon());
		fireEvent.click(name());
		expect(onOpen.mock.calls).toEqual([[SelectorModalTabs.Items], [SelectorModalTabs.Items]]);
	});

	it('is a button only where it opens the selector with no link, and never nests one anchor in another', () => {
		const { rerender } = renderCell({ onOpen: () => {} });
		expect([icon(), name()].map(anchor => anchor.getAttribute('role'))).toEqual([null, null]);
		expect(document.querySelectorAll('a a')).toHaveLength(0);

		rerender(<ItemDetailCellView slot={ItemSlot.ItemSlotMainHand} item={null} iconUrl="" href="" onOpen={() => {}} />);
		expect([icon(), name()].map(anchor => anchor.getAttribute('role'))).toEqual(['button', 'button']);
		expect([icon(), name()].map(anchor => anchor.tabIndex)).toEqual([0, 0]);

		rerender(<ItemDetailCellView slot={ItemSlot.ItemSlotMainHand} item={null} iconUrl="" href="" />);
		expect([icon(), name()].map(anchor => anchor.getAttribute('role'))).toEqual([null, null]);
	});

	it.each([
		['start', undefined],
		['end', 'end'],
	] as const)('sits the name and labels at the %s edge, the start by default', (edge, align) => {
		renderCell({ align });
		expect(screen.getByTestId('item-picker-labels-container').dataset.align).toBe(edge);
	});
});
