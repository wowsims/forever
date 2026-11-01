import { AuraStats as AuraStatsProto, SpellStats as SpellStatsProto, UnitMetadata as UnitMetadataProto } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Debuffs } from '@generated/proto/buffs';
import {
	ActionID as ActionIdProto,
	ConsumesSpec,
	ItemRandomSuffix,
	ItemSlot,
	PseudoStat,
	Stat,
	WeaponSkill,
	WeaponSkillValue,
	WeaponType,
} from '@generated/proto/common';
import { UIItem as Item } from '@generated/proto/ui';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ActionId } from '../proto/action_id';
import { EquippedItem } from '../proto/equipped_item';
import { Gear } from '../proto/gear';
import { Stats } from '../proto/stats';
import { adjustImbues, adjustSwapImbues, getTalentTreePoints } from '../proto/utils';
import { playerTalentStringToProto } from '../talents/factory';
import { Player, UnitMetadata } from './player';
import { PlayerSpecs } from './specs';

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

describe('Player.setWeaponSkills', () => {
	// Against the store's weapon skills, as setConsumes above: the method reads `slice()` and writes through `patch`.
	const withSkills = (skills: WeaponSkillValue[]) => {
		let stored = skills;
		const patch = vi.fn((_field: string, next: WeaponSkillValue[]) => {
			stored = next;
		});
		const player = { slice: () => ({ weaponSkills: stored }), patch } as unknown as Player<any>;
		return { set: (next: WeaponSkillValue[]) => Player.prototype.setWeaponSkills.call(player, next), patch, stored: () => stored };
	};
	const swords = (value: number) => WeaponSkillValue.create({ skill: WeaponSkill.WeaponSkillSwords, value });

	it('changes a skill already set to another value', () => {
		const skills = withSkills([swords(305)]);
		skills.set([swords(310)]);
		expect(skills.stored()).toEqual([swords(310)]);
	});

	it('writes nothing when the skills are the same', () => {
		const skills = withSkills([swords(305)]);
		skills.set([swords(305)]);
		expect(skills.patch).not.toHaveBeenCalled();
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
		expect(debuffStats({ huntersMark: true }).getStat(Stat.StatRangedAttackPower)).toBe(143);
		expect(debuffStats({ huntersMark: false }).getStat(Stat.StatRangedAttackPower)).toBe(0);
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

describe('Player.getResolvedAplRotationForBuild', () => {
	const SPEC = PlayerSpecs.DpsWarrior;
	const LIVE = '30305001102';
	const LIVE_TALENTS = playerTalentStringToProto(SPEC, LIVE);
	const LIVE_GEAR = new Gear({});
	const OTHER_GEAR = new Gear({ [ItemSlot.ItemSlotBack]: new EquippedItem({ item: Item.create({ id: 30 }) }) });
	const CUSTOM = APLRotation.create({ type: APLRotationType.TypeAPL, priorityList: [APLListItem.create({ hide: true })] });
	const simple = (name: string) => APLRotation.create({ type: APLRotationType.TypeSimple, simple: { specRotationJson: JSON.stringify({ name }) } });

	// Only what resolving a rotation reads, with the live talents already parsed into the cache.
	const makePlayer = ({
		type,
		rotation = APLRotation.create({ type }),
		generate = null,
	}: {
		type: APLRotationType;
		rotation?: APLRotation;
		generate?: ((player: Player<any>) => APLRotation) | null;
	}) =>
		Object.assign(Object.create(Player.prototype), {
			slice: () => ({ talentsString: LIVE, gear: LIVE_GEAR }),
			aplRotation_: rotation,
			autoRotationGenerator: type === APLRotationType.TypeAuto ? generate : null,
			simpleRotationGenerator: type === APLRotationType.TypeSimple ? generate : null,
			specTypeFunctions: { rotationCreate: () => ({}), rotationFromJson: (json: unknown) => json },
			playerSpec: SPEC,
			talents: LIVE_TALENTS,
		}) as Player<any>;

	it("resolves an Auto rotation from the given talents and leaves the player's talents alone", () => {
		const generated = APLRotation.create({ type: APLRotationType.TypeAuto, priorityList: [APLListItem.create({ hide: true })] });
		const seen: unknown[] = [];
		const player = makePlayer({
			type: APLRotationType.TypeAuto,
			generate: p => {
				seen.push(p.getTalentsString(), p.getTalentTreePoints(), p.getTalents());
				return generated;
			},
		});

		expect(player.getResolvedAplRotationForBuild({ talentsString: '05-3', gear: LIVE_GEAR, rotation: player.aplRotation })).toEqual(generated);
		expect(seen).toEqual(['05-3', getTalentTreePoints('05-3'), playerTalentStringToProto(SPEC, '05-3')]);
		expect(player.getTalentsString()).toBe(LIVE);
		expect(player.getTalents()).toBe(LIVE_TALENTS);
	});

	it('resolves a Simple rotation from the given talents', () => {
		const seen: string[] = [];
		const player = makePlayer({
			type: APLRotationType.TypeSimple,
			generate: p => {
				seen.push(p.getTalentsString());
				return APLRotation.create();
			},
		});

		player.getResolvedAplRotationForBuild({ talentsString: '05-3', gear: LIVE_GEAR, rotation: player.aplRotation });

		expect(seen).toEqual(['05-3']);
	});

	it('resolves an Auto rotation from the given gear, and leaves the player’s gear alone', () => {
		const seen: Gear[] = [];
		const player = makePlayer({
			type: APLRotationType.TypeAuto,
			generate: p => {
				seen.push(p.getGear());
				return APLRotation.create();
			},
		});

		player.getResolvedAplRotationForBuild({ talentsString: LIVE, gear: OTHER_GEAR, rotation: player.aplRotation });

		expect(seen).toHaveLength(1);
		expect(seen[0]).toBe(OTHER_GEAR);
		expect(player.getGear()).toBe(LIVE_GEAR);
	});

	it('resolves the given Simple rotation, not the player’s own', () => {
		const seen: unknown[] = [];
		const player = makePlayer({
			type: APLRotationType.TypeSimple,
			rotation: simple('live'),
			generate: p => {
				seen.push(p.getSimpleRotation());
				return APLRotation.create();
			},
		});

		player.getResolvedAplRotationForBuild({ talentsString: LIVE, gear: LIVE_GEAR, rotation: simple('picked') });

		expect(seen).toEqual([{ name: 'picked' }]);
		expect(player.getSimpleRotation()).toEqual({ name: 'live' });
	});

	it('returns a given custom APL as it is', () => {
		const player = makePlayer({ type: APLRotationType.TypeAuto, generate: () => APLRotation.create({ type: APLRotationType.TypeAuto }) });

		expect(player.getResolvedAplRotationForBuild({ talentsString: LIVE, gear: LIVE_GEAR, rotation: CUSTOM })).toBe(CUSTOM);
	});
});

describe('UnitMetadata ranks', () => {
	// 2120 and 2121 are two ranks of Flamestrike, and 25304 is the one rank of Frostbolt these tests register.
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
