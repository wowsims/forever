import { AuraStats as AuraStatsProto, SpellStats as SpellStatsProto, UnitMetadata as UnitMetadataProto } from '@generated/proto/api';
import { Debuffs } from '@generated/proto/buffs';
import { ActionID as ActionIdProto, Class, ConsumesSpec, ItemRandomSuffix, PseudoStat, Spec, Stat, WeaponType } from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ActionId } from '../proto/action_id';
import { Stats } from '../proto/stats';
import { adjustImbues, adjustSwapImbues } from '../proto/utils';
import { Player, UnitMetadata } from './player';

// getDebuffStats reads nothing but the raid's debuffs, so the method is called against that one
// dependency rather than against a constructed Player, which needs a whole sim behind it.
const debuffStats = (debuffs: Partial<Debuffs>): Stats =>
	Player.prototype.getDebuffStats.call({
		sim: { raid: { getDebuffs: () => Debuffs.create(debuffs) } },
	} as unknown as Player<any>);

describe('Player.setConsumes', () => {
	it('matches main-hand and item-swap imbues to their weapons', () => {
		const sword = Item.create({ weaponType: WeaponType.WeaponTypeSword });
		const mace = Item.create({ weaponType: WeaponType.WeaponTypeMace });
		let storedConsumes = ConsumesSpec.create();
		const player = {
			slice: () => ({ consumables: storedConsumes }),
			getGear: () => ({ adjustImbues: (consumes: ConsumesSpec) => adjustImbues(consumes, sword, mace) }),
			itemSwapSettings: {
				getGear: () => ({ adjustSwapImbues: (consumes: ConsumesSpec) => adjustSwapImbues(consumes, mace, sword) }),
			},
			patch: (_field: string, consumes: ConsumesSpec) => {
				storedConsumes = consumes;
			},
		} as unknown as Player<any>;

		Player.prototype.setConsumes.call(
			player,
			ConsumesSpec.create({
				mhImbueId: 12643,
				ohImbueId: 12404,
				swapMhImbueId: 12404,
				swapOhImbueId: 12643,
			}),
		);

		expect([storedConsumes.mhImbueId, storedConsumes.ohImbueId, storedConsumes.swapMhImbueId, storedConsumes.swapOhImbueId]).toEqual([
			12404, 12643, 12643, 12404,
		]);
	});
});

describe('Player.getDebuffStats', () => {
	// Judgement of the Crusader raises Holy damage taken and nothing else: Improved Seal of the
	// Crusader, the crit half, has no node in any Forever trait tree, so the sheet credits no crit.
	// Faerie Fire stands for a debuff that moves no stat on the sheet at all.
	it.each<[string, Partial<Debuffs>]>([
		['Judgement of the Crusader', { judgementOfTheCrusader: true }],
		['Faerie Fire', { faerieFire: true }],
	])('credits %s with no crit', (_name, debuffs) => {
		const stats = debuffStats(debuffs);

		expect(stats.getPseudoStat(PseudoStat.PseudoStatMeleeCritPercent)).toBe(0);
		expect(stats.getPseudoStat(PseudoStat.PseudoStatRangedCritPercent)).toBe(0);
		expect(stats.getPseudoStat(PseudoStat.PseudoStatSpellCritPercent)).toBe(0);
	});

	// The number has to be the one HuntersMarkValue states in sim/core/debuffs_auto_gen.go,
	// or the sheet and the simulation disagree about the same debuff.
	it("credits Hunter's Mark with the ranged attack power spell 14325 states", () => {
		expect(debuffStats({ huntersMark: true }).getStat(Stat.StatRangedAttackPower)).toBe(71);
		expect(debuffStats({ huntersMark: false }).getStat(Stat.StatRangedAttackPower)).toBe(0);
	});
});

// The main-hand table the sim sends against a level 63 target, in percent.
interface SheetTable {
	miss: number;
	dodge: number;
	parry: number;
	glance: number;
	suppression: number;
}
const LEVEL_63_AT_300: SheetTable = { miss: 8, dodge: 6.5, parry: 14, glance: 40, suppression: 3 };
const LEVEL_63_AT_285: SheetTable = { miss: 11, dodge: 8, parry: 23, glance: 70, suppression: 6 };

// A rogue with one weapon and no hit, expertise or crit, behind the target unless placed in front.
const critCapInfo = (table: SheetTable, { inFront = false, dualWield = false, playerClass = Class.ClassRogue } = {}) => {
	const pseudoStats: Array<number> = [];
	pseudoStats[PseudoStat.PseudoStatMainHandMissPercent] = table.miss;
	pseudoStats[PseudoStat.PseudoStatMainHandDodgePercent] = table.dodge;
	pseudoStats[PseudoStat.PseudoStatMainHandParryPercent] = table.parry;
	pseudoStats[PseudoStat.PseudoStatMainHandGlancePercent] = table.glance;
	pseudoStats[PseudoStat.PseudoStatMeleeCritSuppressionPercent] = table.suppression;
	return Player.prototype.getMeleeCritCapInfo.call({
		slice: () => ({ currentStats: { finalStats: { pseudoStats } } }),
		getDebuffStats: () => new Stats(),
		isDualWielding: () => dualWield,
		getClass: () => playerClass,
		getInFrontOfTarget: () => inFront,
		getSpec: () => Spec.SpecRogue,
	} as unknown as Player<any>);
};

