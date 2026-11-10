import type { AplRun } from '@features/apl/model/apl_entry';
import { type AplRevealRequest, aplReveals } from '@features/apl/model/reveal';
import { TabActivationContext } from '@features/hooks/useActivateTab';
import type { RowItem } from '@features/results/model/timeline/rotation';
import { APLListItem, APLRotation } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { AplEntryPath } from '@sim/proto/combat_log';
import { fakeHost } from '@testing/sim/fake_host';
import { mockSubscriptions } from '@testing/sim/subscriptions';
import { castItem, castRow, rotationModel } from '@testing/timeline_rotation';
import { act, fireEvent, render } from '@testing-library/react';
import type { ReactNode } from 'react';
import { useState } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { DrStickySlotContext } from '../../DetailedResults/context/DrStickySlotContext';
import { RotationTrack } from './RotationTrack';

vi.mock('@sim/state/subscriptions', async () => mockSubscriptions());

// The row icon is `useActionId` plus `useWowheadDataset`; both are covered where they live, and
// letting them run here would put a Wowhead fetch behind every row.
vi.mock('./RotationRowIcon', () => ({ RotationRowIcon: () => <a data-testid="rotation-row-icon" /> }));

// The player's rotation, unchanged since the run in the cases that click a cast.
const ROTATION = APLRotation.create({ priorityList: [APLListItem.create(), APLListItem.create(), APLListItem.create()] });

const host = fakeHost({
	player: { aplRotation: ROTATION },
	sim: { getShowDamageMetrics: () => true, getShowThreatMetrics: () => true, getShowHealingMetrics: () => true } as never,
});

let frames: Map<number, FrameRequestCallback>;
let nextHandle = 0;

beforeEach(() => {
	frames = new Map();
	nextHandle = 0;
	vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => {
		frames.set(++nextHandle, callback);
		return nextHandle;
	});
	vi.stubGlobal('cancelAnimationFrame', (handle: number) => frames.delete(handle));
	for (const name of ['ResizeObserver', 'IntersectionObserver']) {
		vi.stubGlobal(
			name,
			class {
				observe() {}
				unobserve() {}
				disconnect() {}
				takeRecords() {
					return [];
				}
			},
		);
	}
});

const MODEL = rotationModel([castRow('cast:a', 'Alpha', [castItem(0, 1), castItem(4, 5), castItem(25, 26)]), castRow('cast:b', 'Beta', [castItem(2, 3)])]);

/** Delivers frames the way a browser would — a cancelled one never arrives — until nothing asks for another. */
const settle = async () => {
	for (let generation = 0; frames.size; generation++) {
		if (generation === 8) throw new Error('the rotation view never stopped scheduling frames');
		const due = [...frames.values()];
		frames.clear();
		await act(async () => due.forEach(callback => callback(generation)));
	}
};

const activateTab = vi.fn();

const Wrapper = ({ children }: { children: ReactNode }) => {
	const [slot, setSlot] = useState<HTMLDivElement | null>(null);
	return (
		<SimHostProvider host={host}>
			<TabActivationContext.Provider value={activateTab}>
				<div data-testid="dr-sticky-slot" ref={setSlot} />
				<DrStickySlotContext.Provider value={slot}>{children}</DrStickySlotContext.Provider>
			</TabActivationContext.Provider>
		</SimHostProvider>
	);
};

const mount = async (model = MODEL, aplRun?: AplRun) => {
	const view = render(
		<Wrapper>
			<RotationTrack model={model} active aplRun={aplRun} />
		</Wrapper>,
	);
	const scroller = view.container.querySelector<HTMLElement>('[data-testid="rotation-scroller"]')!;
	// happy-dom reports every element as 0x0, and a zero-width scroller is a frame the window
	// declines to measure.
	Object.defineProperty(scroller, 'clientWidth', { value: 800, configurable: true });
	await settle();
	fireEvent.scroll(scroller);
	await settle();
	return { ...view, scroller };
};

const rowKeys = (container: HTMLElement) => [...container.querySelectorAll('[data-testid="rotation-row"]')].map(row => row.getAttribute('data-row-key'));

