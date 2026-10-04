import { FakeSource } from '@testing/inputs';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectedBooleanPicker } from './ConnectedBooleanPicker';
import type { BooleanPickerConfig } from './types';

type Settings = FakeSource<boolean>;

const configFor = (extra: Partial<BooleanPickerConfig<Settings>> = {}): BooleanPickerConfig<Settings> => ({
	id: 'use-item-swap',
	label: 'Enable Item Swap',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const checkbox = () => screen.getByRole('checkbox') as HTMLButtonElement;
const isChecked = (el: HTMLElement) => el.getAttribute('aria-checked') === 'true';

describe('ConnectedBooleanPicker', () => {
	it('shows the source value and writes back through setValue', () => {
		const settings = new FakeSource(true);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		expect(isChecked(checkbox())).toBe(true);

		fireEvent.click(checkbox());
		expect(settings.value).toBe(false);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new FakeSource(false);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		act(() => settings.set(true));
		expect(isChecked(checkbox())).toBe(true);
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new FakeSource(false);
		const { unmount } = render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new FakeSource(false);
		const { container } = render(<ConnectedBooleanPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new FakeSource(false);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(checkbox().getAttribute('aria-disabled')).toBe('true');
		expect(checkbox().closest('[data-testid="boolean-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('passes the frame and reverse through', () => {
		const settings = new FakeSource(false);
		const { container } = render(<ConnectedBooleanPicker modObject={settings} config={configFor({ reverse: true, layout: 'inline' })} testId="picker" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.hasAttribute('data-reverse')).toBe(true);
		expect(root.querySelector('label')!.getAttribute('for')).toBe('use-item-swap');
		expect(root.querySelector('input')!.id).toBe('use-item-swap');
	});
});
