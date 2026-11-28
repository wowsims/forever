import { RaidSimResult } from '@generated/proto/api';
import { CreatureFamily, MobType, Stat } from '@generated/proto/common';
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
import type { CastBeganLog, CastFailedLog, CastPushbackLog, DamageLog, MovementLog, ParsedLog, SpellQueuedLog, StatChangeLog } from './types';

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

	// flags.go's PartialResistString appends the resist, or the Spell Piercing bonus as a negative resist, to a
	// glance, crit or hit.
	it.each([
		['50% Resist', 50],
		['75% Resist', 75],
		['10% Pierced', -10],
	])('reads (%s) off a hit', async (token, resist) => {
		const log = (await parseLine(
			`[5.00] [Player (#1)] [Target 1] {SpellID: 25368} Hit (${token}) for 1234.500 damage (SpellSchool: 4). (Threat: 100.000)`,
		)) as DamageLog;
		expect(log.kind).toBe('damage');
		expect(log.outcome).toBe('hit');
		expect(log.amount).toBe(1234.5);
		expect(log.resist).toBe(resist);
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

	it('reads a pet walk and names no action for it', async () => {
		// movement.go's registerMovementAction.
		const log = (await parseLine('[-0.50] [Player (#1) - Cat] Walking from 12.00 to 4.00 yards over 0.555s')) as MovementLog;
		expect(log.kind).toBe('movement');
		expect(log.timestamp).toBe(-0.5);
		expect(log.source?.isPet).toBe(true);
		expect(log.source?.name).toBe('Cat');
		expect([log.fromYards, log.toYards, log.duration]).toEqual([12, 4, 0.555]);
		expect(log.actionId).toBeNull();
	});

	it('reads a stop as a walk of no duration on the spot', async () => {
		// movement.go's logStop.
		const log = (await parseLine('[-17.14] [Player (#1)] Stopped walking at 25.00 yards')) as MovementLog;
		expect(log.kind).toBe('movement');
		expect([log.fromYards, log.toYards, log.duration]).toEqual([25, 25, 0]);
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
	// attack_table_helpers.go's versusLogTag: the tag follows "(Threat: X)" at the end of the line.
	describe('the versus tag', () => {
		const hit = '[0.01] [Consumer (#1)] [Target 1] {OtherID: 3, Tag: 14} Hit for 157.393 damage (SpellSchool: 1). (Threat: 157.393)';
		const sourceIds = (log: DamageLog) => log.versus?.sources.map(source => String(source.actionId));

		it('reads the mob type and every source of a direct hit', async () => {
			const log = (await parseLine(`${hit} [Versus Beast: {SpellID: 20557} x1.05; {SpellID: 13650} +6 PhysicalDamage]`)) as DamageLog;
			expect(log.amount).toBe(157.393);
			expect(log.threat).toBe(157.393);
			expect(log.versus?.mobType).toBe(MobType.MobTypeBeast);
			expect(log.versus?.creatureFamily).toBeUndefined();
			expect(sourceIds(log)).toEqual(['{SpellID: 20557}', '{SpellID: 13650}']);
			expect(log.versus?.sources.map(source => source.effect)).toEqual([
				{ kind: 'damage', multiplier: 1.05 },
				{ kind: 'stat', stat: Stat.StatPhysicalDamage, amount: 6 },
			]);
		});

		it('reads the tag of a tick', async () => {
			const log = (await parseLine(
				'[2.00] [Player (#1)] [Target 1] {SpellID: 1943} tick Hit for 80.000 damage (SpellSchool: 1). (Threat: 80.000) [Versus Undead: {ItemID: 12345} -3 AttackPower]',
			)) as DamageLog;
			expect(log.tick).toBe(true);
			expect(log.versus?.mobType).toBe(MobType.MobTypeUndead);
			expect(log.versus?.sources[0].effect).toEqual({ kind: 'stat', stat: Stat.StatAttackPower, amount: -3 });
		});

		it('reads the creature family and the crit and base entries', async () => {
			const log = (await parseLine(
				`${hit.replace('Hit', 'Crit')} [Versus Beast (Wolf): {ItemID: 12345} x1.5 crit; {SpellID: 17670} x2 base]`,
			)) as DamageLog;
			expect(log.outcome).toBe('crit');
			expect(log.versus?.creatureFamily).toBe(CreatureFamily.CreatureFamilyWolf);
			expect(log.versus?.sources.map(source => source.effect)).toEqual([
				{ kind: 'crit', multiplier: 1.5 },
				{ kind: 'base', multiplier: 2 },
			]);
		});

		it('reads the crit chance and proc entries', async () => {
			const log = (await parseLine(
				`${hit} [Versus Humanoid (Orc): {ItemID: 13965} +10% crit chance; {ItemID: 11684} x2 proc chance; {ItemID: 10760} proc only]`,
			)) as DamageLog;
			expect(log.versus?.creatureFamily).toBe(CreatureFamily.CreatureFamilyOrc);
			expect(sourceIds(log)).toEqual(['{ItemID: 13965}', '{ItemID: 11684}', '{ItemID: 10760}']);
			expect(log.versus?.sources.map(source => source.effect)).toEqual([
				{ kind: 'crit_chance', amount: 10 },
				{ kind: 'proc_chance', multiplier: 2 },
				{ kind: 'proc_only' },
			]);
		});

		it('leaves a line without the tag with no versus', async () => {
			expect(((await parseLine(hit)) as DamageLog).versus).toBeUndefined();
		});

		it('skips an entry it cannot read and keeps the rest', async () => {
			const log = (await parseLine(`${hit} [Versus Beast: {SpellID: 1} +5 NotAStat; {SpellID: 2} x1.1; garbage]`)) as DamageLog;
			expect(sourceIds(log)).toEqual(['{SpellID: 2}']);
		});

		it('drops a tag whose mob type it does not know', async () => {
			expect(((await parseLine(`${hit} [Versus Nonsense: {SpellID: 2} x1.1]`)) as DamageLog).versus).toBeUndefined();
		});
	});
});
