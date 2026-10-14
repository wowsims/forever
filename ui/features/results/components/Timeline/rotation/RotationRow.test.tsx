import { actionId, castItem, castRow } from '@testing/timeline_rotation';
import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { AuraItem, CastItem, ContentRow, HeaderRow, ResourceItem, ResourceRow, RowItem, SeparatorRow, TickItem } from '../../../model/timeline/rotation';
import { RotationHeaderRow } from './RotationHeaderRow';
import { RotationRow, RowItemView } from './RotationRow';
import { RotationSeparatorRow } from './RotationSeparatorRow';

vi.mock('./RotationRowIcon', () => ({
	RotationRowIcon: ({ tooltip }: { tooltip?: string }) => <a data-testid="rotation-row-icon" data-tooltip-kind={tooltip ?? 'none'} />,
}));

const ROW = castRow('cast:a', 'Fire Bolt', [castItem(0, 1), castItem(2, 3), castItem(4, 5)]);

const mount = (row: ContentRow = ROW, items: Array<number> = [0, 2]) => {
	const onHide = vi.fn();
	return { onHide, ...render(<RotationRow row={row} items={items} onHide={onHide} />) };
};

describe('RotationRow', () => {
	it('renders only the items it is handed, keyed by their index in the row', () => {
		const { container } = mount();
		const items = [...container.querySelectorAll<HTMLElement>('[data-testid="rotation-item-cast"]')];
		expect(items.map(item => item.dataset.itemIndex)).toEqual(['0', '2']);
	});

	it('carries the row key and height the window measured it at', () => {
		const { container } = mount();
		const row = container.querySelector<HTMLElement>('[data-testid="rotation-row"]')!;
		expect(row.dataset.rowKey).toBe('cast:a');
		expect(row.dataset.rowKind).toBe('cast');
		expect(row.style.getPropertyValue('--row-h')).toBe('32');
	});

	it('asks to hide itself by key, and names the row in the button', () => {
		const { container, onHide } = mount();
		const hide = container.querySelector<HTMLButtonElement>('[data-testid="rotation-row-hide"]')!;
		expect(hide.getAttribute('aria-label')).toBe('results_tab.details.timeline.hide_row_named');
		fireEvent.click(hide);
		expect(onHide).toHaveBeenCalledWith('cast:a');
	});

	it('resolves an aura row’s icon against the buff aura rather than the spell', () => {
		const aura = { ...ROW, kind: 'aura' } as ContentRow;
		expect(mount(aura, []).container.querySelector<HTMLElement>('[data-testid="rotation-row-icon"]')!.dataset.tooltipKind).toBe('buffAura');
		expect(mount(ROW, []).container.querySelector<HTMLElement>('[data-testid="rotation-row-icon"]')!.dataset.tooltipKind).toBe('spell');
	});

	it('paints a resource row’s icon from its own url, since it has no action', () => {
		const resource = {
			kind: 'resource',
			key: 'res:mana',
			section: 'player',
			height: 32,
			label: 'Mana',
			icon: 'mana.png',
			cssName: 'mana',
			items: [],
			maxRightUpTo: [],
		} as ResourceRow;
		const icon = mount(resource, []).container.querySelector<HTMLElement>('[data-testid="rotation-row-icon"]')!;
		expect(icon.style.backgroundImage).toContain('mana.png');
		expect(icon.dataset.tooltipKind).toBeUndefined();
	});
});

describe('RotationHeaderRow', () => {
	it('labels the section and shows an icon only when it has an action', () => {
		const withIcon = { kind: 'header', key: 'header:a', section: 'a', height: 32, label: 'Target 1', actionId: actionId('Boss') } as HeaderRow;
		expect(render(<RotationHeaderRow row={withIcon} />).container.querySelector('[data-testid="rotation-row-icon"]')).not.toBeNull();

		const plain = { ...withIcon, actionId: null } as HeaderRow;
		const { container } = render(<RotationHeaderRow row={plain} />);
		expect(container.querySelector('[data-testid="rotation-row-icon"]')).toBeNull();
		expect(container.querySelector('[data-testid="rotation-label-text"]')!.textContent).toBe('Target 1');
		expect(container.querySelector('[data-testid="rotation-row-hide"]')).toBeNull();
	});
});

