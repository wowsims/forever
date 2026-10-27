import { UnitMetrics } from '@sim/proto/sim_result';
import { distinct, stringComparator } from '@sim/utils/collections';
import { Chart, Plugin } from 'chart.js';

import { actionColors } from '../../../model/color_settings';
import { AnnotationSpec, ScheduledKind } from '../../../model/timeline/chart/types';
import { cssVarColor } from './colors';

const MAX_ALLOWED_DIST = 10;
const ICON_SIZE = 20;
const ICON_ROW_HEIGHT = 25;
const BAND_OPACITY = 0.3;
const SCHEDULED_STRIP_HEIGHT = 6;
const SCHEDULED_ROW_HEIGHT = 9;

// One row of strips along the top edge per kind, so they never cover the cooldown bands or each other.
export const SCHEDULED_ROWS: ReadonlyArray<{ kind: ScheduledKind; token: string }> = [
	{ kind: 'downtime', token: '--designer-lane-boss-bar' },
	{ kind: 'wave', token: '--designer-lane-waves-bar' },
	{ kind: 'amp', token: '--designer-lane-amps-bar' },
	{ kind: 'move', token: '--designer-lane-moves-bar' },
];

const iconCache = new Map<string, HTMLImageElement>();

function iconImage(url: string): HTMLImageElement {
	let img = iconCache.get(url);
	if (!img) {
		img = new Image();
		img.src = url;
		iconCache.set(url, img);
	}
	return img;
}

// A 404 also reports `complete`, with zero intrinsic size.
const isDrawable = (img: HTMLImageElement) => img.complete && img.naturalWidth > 0;

export function majorCooldownAnnotations(unit: UnitMetrics): AnnotationSpec {
	const mcdLogs = unit.majorCooldownLogs;
	const mcdAuraLogs = unit.majorCooldownAuraUptimeLogs;

	// Cooldowns used very close to each other stack their icons upwards instead of overlapping.
	const iconRows = mcdLogs.map(
		(mcdLog, mcdIdx) => mcdLogs.filter((cdLog, cdIdx) => cdIdx < mcdIdx && cdLog.timestamp > mcdLog.timestamp - MAX_ALLOWED_DIST).length,
	);

	const distinctMcdAuras = distinct(mcdAuraLogs, (a, b) => a.actionId!.equalsIgnoringTag(b.actionId!));
	// Sort by name so auras keep their same colors even if timings change.
	distinctMcdAuras.sort((a, b) => stringComparator(a.actionId!.name, b.actionId!.name));

	return {
		bands: mcdAuraLogs.map(log => ({
			start: log.gainedAt,
			end: log.fadedAt,
			color: actionColors[distinctMcdAuras.findIndex(dAura => dAura.actionId!.equalsIgnoringTag(log.actionId!)) % actionColors.length],
		})),
		icons: mcdLogs.map((log, i) => ({ time: log.timestamp, row: iconRows[i], url: log.actionId!.iconUrl })),
		scheduled: [],
	};
}

function drawScheduled(chart: Chart, spec: AnnotationSpec, isVisible: (kind: ScheduledKind) => boolean) {
	const { ctx, chartArea } = chart;
	const scale = chart.scales.x;
	ctx.save();
	ctx.beginPath();
	ctx.rect(chartArea.left, chartArea.top, chartArea.right - chartArea.left, chartArea.bottom - chartArea.top);
	ctx.clip();
	SCHEDULED_ROWS.forEach(({ kind, token }, row) => {
		if (!isVisible(kind)) return;
		const color = cssVarColor(token);
		const top = chartArea.top + row * SCHEDULED_ROW_HEIGHT;
		ctx.fillStyle = color;
		ctx.strokeStyle = color;
		for (const marker of spec.scheduled) {
			if (marker.kind !== kind) continue;
			const left = scale.getPixelForValue(marker.start);
			if (kind === 'move') {
				const mid = top + SCHEDULED_STRIP_HEIGHT / 2;
				ctx.beginPath();
				ctx.moveTo(left, mid - SCHEDULED_STRIP_HEIGHT / 2);
				ctx.lineTo(left + SCHEDULED_STRIP_HEIGHT / 2, mid);
				ctx.lineTo(left, mid + SCHEDULED_STRIP_HEIGHT / 2);
				ctx.lineTo(left - SCHEDULED_STRIP_HEIGHT / 2, mid);
				ctx.closePath();
				ctx.fill();
			} else {
				const right = scale.getPixelForValue(marker.end);
				ctx.fillRect(left, top, Math.max(right - left, 1), SCHEDULED_STRIP_HEIGHT);
			}
		}
	});
	ctx.restore();
}

export function annotationsPlugin(getSpec: () => AnnotationSpec | null, isScheduledVisible: (kind: ScheduledKind) => boolean): Plugin<'line'> {
	const hooked = new WeakSet<HTMLImageElement>();

	const clipToChartArea = (chart: Chart) => {
		const { ctx, chartArea } = chart;
		ctx.save();
		ctx.beginPath();
		ctx.rect(chartArea.left, chartArea.top, chartArea.right - chartArea.left, chartArea.bottom - chartArea.top);
		ctx.clip();
	};

	return {
		id: 'timelineAnnotations',
		beforeDatasetsDraw(chart) {
			const spec = getSpec();
			const scale = chart.scales.x;
			if (!spec?.bands.length || !scale) return;

			const { ctx, chartArea } = chart;
			clipToChartArea(chart);
			ctx.globalAlpha = BAND_OPACITY;
			for (const band of spec.bands) {
				const left = scale.getPixelForValue(band.start);
				const right = scale.getPixelForValue(band.end);
				ctx.fillStyle = band.color;
				ctx.fillRect(left, chartArea.top, Math.max(right - left, 1), chartArea.bottom - chartArea.top);
			}
			ctx.restore();
		},
		afterDatasetsDraw(chart) {
			const spec = getSpec();
			const scale = chart.scales.x;
			if (spec?.scheduled.length && scale) drawScheduled(chart, spec, isScheduledVisible);
			if (!spec?.icons.length || !scale) return;

			const { ctx, chartArea } = chart;
			clipToChartArea(chart);
			for (const icon of spec.icons) {
				const img = iconImage(icon.url);
				if (!isDrawable(img)) {
					if (!img.complete && !hooked.has(img)) {
						hooked.add(img);
						img.addEventListener(
							'load',
							() =>
								requestAnimationFrame(() => {
									// The chart may have been torn down while the icon was in flight.
									if (chart.ctx) chart.render();
								}),
							{ once: true },
						);
					}
					continue;
				}
				ctx.drawImage(
					img,
					scale.getPixelForValue(icon.time) - ICON_SIZE / 2,
					chartArea.bottom - ICON_SIZE / 2 - icon.row * ICON_ROW_HEIGHT,
					ICON_SIZE,
					ICON_SIZE,
				);
			}
			ctx.restore();
		},
	};
}
