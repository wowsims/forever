import i18n from '@i18n/config';
import type { CombatLog } from '@sim/proto/combat_log';
import { Button } from '@ui-kit/Button';
import type { Chart as ChartJS, ChartData, ChartOptions, Plugin } from 'chart.js';
import { Suspense, useEffect, useMemo, useRef, useState } from 'react';

import { Chart } from '../../../../../shared/lazy_charts';
import { useHoverTooltip } from '../../../hooks/useHoverTooltip';
import type { AnnotationSpec, ScheduledKind, TimelineChartSpec, TimelineDataset, TimelinePoint, TooltipSpec } from '../../../model/timeline/chart/types';
import { annotationsPlugin, SCHEDULED_ROWS } from './annotations';
import { ChartSeriesTooltip } from './ChartSeriesTooltip';
import { ChartToolbar } from './ChartToolbar';
import { cssVarColor } from './colors';
import { THREAT_SERIES_ID } from './series';
import { ChartZoom } from './zoom';

const PAN_STEP_PX = 60;
const ZOOM_STEP = 1.2;

export interface TimelineChartProps {
	/** Null until the DPS view has been opened: building the series costs a pass over every log the unit kept. */
	spec: TimelineChartSpec | null;
}

export const TimelineChart = ({ spec }: TimelineChartProps) => {
	const chartRef = useRef<ChartJS<'line'> | undefined>(undefined);
	const annotations = useRef<AnnotationSpec | null>(null);
	// Keyed by the stable series id, not by dataset index or legend label, so a toggle survives both a
	// dataset rebuild and a UI language change.
	const seriesVisible = useRef(new Map<string, boolean>());
	// Local to the chart like series visibility; the state only re-renders the legend.
	const hiddenKinds = useRef(new Set<ScheduledKind>());
	const [, setKindsVersion] = useState(0);
	// datasetIndex + dataIndex, which a new result can reuse for a different log, so a rebuild clears it.
	const tipKey = useRef('');
	const suppressed = useRef(false);

	const { ref: tooltipRef, content: tip, show: showTip, moveTo: moveTip, hide: hideTip } = useHoverTooltip<{ spec: TooltipSpec; log: CombatLog }>();

	const isSeriesVisible = (seriesId: string) => seriesVisible.current.get(seriesId) ?? seriesId !== THREAT_SERIES_ID;

	// Everything about the chart except its scales, which are the only part a new result changes.
	// Held rather than rebuilt so react-chartjs-2's options effect fires once per result, not per render.
	const behaviour = useRef<ChartOptions<'line'>>(undefined as never);
	if (!behaviour.current) {
		behaviour.current = {
			animation: false,
			responsive: true,
			maintainAspectRatio: false,
			parsing: false,
			normalized: true,
			interaction: { mode: 'nearest', axis: 'xy', intersect: false },
			plugins: {
				legend: {
					position: 'top',
					// Threat is off by default in the single-player view — it is rarely what anyone opened
					// the chart for in MoP — but a legend click sticks for the rest of the session.
					onClick: (_event, item) => {
						const chart = chartRef.current;
						const dataset = item.datasetIndex == null ? undefined : (chart?.data.datasets[item.datasetIndex] as TimelineDataset | undefined);
						if (!chart || !dataset) return;
						const next = !isSeriesVisible(dataset.seriesId);
						seriesVisible.current.set(dataset.seriesId, next);
						dataset.hidden = !next;
						chart.update('none');
					},
				},
				tooltip: {
					enabled: false,
					external: context => {
						const { chart, tooltip } = context;
						if (suppressed.current || tooltip.opacity === 0 || !tooltip.dataPoints?.length) {
							tipKey.current = '';
							hideTip();
							return;
						}
						const rect = chart.canvas.getBoundingClientRect();
						const x = rect.left + tooltip.caretX;
						const y = rect.top + tooltip.caretY;
						const key = tooltip.dataPoints.map(point => `${point.datasetIndex}:${point.dataIndex}`).join(',');
						if (key === tipKey.current) {
							moveTip(x, y);
							return;
						}
						tipKey.current = key;
						const point = tooltip.dataPoints[0];
						const dataset = chart.data.datasets[point.datasetIndex] as TimelineDataset | undefined;
						const raw = point.raw as TimelinePoint | undefined;
						if (dataset && raw?.log) showTip({ spec: dataset.tooltipSpec, log: raw.log }, x, y);
						else hideTip();
					},
				},
			},
		};
	}

	const zoom = useRef<ChartZoom>(undefined as never);
	if (!zoom.current) {
		zoom.current = new ChartZoom(
			() => chartRef.current ?? null,
			dragging => {
				suppressed.current = dragging;
				if (dragging) {
					tipKey.current = '';
					hideTip();
				}
			},
		);
	}

	const plugins = useRef<Array<Plugin<'line'>>>(undefined as never);
	if (!plugins.current)
		plugins.current = [
			annotationsPlugin(
				() => annotations.current,
				kind => !hiddenKinds.current.has(kind),
			),
			zoom.current.plugin(),
		];

	// chart.js has no noData equivalent and would draw a pair of empty axes, so the canvas is only
	// mounted once there is a series to draw.
	const hasData = !!spec && spec.datasets.length > 0;

	// Written here rather than from an effect: a child's effects run before its parent's, and the one
	// react-chartjs-2 updates the chart from draws synchronously, so an effect would paint the new
	// result's series under the previous result's cooldown bands.
	annotations.current = spec?.annotations ?? null;
	if (spec) for (const dataset of spec.datasets) dataset.hidden = !isSeriesVisible(dataset.seriesId);

	const scheduledKinds = SCHEDULED_ROWS.filter(row => spec?.annotations?.scheduled.some(marker => marker.kind === row.kind));

	const toggleKind = (kind: ScheduledKind) => {
		if (!hiddenKinds.current.delete(kind)) hiddenKinds.current.add(kind);
		setKindsVersion(version => version + 1);
		chartRef.current?.update('none');
	};

	const data = useMemo<ChartData<'line', Array<TimelinePoint>>>(() => ({ datasets: spec?.datasets ?? [] }), [spec]);
	const options = useMemo<ChartOptions<'line'>>(() => ({ ...behaviour.current, scales: spec?.scales ?? {} }), [spec]);

	useEffect(() => {
		if (spec) zoom.current.setDuration(spec.duration);
	}, [spec]);

	return (
		<div data-testid="timeline-chart" className="m-auto flex h-full flex-col gap-2">
			<ChartToolbar
				onReset={() => zoom.current.reset()}
				onZoomIn={() => zoom.current.zoomBy(ZOOM_STEP)}
				onZoomOut={() => zoom.current.zoomBy(1 / ZOOM_STEP)}
				onPanLeft={() => zoom.current.panBy(-PAN_STEP_PX)}
				onPanRight={() => zoom.current.panBy(PAN_STEP_PX)}
			/>
			{hasData && scheduledKinds.length > 0 && (
				<ul data-testid="timeline-scheduled-legend" className="m-0 flex list-none flex-wrap items-center gap-x-4 p-0 text-sm">
					<li>{i18n.t('results_tab.details.timeline.scheduled.title')}</li>
					{scheduledKinds.map(({ kind, token }) => (
						<li key={kind}>
							<Button
								variant="unstyled"
								aria-pressed={!hiddenKinds.current.has(kind)}
								className={`flex cursor-pointer items-center gap-1 ${hiddenKinds.current.has(kind) ? 'line-through opacity-50' : ''}`}
								onClick={() => toggleKind(kind)}>
								<span aria-hidden className="inline-block size-2" style={{ backgroundColor: cssVarColor(token) }} />
								{i18n.t(`results_tab.details.timeline.scheduled.${kind}`)}
							</Button>
						</li>
					))}
				</ul>
			)}
			{hasData && (
				<div data-testid="timeline-chart-canvas" className="relative min-h-0 flex-1">
					<Suspense>
						<Chart
							type="line"
							ref={chartRef}
							data={data}
							options={options}
							plugins={plugins.current}
							// react-chartjs-2 otherwise matches a new result's datasets to the old ones by
							// `label`, which is translated and can collide; `seriesId` is the stable key.
							datasetIdKey="seriesId"
							updateMode="none"
							role="img"
							aria-label={i18n.t('results_tab.details.timeline.chart_options.chart_label')}
							onPointerDown={event => zoom.current.down(event)}
							onPointerMove={event => zoom.current.move(event)}
							onPointerUp={event => zoom.current.up(event)}
							onPointerCancel={event => zoom.current.cancel(event)}
						/>
					</Suspense>
				</div>
			)}
			{!hasData && (
				<div data-testid="timeline-chart-empty" className="flex flex-1 items-center justify-center">
					{i18n.t('results_tab.details.timeline.chart_options.waiting_for_data')}
				</div>
			)}
			{tip && (
				<div ref={tooltipRef} data-testid="timeline-hover-tooltip" className="ui-timeline-hover-tooltip">
					<ChartSeriesTooltip spec={tip.spec} log={tip.log} />
				</div>
			)}
		</div>
	);
};