describe('Player.getMeleeCritCapInfo', () => {
	it.each<[number, SheetTable]>([
		[300, LEVEL_63_AT_300],
		[285, LEVEL_63_AT_285],
	])('with %i main-hand weapon skill against a level 63 target, caps hit, dodge and crit on the table the sim sends', (_skill, table) => {
		const info = critCapInfo(table);

		expect(info.meleeHitCap).toBeCloseTo(table.miss);
		expect(info.expertiseCap).toBeCloseTo(table.dodge);
		expect(info.glancing).toBeCloseTo(table.glance);
		expect(info.suppression).toBeCloseTo(table.suppression);
		expect(info.baseCritCap).toBeCloseTo(100 - table.glance + table.suppression - table.miss - table.dodge);
	});

	it.each<[number, SheetTable]>([
		[300, LEVEL_63_AT_300],
		[285, LEVEL_63_AT_285],
	])('with %i main-hand weapon skill, adds the parry of a level 63 target to the expertise cap from in front', (_skill, table) => {
		expect(critCapInfo(table, { inFront: true }).expertiseCap).toBeCloseTo(table.dodge + table.parry);
	});

	it('adds the dual-wield miss to the hit cap, except for a warrior', () => {
		expect(critCapInfo(LEVEL_63_AT_300, { dualWield: true }).meleeHitCap).toBeCloseTo(8 + 19);
	});
});

describe('Player.getRandomSuffixes', () => {
	// A spec whose default EP weights are all zero rates every suffix at 0 EP.
	const suffixesOf = (options: Array<number>) =>
		Player.prototype.getRandomSuffixes.call(
			{
				sim: { db: { getRandomSuffixById: (id: number) => (id < 100 ? ItemRandomSuffix.create({ id }) : undefined) } },
				computeRandomSuffixEP: () => 0,
			} as unknown as Player<any>,
			Item.create({ randomSuffixOptions: options }),
		);

	it('lists every suffix the database knows, whatever its EP', () => {
		expect(suffixesOf([1, 2, 3]).map(suffix => suffix.id)).toEqual([1, 2, 3]);
	});

	it('skips an option missing from the database', () => {
		expect(suffixesOf([1, 100, 2]).map(suffix => suffix.id)).toEqual([1, 2]);
	});
});

describe('UnitMetadata ranks', () => {
	// 2120 and 2121 are two ranks of Flamestrike, 25304 the only rank of Frostbolt the sim registers.
	const names: Record<number, string> = { 2120: 'Flamestrike', 2121: 'Flamestrike', 25304: 'Frostbolt' };
	beforeEach(() => {
		vi.spyOn(ActionId, 'getTooltipData').mockImplementation(async id => ({ id: id.spellId, name: names[id.spellId], icon: '' }) as never);
	});

	const actionId = (spellId: number, rank: number, tag = 0) => ActionIdProto.create({ rawId: { oneofKind: 'spellId', spellId }, rank, tag });
	const ranksOf = async (ids: Array<ActionIdProto>) => {
		const metadata = new UnitMetadata();
		await metadata.update(
			UnitMetadataProto.create({
				spells: ids.map(id => SpellStatsProto.create({ id })),
				auras: ids.map(id => AuraStatsProto.create({ id })),
			}),
		);
		const rankMap = (entries: Array<{ id: ActionId }>) => Object.fromEntries(entries.map(({ id }) => [`${id.spellId}|${id.tag}`, id.rank]));
		return { spells: rankMap(metadata.getSpells()), auras: rankMap(metadata.getAuras()) };
	};

	it('keeps the ranks of a spell registered at several', async () => {
		const ranks = { '2120|0': 1, '2121|0': 2 };
		expect(await ranksOf([actionId(2120, 1), actionId(2121, 2)])).toEqual({ spells: ranks, auras: ranks });
	});

	it('drops the rank of a spell registered at one', async () => {
		const ranks = { '25304|0': 0, '2120|0': 1, '2121|0': 2 };
		expect(await ranksOf([actionId(25304, 11), actionId(2120, 1), actionId(2121, 2)])).toEqual({ spells: ranks, auras: ranks });
	});

	// A spell and its DoT, registered as a second spell under another tag, count apart.
	it('counts one spell under two tags as two', async () => {
		const ranks = { '25304|0': 0, '25304|1': 0 };
		expect(await ranksOf([actionId(25304, 11), actionId(25304, 11, 1)])).toEqual({ spells: ranks, auras: ranks });
	});
});
