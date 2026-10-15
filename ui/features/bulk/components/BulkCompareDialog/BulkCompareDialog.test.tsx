// The dialog over a real store holding three finished encounters; chart.js is stubbed. With no locale loaded every string
// is its key, so sets are told apart by their DPS and their changed helm.
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeBulkCompareHost } from '@testing/storybook/fixtures/bulk-compare';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import type { ActiveElement, ChartOptions } from 'chart.js';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { BulkCompareDialog } from './BulkCompareDialog';

vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);

type FakeChart = {
	config: { data: { datasets: Array<{ data: number[]; backgroundColor: string[]; hoverBackgroundColor: string[] }> }; options: ChartOptions<'bar'> };
};
const charts = fakeCharts as FakeChart[];

const rootElem = document.createElement('div');

const mount = () => {
	document.body.appendChild(rootElem);
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={makeBulkCompareHost()}>
				<BulkCompareDialog open onOpenChange={() => {}} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
};

const dialog = () => within(screen.getByRole('dialog'));
const drawn = () => charts.map(chart => chart.config.data.datasets[0].data);
// Which bars are dimmed, by the alpha a dimmed bar's colour ends in.
const dimmed = () => charts.map(chart => chart.config.data.datasets[0].backgroundColor.map(color => color.endsWith('4d')));
const bodyRows = () => within(dialog().getByRole('table')).getAllByRole('row').slice(1);
const hoverBar = async (chart: number, bar: number) => {
	const onHover = charts[chart].config.options.onHover as unknown as (event: unknown, elements: ActiveElement[]) => void;
	await act(async () => onHover(undefined, bar < 0 ? [] : [{ index: bar } as ActiveElement]));
};

beforeEach(() => {
	charts.length = 0;
});

describe('BulkCompareDialog', () => {
	it('draws one chart per encounter, shortest first and health-based last, each with its own ranking and the current gear', () => {
		mount();

		expect(drawn()).toEqual([
			[1030, 1020, 1000, 990],
			[1100, 1090, 1080, 1000],
			[12500, 1150, 1120, 1000],
		]);
		expect(dialog().getAllByRole('img', { name: 'bulk_tab.compare.chart_label' })).toHaveLength(3);
		expect(
			dialog()
				.getAllByRole('heading', { level: 3 })
				.map(heading => heading.firstChild?.textContent),
		).toEqual(['Council', 'Patchwerk', 'Ragnaros']);
	});

	it('picks a set out in every chart and in the table while its bar is hovered, dimming the other sets but not the current gear', async () => {
		mount();

		await hoverBar(1, 0);

		expect(dimmed()).toEqual([
			[true, false, false, true],
			[false, true, true, false],
			[true, true, false, false],
		]);
		expect(charts.map(chart => chart.config.data.datasets[0].hoverBackgroundColor)).toEqual(
			charts.map(chart => chart.config.data.datasets[0].backgroundColor),
		);
		expect(bodyRows().map(row => row.hasAttribute('data-active'))).toEqual([false, false, true, false, false, false, false]);

		await hoverBar(1, -1);

		expect(
			dimmed()
				.flat()
				.every(isDimmed => !isDimmed),
		).toBe(true);
		expect(bodyRows().some(row => row.hasAttribute('data-active'))).toBe(false);
	});

	it('picks a set out in every chart while its table row is hovered', async () => {
		mount();

		await act(async () => fireEvent.mouseEnter(bodyRows()[0]));

		expect(dimmed()).toEqual([
			[true, true, false, true],
			[true, true, true, false],
			[true, true, true, false],
		]);

		await act(async () => fireEvent.mouseLeave(bodyRows()[0]));

		expect(dimmed().flat().some(Boolean)).toBe(false);
	});

	it('tables the current gear, then every set by the number it first got, with a cell only where an encounter ranked it', () => {
		mount();

		const rows = bodyRows();
		const cells = (row: HTMLElement) =>
			within(row)
				.getAllByRole('cell')
				.map(cell => cell.textContent);

		expect(
			rows.map(
				row =>
					within(row)
						.getByRole('rowheader')
						.textContent?.match(/Helm \d+/)?.[0] ?? '',
			),
		).toEqual(['', 'Helm 3', 'Helm 2', 'Helm 6', 'Helm 4', 'Helm 5', 'Helm 7']);
		expect(cells(rows[0]).every(cell => /^1,000\.00 ±\d+$/.test(cell!))).toBe(true);
		expect(cells(rows[3])).toEqual([expect.stringMatching(/^990\.00 ±\d+-10\.00 \(1\.01%\)$/), '', '']);
		expect(cells(rows[2]).map(cell => cell!.split(' ')[0])).toEqual(['1,020.00', '1,100.00', '1,120.00']);
	});
});
