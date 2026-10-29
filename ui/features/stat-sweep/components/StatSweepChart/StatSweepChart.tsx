import { formatAverage, formatStdev } from '@features/results/components/ResultMetricList/utils';
import { AXIS_GRID_COLOR, cssVarColor } from '@features/results/components/Timeline/chart/colors';
import type { ResultMetric } from '@features/results/model/topline_metrics';
import i18n from '@i18n/config';
import type { StatSweepPoint } from '@sim/cache/stat_sweep_history';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { distinct } from '@sim/utils/collections';
import { formatToNumber } from '@sim/utils/format';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import type { ChartData, ChartDataset, ChartOptions, Plugin } from 'chart.js';
import { Suspense, useMemo } from 'react';

import { Line } from '../../../../shared/lazy_charts';

export interface StatSweepSeries {
	id: string;
	label: string;
	colorSlot: number;
	statLabel: string;
	/** The swept stat's total when the run started. */
	baseStatValue: number;
	points: ReadonlyArray<StatSweepPoint>;
	/** The step past which more of the stat left every metric unchanged. */
	capped?: Pick<StatSweepPoint, 'offset' | 'bonus'>;
}

export interface StatSweepChartProps {
	series: ReadonlyArray<StatSweepSeries>;
	metric: StatSweepMetric;
	/** The metrics the toggle offers, in order. */
	metrics: ReadonlyArray<StatSweepMetric>;
	onMetricChange: (metric: StatSweepMetric) => void;
	/** Shades the standard deviation around each line, and fits the y scale to it. */
	showSpread: boolean;
	onShowSpreadChange: (showSpread: boolean) => void;
}

type SweepDatum = { x: number; y: number; stdev: number; point: StatSweepPoint; statLabel: string; capped: boolean };

// `id` is `${run.id}:${role}`, by which the chart keeps each dataset's state across updates.
type SweepDataset = ChartDataset<'line', SweepDatum[]> & { id: string; role: 'upper' | 'lower' | 'line'; runId: string };

// Each run draws the band's upper edge, its lower edge, then the metric's line. The upper edge fills down to the next dataset, so the band depends on this order.
// The band is one standard deviation of a single fight's value either side of the average, the spread the sim card's ± reports.
// Without the band a run draws its line alone.
const isLine = (dataset: object) => (dataset as SweepDataset).role === 'line';
// The x range reaches this share of its span past the outermost value, so the end points draw whole inside the plot.
const X_RANGE_PAD = 0.03;
const BAND_ALPHA = '2e';
// The band's dashed edges keep each run's band traceable where two bands overlap.
const BAND_EDGE_ALPHA = '99';
const LEGEND_LINE_WIDTH = 3;
// The y scale fits its data, with this share of the span as room past the extremes.
const Y_GRACE = '5%';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

const metricLabel = (metric: StatSweepMetric) => i18n.t(`sidebar.results.metrics.${metric}.label`);

const seriesColor = (slot: number) => cssVarColor(`--color-chart-series-${slot + 1}`);

// The sim card's own formatting of a metric and its ±; TMI is the one the card gives in percent.
const cardMetric = (metric: StatSweepMetric, average: number, stdev = 0): ResultMetric => ({
	metric,
	name: '',
	average,
	stdev,
	unit: metric === 'tmi' ? 'percentage' : undefined,
});
const formatValue = (metric: StatSweepMetric, average: number) => formatAverage(cardMetric(metric, average), 'list');
const formatError = (metric: StatSweepMetric, stdev: number) => formatStdev(cardMetric(metric, 0, stdev), 'list');

