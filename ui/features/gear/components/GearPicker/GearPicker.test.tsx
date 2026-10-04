import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { LEFT_ITEM_SLOTS, RIGHT_ITEM_SLOTS } from '../../model/gear_data';

vi.mock('./ItemPickerCell', () => ({
	ItemPickerCell: ({ slot, ready, align }: any) => <div data-cell={slot} data-ready={String(ready)} data-align={align ?? 'start'} />,
}));

const { GearPicker } = await import('./GearPicker');

describe('GearPicker', () => {
	const cells = (container: HTMLElement) =>
		Array.from(container.querySelectorAll<HTMLElement>('[data-cell]')).map(cell => ({ slot: Number(cell.dataset.cell), align: cell.dataset.align }));

	it('lays the left slots out from the start and the right slots against their icons at the end', () => {
		const { container } = render(<GearPicker ready />);
		expect(cells(container)).toEqual([
			...LEFT_ITEM_SLOTS.map(slot => ({ slot, align: 'start' })),
			...RIGHT_ITEM_SLOTS.map(slot => ({ slot, align: 'end' })),
		]);
	});

	it('hands every cell the ready flag', () => {
		const { container } = render(<GearPicker ready={false} />);
		expect(new Set(Array.from(container.querySelectorAll<HTMLElement>('[data-cell]')).map(cell => cell.dataset.ready))).toEqual(new Set(['false']));
	});
});
