import { cssVarColor } from '@features/results/components/Timeline/chart/colors';
import i18n from '@i18n/config';
import { Chart, type ChartData, type ChartDataset, type ChartOptions, type Plugin } from 'chart.js';
import clsx from 'clsx';
import { Bar } from 'react-chartjs-2';

import { dpsAxisRange } from '../../model/compare';
import { formatBulkDps } from '../../model/format';

export interface BulkCompareBar {
	/** The set's gear identity key, the same in every group the set shows in. */
	key: string;
	label: string;
	isCurrent: boolean;
	avg: number;
	/** The line shown under the chart while the set is picked out. */
	details: string;
}

export interface BulkCompareGroup {
	label: string;
	summary: string;
	/** The encounter's own ranking, best first. */
	bars: readonly BulkCompareBar[];
}

export interface BulkCompareChartProps {
	groups: readonly BulkCompareGroup[];
	/** The key of the set picked out in every group; the others dim. */
	highlighted: string | null;
	onHighlight: (key: string | null) => void;
}

// `id` is how the chart keeps the dataset across updates.
type DpsDataset = ChartDataset<'bar', number[]> & { id: string };

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`bulk_tab.compare.${key}`, options);

const RANKED_COLOR = '--color-chart-series-1';
const CURRENT_COLOR = '--color-grey';
const DIM_ALPHA = '4d';
const VALUE_GAP_PX = 4;
// By bar count, so every bar keeps about the same thickness.
const PLOT_HEIGHT_CLASSES = ['h-12', 'h-12', 'h-16', 'h-24', 'h-28', 'h-36', 'h-40'];

const valueFont = () => `${Chart.defaults.font.size}px ${Chart.defaults.font.family}`;

type ValueChart = { ctx: CanvasRenderingContext2D; data: { datasets: ReadonlyArray<{ data: unknown }> } };

const shownValues = (chart: ValueChart) => (chart.data.datasets[0]?.data ?? []) as number[];

// Room right of the plot for the widest printed DPS, so a long value is never clipped.
const valueRoom = (chart: ValueChart): number => {
	const { ctx } = chart;
	ctx.save();
	ctx.font = valueFont();
	const widest = Math.max(0, ...shownValues(chart).map(value => ctx.measureText(formatBulkDps(value)).width));
	ctx.restore();
	return Math.ceil(widest) + 2 * VALUE_GAP_PX;
};

const barValuePlugin: Plugin<'bar'> = {
	id: 'bulkCompareBarValue',
	afterDatasetsDraw: chart => {
		const { ctx } = chart;
		const values = shownValues(chart);
		ctx.save();
		ctx.font = valueFont();
		ctx.fillStyle = cssVarColor('--color-foreground');
		ctx.textBaseline = 'middle';
		chart.getDatasetMeta(0).data.forEach((element, index) => ctx.fillText(formatBulkDps(values[index]), element.x + VALUE_GAP_PX, element.y));
		ctx.restore();
	},
};

const PLUGINS = [barValuePlugin];

const barColor = (bar: BulkCompareBar, highlighted: string | null) => {
	const color = cssVarColor(bar.isCurrent ? CURRENT_COLOR : RANKED_COLOR);
	// The current gear is the baseline, so it never dims.
	return bar.isCurrent || highlighted === null || highlighted === bar.key ? color : color + DIM_ALPHA;
};

interface GroupChartProps {
	group: BulkCompareGroup;
	highlighted: string | null;
	onHighlight: (key: string | null) => void;
}

const BulkCompareGroupChart = ({ group, highlighted, onHighlight }: GroupChartProps) => {
	const range = dpsAxisRange(group.bars.map(bar => bar.avg));
	const colors = group.bars.map(bar => barColor(bar, highlighted));
	const dataset: DpsDataset = {
		id: 'dps',
		data: group.bars.map(bar => bar.avg),
		backgroundColor: colors,
		// The same as at rest, so the hovered bar looks like the same set in every other group.
		hoverBackgroundColor: colors,
		categoryPercentage: 1,
		barPercentage: 0.8,
	};
	const data: ChartData<'bar', number[], string> = { labels: group.bars.map(bar => bar.label), datasets: [dataset] };
	const options: ChartOptions<'bar'> = {
		indexAxis: 'y',
		animation: false,
		responsive: true,
		maintainAspectRatio: false,
		layout: { padding: ({ chart }) => ({ right: valueRoom(chart) }) },
		// Anywhere along a bar's row picks it out, not only the bar itself.
		interaction: { mode: 'nearest', axis: 'y', intersect: false },
		onHover: (_event, elements) => onHighlight(elements.length ? group.bars[elements[0].index].key : null),
		plugins: { legend: { display: false }, tooltip: { enabled: false } },
		scales: {
			x: { type: 'linear', min: range.min, max: range.max, display: false },
			y: { grid: { display: false } },
		},
	};
	const picked = group.bars.find(bar => bar.key === highlighted);

	return (
		<section className="grid gap-1">
			<h3 className="mb-0 text-base">
				{group.label} <span className="text-xs font-normal text-muted">{group.summary}</span>
			</h3>
			<p className="mb-0 text-xs text-muted">{t('scale', { min: formatBulkDps(range.min), max: formatBulkDps(range.max) })}</p>
			<div
				className={clsx('relative w-full', PLOT_HEIGHT_CLASSES[Math.min(group.bars.length, PLOT_HEIGHT_CLASSES.length - 1)])}
				onMouseLeave={() => onHighlight(null)}>
				{group.bars.length > 0 && (
					<Bar
						datasetIdKey="id"
						data={data}
						options={options}
						plugins={PLUGINS}
						role="img"
						aria-label={t('chart_label', { encounter: group.label })}
					/>
				)}
			</div>
			<p className="mb-0 min-h-lh text-sm tabular-nums">{picked?.details}</p>
		</section>
	);
};

export const BulkCompareChart = ({ groups, highlighted, onHighlight }: BulkCompareChartProps) => (
	<figure className="m-0 grid gap-6">
		<figcaption className="grid gap-2 text-sm">
			{t('caption')}
			<span className="flex flex-wrap gap-4 text-xs">
				<span className="flex items-center gap-2">
					<span aria-hidden className="size-2 rounded-full bg-grey" />
					{i18n.t('bulk_tab.results.current_gear')}
				</span>
				<span className="flex items-center gap-2">
					<span aria-hidden className="size-2 rounded-full bg-chart-series-1" />
					{t('key_ranked')}
				</span>
			</span>
		</figcaption>
		{groups.map((group, index) => (
			<BulkCompareGroupChart key={index} group={group} highlighted={highlighted} onHighlight={onHighlight} />
		))}
	</figure>
);
