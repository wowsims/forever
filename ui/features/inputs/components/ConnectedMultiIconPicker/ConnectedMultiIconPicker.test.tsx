import { type PickerStatOptions, relevantStatOptions } from '@features/settings/model/stat_options';
import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { filled } from '@testing/inputs';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { IconPickerConfig } from '../ConnectedIconPicker/types';
import { ConnectedMultiIconPicker } from './ConnectedMultiIconPicker';
import type { MultiIconPickerConfig } from './types';

const kingsId = filled(ActionId.fromSpellId(20217), 'Kings', 'kings.jpg');
const markId = filled(ActionId.fromSpellId(1126), 'Mark', 'mark.jpg');
const emperorId = filled(ActionId.fromSpellId(115921), 'Emperor', 'emperor.jpg');
const improvedId = filled(ActionId.fromSpellId(9000), 'Improved', 'improved.jpg');
const categoryId = filled(ActionId.fromSpellId(9999), 'Category', 'category.jpg');

/** Stands in for the player: one number per buff, plus the `(onChange) => unsubscribe` contract. */
class Buffs {
	private listeners = new Set<() => void>();
	readonly values: Record<string, number> = { kings: 0, mark: 0, emperor: 0 };
	visible = true;
	set(field: string, value: number) {
		this.values[field] = value;
		this.notify();
	}
	setVisible(next: boolean) {
		this.visible = next;
		this.notify();
	}
	notify() {
		this.listeners.forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
}

const input = (field: string, actionId: ActionId, extra: Partial<IconPickerConfig<Buffs, number>> = {}): IconPickerConfig<Buffs, number> => ({
	actionId,
	states: 2,
	storeSubscribe: buffs => buffs.subscribe,
	getValue: buffs => buffs.values[field],
	setValue: (buffs, value) => buffs.set(field, Number(value)),
	...extra,
});

const configFor = (extra: Partial<MultiIconPickerConfig<Buffs>> = {}): MultiIconPickerConfig<Buffs> => ({
	inputs: [input('kings', kingsId), input('mark', markId), input('emperor', emperorId)],
	label: 'Stats',
	...extra,
});

const mount = (buffs: Buffs, config: MultiIconPickerConfig<Buffs> = configFor(), onClear = () => {}) =>
	render(<ConnectedMultiIconPicker modObject={buffs} config={config} subscribe={buffs.subscribe} onClear={onClear} />);

const trigger = () => screen.getByTestId('multi-icon-picker-root') as HTMLAnchorElement;
const menu = () => screen.queryByTestId('multi-icon-picker-menu') as HTMLUListElement;
const open = () =>
	act(() => {
		fireEvent.mouseEnter(trigger());
		fireEvent.mouseMove(trigger());
	});
const backgroundIcon = () => trigger().style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

beforeEach(() => {
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
		return this;
	});
});
afterEach(() => vi.restoreAllMocks());