describe('RotationTrack', () => {
	it('mounts the ordered rows between the two spacers, and only the items inside the track window', async () => {
		const { container } = await mount();
		expect(rowKeys(container)).toEqual(['header:player', 'cast:a', 'cast:b']);
		expect(container.querySelectorAll('[data-testid="rotation-content"] > [data-testid="rotation-vspacer"]')).toHaveLength(2);
		// 800px wide at the default 100 pps, padded 600 each way, is -6s..14s: the 25s cast is out.
		expect(container.querySelector('[data-row-key="cast:a"]')!.querySelectorAll('[data-testid="rotation-item-cast"]')).toHaveLength(2);
	});

	it('mounts only the rows near the scrollport, and keeps the content its full height through a toggle so the page cannot clamp the scroll', async () => {
		const tall = rotationModel(Array.from({ length: 60 }, (_, index) => castRow(`cast:${index}`, `Row ${index}`, [])));
		const { container } = await mount(tall);
		// 61 rows of 32px against a 768px viewport padded by 200: rows 0..30, and 30 × 32px below.
		expect(rowKeys(container)).toHaveLength(31);
		expect(rowKeys(container).at(-1)).toBe('cast:29');
		const spacers = () => [...container.querySelectorAll<HTMLElement>('[data-testid="rotation-content"] > [data-testid="rotation-vspacer"]')];
		expect(spacers()[0].style.getPropertyValue('--vspacer-h')).toBe('0');
		expect(spacers()[1].style.getPropertyValue('--vspacer-h')).toBe('960');

		const height = () => {
			const spacing = spacers().reduce((total, spacer) => total + Number(spacer.style.getPropertyValue('--vspacer-h')), 0);
			return spacing + container.querySelectorAll('[data-testid="rotation-row"]').length * 32;
		};
		expect(height()).toBe(61 * 32);

		// Read in the render the click causes, not after the next frame: emptying the container even
		// for one paint is what let the browser clamp scrollTop and jump the view.
		fireEvent.click(container.querySelector('[data-row-key="cast:20"] [data-testid="rotation-row-hide"]')!);
		expect(container.querySelectorAll('[data-testid="rotation-row"]').length).toBeGreaterThan(0);
		expect(height()).toBe(60 * 32);
	});

	it('names every eye toggle through one shared tooltip rather than one per row', async () => {
		const { container } = await mount();
		const toggles = [...container.querySelectorAll('[data-testid="rotation-row-hide"]')];

		expect(toggles).toHaveLength(2);
		expect(new Set(toggles.map(toggle => toggle.getAttribute('data-tooltip-id'))).size).toBe(1);
		expect(toggles.map(toggle => toggle.getAttribute('data-tooltip-content'))).toEqual([
			'results_tab.details.timeline.hide_row',
			'results_tab.details.timeline.hide_row',
		]);
		expect(toggles.every(toggle => !toggle.hasAttribute('title'))).toBe(true);
	});

	it('takes each row out of the order as its eye toggle is clicked, drops the emptied section, and brings them all back from the floating bar', async () => {
		const { container } = await mount();

		fireEvent.click(container.querySelector('[data-row-key="cast:a"] [data-testid="rotation-row-hide"]')!);
		await settle();
		expect(rowKeys(container)).toEqual(['header:player', 'cast:b']);

		fireEvent.click(container.querySelector('[data-row-key="cast:b"] [data-testid="rotation-row-hide"]')!);
		await settle();
		expect(rowKeys(container)).toEqual([]);

		fireEvent.click(container.querySelector('[data-testid="rotation-floating-action-bar-show-all"]')!);
		await settle();
		expect(rowKeys(container)).toEqual(['header:player', 'cast:a', 'cast:b']);
	});

	it('opens one tooltip for the hovered item, and closes it when the pointer leaves the track', async () => {
		const { container } = await mount();
		const item = container.querySelector('[data-row-key="cast:a"] [data-testid="rotation-item-cast"]')!;
		fireEvent.mouseOver(item);
		await settle();
		expect(container.querySelector('[data-testid="timeline-hover-tooltip"]')!.textContent).toContain('results_tab.details.timeline.cast_tooltip.window');

		fireEvent.mouseOver(container.querySelector('[data-testid="rotation-content"]')!);
		await settle();
		expect(container.querySelector('[data-testid="timeline-hover-tooltip"]')).toBeNull();
	});

	describe('clicking a cast', () => {
		const ENTRY_PATH: AplEntryPath = [{ kind: 'priority', index: 2 }];
		const withEntryPath = (start: number, end: number, aplEntryPath: AplEntryPath | null): RowItem => {
			const item = castItem(start, end) as Extract<RowItem, { kind: 'cast' }>;
			return { ...item, log: { ...item.log, aplEntryPath } } as RowItem;
		};
		const UNPLACED: AplEntryPath = [{ kind: 'priority', index: 7 }];
		const WITH_ENTRY_PATHS = rotationModel([
			castRow('cast:a', 'Alpha', [withEntryPath(0, 1, ENTRY_PATH), withEntryPath(2, 3, null), withEntryPath(4, 5, UNPLACED)]),
		]);
		const run = (changed: boolean): AplRun => ({ rotation: ROTATION, changedSinceRun: () => changed });
		const casts = (container: HTMLElement) => container.querySelectorAll('[data-row-key="cast:a"] [data-testid="rotation-item-cast"]');
		let revealed: Array<AplRevealRequest>;
		let off: () => void;

		beforeEach(() => {
			activateTab.mockClear();
			revealed = [];
			off = aplReveals.on(request => revealed.push(request));
		});

		afterEach(() => off());

		it('opens the Rotation tab on the entry that cast it', async () => {
			const { container } = await mount(WITH_ENTRY_PATHS, run(false));

			fireEvent.click(casts(container)[0]);

			expect(activateTab).toHaveBeenCalledWith('rotation-tab');
			expect(revealed).toHaveLength(1);
			expect(revealed[0].item).toBe(ROTATION.priorityList[2]);
		});

		it('does nothing for a cast no entry launched, a rotation changed since the run, or the end of a pan', async () => {
			const unchanged = await mount(WITH_ENTRY_PATHS, run(false));
			fireEvent.click(casts(unchanged.container)[1]);

			const cast = casts(unchanged.container)[0];
			fireEvent.pointerDown(cast, { button: 0, pointerId: 1, pointerType: 'mouse', clientX: 10, clientY: 10 });
			fireEvent.pointerMove(cast, { pointerId: 1, clientX: 40, clientY: 10 });
			fireEvent.pointerUp(cast, { pointerId: 1, clientX: 40, clientY: 10 });
			fireEvent.click(cast);
			unchanged.unmount();

			const changed = await mount(WITH_ENTRY_PATHS, run(true));
			fireEvent.click(casts(changed.container)[0]);

			expect(activateTab).not.toHaveBeenCalled();
			expect(revealed).toEqual([]);
		});

		it('does nothing for a cast whose entry the run rotation cannot place, as its tooltip shows none', async () => {
			const { container } = await mount(WITH_ENTRY_PATHS, run(false));

			fireEvent.click(casts(container)[2]);

			expect(activateTab).not.toHaveBeenCalled();
			expect(revealed).toEqual([]);
		});
	});

	it('gives the portalled ruler row the pane timeline variables, and keeps its zoom in step', async () => {
		const { container } = await mount();
		const pane = container.querySelector<HTMLElement>('[data-testid="rotation-pane"]')!;
		const rulerRow = container.querySelector<HTMLElement>('[data-testid="rotation-ruler-track"]')!.parentElement!.parentElement!;
		expect(pane.contains(rulerRow)).toBe(false);
		expect(rulerRow.style.getPropertyValue('--duration')).toBe(pane.style.getPropertyValue('--duration'));
		expect(rulerRow.style.getPropertyValue('--label-w')).toBe(pane.style.getPropertyValue('--label-w'));

		fireEvent.click(container.querySelectorAll('[data-testid="rotation-zoom-button"]')[1]);
		await settle();

		expect(pane.style.getPropertyValue('--pps')).not.toBe('100px');
		expect(rulerRow.style.getPropertyValue('--pps')).toBe(pane.style.getPropertyValue('--pps'));
	});

	it('renders nothing but its own wrapper without a model', async () => {
		const { container } = await mount(null as never);
		expect(rowKeys(container)).toEqual([]);
		expect(container.querySelector('[data-testid="rotation-ruler-track"]')).not.toBeNull();
	});
});
