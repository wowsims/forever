// Driven through the sidebar action as a user does. The sim and the history store are stubbed; the settings, the store and the
// run flags are real.
import { ErrorOutcomeType } from '@generated/proto/api';
import { Class, Spec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { StatSweepPoint, StatSweepRecord } from '@sim/cache/stat_sweep_history';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { UnitStat } from '@sim/proto/stats';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { RequestTypes } from '@sim/sim_signal_manager';
import { patchSlice, type SimStore } from '@sim/state/sim_store';
import { Z_90 } from '@sim/utils/math';
import { fakeCharts } from '@testing/chart-js-mock';
import { makeHitSweepPoints, makeStatSweepHost, makeSweepSample, type SweepPass, sweepStepDps } from '@testing/storybook/fixtures/stat-sweep';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { toastManager } from '@ui-kit/Toast';
import type { ChartOptions, Plugin } from 'chart.js';
import { beforeEach, describe, expect, it, type Mock, vi } from 'vitest';

import { StatSweepAction } from '../StatSweepAction';

// happy-dom has no IndexedDB.
const history = vi.hoisted(() => ({
	listRuns: vi.fn(async (_specId: number) => [] as StatSweepRecord[]),
	saveRun: vi.fn(async (_run: StatSweepRecord) => {}),
	deleteRuns: vi.fn(async (_ids: readonly string[]) => {}),
}));
vi.mock('@sim/cache/stat_sweep_history', () => history);
vi.mock('chart.js', async () => (await import('@testing/chart-js-mock')).chartJsMock);
const download = vi.hoisted(() => vi.fn());
vi.mock('@ui-kit/utils/dom', async importOriginal => ({ ...(await importOriginal<typeof import('@ui-kit/utils/dom')>()), downloadString: download }));
vi.mock('../../../../tracking/utils', () => ({ trackPageView: vi.fn() }));

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);
const AGILITY = UnitStat.fromStat(Stat.StatAgility);
// A Target ± step's first pass.
const PROBE = 1000;

type Datum = { x: number; y: number; point: StatSweepPoint };
type Dataset = { label: string; data: Datum[] } & Record<string, unknown>;
type FakeChart = { config: { data: { datasets: Dataset[] }; options: ChartOptions<'line'>; plugins: Array<Plugin<'line'>> } };
const charts = fakeCharts as FakeChart[];
const lastChart = () => charts[charts.length - 1].config;

const rootElem = document.createElement('div');
let store: SimStore;
let settings: StatSweepSettings;
let host: IndividualSimHost<any>;
let simulatePass: Mock<(options: SweepPass) => Promise<object>>;
let abortType: Mock<(typeMask: number) => Promise<void>>;

const setup = (simDisabled = false) => {
	abortType = vi.fn(async () => {});
	simulatePass = vi.fn(async (options: SweepPass) => makeSweepSample(options.iterations, sweepStepDps(options)));
	({ host, store, settings } = makeStatSweepHost({ simulatePass: simulatePass as never, abortType, simDisabled }));
};

const renderAction = () => {
	document.body.appendChild(rootElem);
	render(
		<PortalContainerContext value={rootElem}>
			<SimHostProvider host={host}>
				<StatSweepAction settings={settings} />
			</SimHostProvider>
		</PortalContainerContext>,
	);
};

const action = () => screen.queryByRole<HTMLButtonElement>('button', { name: i18n.t('sidebar.buttons.stat_sweep.title') });
const showPowerTools = async (shown: boolean) => {
	await act(async () => patchSlice(store, 'ui', { showPowerTools: shown }));
};
const openDialog = async () => {
	renderAction();
	await showPowerTools(true);
	await act(async () => {
		fireEvent.click(action()!);
	});
};

const dialog = () => rootElem.querySelector('[data-testid="stat-sweep-dialog"]');
const runButton = () => screen.getByRole<HTMLButtonElement>('button', { name: 'sidebar.buttons.stat_sweep.run' });
const run = async () => {
	await act(async () => {
		fireEvent.click(runButton());
	});
};
const tracker = () => rootElem.querySelector('[data-testid="stat-sweep-progress-dialog"]');
const bonusOf = (call: [SweepPass], stat: Stat) => call[0].bonusStats.getStat(stat);
const strengthSteps = () => simulatePass.mock.calls.map(call => bonusOf(call, Stat.StatStrength));
const saved = (index = 0) => history.saveRun.mock.calls[index][0];
const metricLabel = (metric: string) => i18n.t(`sidebar.results.metrics.${metric}.label`);
const metricToggle = () => within(rootElem.querySelector<HTMLElement>('[data-testid="stat-sweep-metric"]')!);
const offeredMetrics = () =>
	metricToggle()
		.getAllByRole('button')
		.map(button => button.textContent);
const pressedMetric = () =>
	metricToggle()
		.getAllByRole('button')
		.find(button => button.getAttribute('aria-pressed') === 'true')?.textContent;
