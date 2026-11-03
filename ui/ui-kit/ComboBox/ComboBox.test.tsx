import { fireEvent, render, screen, within } from '@testing-library/react';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { describe, expect, it, vi } from 'vitest';

import { ComboBox, type ComboBoxProps } from './ComboBox';

const ITEMS = ['alpha', 'beta', 'gamma'];

type Overrides = Partial<ComboBoxProps<string>>;

// Base UI marks the field inert while the popup is open, which hides the label and the clear
// button from every role/label query, so the shell tests keep the popup closed.
const propsFor = (overrides: Overrides = {}): ComboBoxProps<string> => ({
	value: '',
	onChange: vi.fn(),
	items: ITEMS,
	itemKey: item => item,
	renderItem: item => <span>{item}</span>,
	onItemSelect: vi.fn(),
	open: false,
	onOpenChange: vi.fn(),
	inputTestId: 'combo-box-input',
	...overrides,
});

const mount = (overrides: Overrides = {}) => render(<ComboBox {...propsFor(overrides)} />);

const input = () => screen.getByTestId('combo-box-input') as HTMLInputElement;
const list = () => screen.getByRole('listbox');

describe('ComboBox', () => {
	it('renders a label only when one is given, wired to the input by id', () => {
		const { rerender } = mount({ id: 'the-id' });
		expect(screen.queryByText('Item Name')).toBeNull();

		rerender(<ComboBox {...propsFor({ id: 'the-id', label: 'Item Name' })} />);
		expect(screen.getByText('Item Name').closest('label')!.getAttribute('for')).toBe('the-id');
		expect(input().id).toBe('the-id');
	});

	it('offers no clear button while the value is empty, nor when the box is not clearable', () => {
		const { rerender } = mount({ clearable: true, value: '', clearLabel: 'Clear search' });
		expect(screen.queryByTestId('combo-box-clear-btn')).toBeNull();

		rerender(<ComboBox {...propsFor({ clearable: false, value: 'alp', clearLabel: 'Clear search' })} />);
		expect(screen.queryByTestId('combo-box-clear-btn')).toBeNull();

		rerender(<ComboBox {...propsFor({ clearable: true, value: 'alp', clearLabel: 'Clear search' })} />);
		expect(screen.queryByTestId('combo-box-clear-btn')).not.toBeNull();
	});

	it('labels the clear button and reports an empty value when it is pressed', () => {
		const onChange = vi.fn();
		mount({ clearable: true, value: 'alp', clearLabel: 'Clear search', onChange });

		const clearButton = screen.getByRole('button', { name: 'Clear search' });
		expect(clearButton.getAttribute('data-testid')).toBe('combo-box-clear-btn');

		fireEvent.click(clearButton);
		expect(onChange).toHaveBeenLastCalledWith('');
	});

	it('reports typing through onChange', () => {
		const onChange = vi.fn();
		mount({ onChange });

		fireEvent.change(input(), { target: { value: 'alp' } });
		expect(onChange).toHaveBeenLastCalledWith('alp');
	});

	it('puts the caller renderItem output, with its index, inside each option', () => {
		mount({
			open: true,
			renderItem: (item, index) => <span data-testid={`row-${item}`}>{`${index}:${item}`}</span>,
		});

		const options = screen.getAllByRole('option');
		expect(options).toHaveLength(ITEMS.length);
		ITEMS.forEach((item, index) => {
			expect(within(options[index]).getByTestId(`row-${item}`).textContent).toBe(`${index}:${item}`);
		});
	});

	it('hands the pressed item to onItemSelect without reporting it as typed input', () => {
		const onItemSelect = vi.fn();
		const onChange = vi.fn();
		mount({ open: true, onItemSelect, onChange });

		fireEvent.click(screen.getAllByRole('option')[1]);
		expect(onItemSelect).toHaveBeenCalledWith('beta');
		expect(onChange).not.toHaveBeenCalled();
	});

	it('renders the options as a list, with the list test id', () => {
		mount({ open: true, listTestId: 'the-list' });

		expect(list().tagName).toBe('UL');
		expect(list().getAttribute('data-testid')).toBe('the-list');
		expect(screen.getAllByRole('option').map(option => option.tagName)).toEqual(['LI', 'LI', 'LI']);
	});

	it('renders the footer as a sibling after the scrolling list, not inside it', () => {
		mount({ open: true, footer: <div data-testid="the-footer">3 more</div> });

		const footer = screen.getByTestId('the-footer');
		expect(list().contains(footer)).toBe(false);
		expect(footer.parentElement).toBe(list().parentElement);
		expect(list().nextElementSibling).toBe(footer);
	});

	it('renders a footer button at the default size whatever Field the box sits in', () => {
		render(
			<Field size="xs">
				<ComboBox {...propsFor({ open: true, footer: <Button data-testid="the-footer-button">Add</Button> })} />
			</Field>,
		);

		expect(screen.getByTestId('the-footer-button').getAttribute('data-size')).toBe('sm');
	});

	it('flags the list for multi-column striping only when multiColumn is set', () => {
		const { rerender } = mount({ open: true });
		expect(list().hasAttribute('data-multi-column')).toBe(false);

		rerender(<ComboBox {...propsFor({ open: true, multiColumn: true })} />);
		expect(list().hasAttribute('data-multi-column')).toBe(true);
	});

	it('mounts the popup in the portal container from context rather than the document body', () => {
		const host = document.createElement('div');
		document.body.appendChild(host);

		render(
			<PortalContainerContext value={host}>
				<ComboBox {...propsFor({ open: true })} />
			</PortalContainerContext>,
		);
		expect(host.contains(list())).toBe(true);

		host.remove();
	});
});
