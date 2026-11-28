import { RaidSimResult } from '@generated/proto/api';
import { ResourceType } from '@generated/proto/spell';
import { describe, expect, it, vi } from 'vitest';

// Casts and hits are bucketed by the action's string, so the braced id in the log text stands in
// for the resolved action.
vi.mock('../action_id', () => {
	class ActionId {
		readonly spellId: number;
		readonly itemId: number;
		constructor(readonly logString: string) {
			this.spellId = parseInt(/SpellID: (\d+)/.exec(logString)?.[1] ?? '0');
			this.itemId = parseInt(/ItemID: (\d+)/.exec(logString)?.[1] ?? '0');
		}
		static fromLogString(logString: string) {
			return new ActionId(/\{.*?\}/.exec(logString)?.[0] ?? logString);
		}
		static async replaceAllInString(str: string) {
			return str;
		}
		async fill() {
			return this;
		}
		toString() {
			return this.logString;
		}
		toStringIgnoringTag() {
			return this.logString.replace(/, Tag: \d+/, '');
		}
	}
	return { ActionId };
});

import { buildCastLogs, buildResourceGroups } from './derive';
import { parseAll } from './parse';

const MISSILE = '{SpellID: 25346}';
const THORNS = '{SpellID: 9910}';

const cast = (time: number, spell = MISSILE) => [
	`[${time.toFixed(2)}] [Player (#1)] Casting ${spell} (Cost = 0.000, Cast Time = 0s, GCD = 0s, Effective Time = 0s)`,
	`[${time.toFixed(2)}] [Player (#1)] Completed cast ${spell}`,
];
const hit = (time: number, amount: number, spell = MISSILE) =>
	`[${time.toFixed(2)}] [Player (#1)] [Target 1] ${spell} Hit for ${amount.toFixed(3)} damage (SpellSchool: 64). (Threat: 0.000)`;

const castLogsOf = async (lines: Array<string>) => buildCastLogs(await parseAll(RaidSimResult.create({ logs: lines.join('\n') })));

// Lines as sim/core prints them for an instant cast and its hit.
describe('buildCastLogs', () => {
	// A Missile Barrage channel at 20 yards: a missile every 0.5 sec, each landing 1 sec after it fires.
	it('gives each missile in flight its own hit', async () => {
		const lines = [
			...cast(1.5),
			...cast(2.0),
			hit(2.5, 100),
			...cast(2.5),
			hit(3.0, 200),
			...cast(3.0),
			hit(3.5, 300),
			...cast(3.5),
			hit(4.0, 400),
			hit(4.5, 500),
		];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[100], [200], [300], [400], [500]]);
		expect(castLogs.map(log => log.travelTime)).toEqual([1, 1, 1, 1, 1]);
	});

	it('leaves the missile still flying at the end of the log without a hit', async () => {
		const lines = [...cast(1.5), ...cast(2.0), hit(2.5, 100), ...cast(2.5), hit(3.0, 200)];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[100], [200], []]);
	});

	it('pairs by window when two casts complete at one timestamp', async () => {
		// Thorns on two attackers: two casts at 4.48, both hits with them, and nothing in flight.
		const lines = [
			...cast(4.48, THORNS),
			...cast(4.48, THORNS),
			hit(4.48, 100, THORNS),
			hit(4.48, 200, THORNS),
			...cast(4.77, THORNS),
			hit(4.77, 300, THORNS),
		];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[], [100, 200], [300]]);
	});

	it('pairs by window when a cast deals nothing and the rest land as they complete', async () => {
		// Evenly spaced, so pairing by order would read each hit as the earlier cast's after a steady flight.
		const lines = [...cast(1.0), ...cast(2.0), hit(2.0, 200), ...cast(3.0), hit(3.0, 300)];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[], [200], [300]]);
		expect(castLogs.map(log => log.travelTime)).toEqual([0, 0, 0]);
	});

	it('keeps a hit that lands before the next cast on its own cast', async () => {
		const lines = [...cast(1.0), hit(1.2, 100), ...cast(2.0), hit(2.2, 200)];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[100], [200]]);
		expect(castLogs.map(log => log.travelTime.toFixed(2))).toEqual(['0.20', '0.20']);
	});

	it("puts Demonic Rune's hit on its user under the rune's use", async () => {
		const lines = [...cast(1.0, '{ItemID: 12662}'), hit(1.0, 1000, '{ItemID: 12662, Tag: 1}')];

		const castLogs = await castLogsOf(lines);

		expect(castLogs.map(log => log.damageDealtLogs.map(ddl => ddl.amount))).toEqual([[1000]]);
	});
});

