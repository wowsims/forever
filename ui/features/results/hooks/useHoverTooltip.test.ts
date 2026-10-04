import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useHoverTooltip } from './useHoverTooltip';

const host = (width: number, height: number) => {
	const element = document.createElement('div');
	element.getBoundingClientRect = () => ({ width, height }) as DOMRect;
	return element;
};

const mount = (width = 100, height = 50) => {
	const view = renderHook(() => useHoverTooltip<string>());
	// The hook only positions the element it is given; nothing renders one for it.
	view.result.current.ref.current = host(width, height);
	return view;
};

describe('tooltip placement', () => {
	it('sits below and to the right of the anchor', () => {
		const { result } = mount(100, 50);
		act(() => result.current.moveTo(200, 300));
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-x')).toBe('212');
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-y')).toBe('312');
	});

	it('flips to the left of the anchor rather than overflowing the right edge', () => {
		const { result } = mount(300, 50);
		act(() => result.current.moveTo(window.innerWidth - 20, 100));
		expect(Number(result.current.ref.current!.style.getPropertyValue('--tooltip-x'))).toBe(window.innerWidth - 20 - 12 - 300);
	});

	it('lifts a tooltip taller than the space below the anchor back inside the viewport', () => {
		const { result } = mount(100, 400);
		act(() => result.current.moveTo(10, window.innerHeight - 50));
		expect(Number(result.current.ref.current!.style.getPropertyValue('--tooltip-y'))).toBe(window.innerHeight - 8 - 400);
	});

	it('does nothing without a host', () => {
		const { result } = renderHook(() => useHoverTooltip<string>());
		expect(() => act(() => result.current.show('Fireball', 0, 0))).not.toThrow();
		expect(() => act(() => result.current.moveTo(0, 0))).not.toThrow();
	});
});

describe('useHoverTooltip', () => {
	it('opens on the content it is shown, at the point it is given', () => {
		const { result } = mount();
		expect(result.current.content).toBeNull();

		act(() => result.current.show('Fireball', 200, 300));
		expect(result.current.content).toBe('Fireball');
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-x')).toBe('212');
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-y')).toBe('312');
	});

	it('follows the cursor horizontally, keeping the anchor it was opened at', () => {
		const { result } = mount();
		act(() => result.current.show('Fireball', 200, 300));
		act(() => result.current.moveTo(260));
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-x')).toBe('272');
		expect(result.current.ref.current!.style.getPropertyValue('--tooltip-y')).toBe('312');
		expect(result.current.content).toBe('Fireball');
	});

	it('closes on hide', () => {
		const { result } = mount();
		act(() => result.current.show('Fireball', 10, 10));
		act(() => result.current.hide());
		expect(result.current.content).toBeNull();
	});

	it('keeps its callbacks stable, so a chart.js options object built once can hold them', () => {
		const { result } = mount();
		const { show, moveTo, hide } = result.current;
		act(() => result.current.show('Fireball', 10, 10));
		expect(result.current.show).toBe(show);
		expect(result.current.moveTo).toBe(moveTo);
		expect(result.current.hide).toBe(hide);
	});
});