const precisionMode = (mode: 'iterations' | 'target') =>
	// Also behind the modal progress dialog.
	screen.getByRole<HTMLButtonElement>('button', { name: `common.precision.mode_${mode}`, hidden: true });
const HISTORY_EYE = /^sidebar\.buttons\.stat_sweep\.history\.(show|hide|show_hidden)$/;
// The eyes also include the runs inside a closed batch; a batch's own eye is the one whose nearest row is the batch's.
const eyes = (inBatch: boolean) =>
	screen
		.getAllByRole<HTMLButtonElement>('button', { name: HISTORY_EYE, hidden: true })
		.filter(eye => eye.closest('li')!.matches('[data-testid="stat-sweep-history-batch"]') === inBatch);
const drawnState = (eye: HTMLElement) => eye.getAttribute('aria-pressed');
// Each run's eye, in list order.
const overlayTicks = () => eyes(false).map(drawnState);
const batchTicks = () => eyes(true).map(drawnState);
const batchEye = (index = 0) => eyes(true)[index];
const runEye = (index: number) => eyes(false)[index];
// The title is the first text of the header button, which also holds the chevron and the date.
const batchLabels = () =>
	[...rootElem.querySelectorAll('[data-testid="stat-sweep-history-batch"] > button[aria-expanded] > span:first-of-type')].map(title => title.textContent);
const deleteButtons = (key: 'delete' | 'delete_batch') => screen.getAllByRole('button', { name: `sidebar.buttons.stat_sweep.history.${key}`, hidden: true });
const STRENGTH_NAME = STRENGTH.getFullName(Class.ClassWarrior);
const AGILITY_NAME = AGILITY.getFullName(Class.ClassWarrior);
const expanders = () => screen.getAllByRole('button').filter(button => button.matches('[data-testid="stat-sweep-history-batch"] > button'));
// Iterations of each pass, by Strength added.
const passesByStep = () => {
	const passes: Record<number, number[]> = {};
	for (const call of simulatePass.mock.calls) (passes[bonusOf(call, Stat.StatStrength)] ??= []).push(call[0].iterations);
	return passes;
};
// Every metric stops at 1010 DPS, which Strength reaches at +10.
const cappedAtTen = () => simulatePass.mockImplementation(async options => makeSweepSample(options.iterations, Math.min(sweepStepDps(options), 1010)));
const storedRun = (id: string, createdAt: number, colorSlot: number, batchId?: string, unitStat = STRENGTH): StatSweepRecord => ({
	id,
	createdAt,
	specId: Spec.SpecUnknown,
	label: id,
	colorSlot,
	unitStat: unitStat.toJson(),
	metric: 'dps',
	baseStatValue: 120,
	points: makeHitSweepPoints(120),
	...(batchId && { batchId, batchLabel: `${batchId} label` }),
});

beforeEach(() => {
	setup();
	charts.length = 0;
	history.listRuns.mockClear();
	history.saveRun.mockClear();
	history.deleteRuns.mockClear();
	download.mockClear();
});

