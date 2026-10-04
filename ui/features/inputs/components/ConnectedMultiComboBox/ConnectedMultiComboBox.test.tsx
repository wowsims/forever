import { FakeSource } from '@testing/inputs';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ConnectedMultiComboBox } from './ConnectedMultiComboBox';
import type { MultiComboBoxConfig } from './types';

type Model = FakeSource<Array<number>>;

const configFor = (overrides: Partial<MultiComboBoxConfig<Model>> = {}): MultiComboBoxConfig<Model> => ({
	id: 'fruit',
	label: 'Fruit',
	placeholder: 'Pick fruit',
	emptyText: 'Nothing matches',
	removeLabel: 'Remove',
	openLabel: 'Show all',
	values: [
		{ name: 'Apple', value: 1 },
		{ name: 'Banana', value: 2 },
		{ name: 'Cherry', value: 3 },
	],
	storeSubscribe: (model: Model) => model.subscribe,
	getValue: (model: Model) => model.value,
	setValue: (model: Model, next: Array<number>) => model.set(next),
	...overrides,
});

let model: Model;
const mount = (config = configFor(), props: { testId?: string } = {}) => render(<ConnectedMultiComboBox modObject={model} config={config} {...props} />);
const chips = () => screen.queryAllByTestId('multi-combo-box-chip').map(chip => chip.textContent);
const options = () => within(screen.getByTestId('multi-combo-box-list')).getAllByRole('option');

beforeEach(() => {
	model = new FakeSource<Array<number>>([]);
});

describe('ConnectedMultiComboBox', () => {
	it('shows the source picks and follows the source when it changes', () => {
		model.value = [1];
		mount(configFor(), { testId: 'bound-root' });
		expect(chips()).toEqual(['Apple']);
		expect(screen.getByTestId('bound-root').getAttribute('data-testid')).toBe('bound-root');

		act(() => model.set([2, 3]));
		expect(chips()).toEqual(['Banana', 'Cherry']);
	});

	it('writes a pick back through setValue', () => {
		model.value = [1];
		mount();

		fireEvent.click(screen.getByTestId('multi-combo-box-trigger'));
		fireEvent.click(options()[2]);
		expect(model.value).toEqual([1, 3]);
	});

	it('renders nothing when hidden', () => {
		const { container } = mount(configFor({ showWhen: () => false }));
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		mount(configFor({ enableWhen: () => false }));
		expect((screen.getByTestId('multi-combo-box-input') as HTMLInputElement).disabled).toBe(true);
		expect(screen.getByTestId('multi-combo-box-root').hasAttribute('data-disabled')).toBe(true);
	});
});

vi.mock('@ui-kit/hooks/usePortalContainer', () => ({ usePortalContainer: () => null }));
