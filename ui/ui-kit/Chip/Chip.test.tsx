import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { Chip, type ChipProps } from './Chip';

const chip = (props: Partial<ChipProps> = {}) =>
	render(<Chip label="Loadout 1" testId="loadout-chip" nameTestId="loadout-name" deleteTestId="loadout-delete" deleteLabel="Delete" {...props} />);

describe('Chip', () => {
	it('renders the label', () => {
		chip();
		expect(screen.getByTestId('loadout-name').textContent).toBe('Loadout 1');
	});

	it('carries no test ids unless the caller names them', () => {
		const { container } = render(<Chip label="Loadout 1" onDelete={() => {}} confirmDelete={false} />);
		expect(container.querySelectorAll('[data-testid]')).toHaveLength(0);
	});

	it('sets data-active when active', () => {
		const { rerender } = chip({ active: false });
		expect(screen.getByTestId('loadout-chip').hasAttribute('data-active')).toBe(false);

		rerender(<Chip label="Loadout 1" testId="loadout-chip" active />);
		expect(screen.getByTestId('loadout-chip').getAttribute('data-active')).toBe('');
	});

	it('sets data-disabled when disabled', () => {
		chip({ disabled: true });
		expect(screen.getByTestId('loadout-chip').getAttribute('data-disabled')).toBe('');
	});

	it('calls onSelect when the name is clicked', () => {
		const onSelect = vi.fn();
		chip({ onSelect });
		fireEvent.click(screen.getByRole('button', { name: 'Loadout 1' }));
		expect(onSelect).toHaveBeenCalledTimes(1);
	});

	it('omits the delete trigger when onDelete is not given', () => {
		chip();
		expect(screen.queryByRole('button', { name: 'Delete' })).toBeNull();
	});

	it('does not call onDelete until the confirm popover is confirmed', () => {
		const onDelete = vi.fn();
		chip({ onDelete, deleteConfirmLabel: 'Confirm' });
		fireEvent.click(screen.getByTestId('loadout-delete'));
		expect(onDelete).not.toHaveBeenCalled();

		fireEvent.click(screen.getByRole('button', { name: 'Confirm' }));
		expect(onDelete).toHaveBeenCalledTimes(1);
	});

	it('calls onDelete immediately when confirmDelete is false', () => {
		const onDelete = vi.fn();
		chip({ onDelete, confirmDelete: false });
		const deleteButton = screen.getByTestId('loadout-delete');
		expect(deleteButton.getAttribute('aria-label')).toBe('Delete');
		fireEvent.click(deleteButton);
		expect(onDelete).toHaveBeenCalledTimes(1);
	});
});