// A regen line as mana.go logs it: the gain, then the stretch it covers and the 100 ms ticks that show it.
const regen = (time: number, amount: number, before: number, after: number, from: number, to: number, ticks: number, firstTick: number, tag = 2) =>
	`[${time.toFixed(2)}] [Player (#1)] Gained ${amount.toFixed(3)} mana from {OtherID: 2, Tag: ${tag}} (${before.toFixed(3)} --> ${after.toFixed(3)}) of 1000 total, ` +
	`regenerating from ${from.toFixed(6)}s to ${to.toFixed(6)}s in ${ticks} ticks from ${firstTick.toFixed(6)}s.`;
const spend = (time: number, amount: number, before: number) =>
	`[${time.toFixed(2)}] [Player (#1)] Spent ${amount.toFixed(3)} mana from {SpellID: 25346} (${before.toFixed(3)} --> ${(before - amount).toFixed(3)}) of 1000 total.`;

const manaStepsOf = async (lines: Array<string>) =>
	buildResourceGroups(await parseAll(RaidSimResult.create({ logs: lines.join('\n') })))[ResourceType.ResourceTypeMana].map(group => [
		group.timestamp,
		Number(group.valueAfter.toFixed(3)),
	]);

describe('buildResourceGroups', () => {
	// A read at 1.25 credits the regen up to the tick at 1.2: 2 mana per second up to the end of the
	// five-second rule at 1.05, 100 per second after it, which max mana stops at 1000.
	it('gives a regen stretch a step per tick, split at the five-second rule and capped at max mana', async () => {
		const lines = [regen(1.25, 1.5, 990, 991.5, 0.3, 1.05, 8, 0.4, 1), regen(1.25, 15, 991.5, 1000, 1.05, 1.2, 2, 1.1), spend(1.25, 100, 1000)];

		expect(await manaStepsOf(lines)).toEqual([
			[0.4, 990.2],
			[0.5, 990.4],
			[0.6, 990.6],
			[0.7, 990.8],
			[0.8, 991],
			[0.9, 991.2],
			[1, 991.4],
			[1.1, 996.5],
			[1.2, 1000],
			[1.25, 900],
		]);
	});

	// A rate change at 1.37 credits the regen up to that moment, past the tick at 1.3.
	it('puts the last step of a stretch that ends between two ticks at the line', async () => {
		expect(await manaStepsOf([regen(1.37, 17, 900, 917, 1.2, 1.37, 2, 1.3)])).toEqual([
			[1.3, 910],
			[1.37, 917],
		]);
	});

	// An energy read at 1.3 credits the three ticks from 1.05 to 1.25, of which max energy lets two through.
	it('gives an energy stretch a step per tick, capped at max energy', async () => {
		const line =
			'[1.30] [Player (#1)] Gained 3.000 energy from {OtherID: 5} (98.000 --> 100.000) of 100 total, ' +
			'regenerating from 0.950000s to 1.250000s in 3 ticks from 1.050000s.';
		const groups = buildResourceGroups(await parseAll(RaidSimResult.create({ logs: line })))[ResourceType.ResourceTypeEnergy];

		expect(groups.map(group => [group.timestamp, Number(group.valueAfter.toFixed(3))])).toEqual([
			[1.05, 99],
			[1.15, 100],
			[1.25, 100],
		]);
	});
});