describe('RotationSeparatorRow', () => {
	it('is an empty row of its own height', () => {
		const { container } = render(<RotationSeparatorRow row={{ kind: 'separator', key: 'sep:a', section: 'a', height: 17 } as SeparatorRow} />);
		const row = container.querySelector<HTMLElement>('[data-row-kind="separator"]')!;
		expect(row.style.getPropertyValue('--row-h')).toBe('17');
		expect(row.textContent).toBe('');
	});
});

const mountItem = (item: RowItem, index = 3, iconUrl = 'bolt.png', cssName = 'mana') =>
	render(<RowItemView item={item} index={index} iconUrl={iconUrl} cssName={cssName} />).container.firstElementChild as HTMLElement;

const cast = (item: Partial<CastItem>) =>
	({ kind: 'cast', outcome: 'hit', cancelled: false, travelStart: null, travelDuration: null, stack: [], ...item }) as CastItem;

const vars = (element: HTMLElement) => ({ t: element.style.getPropertyValue('--t'), dur: element.style.getPropertyValue('--dur') });

describe('RowItemView', () => {
	it('draws a cast as a bar carrying its outcome, its span and the row’s icon', () => {
		const element = mountItem(cast({ start: 2, end: 3.5, outcome: 'crit' }));
		expect(element.dataset.outcome).toBe('crit');
		expect(vars(element)).toEqual({ t: '2', dur: '1.5' });
		expect(element.dataset.itemIndex).toBe('3');
		expect(element.querySelector<HTMLElement>('[data-testid="rotation-item-icon"]')!.style.backgroundImage).toContain('bolt.png');
		expect(element.querySelector<HTMLElement>('[data-testid="rotation-item-travel"]')!.hidden).toBe(true);
	});

	it('marks a cancelled cast and shows the travel leg when there is one', () => {
		const element = mountItem(cast({ start: 0, end: 1, outcome: 'none', cancelled: true, travelStart: 1, travelDuration: 0.4 }));
		expect(element.dataset.outcome).toBe('cancelled');
		const travel = element.querySelector<HTMLElement>('[data-testid="rotation-item-travel"]')!;
		expect(travel.hidden).toBe(false);
		expect(vars(travel)).toEqual({ t: '1', dur: '0.4' });
	});

	it('counts the casts a stack hides under its icon, and leaves a lone cast unmarked', () => {
		const stacked = (count: number) =>
			mountItem(cast({ start: 2, end: 2, stack: Array(count).fill({}) })).querySelector('[data-testid="rotation-item-stack-count"]');
		expect(stacked(3)!.textContent).toBe('3');
		expect(stacked(1)).toBeNull();
	});

	it('draws a tick as a mark at its own time, with no span', () => {
		const element = mountItem({ kind: 'tick', start: 7, end: 7 } as TickItem);
		expect(element.dataset.testid).toBe('rotation-item-tick');
		expect(vars(element)).toEqual({ t: '7', dur: '' });
	});

	it('draws an aura as a span, with one block per stack segment', () => {
		const element = mountItem({
			kind: 'aura',
			start: 1,
			end: 5,
			sharesRowWithCast: true,
			stacks: [
				{ offset: 0, duration: 2, stacks: 3 },
				{ offset: 2, duration: 2, stacks: 5 },
			],
		} as AuraItem);
		expect(element.hasAttribute('data-shares-row')).toBe(true);
		const stacks = [...element.querySelectorAll<HTMLElement>('[data-testid="rotation-item-stacks"]')];
		expect(stacks.map(stack => stack.textContent)).toEqual(['3', '5']);
		expect(vars(stacks[1])).toEqual({ t: '2', dur: '2' });
	});

	it('draws a resource block, filling it only when the row asks for a fill', () => {
		const numeric = mountItem({ kind: 'resource', start: 0, end: 2, display: 'number', fillPercent: 0, text: '42' } as ResourceItem);
		expect(numeric.dataset.testid).toBe('series-color');
		expect(numeric.querySelector<HTMLElement>('[data-testid="rotation-item-resource-fill"]')!.hidden).toBe(true);
		expect(numeric.querySelector('[data-testid="rotation-item-resource-text"]')!.textContent).toBe('42');

		const filled = mountItem({ kind: 'resource', start: 0, end: 2, display: 'fill', fillPercent: 65, text: '65%' } as ResourceItem);
		const fill = filled.querySelector<HTMLElement>('[data-testid="rotation-item-resource-fill"]')!;
		expect(fill.hidden).toBe(false);
		expect(fill.style.getPropertyValue('--fill')).toBe('65');
	});
});
