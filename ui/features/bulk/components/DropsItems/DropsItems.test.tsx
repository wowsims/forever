import { dropsItemListFixture, dropsNamesFixture, dropsTreeFixture, UNTICKED_ITEM_ID, WORN_ITEM_ID } from '@testing/storybook/fixtures/bulk-drops';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { DropsItemList } from '../../model/drops/items';
import { DropsItemsView, type DropsItemsViewProps } from './DropsItems';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: '', ready: true }) }));

const FELSTRIKER = 2001;
const REND = 'Warchief Rend Blackhand';
const WYRMTHALAK = 'Overlord Wyrmthalak';
const BARTHILAS = 'Magistrate Barthilas';

const onIncludedChange = vi.fn();
const onGroupingChange = vi.fn();

const mount = (list: DropsItemList = dropsItemListFixture(), props: Partial<DropsItemsViewProps> = {}) =>
	render(
		<DropsItemsView
			list={list}
			grouping="source"
			onGroupingChange={onGroupingChange}
			tree={dropsTreeFixture()}
			names={dropsNamesFixture()}
			onIncludedChange={onIncludedChange}
			{...props}
		/>,
	);

const withExcluded = (list: DropsItemList, ids: number[]): DropsItemList => ({
	items: list.items.map(entry => (ids.includes(entry.item.id) ? { ...entry, included: false } : entry)),
});

const group = (title: string) => screen.getAllByTestId('drops-group').find(item => within(item).queryByRole('checkbox', { name: title }))!;
const open = (title: string) => fireEvent.click(within(group(title)).getByTestId('drops-group-trigger'));
const box = (title: string) => within(group(title)).getByRole('checkbox', { name: title });
const rowIds = (title: string) =>
	within(group(title))
		.queryAllByTestId('drops-item-row')
		.map(row => Number(row.getAttribute('data-item-id')));

beforeEach(() => {
	onIncludedChange.mockClear();
});

describe('DropsItemsView', () => {
	it('starts with every section collapsed and no row mounted', () => {
		mount();

		expect(screen.getAllByTestId('drops-group')).toHaveLength(5);
		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(0);
		for (const trigger of screen.getAllByTestId('drops-group-trigger')) expect(trigger.getAttribute('aria-expanded')).toBe('false');
	});

	it('ticks a row from a press on its sources', () => {
		mount();
		open(REND);

		const row = within(group(REND))
			.getAllByTestId('drops-item-row')
			.find(r => Number(r.getAttribute('data-item-id')) === FELSTRIKER)!;
		fireEvent.click(within(row).getByText('Baron Rivendare'));
		expect(onIncludedChange).toHaveBeenCalledWith([FELSTRIKER], false);
	});

	it('shows a section’s rows when it opens, and drops them when it closes', () => {
		mount();

		open(REND);

		expect(rowIds(REND).sort()).toEqual([FELSTRIKER, WORN_ITEM_ID, 2011]);
		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(3);

		open(REND);

		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(0);
	});

	it('badges the included non-worn items of a section, and none when nothing is included', () => {
		mount(withExcluded(dropsItemListFixture(), [FELSTRIKER]));

		expect(within(group(REND)).getByTestId('drops-group-count').textContent).toBe('1');
		expect(within(group(WYRMTHALAK)).getByTestId('drops-group-count').textContent).toBe('2');
		expect(within(group(BARTHILAS)).queryByTestId('drops-group-count')).toBeNull();
	});

	it('shows the empty-slot icon of a slot section and none for a source section', () => {
		const view = mount(dropsItemListFixture(), { grouping: 'slot' });
		const art = (item: HTMLElement) => item.querySelector<HTMLElement>('[style*="/item_slots/"]');
		expect(screen.getAllByTestId('drops-group').every(art)).toBe(true);

		view.unmount();
		mount();
		expect(screen.getAllByTestId('drops-group').some(art)).toBe(false);
	});

	it('reads the section box as checked, mixed or unchecked', () => {
		mount(withExcluded(dropsItemListFixture(), [FELSTRIKER]));

		expect(box(WYRMTHALAK).getAttribute('aria-checked')).toBe('true');
		expect(box(REND).getAttribute('aria-checked')).toBe('mixed');
		expect(box(BARTHILAS).getAttribute('aria-checked')).toBe('false');
	});

	it('includes the non-worn items of a section from its box, without opening it', () => {
		mount(withExcluded(dropsItemListFixture(), [FELSTRIKER]));

		fireEvent.click(box(REND));

		expect(onIncludedChange).toHaveBeenCalledWith(expect.arrayContaining([FELSTRIKER, 2011]), true);
		expect(onIncludedChange.mock.calls[0][0]).toHaveLength(2);
		expect(onIncludedChange.mock.calls[0][0]).not.toContain(WORN_ITEM_ID);
		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(0);
	});

	it('excludes a whole ticked section and includes an unticked one', () => {
		mount();

		fireEvent.click(box(WYRMTHALAK));
		expect(onIncludedChange).toHaveBeenLastCalledWith(expect.arrayContaining([2005, 2006]), false);

		fireEvent.click(box(BARTHILAS));
		expect(onIncludedChange).toHaveBeenLastCalledWith([UNTICKED_ITEM_ID], true);
	});

	it('collapses every section again when the grouping changes', () => {
		const view = mount();
		open(REND);
		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(3);

		view.rerender(
			<DropsItemsView
				list={dropsItemListFixture()}
				grouping="slot"
				onGroupingChange={onGroupingChange}
				tree={dropsTreeFixture()}
				names={dropsNamesFixture()}
				onIncludedChange={onIncludedChange}
			/>,
		);

		expect(screen.queryAllByTestId('drops-item-row')).toHaveLength(0);
		for (const trigger of screen.getAllByTestId('drops-group-trigger')) expect(trigger.getAttribute('aria-expanded')).toBe('false');
	});
});
