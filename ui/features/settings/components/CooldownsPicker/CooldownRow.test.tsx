import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CooldownRowView, type CooldownRowViewProps } from './CooldownRow';

const renderRow = (props: Partial<CooldownRowViewProps> = {}) => {
	const onDelete = vi.fn();
	render(
		<CooldownRowView
			name="Recklessness"
			isAdd={false}
			deleteTooltipId="delete-tooltip"
			onDelete={onDelete}
			actionPicker={<span data-testid="action-slot" />}
			timingsPicker={<span data-testid="timings-slot" />}
			{...props}
		/>,
	);
	return onDelete;
};

describe('CooldownRowView', () => {
	it('renders the action picker, the name, the timings picker and the delete button in order', () => {
		renderRow();
		const row = screen.getByTestId('cooldown-picker');

		expect(Array.from(row.children).map(child => child.getAttribute('data-testid'))).toEqual([
			'action-slot',
			'cooldown-picker-label',
			'timings-slot',
			'delete-cooldown',
		]);
		expect(screen.getByTestId('cooldown-picker-label').textContent).toBe('Recklessness');
	});

	it('deletes the row from its delete button', () => {
		const onDelete = renderRow();

		fireEvent.click(screen.getByRole('button', { name: 'rotation_tab.cooldowns.delete_tooltip' }));

		expect(onDelete).toHaveBeenCalledOnce();
	});

	it('marks the trailing add row and hides its delete button', () => {
		renderRow({ isAdd: true, name: undefined });

		expect(screen.getByTestId('cooldown-picker').hasAttribute('data-add')).toBe(true);
		expect(screen.getByTestId('delete-cooldown').hasAttribute('data-hidden')).toBe(true);
		expect(screen.getByTestId('cooldown-picker-label').textContent).toBe('');
	});

	it('anchors the delete button to the shared tooltip', () => {
		renderRow();

		expect(screen.getByTestId('delete-cooldown').getAttribute('data-tooltip-id')).toBe('delete-tooltip');
	});
});
