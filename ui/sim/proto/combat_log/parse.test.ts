import { RaidSimResult } from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

// Resolving an ActionId goes to the item/spell database; every assertion here is about the text
// around the id, so the log string stands in for the resolved action.
vi.mock('../action_id', () => {
	class ActionId {
		constructor(readonly logString: string) {}
		static fromLogString(logString: string) {
			return new ActionId(logString);
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
	}
	return { ActionId };
});

import { parseAll } from './parse';
import type { CastBeganLog, CastFailedLog, CastPushbackLog, DamageLog, ParsedLog, SpellQueuedLog, StatChangeLog } from './types';

const parseLine = async (line: string): Promise<ParsedLog> => (await parseAll(RaidSimResult.create({ logs: line })))[0];

// Lines as sim/core prints them: sim.go:304 stamps the time, unit.go:245 the entity label.
describe('parseAll on sim log lines', () => {
	it('reads the GCD out of a cast line', async () => {
		// cast.go:182 - five call sites, all with the GCD clause MoP's sim does not print.
		const log = (await parseLine(
			'[0.00] [Player (#1)] Casting {SpellID: 11605} (Cost = 465.000, Cast Time = 2.5s, GCD = 1.5s, Effective Time = 2.5s)',
		)) as CastBeganLog;
		expect(log.kind).toBe('cast-began');
		expect(log.manaCost).toBe(465);
		expect(log.castTime).toBe(2.5);
		expect(log.gcd).toBe(1.5);
		expect(log.effectiveTime).toBe(2.5);
	});

	it.each([
		['BlockedCrit', 'critical-block', 1234.5],
		['Crush', 'crush', 2000],
	])('reads %s as %s', async (token, outcome, amount) => {
		const log = (await parseLine(
			`[6.00] [Target 1] [Player (#1)] {SpellID: 12345} ${token} for ${amount.toFixed(3)} damage (SpellSchool: 1). (Threat: 0.000)`,
		)) as DamageLog;
		expect(log.kind).toBe('damage');
		expect(log.outcome).toBe(outcome);
		expect(log.amount).toBe(amount);
		expect(log.effect).toBe('damage');
		expect(log.resist).toBe(0);
	});

	it('reads a partial resist off a hit', async () => {
		// flags.go:156 appends the resist to a glance, crit or hit.
		const log = (await parseLine(
			'[5.00] [Player (#1)] [Target 1] {SpellID: 25368} Hit (50% Resist) for 1234.500 damage (SpellSchool: 4). (Threat: 100.000)',
		)) as DamageLog;
		expect(log.kind).toBe('damage');
		expect(log.outcome).toBe('hit');
		expect(log.amount).toBe(1234.5);
		expect(log.resist).toBe(50);
	});

	it('reads an auto attack delay', async () => {
		// attack.go:528.
		const log = await parseLine('[1.00] [Player (#1)] {SpellID: 6603} delayed by 250ms, was ready at 1.5s');
		expect(log.kind).toBe('auto-delay');
	});

	it('reads a queued spell', async () => {
		// spell_queueing.go:66.
		const log = (await parseLine('[2.00] [Player (#1)] Queueing up {SpellID: 11605} to cast at 2.5s.')) as SpellQueuedLog;
		expect(log.kind).toBe('spell-queued');
		expect(log.fireAt).toBe(2.5);
	});

	it('reads a cast failure and drops the redundant curTime', async () => {
		// cast.go:83.
		const log = (await parseLine('[3.00] [Player (#1)] {SpellID: 11605} failed to cast: not enough mana, curTime = 3s')) as CastFailedLog;
		expect(log.kind).toBe('cast-failed');
		expect(log.reason).toBe('not enough mana');
	});

	it('reads a cast pushback', async () => {
		// character.go:455.
		const log = (await parseLine('[4.00] [Player (#1)] {SpellID: 11605} pushed back 500ms while casting')) as CastPushbackLog;
		expect(log.kind).toBe('cast-pushback');
		expect(log.pushback).toBe(0.5);
		expect(log.isChanneling).toBe(false);
	});

	it.each([
		// aura_helpers.go, with the stats from Stats.FlatString().
		['[1.00] [Player (#1)] Gained {"Strength": 20.000,"CritRating": 14.000,} from {SpellID: 1234}.', true, '{SpellID: 1234}.'],
		['[1.00] [Player (#1)] Lost {"Strength": 20.000,"CritRating": 14.000,} from fading {ItemID: 5678}.', false, '{ItemID: 5678}.'],
	])('reads a temporary stat line: %s', async (line, isGain, actionId) => {
		const log = (await parseLine(line)) as StatChangeLog;
		expect(log.kind).toBe('stat-change');
		expect(log.isGain).toBe(isGain);
		expect(log.stats).toBe('{"Strength": 20.000,"CritRating": 14.000,}');
		expect(log.actionIdAsString).toBe(actionId);
	});

	it.each([
		// unit.go, with the parenthesis from StatDependency.String() and the stats from Stats.FlatString().
		'[1.00] [Player (#1)] Dynamic stat change: {"DodgeRating": 12.000,"DodgePercent": 1.000,}',
		'[1.00] [Player (#1)] Dynamic dep enabled (DodgeRating += Intellect * 0.50): {"DodgeRating": 12.000,"DodgePercent": 1.000,}',
	])('keeps a stat line as the sim wrote it for the exporter and search: %s', async line => {
		const log = await parseLine(line);
		expect(log.kind).toBe('plain');
		expect(log.raw).toBe(line);
	});
});
