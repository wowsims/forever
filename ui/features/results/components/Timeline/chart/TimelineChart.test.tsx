import { fakeCharts } from '@testing/chart-js-mock';
import { act, fireEvent, render } from '@testing-library/react';
import { StrictMode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { TimelineChartSpec, TimelineDataset } from '../../../model/timeline/chart/types';
import { TimelineChart } from './TimelineChart';

interface FakeChart {
	config: { data: { datasets: Array<TimelineDataset> }; options: { scales: unknown }; plugins: Array<unknown> };
	data: { datasets: Array<TimelineDataset> };
	canvas: HTMLCanvasElement;
	updates: number;
	destroyed: boolean;
	update: () => void;
	destroy: () => void;
}

const charts = fakeCharts as Array<FakeChart>;

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

const dataset = (seriesId: string, label: string): TimelineDataset =>
	({
		seriesId,
		label,
		tooltipSpec: { kind: 'dps' },
		data: [{ x: 1, y: 2, log: { timestamp: 1, dps: 2, damageLogs: [], activeAuras: [] } }],
	}) as unknown as TimelineDataset;

const spec = (...datasets: Array<TimelineDataset>): TimelineChartSpec => ({ datasets, scales: { x: {} }, annotations: null, duration: 30 });

const withScheduled = (...kinds: Array<'downtime' | 'wave' | 'amp' | 'move'>): TimelineChartSpec => ({
	...spec(dataset('dps', 'DPS')),
	annotations: { bands: [], icons: [], scheduled: kinds.map(kind => ({ kind, start: 5, end: 10 })) },
});

const legend = (container: HTMLElement) => container.querySelector('[data-testid="timeline-scheduled-legend"]');

// The strips a plugin pass fills, from a canvas context that records them.
const drawnStrips = (chart: FakeChart) => {
	const fills: Array<unknown> = [];
	const ctx = new Proxy({ fillRect: (...args: Array<unknown>) => fills.push(args) } as Record<string, unknown>, {
		get: (target, key: string) => target[key] ?? (() => {}),
		set: () => true,
	});
	const fake = { ctx, chartArea: { left: 0, right: 100, top: 0, bottom: 100 }, scales: { x: { getPixelForValue: (value: number) => value } } };
	for (const plugin of chart.config.plugins as Array<{ afterDatasetsDraw?: (chart: unknown) => void }>) plugin.afterDatasetsDraw?.(fake);
	return fills.length;
};

const options = (chart: FakeChart) =>
	chart.config.options as unknown as {
		plugins: { legend: { onClick: (event: unknown, item: { datasetIndex: number }) => void }; tooltip: { external: (context: unknown) => void } };
	};

beforeEach(() => {
	charts.length = 0;
});

describe('TimelineChart', () => {
	it.each([
		['with no spec at all', null],
		['for a spec that carries no series', spec()],
	])('shows the waiting state %s, rather than a blank canvas, and builds no chart', (_name, chartSpec) => {
		const { container } = render(<TimelineChart spec={chartSpec} />);
		expect(container.querySelector('[data-testid="timeline-chart-empty"]')).toBeTruthy();
		expect(container.querySelector('[data-testid="timeline-chart-canvas"]')).toBeNull();
		expect(charts).toHaveLength(0);
	});

	it('goes back to the waiting state when a result with data is replaced by no spec', () => {
		const { container, rerender } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		expect(container.querySelector('[data-testid="timeline-chart-canvas"]')).toBeTruthy();

		rerender(<TimelineChart spec={null} />);

		expect(container.querySelector('[data-testid="timeline-chart-empty"]')).toBeTruthy();
		expect(container.querySelector('[data-testid="timeline-chart-canvas"]')).toBeNull();
		expect(charts[0].destroyed).toBe(true);
	});

	it('builds one chart from the spec and swaps the empty state out', () => {
		const { container } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		expect(charts).toHaveLength(1);
		expect(charts[0].config.data.datasets.map(entry => entry.label)).toEqual(['DPS']);
		expect(container.querySelector('[data-testid="timeline-chart-empty"]')).toBeNull();
		expect(container.querySelector('[data-testid="timeline-chart-canvas"]')).toBeTruthy();
	});

	it('starts with threat off and every other series on', () => {
		render(<TimelineChart spec={spec(dataset('dps', 'DPS'), dataset('threat', 'Threat'))} />);
		expect(charts[0].config.data.datasets.map(entry => entry.hidden)).toEqual([false, true]);
	});

	it('remembers a legend toggle across a rebuild, keyed by series rather than by index', () => {
		const { rerender } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'), dataset('threat', 'Threat'))} />);
		act(() => options(charts[0]).plugins.legend.onClick(null, { datasetIndex: 1 }));
		expect(charts[0].data.datasets[1].hidden).toBe(false);
		expect(charts[0].updates).toBe(1);

		// A result that reorders the series must still bring threat back visible.
		rerender(<TimelineChart spec={spec(dataset('threat', 'Threat'), dataset('dps', 'DPS'))} />);
		expect(charts[0].config.data.datasets.map(entry => entry.hidden)).toEqual([false, false]);
	});

	it('updates the one chart when the result changes, and destroys it on unmount', () => {
		const { rerender, unmount } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		const next = { ...spec(dataset('threat', 'Threat')), scales: { x: { max: 90 } } };
		rerender(<TimelineChart spec={next} />);
		expect(charts).toHaveLength(1);
		expect(charts[0].destroyed).toBe(false);
		expect(charts[0].updates).toBe(1);
		expect(charts[0].config.data.datasets.map(entry => entry.label)).toEqual(['Threat']);
		expect(charts[0].config.options.scales).toEqual({ x: { max: 90 } });

		unmount();
		expect(charts[0].destroyed).toBe(true);
	});

	// StrictMode replays the effect over the same options object the first chart was built from —
	// the same path a second result takes, and chart.js keeps a reference to what it is handed.
	it('leaves exactly one live chart under StrictMode', () => {
		render(
			<StrictMode>
				<TimelineChart spec={spec(dataset('dps', 'DPS'))} />
			</StrictMode>,
		);
		expect(charts).toHaveLength(2);
		expect(charts.filter(chart => !chart.destroyed)).toHaveLength(1);
	});

	it('renders the hovered point’s tooltip, and takes it away when the pointer leaves', () => {
		const { container } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		const chart = charts[0];
		chart.canvas.getBoundingClientRect = () => ({ left: 0, top: 0 }) as DOMRect;
		const hover = {
			chart,
			tooltip: { opacity: 1, caretX: 10, caretY: 20, dataPoints: [{ datasetIndex: 0, dataIndex: 0, raw: chart.data.datasets[0].data[0] }] },
		};

		act(() => options(chart).plugins.tooltip.external(hover));
		expect(container.querySelector('[data-testid="timeline-hover-tooltip"]')!.textContent).toContain('1.00s');

		act(() => options(chart).plugins.tooltip.external({ ...hover, tooltip: { ...hover.tooltip, opacity: 0 } }));
		expect(container.querySelector('[data-testid="timeline-hover-tooltip"]')).toBeNull();
	});

	it('drives the zoom from its toolbar', () => {
		const { container } = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		const buttons = container.querySelectorAll<HTMLButtonElement>('[data-testid="timeline-chart-toolbar"] button');
		expect(buttons).toHaveLength(5);

		fireEvent.click(buttons[2]);
		expect(charts[0].updates).toBe(1);
		fireEvent.click(buttons[0]);
		expect(charts[0].updates).toBe(2);
	});

	it('lists only the scheduled kinds present, and nothing without a design', () => {
		const withoutDesign = render(<TimelineChart spec={spec(dataset('dps', 'DPS'))} />);
		expect(legend(withoutDesign.container)).toBeNull();
		withoutDesign.unmount();

		const { container } = render(<TimelineChart spec={withScheduled('wave', 'move')} />);
		const entries = Array.from(legend(container)!.querySelectorAll('button')).map(button => button.textContent);
		expect(entries).toEqual(['results_tab.details.timeline.scheduled.wave', 'results_tab.details.timeline.scheduled.move']);
	});

	it('hides and shows a scheduled kind from its legend entry', () => {
		const { container } = render(<TimelineChart spec={withScheduled('downtime', 'wave')} />);
		const [downtime] = Array.from(legend(container)!.querySelectorAll('button'));
		expect(drawnStrips(charts[0])).toBe(2);

		fireEvent.click(downtime);
		expect(downtime.getAttribute('aria-pressed')).toBe('false');
		expect(charts[0].updates).toBe(1);
		expect(drawnStrips(charts[0])).toBe(1);

		fireEvent.click(downtime);
		expect(downtime.getAttribute('aria-pressed')).toBe('true');
		expect(drawnStrips(charts[0])).toBe(2);
	});
});
