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

	it('takes the button image from the first active child, not the last', () => {
		const buffs = new Buffs();
		mount(buffs);
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('');

		act(() => buffs.set('emperor', 1));
		expect(trigger().hasAttribute('data-active')).toBe(true);
		expect(backgroundIcon()).toBe('emperor.jpg');

		// Kings is earlier in the list, so it takes the icon over the one already on.
		act(() => buffs.set('kings', 1));
		expect(backgroundIcon()).toBe('kings.jpg');

		act(() => buffs.set('kings', 0));
		act(() => buffs.set('emperor', 0));
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(backgroundIcon()).toBe('');
	});

	it('uses the improved id at state 2', () => {
		const buffs = new Buffs();
		mount(buffs, configFor({ inputs: [input('kings', kingsId, { states: 3, improvedId })] }));

		act(() => buffs.set('kings', 1));
		expect(backgroundIcon()).toBe('kings.jpg');
		act(() => buffs.set('kings', 2));
		expect(backgroundIcon()).toBe('improved.jpg');
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
});