function datasets(series: ReadonlyArray<StatSweepSeries>, metric: StatSweepMetric, showSpread: boolean): SweepDataset[] {
	const surface = cssVarColor('--color-background');
	return series.flatMap((run): SweepDataset[] => {
		const color = seriesColor(run.colorSlot);
		const at = (sign: number) =>
			run.points.map(point => {
				const x = point.offset;
				const { avg, stdev } = point.metrics[metric];
				return { x, y: avg + sign * stdev, stdev, point, statLabel: run.statLabel, capped: run.capped !== undefined && x >= run.capped.offset };
			});
		const edge = {
			pointRadius: 0,
			pointHoverRadius: 0,
			borderWidth: 1,
			borderDash: [4, 3],
			borderColor: color + BAND_EDGE_ALPHA,
			label: run.label,
			runId: run.id,
		};
		const band: SweepDataset[] = showSpread
			? [
					{ ...edge, id: `${run.id}:upper`, role: 'upper', data: at(1), backgroundColor: color + BAND_ALPHA, fill: '+1' },
					{ ...edge, id: `${run.id}:lower`, role: 'lower', data: at(-1), fill: false },
				]
			: [];
		return [
			...band,
			{
				id: `${run.id}:line`,
				role: 'line',
				runId: run.id,
				label: run.label,
				data: at(0),
				borderColor: color,
				borderWidth: 2,
				borderJoinStyle: 'round',
				borderCapStyle: 'round',
				fill: false,
				backgroundColor: color,
				pointRadius: 4,
				pointHoverRadius: 6,
				pointBorderColor: surface,
				pointBorderWidth: 2,
			},
		];
	});
}

// The current value sits at 0, so the range always reaches it.
function xRange(series: ReadonlyArray<StatSweepSeries>) {
	const offsets = series.flatMap(run => run.points.map(point => point.offset));
	const min = Math.min(0, ...offsets);
	const max = Math.max(0, ...offsets);
	const pad = (max - min) * X_RANGE_PAD || 1;
	return { min: min - pad, max: max + pad };
}

const shown = (value: number) => formatToNumber(value, { useGrouping: false });

const signed = (added: number) => formatToNumber(added, { signDisplay: 'exceptZero', useGrouping: false });

/** The run's label, with the stat added at which more of the stat stopped changing any metric, in the stat's own units as the tooltip gives it. */
export const runLabel = ({ label, capped }: Pick<StatSweepSeries, 'label' | 'capped'>) =>
	capped === undefined ? label : t('capped_label', { label, added: signed(capped.bonus) });

const markerPlugin: Plugin<'line'> = {
	id: 'statSweepMarker',
	afterDatasetsDraw: chart => {
		const { ctx, chartArea, scales } = chart;
		const line = (x: number, dash: number[]) => {
			if (x < chartArea.left || x > chartArea.right) return;
			ctx.save();
			ctx.strokeStyle = cssVarColor('--color-muted');
			ctx.lineWidth = 1;
			ctx.setLineDash(dash);
			ctx.beginPath();
			ctx.moveTo(x, chartArea.top);
			ctx.lineTo(x, chartArea.bottom);
			ctx.stroke();
			ctx.restore();
		};
		line(scales.x.getPixelForValue(0), [4, 4]);
		const hovered = chart.tooltip?.getActiveElements()[0];
		if (hovered) line(hovered.element.x, []);
	},
};

const PLUGINS = [markerPlugin];

