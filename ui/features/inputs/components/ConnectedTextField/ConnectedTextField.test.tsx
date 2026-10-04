import { FakeSource } from '@testing/inputs';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectedNumberField, ConnectedNumberListField, ConnectedStringField } from './ConnectedTextField';
import type { NumberFieldConfig, NumberListFieldConfig, StringFieldConfig } from './types';

class Settings<V> extends FakeSource<V> {
	writes = 0;
	constructor(
		value: V,
		private readonly quiet = false,
	) {
		super(value);
	}
	override set(next: V) {
		const changed = JSON.stringify(next) !== JSON.stringify(this.value);
		this.value = next;
		// `quiet` is the APL row source: it relays a change and swallows a write that moves nothing.
		if (changed || !this.quiet) this.notify();
	}
}

const binding = <V,>() => ({
	storeSubscribe: (settings: Settings<V>) => settings.subscribe,
	getValue: (settings: Settings<V>) => settings.value,
	setValue: (settings: Settings<V>, value: V) => {
		settings.writes++;
		settings.set(value);
	},
});

const numberConfig = (extra: Partial<NumberFieldConfig<Settings<number>>> = {}): NumberFieldConfig<Settings<number>> => ({
	id: 'cast-delay',
	label: 'Cast Delay',
	...binding<number>(),
	...extra,
});

const stringConfig = (extra: Partial<StringFieldConfig<Settings<string>>> = {}): StringFieldConfig<Settings<string>> => ({
	id: 'custom-name',
	label: 'Custom Name',
	...binding<string>(),
	...extra,
});

const listConfig = (extra: Partial<NumberListFieldConfig<Settings<Array<number>>>> = {}): NumberListFieldConfig<Settings<Array<number>>> => ({
	id: 'thresholds',
	label: 'Thresholds',
	...binding<Array<number>>(),
	...extra,
});

const input = () => screen.getByRole('textbox') as HTMLInputElement;

describe('ConnectedNumberField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings(5);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		expect(input().value).toBe('5');

		fireEvent.change(input(), { target: { value: '7' } });
		fireEvent.blur(input());
		expect(settings.value).toBe(7);
		expect(input().value).toBe('7');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings(1);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		act(() => settings.set(9));
		expect(input().value).toBe('9');
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings(0);
		const { container } = render(<ConnectedNumberField modObject={settings} config={numberConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings(0);
		render(<ConnectedNumberField modObject={settings} config={numberConfig({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(screen.getByTestId('number-picker-root').hasAttribute('data-disabled')).toBe(true);
	});

	it('passes the frame and the options through', () => {
		const settings = new Settings(3);
		render(<ConnectedNumberField modObject={settings} config={numberConfig({ float: true, layout: 'inline' })} testId="iterations-picker" />);
		const root = screen.getByTestId('iterations-picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(screen.getByLabelText('Cast Delay')).toBe(input());
		expect(input().id).toBe('cast-delay');
		expect(input().value).toBe('3.00');
	});

	it('shows the parsed value after a commit a quiet source has nothing to report about', () => {
		const settings = new Settings(12, true);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);

		fireEvent.change(input(), { target: { value: '12abc' } });

		expect(settings.value).toBe(12);
		expect(input().value).toBe('12');
	});

	it('shows the source after a commit the setter clamped to the displayed value', () => {
		const settings = new Settings(500, true);
		render(
			<ConnectedNumberField
				modObject={settings}
				config={numberConfig({
					setValue: (subject, value) => {
						subject.writes++;
						subject.set(Math.min(value, 500));
					},
				})}
			/>,
		);

		fireEvent.input(input(), { target: { value: '999' } });
		fireEvent.change(input());

		expect(settings.writes).toBe(1);
		expect(settings.listenerCount).toBe(1);
		expect(settings.value).toBe(500);
		expect(input().value).toBe('500');
	});
});

describe('ConnectedStringField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings('abc');
		render(<ConnectedStringField modObject={settings} config={stringConfig()} />);
		expect(input().value).toBe('abc');

		fireEvent.change(input(), { target: { value: 'xyz' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toBe('xyz');
		expect(input().value).toBe('xyz');
	});

	it('forwards the config frame to the root', () => {
		const settings = new Settings('');
		render(<ConnectedStringField modObject={settings} config={stringConfig({ layout: 'inline' })} />);
		const root = screen.getByTestId('adaptive-string-picker-root');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(screen.getByLabelText('Custom Name')).toBe(input());
		expect(input().id).toBe('custom-name');
	});

	it('renders nothing when showWhen is false, and disables the input when enableWhen is false', () => {
		const settings = new Settings('');
		const { container, rerender } = render(<ConnectedStringField modObject={settings} config={stringConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();

		rerender(<ConnectedStringField modObject={settings} config={stringConfig({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(screen.getByTestId('adaptive-string-picker-root').hasAttribute('data-disabled')).toBe(true);
	});
});

describe('ConnectedNumberListField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings([1, 2, 3]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig()} />);
		expect(input().value).toBe('1,2,3');

		fireEvent.change(input(), { target: { value: '4,5' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toEqual([4, 5]);
		expect(input().value).toBe('4,5');
	});

	it('forwards the config frame and placeholder', () => {
		const settings = new Settings<Array<number>>([]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig({ placeholder: 'e.g. 1,2,3' })} />);
		expect(screen.getByLabelText('Thresholds')).toBe(input());
		expect(input().id).toBe('thresholds');
		expect(input().placeholder).toBe('e.g. 1,2,3');
	});

	it('disables the input when enableWhen is false', () => {
		const settings = new Settings<Array<number>>([]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig({ label: undefined, enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
	});
});
