import type { SimResultData } from '@features/results/model/result_data';
import { ActionMetrics } from '@sim/proto/sim_result';
import { vi } from 'vitest';

const NONE = { value: 0, percentage: 0, average: 0 };

/**
 * An `ActionMetrics` stand-in carrying every field the damage, damage-taken, threat and healing
 * tables read. `forTarget` returns the metric itself, so a filtered read sees the same numbers.
 */
export const actionMetric = (name: string, overrides: Record<string, unknown> = {}): ActionMetrics => {
	const base: Record<string, unknown> = {
		name,
		actionId: { toStringIgnoringTag: () => name, toString: () => name },
		unit: { isPet: false, petActionId: null },
		spellSchool: null,
		casts: 10,
		isPassiveAction: false,
		landedHits: 10,
		landedTicks: 0,
		hitAttempts: 11,
		critPercent: 20,
		blockedCritPercent: 0,
		critTickPercent: 0,
		totalMisses: 1,
		totalMissesPercent: 10,
		misses: 1,
		missPercent: 10,
		parries: 0,
		parryPercent: 0,
		dodges: 0,
		dodgePercent: 0,
		hits: 8,
		crits: 2,
		glances: 0,
		blocks: 0,
		blockedCrits: 0,
		blockPercent: 0,
		crushes: 0,
		crushPercent: 0,
		resistedHits: 0,
		resistedCrits: 0,
		resistedTicks: 0,
		resistedCritTicks: 0,
		ticks: 0,
		critTicks: 0,
		damage: 100,
		avgDamage: 100,
		totalDamagePercent: 50,
		damageThroughput: 20,
		avgCastHit: 10,
		avgCastTick: 0,
		avgHit: 10,
		avgTick: 0,
		dps: 10,
		damageDone: {
			hit: { value: 80, percentage: 80, average: 10 },
			critHit: { value: 20, percentage: 20, average: 10 },
			resistedHit: NONE,
			resistedCritHit: NONE,
			tick: NONE,
			resistedTick: NONE,
			critTick: NONE,
			resistedCritTick: NONE,
			glance: NONE,
			block: NONE,
			blockedCrit: NONE,
			crush: NONE,
		},
		threat: 1000,
		avgThreat: 1000,
		totalThreatPercent: 50,
		threatThroughput: 40,
		avgCastThreat: 5,
		avgHitThreat: 5,
		tps: 5,
		healing: 100,
		avgHealing: 100,
		avgCritHealing: 20,
		healingPercent: 80,
		healingCritPercent: 20,
		shielding: 40,
		totalHealingPercent: 50,
		castsPerMinute: 20,
		avgCastTimeMs: 1500,
		avgCastHealing: 10,
		avgHitHealing: 10,
		hpm: 3,
		healingThroughput: 40,
		hps: 10,
		...overrides,
	};
	base.forTarget = () => base;
	return base as unknown as ActionMetrics;
};

export const petActionMetric = (name: string, overrides: Record<string, unknown> = {}) =>
	actionMetric(name, { unit: { isPet: true, petActionId: null }, ...overrides });

/**
 * `ActionMetrics.merge` as the first metric with the named fields summed. The real one builds an
 * ActionMetrics out of protos; a parent row only needs to be a row.
 */
export const mockActionMetricsMerge = (...summed: Array<string>) =>
	vi.spyOn(ActionMetrics, 'merge').mockImplementation(((metrics: Array<Record<string, number>>) => ({
		...metrics[0],
		...Object.fromEntries(summed.map(field => [field, metrics.reduce((sum, entry) => sum + entry[field], 0)])),
	})) as unknown as typeof ActionMetrics.merge);

type ActionsGetter = 'getDamageActions' | 'getThreatActions' | 'getHealingActions';

export const playerMetricsResult =
	(getter: ActionsGetter) =>
	(actions: Array<ActionMetrics>, petActions: Array<ActionMetrics> = []) =>
		({
			filter: {},
			result: {
				getRaidIndexedPlayers: () => [{ [getter]: () => actions, pets: petActions.length ? [{ name: 'Wolf', [getter]: () => petActions }] : [] }],
			},
		}) as unknown as SimResultData;

export const metricRows = (container: HTMLElement) => [...container.querySelectorAll<HTMLTableRowElement>('tbody tr')];

export const metricsTooltipTable = () => document.querySelector('.sim-tooltip table[data-testid="metrics-table"]');

export const metricsTooltipRowNames = () => [...metricsTooltipTable()!.querySelectorAll<HTMLTableRowElement>('tbody tr')].map(row => row.cells[0].textContent);