describe('ConnectedMultiIconPicker', () => {
	it('renders a connected icon picker per input, each toggling its own field', () => {
		const buffs = new Buffs();
		mount(buffs);
		open();
		const options = Array.from(menu().children).slice(1);
		expect(options.every(option => within(option as HTMLElement).queryByTestId('icon-picker-root'))).toBe(true);

		fireEvent.click(within(options[1] as HTMLElement).getByTestId('icon-picker-root'));
		expect(buffs.values.mark).toBe(1);
	});

	it("renders each input as a bare swatch, without the input's label", () => {
		const buffs = new Buffs();
		mount(buffs, configFor({ inputs: [input('kings', kingsId, { label: 'Blessing of Kings' }), input('mark', markId)] }));
		open();

		expect(within(menu()).getAllByTestId('icon-picker-root')).toHaveLength(2);
		expect(within(menu()).queryByText('Blessing of Kings')).toBeNull();
	});

	it('takes the button image from the first active child, not the last, and from the first child while none is on', () => {
		const buffs = new Buffs();
		mount(buffs);
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('kings.jpg');

		act(() => buffs.set('emperor', 1));
		expect(trigger().hasAttribute('data-active')).toBe(true);
		expect(backgroundIcon()).toBe('emperor.jpg');

		// Kings is earlier in the list, so it takes the icon over the one already on.
		act(() => buffs.set('kings', 1));
		expect(backgroundIcon()).toBe('kings.jpg');

		act(() => buffs.set('kings', 0));
		act(() => buffs.set('emperor', 0));
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('kings.jpg');
	});

	describe('state badges on the trigger', () => {
		const improved2Id = filled(ActionId.fromSpellId(9001), 'Improved2', 'improved2.jpg');
		const badge = (n: 1 | 2) => screen.queryByTestId(`multi-icon-picker-improved${n}`);
		const quad = () => configFor({ inputs: [input('kings', kingsId, { states: 4, improvedId, improvedId2: improved2Id })] });

		it("keeps the active input's own icon and lights its improved badges by its value", () => {
			const buffs = new Buffs();
			buffs.set('kings', 2);
			mount(buffs, quad());
			expect(backgroundIcon()).toBe('kings.jpg');
			expect(badge(1)!.style.backgroundImage).toContain('improved.jpg');
			expect(badge(1)!.hasAttribute('data-active')).toBe(true);
			expect(badge(2)!.style.backgroundImage).toContain('improved2.jpg');
			expect(badge(2)!.hasAttribute('data-active')).toBe(false);
		});

		it('leaves the badges unlinked and still clears on a right click with them on', () => {
			const buffs = new Buffs();
			buffs.set('kings', 2);
			const onClear = vi.fn();
			mount(buffs, quad(), onClear);
			expect(badge(1)!.hasAttribute('href')).toBe(false);
			fireEvent.mouseDown(trigger(), { button: 2 });
			expect(onClear).toHaveBeenCalledTimes(1);
		});

		it('carries none under a category icon', () => {
			const buffs = new Buffs();
			buffs.set('kings', 2);
			mount(buffs, { ...quad(), categoryId });
			expect(badge(1)).toBeNull();
			expect(backgroundIcon()).toBe('category.jpg');
		});
	});

	it('prefers categoryId over the active child, and keeps it while nothing is on', () => {
		const buffs = new Buffs();
		mount(buffs, configFor({ categoryId }));

		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('category.jpg');

		act(() => buffs.set('kings', 1));
		expect(trigger().hasAttribute('data-active')).toBe(true);
		expect(backgroundIcon()).toBe('category.jpg');
	});

	it('unmounts when showWhen goes false and mounts again when it comes back', () => {
		const buffs = new Buffs();
		mount(buffs, configFor({ showWhen: (player: any) => (player as Buffs).visible }));
		expect(screen.queryByTestId('multi-icon-picker-root')).toBeTruthy();

		act(() => buffs.setVisible(false));
		expect(screen.queryByTestId('multi-icon-picker-root')).toBeNull();

		act(() => buffs.setVisible(true));
		expect(screen.queryByTestId('multi-icon-picker-root')).toBeTruthy();
	});

	it('hands the clear to its caller', () => {
		const onClear = vi.fn();
		mount(new Buffs(), configFor(), onClear);
		fireEvent.mouseDown(trigger(), { button: 2 });
		expect(onClear).toHaveBeenCalledTimes(1);
	});

	it('marks the trigger kept while keptBy holds, and follows the source', () => {
		const buffs = new Buffs();
		buffs.set('kings', 1);
		mount(buffs, configFor({ keptBy: obj => obj.values.kings > 0 }));
		expect(trigger().parentElement!.hasAttribute('data-kept')).toBe(true);
		expect(screen.getByTestId('icon-swatch-kept').getAttribute('data-tooltip-content')).toBe(i18n.t('stat_filters.kept_tooltip'));

		act(() => buffs.set('kings', 0));
		expect(trigger().parentElement!.hasAttribute('data-kept')).toBe(false);
		expect(screen.queryByTestId('icon-swatch-kept')).toBeNull();
	});

	describe('as a stat option', () => {
		const statOption = (shown: Array<Stat>) => {
			const option = { stats: [Stat.StatStamina], config: configFor() } as unknown as PickerStatOptions;
			const [row] = relevantStatOptions([option], { includeBuffDebuffInputs: [], excludeBuffDebuffInputs: [] }, new Set(shown));
			return row.config as unknown as MultiIconPickerConfig<Buffs>;
		};
		const offeredIcons = () => within(menu()).queryAllByTestId('icon-picker-root');

		it('offers only the picked inputs of a group the stat filter keeps, each marked kept', () => {
			const buffs = new Buffs();
			buffs.set('mark', 1);
			mount(buffs, statOption([]));
			open();

			expect(offeredIcons().map(icon => icon.style.backgroundImage)).toEqual([expect.stringContaining('mark.jpg')]);
			expect(offeredIcons()[0].parentElement!.hasAttribute('data-kept')).toBe(true);
			expect(within(menu()).getByTestId('icon-swatch-kept').getAttribute('data-tooltip-content')).toBe(i18n.t('stat_filters.kept_tooltip'));
			expect(menu().querySelectorAll('li:not(:empty)')).toHaveLength(2);
		});

		it('offers every input of a group whose stat is shown, none marked kept', () => {
			const buffs = new Buffs();
			buffs.set('mark', 1);
			mount(buffs, statOption([Stat.StatStamina]));
			open();

			expect(offeredIcons()).toHaveLength(3);
			expect(within(menu()).queryByTestId('icon-swatch-kept')).toBeNull();
		});
	});

	it('follows a new config without a store notification', () => {
		const buffs = new Buffs();
		const { rerender } = mount(buffs, configFor({ keptBy: () => true }));
		expect(trigger().parentElement!.hasAttribute('data-kept')).toBe(true);

		rerender(<ConnectedMultiIconPicker modObject={buffs} config={configFor({ keptBy: () => false })} subscribe={buffs.subscribe} onClear={() => {}} />);
		expect(trigger().parentElement!.hasAttribute('data-kept')).toBe(false);
	});

	it('follows a new showWhen and new inputs without a store notification', () => {
		const buffs = new Buffs();
		buffs.set('mark', 1);
		const { rerender } = mount(buffs, configFor({ showWhen: () => false }));
		expect(screen.queryByTestId('multi-icon-picker-root')).toBeNull();

		rerender(<ConnectedMultiIconPicker modObject={buffs} config={configFor({ showWhen: () => true })} subscribe={buffs.subscribe} onClear={() => {}} />);
		expect(backgroundIcon()).toBe('mark.jpg');

		const kingsOnly = configFor({ inputs: [input('kings', kingsId)] });
		rerender(<ConnectedMultiIconPicker modObject={buffs} config={kingsOnly} subscribe={buffs.subscribe} onClear={() => {}} />);
		expect(trigger().hasAttribute('data-active')).toBe(false);
	});
});
