import { FakeSource } from '@testing/inputs';
import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import { describe, expect, it } from 'vitest';

import type { InputConfig } from '../../input';
import { ConnectedChoiceField } from './ConnectedChoiceField';
import type { ChoiceFieldConfig } from './types';

const values = [
	{ name: 'Passive', value: 0, tooltip: 'Do nothing' },
	{ name: 'Aggressive', value: 1, tooltip: 'Attack everything' },
	{ name: 'Defensive', value: 2, tooltip: 'Attack when attacked' },
];

const enumConfig = (extra: Partial<ChoiceFieldConfig<FakeSource<number>>> = {}): ChoiceFieldConfig<FakeSource<number>> => ({
	id: 'pet-mode',
	label: 'Pet Mode',
	values,
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const KINDS: Array<ChoiceOption<number>> = [
	{ value: 0, label: 'None' },
	{ value: 1, label: 'Cast' },
	{ value: 2, label: 'Wait' },
];

const trigger = () => screen.getByRole('combobox');
const pick = async (label: string) => {
	await userEvent.click(trigger());
	await userEvent.click(screen.getByRole('option', { name: label }));
};

describe('ConnectedChoiceField', () => {
	describe('an enum config', () => {
		it('offers the config values, shows the source value and writes back a number through setValue', async () => {
			const settings = new FakeSource(1);
			render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
			expect(trigger().textContent).toBe('Aggressive');

			await userEvent.click(trigger());
			expect(screen.getAllByRole('option').map(option => [option.textContent, option.getAttribute('data-tooltip-content')])).toEqual([
				['Passive', 'Do nothing'],
				['Aggressive', 'Attack everything'],
				['Defensive', 'Attack when attacked'],
			]);
			await userEvent.click(screen.getByRole('option', { name: 'Defensive' }));

			expect(settings.value).toBe(2);
			expect(trigger().textContent).toBe('Defensive');
		});

		it('re-renders when the source changes underneath it', () => {
			const settings = new FakeSource(0);
			render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
			act(() => settings.set(2));
			expect(trigger().textContent).toBe('Defensive');
		});

		it('passes the frame, the aria label and the test id through', () => {
			const settings = new FakeSource(0);
			const { container } = render(
				<ConnectedChoiceField modObject={settings} config={enumConfig({ layout: 'inline' })} ariaLabel="Mode" testId="picker" />,
			);
			const root = container.firstElementChild!;
			expect(root.getAttribute('data-testid')).toBe('picker');
			expect(root.getAttribute('data-layout')).toBe('inline');
			expect(root.querySelector('label')!.getAttribute('for')).toBe('pet-mode');
			expect(trigger().id).toBe('pet-mode');
			expect(trigger().getAttribute('aria-label')).toBe('Mode');
		});
	});

	describe('options', () => {
		const configFor = (extra: Partial<InputConfig<FakeSource<number>, number>> = {}): InputConfig<FakeSource<number>, number> & { id: string } => ({
			id: 'kind',
			storeSubscribe: settings => settings.subscribe,
			getValue: settings => settings.value,
			setValue: (settings, value) => settings.set(value),
			...extra,
		});

		it('shows the source selection, follows the source, and writes the pick back', async () => {
			const settings = new FakeSource(1);
			render(<ConnectedChoiceField modObject={settings} config={configFor()} options={KINDS} variant="link" />);
			expect(trigger().textContent).toBe('Cast');
			expect(trigger().getAttribute('data-variant')).toBe('link');

			act(() => settings.set(2));
			expect(trigger().textContent).toBe('Wait');

			await pick('None');
			expect(settings.value).toBe(0);
		});

		it('bridges a stored type to the option type through sourceToValue and valueToSource', async () => {
			interface Stored {
				kind: number;
			}
			const settings = new FakeSource<Stored>({ kind: 1 });
			const config: InputConfig<FakeSource<Stored>, Stored, number> & { id: string } = {
				id: 'kind',
				storeSubscribe: target => target.subscribe,
				getValue: target => target.value,
				setValue: (target, value) => target.set(value),
				sourceToValue: stored => stored.kind,
				valueToSource: kind => ({ kind }),
			};
			render(<ConnectedChoiceField modObject={settings} config={config} options={KINDS} />);
			expect(trigger().textContent).toBe('Cast');

			await pick('Wait');
			expect(settings.value).toEqual({ kind: 2 });
		});
	});

	it('renders nothing when showWhen is false', () => {
		const { container } = render(<ConnectedChoiceField modObject={new FakeSource(0)} config={enumConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the trigger and marks the root when enableWhen is false', () => {
		const { container } = render(<ConnectedChoiceField modObject={new FakeSource(0)} config={enumConfig({ enableWhen: () => false })} />);
		expect((trigger() as HTMLButtonElement).disabled).toBe(true);
		expect(container.firstElementChild!.hasAttribute('data-disabled')).toBe(true);
	});
});
