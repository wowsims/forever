import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { FloatingActionBar } from './FloatingActionBar';

const observed: Array<{ callback: IntersectionObserverCallback; options?: IntersectionObserverInit }> = [];

beforeEach(() => {
	observed.length = 0;
	vi.stubGlobal(
		'IntersectionObserver',
		class {
			constructor(
				public callback: IntersectionObserverCallback,
				public options?: IntersectionObserverInit,
			) {
				observed.push({ callback, options });
			}
			observe() {}
			disconnect() {}
			unobserve() {}
			takeRecords() {
				return [];
			}
		},
	);
});

const mount = () =>
	render(
		<FloatingActionBar
			testIdPrefix="bar"
			rootClassName=""
			sheetClassName=""
			toggleLabel="Filters"
			icon={null}
			summary="No filters"
			preview=""
			clear={{ testId: 'bar-clear', icon: null, label: 'Clear', hidden: true, onClick: () => {} }}>
			<div>The drawer's contents.</div>
		</FloatingActionBar>,
	);

const root = () => screen.getByTestId('bar-floating-action-bar-root');
const toggle = () => screen.getByTestId('bar-floating-action-bar-toggle');
const panel = () => screen.queryByTestId('bar-floating-action-bar-panel-inner');

describe('FloatingActionBar', () => {
	it('starts collapsed, with the drawer out of the tab order', () => {
		mount();

		expect(toggle().getAttribute('aria-expanded')).toBe('false');
		expect(panel()).toBeNull();
	});

	it('opens and closes on the toggle, and lets the drawer back into the tab order', () => {
		mount();

		fireEvent.click(toggle());
		expect(toggle().getAttribute('aria-expanded')).toBe('true');
		expect(panel()).not.toBeNull();

		fireEvent.click(toggle());
		expect(toggle().getAttribute('aria-expanded')).toBe('false');
	});

	// Built inside a hidden tab, the ratio goes 0 -> pinned without ever passing through 1, so a
	// threshold of [1] alone never fires again.
	it('watches for the pin with both thresholds', () => {
		mount();

		expect(observed).toHaveLength(1);
		expect(observed[0].options).toEqual({ rootMargin: '0px 0px -1px 0px', threshold: [0, 1] });
	});

	describe('the pinned state', () => {
		const deliver = (ratios: Array<number>, clientHeight = 48) => {
			const target = root();
			Object.defineProperty(target, 'clientHeight', { value: clientHeight, configurable: true });
			const records = ratios.map(intersectionRatio => ({ target, intersectionRatio })) as unknown as Array<IntersectionObserverEntry>;
			act(() => observed[0].callback(records, null as never));
		};

		it('pins the bar once it is clipped by the viewport edge', () => {
			mount();

			deliver([0.98]);

			expect(root().hasAttribute('data-stuck')).toBe(true);
		});

		it('lets go once the bar is fully in view again', () => {
			mount();
			deliver([0.98]);

			deliver([1]);

			expect(root().hasAttribute('data-stuck')).toBe(false);
		});

		// A pane the tab switch closed has no box at all, and a ratio of 0 would otherwise read as
		// "clipped" and pin a bar nobody can see.
		it('stays unpinned while the pane it lives in is closed', () => {
			mount();
			deliver([0.98]);

			deliver([0], 0);

			expect(root().hasAttribute('data-stuck')).toBe(false);
		});

		// One delivery can carry several records, oldest first. The list growing under a bar that was
		// already pinned produces exactly this pair, and reading the first leaves it unpinned for good
		// because nothing moves again to produce another record.
		it('reads the newest record in a delivery, not the oldest', () => {
			mount();

			deliver([1, 0.98]);
			expect(root().hasAttribute('data-stuck')).toBe(true);

			deliver([0.98, 1]);
			expect(root().hasAttribute('data-stuck')).toBe(false);
		});
	});
});
