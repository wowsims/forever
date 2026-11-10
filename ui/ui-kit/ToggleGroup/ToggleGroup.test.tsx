import { fireEvent, render, screen } from '@testing-library/react';
import type { ControlSize } from '@ui-kit/Field';
import { describe, expect, it, vi } from 'vitest';

import { ToggleGroup } from './ToggleGroup';

const OPTIONS = [
	{ value: 'one', label: 'One' },
	{ value: 'two', label: 'Two' },
	{ value: 'three', label: 'Three' },
] as const;

const mount = (value: 'one' | 'two' | 'three', size?: ControlSize) => {
	const onChange = vi.fn();
	return { onChange, ...render(<ToggleGroup options={OPTIONS} value={value} onChange={onChange} aria-label="Pick one" size={size} />) };
};

const pressed = () => screen.getAllByRole('button').map(button => button.getAttribute('aria-pressed'));

describe('ToggleGroup', () => {
	it('renders one button per option inside a labelled group', () => {
		mount('one');

		const group = screen.getByRole('group', { name: 'Pick one' });
		expect(screen.getAllByRole('button').map(button => [button.parentElement, button.textContent])).toEqual([
			[group, 'One'],
			[group, 'Two'],
			[group, 'Three'],
		]);
	});

	it('presses the option it is given, not the one that was clicked last', () => {
		const { onChange } = mount('two');

		expect(pressed()).toEqual(['false', 'true', 'false']);
		fireEvent.click(screen.getByRole('button', { name: 'Three' }));
		expect(onChange).toHaveBeenCalledWith('three');
		expect(pressed()).toEqual(['false', 'true', 'false']);
	});

	it('reports nothing for the option already pressed', () => {
		const { onChange } = mount('one');

		fireEvent.click(screen.getByRole('button', { name: 'One' }));
		expect(onChange).not.toHaveBeenCalled();
	});

	it('disables every option while disabled', () => {
		const onChange = vi.fn();
		render(<ToggleGroup options={OPTIONS} value="one" onChange={onChange} aria-label="Pick one" disabled />);

		expect(screen.getAllByRole<HTMLButtonElement>('button').map(button => button.disabled)).toEqual([true, true, true]);
		fireEvent.click(screen.getByRole('button', { name: 'Two' }));
		expect(onChange).not.toHaveBeenCalled();
	});

	it.each(['xs', 'sm', 'md'] as const)('passes the %s size to the group and every button', size => {
		mount('one', size);

		expect(screen.getByRole('group').getAttribute('data-size')).toBe(size);
		expect(screen.getAllByRole('button').map(button => button.getAttribute('data-size'))).toEqual([size, size, size]);
	});
});