describe('StatSweepDialog', () => {
	describe('the sidebar action', () => {
		it('shows only while Show power tools is on, as an icon button named for the tool, and opens the dialog', async () => {
			renderAction();
			expect(action()).toBeNull();

			await showPowerTools(true);
			expect(action()!.textContent).toBe('');
			expect(dialog()).toBeNull();
			expect(history.listRuns).not.toHaveBeenCalled();
			await act(async () => {
				fireEvent.click(action()!);
			});
			expect(dialog()).not.toBeNull();
			expect(history.listRuns).toHaveBeenCalledTimes(1);
		});

		it('stays hidden in a gear planner, which runs no sims', async () => {
			setup(true);
			renderAction();
			await showPowerTools(true);

			expect(action()).toBeNull();
		});
	});

	it('runs one plain sim per step at one seed with the step as bonus stats, and saves every metric of each step', async () => {
		settings.set({ range: 20, step: 10 });
		settings.set({ iterations: 3000 });
		await openDialog();

		await run();

		expect(strengthSteps()).toEqual([0, 10, 20]);
		expect(simulatePass.mock.calls.map(call => [call[0].iterations, call[0].seed])).toEqual([
			[3000, 7],
			[3000, 7],
			[3000, 7],
		]);
		expect(history.saveRun).toHaveBeenCalledTimes(1);
		expect(saved().metric).toBe('dps');
		expect(saved().points.map(point => [point.statValue, point.metrics.dps.avg, point.metrics.dps.n])).toEqual([
			[0, 1000, 3000],
			[10, 1010, 3000],
			[20, 1020, 3000],
		]);
		expect(saved().points[1].metrics).toEqual({
			dps: { avg: 1010, stdev: 100, n: 3000 },
			hps: { avg: 10, stdev: 5, n: 3000 },
			tps: { avg: 2020, stdev: 400, n: 3000 },
			dtps: { avg: 1990, stdev: 300, n: 3000 },
			tmi: { avg: 10.1, stdev: 2, n: 3000 },
		});
		expect(saved().cappedAt).toBeUndefined();
	});

	it('stores an entered range and step, and the stats picked and removed by key', async () => {
		await openDialog();

		fireEvent.change(screen.getByLabelText('sidebar.buttons.stat_sweep.range'), { target: { value: '50' } });
		fireEvent.change(screen.getByLabelText('sidebar.buttons.stat_sweep.step'), { target: { value: '5' } });
		expect(settings.values).toMatchObject({ range: 50, step: 5 });

		expect(screen.getAllByTestId('multi-combo-box-chip').map(chip => chip.textContent)).toEqual([STRENGTH.getFullName(Class.ClassWarrior)]);
		fireEvent.click(screen.getByRole('button', { name: 'sidebar.buttons.stat_sweep.stats.remove' }));
		expect(settings.values.stats).toEqual([]);
		expect(runButton().disabled).toBe(true);

		fireEvent.click(screen.getByTestId('multi-combo-box-trigger'));
		const agility = within(screen.getByTestId('multi-combo-box-list'))
			.getAllByRole('option')
			.find(option => option.textContent?.includes(AGILITY.getFullName(Class.ClassWarrior)))!;
		fireEvent.click(agility);
		expect(settings.values.stats!.map(stat => stat.getKey())).toEqual([AGILITY.getKey()]);
	});

	it('locks the setup and the precision options while a run goes', async () => {
		simulatePass.mockImplementation(() => new Promise(() => {}));
		await openDialog();

		await run();
		expect(screen.getByLabelText<HTMLInputElement>('sidebar.buttons.stat_sweep.range').disabled).toBe(true);
		expect(screen.getByLabelText<HTMLInputElement>('sidebar.buttons.stat_sweep.step').disabled).toBe(true);
		expect(precisionMode('target').disabled).toBe(true);
		expect(screen.getByLabelText<HTMLInputElement>('common.precision.iterations').disabled).toBe(true);
	});

	it('shows the band key while the spread is on, and hides it with the spread checkbox, which stores the pick', async () => {
		await openDialog();
		const bandKey = () => screen.queryByText(i18n.t('sidebar.buttons.stat_sweep.band_key', { metric: metricLabel('dps') }));
		expect(bandKey()).toBeTruthy();

		fireEvent.click(within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox'));
		expect(settings.values.showSpread).toBe(false);
		expect(bandKey()).toBeNull();

		fireEvent.click(within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox'));
		expect(settings.values.showSpread).toBe(true);
		expect(bandKey()).toBeTruthy();
	});

	it('offers DPS and TPS, adds DTPS and TMI with the threat and tank options, and HPS with the healing metrics', async () => {
		await openDialog();
		expect(offeredMetrics()).toEqual([metricLabel('dps'), metricLabel('tps')]);

		act(() => patchSlice(store, 'ui', { showThreatMetrics: true }));
		expect(offeredMetrics()).toEqual(['dps', 'tps', 'dtps', 'tmi'].map(metricLabel));

		act(() => patchSlice(store, 'ui', { showHealingMetrics: true }));
		expect(offeredMetrics()).toEqual(['dps', 'hps', 'tps', 'dtps', 'tmi'].map(metricLabel));
	});

	it('stores the picked metric, and shows DPS while the stored one is not offered', async () => {
		act(() => patchSlice(store, 'ui', { showThreatMetrics: true }));
		await openDialog();

		fireEvent.click(metricToggle().getByRole('button', { name: metricLabel('tmi') }));
		expect(settings.values.metric).toBe('tmi');
		expect(pressedMetric()).toBe(metricLabel('tmi'));

		act(() => patchSlice(store, 'ui', { showThreatMetrics: false }));
		expect(pressedMetric()).toBe(metricLabel('dps'));
		expect(settings.values.metric).toBe('tmi');
	});

	describe('the chart', () => {
		it('plots each run over the stat added as a line in a band of ± one standard deviation, and the line alone with the spread off', async () => {
			settings.set({ range: 10, step: 10 });
			await openDialog();
			await run();

			const atTen = (datasets: Dataset[]) => datasets.map(dataset => dataset.data.find(datum => datum.x === 10)!.y);
			expect(lastChart().data.datasets[2].data.map(datum => datum.x)).toEqual([0, 10]);
			expect(atTen(lastChart().data.datasets)).toEqual([1110, 910, 1010]);
			expect(lastChart().options.plugins!.legend!.display).toBe(false);

			fireEvent.click(within(screen.getByTestId('stat-sweep-show-spread')).getByRole('checkbox'));
			expect(atTen(lastChart().data.datasets)).toEqual([1010]);
		});

		it('marks the current value at 0', async () => {
			settings.set({ range: 10, step: 10 });
			await openDialog();
			await run();
			const marker = lastChart().plugins.find(plugin => plugin.id === 'statSweepMarker')!;
			const getPixelForValue = vi.fn(() => 40);
			const moveTo = vi.fn();
			const ctx = { save: vi.fn(), restore: vi.fn(), setLineDash: vi.fn(), beginPath: vi.fn(), moveTo, lineTo: vi.fn(), stroke: vi.fn() };

			(marker.afterDatasetsDraw as (chart: unknown) => void)({
				ctx,
				chartArea: { left: 0, right: 100, top: 0, bottom: 50 },
				scales: { x: { getPixelForValue } },
			});

			expect(getPixelForValue).toHaveBeenCalledWith(0);
			expect(moveTo).toHaveBeenCalledWith(40, 0);
		});

		it('lists each run once in the legend from two runs on, and titles the tooltip with each stat total and the stat added', async () => {
			settings.set({ stats: [STRENGTH, AGILITY], range: 10, step: 10 });
			await openDialog();
			await run();

			const { data, options } = lastChart();
			expect(options.plugins!.legend!.display).toBe(true);
			const items = options.plugins!.legend!.labels!.generateLabels!({ data, isDatasetVisible: () => true } as never);
			expect(items.map(item => item.datasetIndex)).toEqual([2, 5]);
			const tooltipFilter = options.plugins!.tooltip!.filter!;
			const lines = data.datasets
				.map((_, datasetIndex) => datasetIndex)
				.filter(datasetIndex => tooltipFilter({ datasetIndex, dataset: data.datasets[datasetIndex] } as never, 0, [], {} as never));
			expect(lines).toEqual([2, 5]);

			// The newest run is drawn first.
			const expected = [AGILITY, STRENGTH].map(stat => `${stat.getFullName(Class.ClassWarrior)} 10 (+10)`);
			const title = options.plugins!.tooltip!.callbacks!.title! as unknown as (items: object[]) => string[];
			const at = (datasetIndex: number, added: number) => ({ raw: data.datasets[datasetIndex].data.find(datum => datum.x === added)! });
			const t = vi
				.spyOn(i18n, 't')
				.mockImplementation(
					((_key: string, values: { stat: string; total: number; added: string }) => `${values.stat} ${values.total} (${values.added})`) as never,
				);
			expect(title([at(2, 10), at(5, 10)])).toEqual(expected);
			t.mockRestore();
		});

		it('reads the hovered value and its ± as the sim card does, and marks each point from the cap on as capped', async () => {
			cappedAtTen();
			settings.set({ range: 40, step: 10 });
			await openDialog();
			await run();

			const { data, options } = lastChart();
			const label = options.plugins!.tooltip!.callbacks!.label! as unknown as (item: object) => string;
			const t = vi.spyOn(i18n, 't');
			const valueAt = (added: number) => {
				t.mockClear();
				label({ dataset: data.datasets[2], raw: data.datasets[2].data.find(datum => datum.x === added) });
				return t.mock.calls.find(([key]) => String(key).startsWith('sidebar.buttons.stat_sweep.tooltip_value'));
			};
			expect(valueAt(0)).toEqual(['sidebar.buttons.stat_sweep.tooltip_value', { value: (1000).toFixed(2), stdev: (100).toFixed(0) }]);
			expect([10, 20, 30].map(added => valueAt(added)![0])).toEqual([
				'sidebar.buttons.stat_sweep.tooltip_value_capped',
				'sidebar.buttons.stat_sweep.tooltip_value_capped',
				'sidebar.buttons.stat_sweep.tooltip_value_capped',
			]);
			t.mockRestore();
		});
	});

	it('stops a stat going up after two steps in a row give every metric of the step below, and saves where it capped', async () => {
		cappedAtTen();
		settings.set({ range: 50, step: 10 });
		await openDialog();

		await run();

		expect(strengthSteps()).toEqual([0, 10, 20, 30]);
		expect(saved().points.map(point => point.statValue)).toEqual([0, 10, 20, 30]);
		expect(saved().cappedAt).toBe(10);
		expect(screen.getByText('sidebar.buttons.stat_sweep.capped_label')).toBeTruthy();
	});

	it('keeps a stat going up while it changes only TMI, whatever metric the run follows, and records that metric on the run', async () => {
		// Every metric but TMI stays put across the steps.
		simulatePass.mockImplementation(async options => makeSweepSample(options.iterations, 1000, 10 + sweepStepDps(options) / 100));
		settings.set({ range: 40, step: 10 });
		settings.set({ metric: 'tps' });
		await openDialog();

		await run();

		expect(strengthSteps()).toEqual([0, 10, 20, 30, 40]);
		expect(saved().metric).toBe('tps');
		expect(saved().cappedAt).toBeUndefined();
		expect(screen.getByTestId('stat-sweep-history-metric').textContent).toBe(metricLabel('tps'));
	});

	it("runs each picked stat on its own in the order offered, saving each as its own run, and works out each stat's totals but the current one as it starts", async () => {
		const totals = vi.spyOn(host.sim, 'computeRaidStats');
		const totalsBefore: number[] = [];
		simulatePass.mockImplementation(async options => {
			totalsBefore.push(totals.mock.calls.length);
			return makeSweepSample(options.iterations, sweepStepDps(options));
		});
		settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
		await openDialog();

		await run();

		expect(simulatePass.mock.calls.map(call => [bonusOf(call, Stat.StatStrength), bonusOf(call, Stat.StatAgility)])).toEqual([
			[0, 0],
			[10, 0],
			[0, 0],
			[0, 10],
		]);
		expect(totalsBefore).toEqual([2, 2, 4, 4]);
		const runs = history.saveRun.mock.calls.map(call => call[0]);
		expect(runs.map(record => UnitStat.fromJson(record.unitStat).getKey())).toEqual([STRENGTH.getKey(), AGILITY.getKey()]);
		expect(runs[0].colorSlot).not.toBe(runs[1].colorSlot);
	});

	describe('the history', () => {
		it('unticks the earlier runs when a run starts, keeping every run of the new batch ticked', async () => {
			settings.set({ stats: [STRENGTH], range: 10, step: 10 });
			await openDialog();

			await run();
			expect(overlayTicks()).toEqual(['true']);

			act(() => settings.set({ stats: [AGILITY, STRENGTH] }));
			await run();
			expect(overlayTicks()).toEqual(['true', 'true', 'false']);
		});

		it('saves the runs of one sweep as one batch, listed under one header naming its stats in run order, each on a line of its tooltip', async () => {
			settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
			await openDialog();

			await run();

			const [first, second] = history.saveRun.mock.calls.map(call => call[0]);
			expect(first.batchId).toBeTruthy();
			expect(second.batchId).toBe(first.batchId);
			expect(second.batchLabel).toBe(first.batchLabel);
			expect(batchLabels()).toEqual([`${STRENGTH_NAME}, ${AGILITY_NAME}`]);
			expect(batchTicks()).toEqual(['true']);
			// The tooltip ignores a hover in the 50 ms after focus moves, as it does when the progress dialog closes.
			const tooltip = await vi.waitFor(() => {
				fireEvent.mouseEnter(expanders()[0]);
				return screen.getByRole('tooltip');
			});
			const statList = within(tooltip).getAllByRole('listitem');
			expect(statList.map(stat => stat.textContent)).toEqual([STRENGTH_NAME, AGILITY_NAME]);

			act(() => settings.set({ stats: [STRENGTH] }));
			await run();
			expect(saved(2).batchId).not.toBe(first.batchId);
			// The one-stat sweep gets a header of its own, open above the earlier sweep's closed one.
			expect(batchLabels()).toEqual([STRENGTH_NAME, `${STRENGTH_NAME}, ${AGILITY_NAME}`]);
			expect(expanders().map(button => button.getAttribute('aria-expanded'))).toEqual(['true', 'false']);
			expect(overlayTicks()).toEqual(['true', 'false', 'false']);
			expect(batchTicks()).toEqual(['true', 'false']);
		});

		it("draws and hides every run of a batch with the batch's eye, which reads mixed while only some are drawn", async () => {
			settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
			await openDialog();
			await run();

			fireEvent.click(runEye(1));
			expect(overlayTicks()).toEqual(['true', 'false']);
			expect(batchTicks()).toEqual(['mixed']);

			fireEvent.click(batchEye());
			expect(overlayTicks()).toEqual(['true', 'true']);
			expect(batchTicks()).toEqual(['true']);
			expect(lastChart().data.datasets).toHaveLength(6);

			fireEvent.click(batchEye());
			expect(overlayTicks()).toEqual(['false', 'false']);
			expect(batchTicks()).toEqual(['false']);
			expect(screen.getByRole<HTMLButtonElement>('button', { name: 'sidebar.buttons.stat_sweep.export_csv' }).disabled).toBe(true);
		});

		it('deletes one run inside a batch, and every run of a batch with its delete', async () => {
			history.listRuns.mockResolvedValueOnce([
				storedRun('third', 3, 2, 'sweep', AGILITY),
				storedRun('second', 2, 1, 'sweep'),
				storedRun('first', 1, 0, 'sweep', AGILITY),
				storedRun('older', 0, 3),
			]);
			await openDialog();
			expect(overlayTicks()).toEqual(['true', 'true', 'true', 'false']);

			fireEvent.click(deleteButtons('delete')[1]);
			expect(history.deleteRuns).toHaveBeenCalledExactlyOnceWith(['second']);
			expect(overlayTicks()).toEqual(['true', 'true', 'false']);
			expect(batchLabels()).toEqual([`${AGILITY_NAME}, ${AGILITY_NAME}`, STRENGTH_NAME]);

			history.deleteRuns.mockClear();
			fireEvent.click(deleteButtons('delete_batch')[0]);
			expect(history.deleteRuns).toHaveBeenCalledExactlyOnceWith(['first', 'third']);
			expect(overlayTicks()).toEqual(['false']);
			expect(batchLabels()).toEqual([STRENGTH_NAME]);
		});

		it('lists a run saved with no batch as a batch of its own, and opens only the newest batch', async () => {
			history.listRuns.mockResolvedValueOnce([
				storedRun('new-b', 5, 1, 'new', AGILITY),
				storedRun('new-a', 4, 0, 'new'),
				storedRun('unbatched', 3, 2),
				storedRun('old-b', 2, 4, 'old', AGILITY),
				storedRun('old-a', 1, 3, 'old'),
			]);
			await openDialog();

			expect(overlayTicks()).toEqual(['true', 'true', 'false', 'false', 'false']);
			expect(batchLabels()).toEqual([`${STRENGTH_NAME}, ${AGILITY_NAME}`, STRENGTH_NAME, `${STRENGTH_NAME}, ${AGILITY_NAME}`]);
			expect(expanders().map(button => button.getAttribute('aria-expanded'))).toEqual(['true', 'false', 'false']);
			// The closed batches' runs are not reachable.
			expect(screen.getAllByRole('button', { name: HISTORY_EYE })).toHaveLength(5);

			fireEvent.click(expanders()[1]);
			expect(expanders()[1].getAttribute('aria-expanded')).toBe('true');
			expect(screen.getAllByRole('button', { name: HISTORY_EYE })).toHaveLength(6);
			fireEvent.click(expanders()[2]);
			expect(screen.getAllByRole('button', { name: HISTORY_EYE })).toHaveLength(8);
			fireEvent.click(runEye(4));
			expect(batchTicks()).toEqual(['true', 'false', 'mixed']);
		});

		it("names a batch's header by its stats, date and metric, and tells whether it is open", async () => {
			history.listRuns.mockResolvedValueOnce([
				{ ...storedRun('b', 5, 1, 'sweep', AGILITY), metric: 'tps' },
				{ ...storedRun('a', 4, 0, 'sweep'), metric: 'tps' },
			]);
			await openDialog();

			const date = new Intl.DateTimeFormat(i18n.language, { dateStyle: 'short', timeStyle: 'short' }).format(4);
			const shown = [`${STRENGTH_NAME}, ${AGILITY_NAME}`, date, metricLabel('tps')].map(text => text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('\\s*');
			const header = screen.getByRole('button', { name: new RegExp(`^${shown}$`) });
			expect(header.getAttribute('aria-expanded')).toBe('true');
			fireEvent.click(header);
			expect(header.getAttribute('aria-expanded')).toBe('false');
		});

		it('leaves the batches open or closed as they are when the newest is deleted, and opens only the newest when a newer sweep arrives', async () => {
			history.listRuns.mockResolvedValueOnce([
				storedRun('new-b', 6, 1, 'new', AGILITY),
				storedRun('new-a', 5, 0, 'new'),
				storedRun('mid-b', 4, 3, 'mid', AGILITY),
				storedRun('mid-a', 3, 2, 'mid'),
				storedRun('old-b', 2, 5, 'old', AGILITY),
				storedRun('old-a', 1, 4, 'old'),
			]);
			settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
			await openDialog();
			const opened = () => expanders().map(button => button.getAttribute('aria-expanded'));
			fireEvent.click(expanders()[2]);
			expect(opened()).toEqual(['true', 'false', 'true']);

			fireEvent.click(deleteButtons('delete_batch')[0]);
			expect(opened()).toEqual(['false', 'true']);

			await run();
			expect(opened()).toEqual(['true', 'false', 'false']);
		});

		it('keeps the history behind the progress dialog while a sweep runs, so no run can be deleted until it ends', async () => {
			history.listRuns.mockResolvedValueOnce([storedRun('older', 1, 0)]);
			let settleThird!: (result: object) => void;
			simulatePass
				.mockResolvedValueOnce(makeSweepSample(1000))
				.mockResolvedValueOnce(makeSweepSample(1000))
				.mockImplementationOnce(() => new Promise(resolve => (settleThird = resolve)));
			settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
			await openDialog();
			const deletes = (hidden: boolean) => screen.queryAllByRole('button', { name: 'sidebar.buttons.stat_sweep.history.delete_batch', hidden });

			await run();
			// The sweep's batch is listed with its finished Strength run, but only behind the modal progress dialog.
			expect(deletes(true)).toHaveLength(2);
			expect(deletes(false)).toHaveLength(0);

			await act(async () => {
				fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
			});
			await act(async () => {
				settleThird({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
			});
			expect(tracker()).toBeNull();
			expect(deletes(false)).toHaveLength(2);
		});

		it('draws the newest saved run, overlays another when it is ticked, and deletes a run', async () => {
			history.listRuns.mockResolvedValueOnce([storedRun('new', 2, 0), storedRun('old', 1, 1)]);
			await openDialog();
			expect(overlayTicks()).toEqual(['true', 'false']);
			// A saved run's points sit at the stat added, not at its total, which starts at 120 here.
			const stored = makeHitSweepPoints(120);
			expect(stored[0].statValue).not.toBe(stored[0].offset);
			expect(lastChart().data.datasets[2].data.map(datum => datum.x)).toEqual(stored.map(point => point.offset));

			fireEvent.click(expanders()[1]);
			fireEvent.click(runEye(1));
			expect(overlayTicks()).toEqual(['true', 'true']);
			expect(lastChart().data.datasets.map(dataset => dataset.label)).toEqual(['new', 'new', 'new', 'old', 'old', 'old']);

			fireEvent.click(screen.getAllByRole('button', { name: 'sidebar.buttons.stat_sweep.history.delete' })[1]);
			expect(history.deleteRuns).toHaveBeenCalledExactlyOnceWith(['old']);
			expect(overlayTicks()).toEqual(['true']);
		});

		it('exports the drawn runs as CSV with every metric, and offers no export with nothing drawn', async () => {
			settings.set({ range: 10, step: 10 });
			await openDialog();
			const exportButton = () => screen.getByRole<HTMLButtonElement>('button', { name: 'sidebar.buttons.stat_sweep.export_csv' });
			expect(exportButton().disabled).toBe(true);

			await run();
			fireEvent.click(exportButton());

			expect(download).toHaveBeenCalledTimes(1);
			const [csv, fileName, mimeType] = download.mock.calls[0];
			const lines = csv.trim().split('\n');
			expect(lines).toHaveLength(3);
			expect(lines[0]).toBe('run,batch,stat,base_total,added,total,dps,dps_stdev,hps,hps_stdev,tps,tps_stdev,dtps,dtps_stdev,tmi,tmi_stdev,samples');
			expect(lines.slice(1).map((line: string) => line.split(',')[1])).toEqual([saved().batchLabel, saved().batchLabel]);
			expect([fileName, mimeType]).toEqual(['wowsims-stat-sweep.csv', 'text/csv']);
		});
	});

	describe('Target ±', () => {
		it('probes each step and tops it up at the seed after the probe', async () => {
			settings.set({ range: 5, step: 10 });
			settings.set({ precisionMode: 'target', targetCi: 2 });
			await openDialog();

			await run();

			const calls = simulatePass.mock.calls.map(call => [call[0].iterations, call[0].seed]);
			expect(calls[0]).toEqual([PROBE, 7]);
			expect(calls[1][1]).toBe(7 + PROBE);
			expect(saved().points[0].metrics.dps.n).toBe(PROBE + calls[1][0]);
		});

		it("sizes the run by the picked metric's spread, with the target in that metric's units", async () => {
			settings.set({ range: 5, step: 10 });
			settings.set({ metric: 'tps' });
			settings.set({ precisionMode: 'target', targetCi: 20 });
			await openDialog();
			expect(rootElem.querySelector('label[for="stat-sweep-precision-target"]')!.textContent).toContain('sidebar.buttons.stat_sweep.precision.target_ci');

			await run();

			// TPS spreads 400 a fight, where DPS spreads 100: a ± of 20 needs that many iterations of TPS.
			const total = Math.ceil(((Z_90 * 400) / 20) ** 2);
			expect(simulatePass.mock.calls.map(call => call[0].iterations)).toEqual([PROBE, total - PROBE]);
			expect(saved().points[0].metrics.tps.n).toBe(total);
		});

		it('tops up the first step that repeats the step below, and runs only the probe of the step that confirms the cap', async () => {
			cappedAtTen();
			settings.set({ range: 40, step: 10 });
			settings.set({ precisionMode: 'target', targetCi: 2 });
			await openDialog();

			await run();

			const topUp = Math.ceil(((Z_90 * 100) / 2) ** 2) - PROBE;
			expect(passesByStep()).toEqual({ 0: [PROBE, topUp], 10: [PROBE, topUp], 20: [PROBE, topUp], 30: [PROBE] });
			expect(saved().cappedAt).toBe(10);
		});

		it('disables Run while the target is empty, and enables it in iterations mode', async () => {
			settings.set({ precisionMode: 'target', targetCi: 2 });
			await openDialog();
			expect(runButton().disabled).toBe(false);

			fireEvent.change(rootElem.querySelector('#stat-sweep-precision-target')!, { target: { value: '' } });
			expect(settings.values.targetCi).toBe(0);
			expect(runButton().disabled).toBe(true);

			fireEvent.click(precisionMode('iterations'));
			expect(runButton().disabled).toBe(false);
		});
	});

	describe('the progress dialog', () => {
		it('shows the run across every step with the time left, and cancels it there, keeping the finished steps', async () => {
			const toast = vi.spyOn(toastManager, 'add');
			let settleSecond!: (result: object) => void;
			simulatePass.mockResolvedValueOnce(makeSweepSample(1000)).mockImplementationOnce(options => {
				options.onProgress({ completedIterations: 250, totalIterations: 1000 } as never);
				return new Promise(resolve => (settleSecond = resolve));
			});
			settings.set({ range: 20, step: 10 });
			await openDialog();
			// The progress dialog is modal, so the button is found before it opens.
			const runAgain = runButton();

			await run();
			// The second of three steps, a quarter done: the bar reads across the whole run.
			expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-title"]')!.textContent).toBe('sidebar.buttons.stat_sweep.progress');
			expect(tracker()!.querySelector('[data-testid="progress-tracker-modal-progress-text"]')!.textContent).toBe('1250/3000');
			expect(tracker()!.querySelector('[data-testid="progress-tracker-time-remaining"]')!.textContent).toBe('sidebar.results.progress.time_remaining');
			expect(runAgain.disabled).toBe(true);

			abortType.mockClear();
			await act(async () => {
				fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
			});
			expect(abortType).toHaveBeenCalledExactlyOnceWith(RequestTypes.StatSweep);
			await act(async () => {
				settleSecond({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
			});

			expect(tracker()).toBeNull();
			expect(simulatePass).toHaveBeenCalledTimes(2);
			expect(saved().points).toHaveLength(1);
			expect(toast).not.toHaveBeenCalled();
			toast.mockRestore();
		});

		it("lists the run's stats, done up to the running one, which holds the spinner", async () => {
			let settleThird!: (result: object) => void;
			simulatePass
				.mockResolvedValueOnce(makeSweepSample(1000))
				.mockResolvedValueOnce(makeSweepSample(1000))
				.mockImplementationOnce(options => {
					options.onProgress({ completedIterations: 250, totalIterations: 1000 } as never);
					return new Promise(resolve => (settleThird = resolve));
				});
			settings.set({ stats: [AGILITY, STRENGTH], range: 10, step: 10 });
			await openDialog();

			await run();
			const rows = () => [...tracker()!.querySelectorAll('[data-testid="progress-step"]')];
			expect(rows().map(row => [row.textContent, row.getAttribute('data-state'), row.getAttribute('aria-current')])).toEqual([
				[STRENGTH.getFullName(Class.ClassWarrior), 'done', null],
				[AGILITY.getFullName(Class.ClassWarrior), 'running', 'step'],
			]);
			expect(tracker()!.querySelectorAll('[data-testid="progress-step-spinner"]')).toHaveLength(1);
			expect(rows()[1].querySelector('[data-testid="progress-step-spinner"]')).not.toBeNull();

			await act(async () => {
				fireEvent.click(tracker()!.querySelector('[data-testid="progress-tracker-modal-cancel-btn"]')!);
			});
			await act(async () => {
				settleThird({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
			});
			expect(tracker()).toBeNull();
		});
	});

	it('tells the user when another run stops the sweep, keeping the finished steps', async () => {
		const toast = vi.spyOn(toastManager, 'add');
		simulatePass.mockResolvedValueOnce(makeSweepSample(1000)).mockResolvedValueOnce({ type: ErrorOutcomeType.ErrorOutcomeAborted, message: '' });
		settings.set({ range: 20, step: 10 });
		await openDialog();

		await run();

		expect(toast).toHaveBeenCalledExactlyOnceWith({ variant: 'info', body: 'sidebar.buttons.stat_sweep.stopped' });
		expect(simulatePass).toHaveBeenCalledTimes(2);
		expect(saved().points).toHaveLength(1);
		toast.mockRestore();
	});

	it('disables Run with every stat removed', async () => {
		settings.set({ stats: [] });
		await openDialog();

		expect(runButton().disabled).toBe(true);
	});

	it('cancels a run when the dialog closes', async () => {
		simulatePass.mockImplementation(() => new Promise(() => {}));
		await openDialog();
		await run();
		abortType.mockClear();

		await act(async () => {
			fireEvent.click(rootElem.querySelector('[data-testid="stat-sweep-dialog"] [data-testid="sim-dialog-close"]')!);
		});

		expect(abortType).toHaveBeenCalledWith(RequestTypes.StatSweep);
	});

	it('cancels a run when Show power tools is turned off, which takes the dialog away', async () => {
		simulatePass.mockImplementation(() => new Promise(() => {}));
		await openDialog();
		await run();
		abortType.mockClear();

		await showPowerTools(false);

		expect(action()).toBeNull();
		expect(dialog()).toBeNull();
		expect(abortType).toHaveBeenCalledExactlyOnceWith(RequestTypes.StatSweep);
	});
});
