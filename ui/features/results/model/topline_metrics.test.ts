import { PlayerSpecs } from '@sim/player/specs/index';
import { DANGER_TEXT } from '@ui-kit/utils/colors';
import { describe, expect, it, vi } from 'vitest';

import { showsOutOfMana, toplineResultMetrics } from './topline_metrics';

vi.mock('@sim/proto/sim_result', async importOriginal => {
	const actual = await importOriginal<typeof import('@sim/proto/sim_result')>();
	const sum = (actions: Array<{ dps: number; tps: number }>, key: 'dps' | 'tps') => actions.reduce((total, action) => total + action[key], 0);
	return {
		...actual,
		ActionMetrics: {
			merge: (actions: Array<{ dps: number; tps: number }>) => ({ dps: sum(actions, 'dps'), tps: sum(actions, 'tps') }),
		},
	};
});

const dist = (avg: number, stdev = 0) => ({ avg, stdev });

const player = (overrides: Record<string, unknown> = {}) => ({
	spec: PlayerSpecs.Mage,
	unitIndex: 0,
	getTargetIndex: () => null,
	getPlayerAndPetActions: () => [],
	getHealingActionsOn: () => [],
	dps: dist(1000, 10),
	tps: dist(500, 5),
	dtps: dist(50, 1),
	tmi: dist(2, 0.5),
	hps: dist(20, 2),
	tto: dist(30, 3),
	chanceOfDeath: dist(1, 0.1),
	secondsOomAvg: 0,
	actionsPerMinute: () => 42,
	...overrides,
});

const simResult = (players: Array<ReturnType<typeof player>>, overrides: Record<string, unknown> = {}) =>
	({
		getRaidIndexedPlayers: () => players,
		getRaidIndexedActionMetrics: () => [],
		getTargets: () => [],
		request: { encounter: { useHealth: false } },
		result: { avgIterationDuration: 300 },
		encounterMetrics: { durationSeconds: 300 },
		...overrides,
	}) as never;

const keys = (result: never, options?: { showOutOfMana?: boolean }) => toplineResultMetrics(result, undefined, options).map(metric => metric.metric);

describe('toplineResultMetrics', () => {
	it('lists the whole-encounter metrics for a non-tank, with tto and hps last', () => {
		expect(keys(simResult([player()]))).toEqual(['dps', 'tps', 'apm', 'dtps', 'tmi', 'cod', 'tto', 'hps']);
	});

	it('drops tto and hps for a tank spec', () => {
		expect(keys(simResult([player({ spec: PlayerSpecs.FeralBearDruid })]))).toEqual(['dps', 'tps', 'apm', 'dtps', 'tmi', 'cod']);
	});

	// The tank list is the bear and the protection paladin only, so a protection warrior keeps the
	// mana rows a tank would otherwise drop.
	it('keeps tto and hps for a protection warrior', () => {
		expect(keys(simResult([player({ spec: PlayerSpecs.ProtectionWarrior })]))).toEqual(['dps', 'tps', 'apm', 'dtps', 'tmi', 'cod', 'tto', 'hps']);
	});

	it('suppresses chance of death for a protection paladin', () => {
		expect(keys(simResult([player({ spec: PlayerSpecs.ProtectionPaladin })]))).toEqual(['dps', 'tps', 'apm', 'dtps', 'tmi']);
	});

	it('carries the metric, its label and its unit', () => {
		const metrics = toplineResultMetrics(simResult([player()]));

		expect(metrics[0]).toMatchObject({ metric: 'dps', name: 'sidebar.results.metrics.dps.label', average: 1000, stdev: 10 });
		expect(metrics.find(metric => metric.metric === 'tmi')).toMatchObject({ unit: 'percentage' });
	});

	it('reads actions per minute for the filtered target', () => {
		const filter = { target: 0 };
		const actionsPerMinute = vi.fn(() => 37.5);
		const metrics = toplineResultMetrics(simResult([player({ actionsPerMinute })]), filter);

		expect(actionsPerMinute).toHaveBeenCalledWith(filter);
		expect(metrics.find(metric => metric.metric === 'apm')).toMatchObject({ average: 37.5, stdev: undefined });
	});

	it('keeps apm whatever the display metrics are', () => {
		const displayMetrics = { damage: false, threat: false, healing: false };

		expect(toplineResultMetrics(simResult([player()]), undefined, { displayMetrics }).map(metric => metric.metric)).toEqual(['tps', 'apm']);
	});

	it('merges the filtered actions when a target is selected', () => {
		const result = simResult([player({ getTargetIndex: () => 0 })], {
			getRaidIndexedActionMetrics: () => [{ dps: 400, tps: 200 }],
			getTargets: () => [{ actions: [{ forTarget: () => ({ dps: 30, tps: 0 }) }] }],
		});

		expect(keys(result)).toEqual(['dps', 'tps', 'apm', 'dtps', 'tto', 'hps']);
		expect(toplineResultMetrics(result)[0]).toMatchObject({ metric: 'dps', average: 400, stdev: undefined });
	});

	it('keeps tps while the other threat metrics follow the setting', () => {
		const result = simResult([player()]);
		const displayMetrics = { damage: true, threat: false, healing: true };

		expect(toplineResultMetrics(result, undefined, { displayMetrics }).map(metric => metric.metric)).toEqual(['dps', 'tps', 'apm', 'tto', 'hps']);
	});

	it('keeps HPS variance and tps for a healer', () => {
		const result = simResult([player({ spec: PlayerSpecs.HealerPriest, getHealingActionsOn: () => [] })]);
		const metrics = toplineResultMetrics(result, undefined, { displayMetrics: { damage: false, threat: false, healing: true } });
		expect(metrics.map(metric => metric.metric)).toContain('tps');
		expect(metrics[0]).toMatchObject({ metric: 'hps', average: 20, stdev: 2 });
	});

	it('omits the merged columns a filtered encounter has no actions for', () => {
		expect(keys(simResult([player({ getTargetIndex: () => 0 })]))).toEqual(['tto', 'hps']);
	});

	it('appends dur only when the encounter uses health', () => {
		const result = simResult([player()], { request: { encounter: { useHealth: true } } });

		expect(keys(result).at(-1)).toBe('dur');
		expect(toplineResultMetrics(result).at(-1)).toMatchObject({ metric: 'dur', average: 300, unit: 'seconds' });
	});

	it.each([
		[1, 'safe'],
		[10, 'warning'],
		[100, 'danger'],
	] as const)('grades %d seconds out of mana as %s', (secondsOomAvg, danger) => {
		const metrics = toplineResultMetrics(simResult([player({ secondsOomAvg })]), undefined, { showOutOfMana: true });
		const dangerClass = DANGER_TEXT[danger];

		expect(metrics.at(-1)).toMatchObject({ metric: 'oom', extraClass: dangerClass, unit: 'seconds' });
	});

	it('leaves out-of-mana out unless it is asked for', () => {
		expect(keys(simResult([player({ secondsOomAvg: 100 })]))).not.toContain('oom');
	});
});

describe('showsOutOfMana', () => {
	it.each([
		['on for a lone mana user', [player()], true],
		['off for a class with no mana bar', [player({ spec: PlayerSpecs.DpsWarrior })], false],
		['off for a raid', [player(), player()], false],
		['off for a player with no spec', [player({ spec: null })], false],
	])('is %s', (_name, players, shows) => {
		expect(showsOutOfMana(simResult(players))).toBe(shows);
	});
});
