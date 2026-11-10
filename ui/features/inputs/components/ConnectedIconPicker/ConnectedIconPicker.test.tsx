import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { FakeSource, filled } from '@testing/inputs';
import { deferFill } from '@testing/sim/hooks';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { StrictMode } from 'react';
import { afterEach, beforeEach, describe, expect, expectTypeOf, it, vi } from 'vitest';

import { ConnectedIconPicker } from './ConnectedIconPicker';
import type { IconPickerConfig } from './types';

const buffId = filled(ActionId.fromSpellId(1), 'Buff', 'buff.jpg');
const improvedId = filled(ActionId.fromSpellId(2), 'Improved Buff', 'improved.jpg');

class Settings extends FakeSource<number | boolean> {
	visible = true;
	enabled = true;
	setVisible(next: boolean) {
		this.visible = next;
		this.notify();
	}
}

const configFor = (extra: Partial<IconPickerConfig<Settings, number | boolean>> = {}): IconPickerConfig<Settings, number | boolean> => ({
	id: 'buff-icon',
	actionId: buffId,
	states: 2,
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const mainAnchor = () => screen.getByTestId('icon-picker-root');

afterEach(() => vi.restoreAllMocks());

beforeEach(() => {
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
		return this;
	});
});

describe('ConnectedIconPicker', () => {
	it('resolves the action ids into the icons the picker paints', () => {
		render(<ConnectedIconPicker modObject={new Settings(0)} config={configFor({ states: 3, improvedId })} />);
		expect(mainAnchor().style.backgroundImage).toContain('buff.jpg');
		expect(mainAnchor().getAttribute('aria-label')).toBe('Buff');
		expect(mainAnchor().getAttribute('href')).toBe(ActionId.makeSpellUrl(1));
		expect(screen.getByTestId('icon-input-improved1').getAttribute('href')).toBe(ActionId.makeSpellUrl(2));
	});

	it('draws a blank improved badge, never the counter, while the improved icon loads', async () => {
		const { resolvers } = deferFill();
		const loadingId = ActionId.fromSpellId(2);
		render(<ConnectedIconPicker modObject={new Settings(2)} config={configFor({ states: 4, improvedId: loadingId })} />);
		expect(mainAnchor().textContent).toBe('');
		expect(screen.getByTestId('icon-input-improved1').style.backgroundImage).not.toContain('improved.jpg');

		await act(async () => resolvers[0](filled(loadingId, 'Improved Buff', 'improved.jpg')));
		expect(mainAnchor().textContent).toBe('');
		expect(screen.getByTestId('icon-input-improved1').style.backgroundImage).toContain('improved.jpg');
	});

	it('reads the source as a number and follows it', () => {
		const settings = new Settings(true);
		render(<ConnectedIconPicker modObject={settings} config={configFor()} />);
		expect(mainAnchor().hasAttribute('data-active')).toBe(true);

		act(() => settings.set(false));
		expect(mainAnchor().hasAttribute('data-active')).toBe(false);
	});

	it('writes a boolean, not a number, back to the source at states 2', () => {
		const settings = new Settings(false);
		const setValue = vi.fn((s: Settings, v: number | boolean) => s.set(v));
		render(<ConnectedIconPicker modObject={settings} config={configFor({ setValue })} />);
		fireEvent.click(mainAnchor());
		expect(setValue).toHaveBeenCalledWith(settings, true);
	});

	it('writes a number back to the source above states 2', () => {
		const settings = new Settings(1);
		render(<ConnectedIconPicker modObject={settings} config={configFor({ states: 3 })} />);
		fireEvent.click(mainAnchor());
		expect(settings.value).toBe(2);
	});

	it('renders nothing without an action id', () => {
		render(<ConnectedIconPicker modObject={new Settings(0)} config={configFor({ actionId: null as unknown as ActionId })} />);
		expect(screen.queryByTestId('icon-picker-root')).toBeNull();
	});

	it('carries the config frame onto the field and the testId onto the swatch', () => {
		render(<ConnectedIconPicker modObject={new Settings(0)} config={configFor({ label: 'Buff' })} testId="bound-root" />);
		expect(screen.getByTestId('bound-root').tagName).toBe('A');
		expect(screen.getByTestId('form-label').textContent).toBe('Buff');
	});

	it('disables the picker on enableWhen', () => {
		const settings = new Settings(1);
		render(<ConnectedIconPicker modObject={settings} config={configFor({ states: 3, enableWhen: s => s.enabled })} />);
		expect(screen.getByTestId('icon-picker-root').hasAttribute('data-disabled')).toBe(false);

		act(() => {
			settings.enabled = false;
			settings.notify();
		});
		expect(screen.getByTestId('icon-picker-root').hasAttribute('data-disabled')).toBe(true);
	});

	it('stores the value and zeroes the source when showWhen goes false, and restores it when true again', () => {
		const settings = new Settings(2);
		render(<ConnectedIconPicker modObject={settings} config={configFor({ states: 3, showWhen: s => s.visible })} />);
		expect(settings.value).toBe(2);

		act(() => settings.setVisible(false));
		expect(screen.queryByTestId('icon-picker-root')).toBeNull();
		expect(settings.value).toBe(0);

		act(() => settings.setVisible(true));
		expect(screen.queryByTestId('icon-picker-root')).toBeTruthy();
		expect(settings.value).toBe(2);
	});

	// It stores from its source subscription, so it acts on any notification while hidden — not on
	// a transition — and never during construction.
	it('zeroes a picker that mounts hidden at the first notification, not at mount', () => {
		const settings = new Settings(2);
		settings.visible = false;
		render(<ConnectedIconPicker modObject={settings} config={configFor({ states: 3, showWhen: s => s.visible })} />);
		expect(settings.value).toBe(2);

		act(() => settings.notify());
		expect(settings.value).toBe(0);

		act(() => settings.setVisible(true));
		expect(settings.value).toBe(2);
	});

	it('still writes nothing at mount under StrictMode', () => {
		const settings = new Settings(2);
		settings.visible = false;
		render(
			<StrictMode>
				<ConnectedIconPicker modObject={settings} config={configFor({ states: 3, showWhen: s => s.visible })} />
			</StrictMode>,
		);
		expect(settings.value).toBe(2);

		act(() => settings.notify());
		expect(settings.value).toBe(0);
	});

	it('restores the normalised value, so a numeric source is written back as a boolean at states 2', () => {
		const settings = new Settings(1);
		const setValue = vi.fn((s: Settings, v: number | boolean) => s.set(v));
		render(<ConnectedIconPicker modObject={settings} config={configFor({ setValue, showWhen: s => s.visible })} />);

		act(() => settings.setVisible(false));
		expect(setValue).toHaveBeenLastCalledWith(settings, false);

		act(() => settings.setVisible(true));
		expect(setValue).toHaveBeenLastCalledWith(settings, true);
	});

	it('keeps a value the source took while hidden instead of restoring the stored one over it', () => {
		const settings = new Settings(true);
		render(<ConnectedIconPicker modObject={settings} config={configFor({ showWhen: s => !!s.value })} />);

		fireEvent.click(mainAnchor());
		expect(screen.queryByTestId('icon-picker-root')).toBeNull();
		expect(settings.value).toBe(false);

		act(() => settings.set(true));
		expect(screen.queryByTestId('icon-picker-root')).toBeTruthy();
		expect(settings.value).toBe(true);
	});

	it('marks the swatch kept while keptBy holds, and follows the source', () => {
		const settings = new Settings(true);
		render(<ConnectedIconPicker modObject={settings} config={configFor({ keptBy: s => s.value === true })} />);
		expect(mainAnchor().parentElement!.hasAttribute('data-kept')).toBe(true);
		expect(screen.getByTestId('icon-swatch-kept').getAttribute('data-tooltip-content')).toBe(i18n.t('stat_filters.kept_tooltip'));

		act(() => settings.set(false));
		expect(mainAnchor().parentElement!.hasAttribute('data-kept')).toBe(false);
		expect(screen.queryByTestId('icon-swatch-kept')).toBeNull();
	});
});

describe('IconPickerConfig', () => {
	type WithoutBinding = Omit<IconPickerConfig<Settings, number>, 'storeSubscribe' | 'storeField'>;
	type FieldOnly = WithoutBinding & { storeField: 'player:*' };

	it('will not accept a config that names neither a store subscription nor a store field', () => {
		expectTypeOf<WithoutBinding>().not.toExtend<IconPickerConfig<Settings, number>>();
		expectTypeOf<FieldOnly>().toExtend<IconPickerConfig<Settings, number>>();
	});
});
