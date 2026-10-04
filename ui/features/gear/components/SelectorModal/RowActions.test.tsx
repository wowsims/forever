import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { RowActions } from './RowActions';

describe('RowActions', () => {
	it('toggles the favourite without the click reaching the row', () => {
		const onToggleFavourite = vi.fn();
		const onRow = vi.fn();
		render(
			<div onClick={onRow}>
				<RowActions favourited={false} favouriteTooltipId="favourite" onToggleFavourite={onToggleFavourite} />
			</div>,
		);

		fireEvent.click(screen.getByTestId('selector-modal-list-item-favorite'));
		expect(onToggleFavourite).toHaveBeenCalledOnce();
		expect(onRow).not.toHaveBeenCalled();
	});

	it.each([true, false])('marks favourited=%s on the star, for the shared tooltip', favourited => {
		render(<RowActions favourited={favourited} favouriteTooltipId="favourite-7" onToggleFavourite={vi.fn()} />);
		const button = screen.getByTestId('selector-modal-list-item-favorite');
		expect(button.getAttribute('data-favourited')).toBe(String(favourited));
		expect(button.getAttribute('data-tooltip-id')).toBe('favourite-7');
	});

	it('leaves the batch toggle out unless it is given', () => {
		render(<RowActions favourited={false} favouriteTooltipId="favourite" onToggleFavourite={vi.fn()} />);
		expect(screen.queryByTestId('selector-modal-list-item-compare')).toBeNull();
	});

	it('marks and toggles the batch state', () => {
		const onToggle = vi.fn();
		render(
			<RowActions
				favourited={false}
				favouriteTooltipId="favourite"
				onToggleFavourite={vi.fn()}
				compare={{ inBatch: true, tooltipId: 'compare', onToggle }}
			/>,
		);
		const button = screen.getByTestId('selector-modal-list-item-compare');
		expect(button.getAttribute('data-in-batch')).toBe('true');
		expect(button.getAttribute('data-tooltip-id')).toBe('compare');

		fireEvent.click(button);
		expect(onToggle).toHaveBeenCalledOnce();
	});
});