function options(series: ReadonlyArray<StatSweepSeries>, metric: StatSweepMetric): ChartOptions<'line'> {
	const range = xRange(series);
	const runs = new Map(series.map(run => [run.id, run]));
	return {
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		// Runs of different ranges or steps hold different changes at one index, so the tooltip lists every run at the hovered change, not at the hovered index.
		interaction: { mode: 'nearest', axis: 'x', intersect: false },
		plugins: {
			legend: {
				display: series.length >= 2,
				position: 'bottom',
				align: 'start',
				labels: {
					usePointStyle: true,
					pointStyle: 'line',
					// The default items take their stroke from the point border, which is the surface colour, so the swatch would vanish.
					generateLabels: chart =>
						chart.data.datasets.flatMap((dataset, datasetIndex) => {
							if (!isLine(dataset)) return [];
							const color = (dataset as SweepDataset).borderColor as string;
							return [
								{
									text: runLabel(runs.get((dataset as SweepDataset).runId)!),
									datasetIndex,
									fillStyle: color,
									strokeStyle: color,
									lineWidth: LEGEND_LINE_WIDTH,
									pointStyle: 'line',
									fontColor: cssVarColor('--color-foreground'),
									hidden: !chart.isDatasetVisible(datasetIndex),
								},
							];
						}),
				},
			},
			tooltip: {
				filter: item => isLine(item.dataset),
				callbacks: {
					// One line per distinct stat total, so runs of one stat from one base value share a line.
					title: items =>
						distinct(
							items.map(item => {
								const { point, statLabel } = item.raw as SweepDatum;
								return t('tooltip_total', { stat: statLabel, total: shown(point.statValue), added: signed(point.bonus) });
							}),
						),
					label: item => {
						const { y, stdev, capped } = item.raw as SweepDatum;
						const value = { value: formatValue(metric, y), stdev: formatError(metric, stdev) };
						return `${item.dataset.label}: ${t(capped ? 'tooltip_value_capped' : 'tooltip_value', value)}`;
					},
				},
			},
		},
		scales: {
			x: {
				type: 'linear',
				min: range.min,
				max: range.max,
				title: { display: true, text: t('added_axis') },
				ticks: { includeBounds: false },
				grid: { color: AXIS_GRID_COLOR },
			},
			y: {
				type: 'linear',
				beginAtZero: false,
				grace: Y_GRACE,
				title: { display: true, text: metricLabel(metric), color: cssVarColor('--color-foreground') },
				grid: { color: AXIS_GRID_COLOR },
			},
		},
	};
}

export const StatSweepChart = ({ series, metric, metrics, onMetricChange, showSpread, onShowSpreadChange }: StatSweepChartProps) => {
	const data = useMemo(() => ({ datasets: datasets(series, metric, showSpread) }) satisfies ChartData<'line', SweepDatum[]>, [series, metric, showSpread]);
	const chartOptions = useMemo(() => options(series, metric), [series, metric]);
	const metricOptions = useMemo(() => metrics.map(value => ({ value, label: metricLabel(value) })), [metrics]);
	const hasPoints = series.some(run => run.points.length);
	const label = metricLabel(metric);

	// The empty text sits in the chart's box, so the first point does not resize the dialog.
	return (
		<figure className="m-0 flex min-h-0 min-w-0 flex-1 flex-col gap-2">
			<div className="flex flex-wrap items-center gap-x-4 gap-y-2">
				<ToggleGroup<StatSweepMetric>
					size="sm"
					aria-label={t('metric')}
					options={metricOptions}
					value={metric}
					onChange={onMetricChange}
					testId="stat-sweep-metric"
				/>
				<BooleanPicker
					id="stat-sweep-show-spread"
					label={t('show_spread')}
					labelTooltip={t('show_spread_tooltip')}
					value={showSpread}
					onChange={onShowSpreadChange}
					testId="stat-sweep-show-spread"
				/>
			</div>
			<div className="relative flex min-h-112 w-full min-w-0 flex-1 items-center justify-center lg:min-h-64">
				{hasPoints ? (
					<Suspense>
						<Line datasetIdKey="id" data={data} options={chartOptions} plugins={PLUGINS} />
					</Suspense>
				) : (
					<p className="mb-0 text-muted">{t('empty', { metric: label })}</p>
				)}
			</div>
			<figcaption className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted">
				<span className="flex items-center gap-2">
					<span className="h-4 border-l border-dashed border-muted" />
					{t('current_value')}
				</span>
				{showSpread && (
					<span className="flex items-center gap-2">
						<span className="h-4 w-4 rounded-sm bg-muted/30" />
						{t('band_key', { metric: label })}
					</span>
				)}
			</figcaption>
		</figure>
	);
};
