import { RaidSimResult } from '@generated/proto/api';
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

import { buildCastLogs } from './derive';
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
