import type { SimResultData } from '@features/results/model/result_data';
import { ActionMetrics as ActionMetricsProto, TargetedActionMetrics as TargetedActionMetricsProto } from '@generated/proto/api';
import { ActionID } from '@generated/proto/common';
import { IconData } from '@generated/proto/ui';
import { Database } from '@sim/proto/database';
import { ActionMetrics, UnitMetrics } from '@sim/proto/sim_result';
import { makeSimRun } from '@testing/storybook/fixtures/results';
import { vi } from 'vitest';

const NONE = { value: 0, percentage: 0, average: 0 };

/**
 * An `ActionMetrics` stand-in carrying every field the damage, damage-taken, threat, healing and cast
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
		manualCasts: 10,
		landedHits: 10,
		landedTicks: 0,
		landedRolls: 10,
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
			hit: { value: 80, percentage: 80, average: 10, min: 8, max: 12 },
			critHit: { value: 20, percentage: 20, average: 10, min: 16, max: 24 },
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
		healingDone: {
			heal: { value: 80, percentage: 80, average: 10, min: 8, max: 12 },
			critHeal: { value: 20, percentage: 20, average: 20, min: 16, max: 24 },
			tick: NONE,
			critTick: NONE,
			shield: NONE,
		},
		customMetrics: [],
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
		manualCastsPerMinute: 20,
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

/** `playerMetricsResult` for the cast table, which reads the player's and the pet's `actions` lists and the player's real `apmActions`. */
export const castMetricsResult = (actions: Array<ActionMetrics>, petActions: Array<ActionMetrics> = []) =>
	({
		filter: {},
		result: {
			getRaidIndexedPlayers: () => [
				{ actions, apmActions: UnitMetrics.prototype.apmActions, pets: petActions.length ? [{ name: 'Wolf', actions: petActions }] : [] },
			],
		},
	}) as unknown as SimResultData;

export const metricRows = (container: HTMLElement) => [...container.querySelectorAll<HTMLTableRowElement>('tbody tr')];

export const metricsTooltipTable = () => document.querySelector('.sim-tooltip table[data-testid="metrics-table"]');

export const metricsTooltipRowNames = () => [...metricsTooltipTable()!.querySelectorAll<HTMLTableRowElement>('tbody tr')].map(row => row.cells[0].textContent);

export interface SpellActionFixture {
	spellId: number;
	tag?: number;
	damage: number;
}

/** `makeSimRun` with the warrior's auto attack swapped for one spell action per entry, each with the auto attack's outcomes on the target. */
export const spellActionsRun = (spells: Array<SpellActionFixture>) => {
	const run = makeSimRun();
	const player = run.result.raidMetrics!.parties[0].players[0];
	const template = player.actions[0];
	player.actions = spells.map(({ spellId, tag = 0, damage }) =>
		ActionMetricsProto.create({
			...template,
			isMelee: false,
			id: ActionID.create({ rawId: { oneofKind: 'spellId', spellId }, tag }),
			targets: [TargetedActionMetricsProto.create({ ...template.targets[0], damage })],
		}),
	);
	return run;
};

/** Answers the spell lookups `ActionId.fill` makes with the given name and rank per spell id. */
export const stubSpellIcons = (spells: Record<number, { name: string; rank?: number }>) =>
	vi
		.spyOn(Database, 'getSpellIconData')
		.mockImplementation(spellId => Promise.resolve(IconData.create({ id: spellId, icon: 'spell_fire_flamebolt', ...spells[spellId] })));

/** Two ranks of one spell and a spell with no rank, for the rank grouping tests. */
export const RANKED_SPELLS = { FIREBALL_1: 133, FIREBALL_2: 143, FROSTBOLT: 116 };

export const stubRankedSpellIcons = () =>
	stubSpellIcons({
		[RANKED_SPELLS.FIREBALL_1]: { name: 'Fireball', rank: 1 },
		[RANKED_SPELLS.FIREBALL_2]: { name: 'Fireball', rank: 2 },
		[RANKED_SPELLS.FROSTBOLT]: { name: 'Frostbolt' },
	});
