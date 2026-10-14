import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { MultiComboBox } from './MultiComboBox';
import type { MultiComboBoxProps } from './types';

interface HarnessHandle {
	set: (next: Array<number>) => void;
}

type HarnessProps = Partial<Omit<MultiComboBoxProps, 'value' | 'onChange'>> & {
	initial?: Array<number>;
	onChange?: (next: Array<number>) => void;
	handle?: Ref<HarnessHandle>;
};

const Harness = ({ initial = [], onChange = () => {}, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }));
	return (
		<MultiComboBox
			id="fruit"
			label="Fruit"
			placeholder="Pick fruit"
			emptyText="Nothing matches"
			removeLabel="Remove"
			openLabel="Show all"
			values={[
				{ name: 'Apple', value: 1 },
				{ name: 'Banana', value: 2 },
				{ name: 'Cherry', value: 3 },
			]}
			{...props}
			value={value}
			onChange={next => {
				onChange(next);
				setValue(next);
			}}
		/>
	);
};

const input = () => screen.getByTestId('multi-combo-box-input') as HTMLInputElement;
const chips = () => screen.queryAllByTestId('multi-combo-box-chip').map(chip => chip.textContent);
const options = () => within(screen.getByTestId('multi-combo-box-list')).getAllByRole('option');
const open = () => fireEvent.click(screen.getByTestId('multi-combo-box-trigger'));

describe('MultiComboBox', () => {
	it('labels its input, keeps its placeholder, and shows the picks as chips below the field', () => {
		const handle = { current: null as HarnessHandle | null };
		render(<Harness handle={handle} />);

		expect(screen.getByText('Fruit').closest('label')!.getAttribute('for')).toBe('fruit');
		expect(input().id).toBe('fruit');
		expect(input().placeholder).toBe('Pick fruit');
		expect(chips()).toEqual([]);

		act(() => handle.current!.set([2]));
		expect(chips()).toEqual(['Banana']);
		expect(input().placeholder).toBe('Pick fruit');
		const field = screen.getByTestId('multi-combo-box-field');
		const chipRow = screen.getByTestId('multi-combo-box-chips');
		expect(field.contains(chipRow)).toBe(false);
		expect(field.compareDocumentPosition(chipRow) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
	});

	it('lists every option from the chevron, picked ones included, and toggles one per click', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1]} onChange={onChange} />);

		open();
		expect(options().map(option => option.textContent)).toEqual(['Apple', 'Banana', 'Cherry']);
		expect(options()[0].getAttribute('aria-selected')).toBe('true');

		fireEvent.click(options()[2]);
		expect(onChange).toHaveBeenLastCalledWith([1, 3]);
		fireEvent.click(options()[0]);
		expect(onChange).toHaveBeenLastCalledWith([3]);
	});

	it('narrows the list as the input is typed into, and says so when nothing matches', () => {
		render(<Harness />);

		open();
		fireEvent.change(input(), { target: { value: 'an' } });
		expect(options().map(option => option.textContent)).toEqual(['Banana']);

		fireEvent.change(input(), { target: { value: 'zz' } });
		expect(screen.getByText('Nothing matches')).toBeTruthy();
	});

	it('drops a pick through its chip', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1, 2]} onChange={onChange} />);

		fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0]);
		expect(onChange).toHaveBeenCalledExactlyOnceWith([2]);
		expect(chips()).toEqual(['Banana']);
	});

	it('takes the root test id from testId', () => {
		render(<Harness testId="custom-root" />);
		expect(screen.getByTestId('custom-root').contains(input())).toBe(true);
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(input().disabled).toBe(true);
		expect(screen.getByTestId('multi-combo-box-root').hasAttribute('data-disabled')).toBe(true);
	});

	it('keeps its picks in view while disabled, with no way to remove one', () => {
		render(<Harness initial={[1, 3]} disabled />);

		expect(chips()).toEqual(['Apple', 'Cherry']);
		for (const chip of screen.getAllByTestId('multi-combo-box-chip')) {
			expect(chip.hasAttribute('data-disabled')).toBe(false);
			expect(chip.getAttribute('aria-disabled')).toBe('true');
		}
		expect(screen.queryAllByRole('button', { name: 'Remove' })).toEqual([]);
	});
});

vi.mock('@ui-kit/hooks/usePortalContainer', () => ({ usePortalContainer: () => null }));
