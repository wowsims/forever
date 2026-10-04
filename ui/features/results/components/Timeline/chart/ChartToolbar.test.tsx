import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ChartToolbar } from './ChartToolbar';

const KEYS = ['reset', 'zoom_out', 'zoom_in', 'pan_left', 'pan_right'];

describe('ChartToolbar', () => {
	it('names each control, and labels them all through one tooltip', () => {
		render(<ChartToolbar onReset={() => {}} onZoomIn={() => {}} onZoomOut={() => {}} onPanLeft={() => {}} onPanRight={() => {}} />);
		const buttons = screen.getAllByRole('button');

		expect(buttons.map(button => button.getAttribute('aria-label'))).toEqual(KEYS.map(key => `results_tab.details.timeline.chart_options.${key}`));
		expect(buttons.map(button => button.getAttribute('data-tooltip-content'))).toEqual(buttons.map(button => button.getAttribute('aria-label')));
		expect(new Set(buttons.map(button => button.getAttribute('data-tooltip-id'))).size).toBe(1);
		expect(buttons.every(button => !button.hasAttribute('title'))).toBe(true);
	});

	it('runs the handler each control stands for', () => {
		const handlers = { onReset: vi.fn(), onZoomOut: vi.fn(), onZoomIn: vi.fn(), onPanLeft: vi.fn(), onPanRight: vi.fn() };
		render(<ChartToolbar {...handlers} />);

		screen.getAllByRole('button').forEach(button => fireEvent.click(button));

		expect(Object.values(handlers).map(handler => handler.mock.calls.length)).toEqual([1, 1, 1, 1, 1]);
	});
});
