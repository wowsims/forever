import { APLGroup, APLRotation } from '@generated/proto/apl';
import { Cooldowns, Spec } from '@generated/proto/common';
import { PaladinAura, PaladinJudgement, PaladinSealTwist, RetributionPaladin_Rotation } from '@generated/proto/paladin';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import DefaultApl from './apls/default.apl.siml';
import HolyShockApl from './apls/holy_shock.apl.siml';
import UwSorSofApl from './apls/uw_sor_sof.apl.siml';
import { Defaults } from './presets';
import spec from './spec';

type Talents = { holyShock?: boolean; unsealedWrath?: boolean; sealOfCommand?: boolean };

// The generator reads nothing from the player but these talents.
const player = (talents: Talents) =>
	({ getTalents: () => ({ holyShock: false, unsealedWrath: false, sealOfCommand: false, ...talents }) }) as unknown as Player<Spec.SpecRetributionPaladin>;

const generateWith = (talents: Talents, simple: Partial<RetributionPaladin_Rotation>, cooldowns = Cooldowns.create()) =>
	spec.simpleRotation!(player(talents), RetributionPaladin_Rotation.create({ ...Defaults.simpleRotation, ...simple }), cooldowns);

const generate = (simple: Partial<RetributionPaladin_Rotation>, holyShock = false, cooldowns = Cooldowns.create()) =>
	generateWith({ holyShock }, simple, cooldowns);

const groupNames = (apl: { groups: Array<{ name: string }> }) => apl.groups.map(g => g.name);

const variable = (rotation: APLRotation, name: string) => {
	const value = rotation.valueVariables.find(v => v.name === name)?.value?.value;
	return value?.oneofKind === 'const' ? value.const.val : undefined;
};

const group = (rotation: APLRotation, name: string) => APLGroup.toJson(rotation.groups.find(g => g.name === name)!);

const castGroup = (name: string, spellId: object) => ({ name, actions: [{ action: { castSpell: { spellId } } }] });

const aplGroup = (apl: any, name: string) => APLGroup.toJson(APLGroup.fromJson(apl.groups.find((g: any) => g.name === name)));

describe('Retribution simple rotation', () => {
	it('builds on the Holy Shock APL with the talent and the Default APL without it', () => {
		expect(generate({}, true).groups.map(g => g.name)).toEqual(HolyShockApl.groups.map((g: { name: string }) => g.name));
		expect(generate({}, false).groups.map(g => g.name)).toEqual(DefaultApl.groups.map((g: { name: string }) => g.name));
	});

	it.each([
		[PaladinJudgement.JudgementOfTheCrusader, { spellId: 20308, rank: 6 }],
		[PaladinJudgement.JudgementOfLight, { spellId: 20349, rank: 4 }],
		[PaladinJudgement.JudgementOfWisdom, { spellId: 20357, rank: 3 }],
	])('casts assigned judgement %i in Cast assigned judgement', (assignedJudgement, spellId) => {
		const rotation = generate({ assignedJudgement });
		expect(variable(rotation, 'Has assigned judgement')).toBe('true');
		expect(group(rotation, 'Cast assigned judgement')).toEqual(castGroup('Cast assigned judgement', spellId));
	});

	it('leaves Cast assigned judgement as the APL has it when no judgement is assigned', () => {
		const rotation = generate({ assignedJudgement: PaladinJudgement.JudgementNone });
		expect(variable(rotation, 'Has assigned judgement')).toBe('false');
		expect(group(rotation, 'Cast assigned judgement')).toEqual(aplGroup(DefaultApl, 'Cast assigned judgement'));
	});

	it('leaves Activate assigned aura as the APL has it when no aura is set', () => {
		expect(group(generate({ aura: PaladinAura.AuraUnknown }), 'Activate assigned aura')).toEqual(aplGroup(DefaultApl, 'Activate assigned aura'));
	});

	it.each([
		[PaladinAura.RetributionAura, { spellId: 10301, rank: 5 }],
		[PaladinAura.DevotionAura, { spellId: 10293, rank: 7 }],
		[PaladinAura.ConcentrationAura, { spellId: 19746 }],
		[PaladinAura.FireResistanceAura, { spellId: 19900, rank: 3 }],
		[PaladinAura.FrostResistanceAura, { spellId: 19898, rank: 3 }],
		[PaladinAura.ShadowResistanceAura, { spellId: 19896, rank: 3 }],
	])('casts aura %i in Activate assigned aura', (aura, spellId) => {
		expect(group(generate({ aura }), 'Activate assigned aura')).toEqual(castGroup('Activate assigned aura', spellId));
	});

	it.each([
		[0, '25'],
		[6, '11'],
		[10, '11'],
		[11, '11'],
		[30, '30'],
	])('sets Pull range from a starting range of %i to %s', (pullRange, expected) => {
		expect(variable(generate({ pullRange }), 'Pull range')).toBe(expected);
	});

	it.each([[{ unsealedWrath: true, sealOfCommand: true }], [{ unsealedWrath: true, sealOfCommand: false }]])(
		'builds on the Seal of Fury twist APL when Unsealed Wrath twists Seal of Fury, with talents %o',
		talents => {
			const rotation = generateWith(talents, { uwTwist: PaladinSealTwist.SealTwistFury });
			expect(groupNames(rotation)).toEqual(groupNames(UwSorSofApl));
		},
	);

	it.each([
		['Seal of Command is twisted', { unsealedWrath: true, sealOfCommand: true }, PaladinSealTwist.SealTwistCommand, DefaultApl],
		['no twist is picked', { unsealedWrath: true, sealOfCommand: true }, PaladinSealTwist.SealTwistUnknown, DefaultApl],
		['Unsealed Wrath is not talented', { sealOfCommand: true }, PaladinSealTwist.SealTwistFury, DefaultApl],
		['Holy Shock is talented', { holyShock: true, unsealedWrath: true }, PaladinSealTwist.SealTwistFury, HolyShockApl],
	])("keeps the talents' APL when %s", (_, talents, uwTwist, apl) => {
		expect(groupNames(generateWith(talents, { uwTwist }))).toEqual(groupNames(apl));
	});

	it("sets the Seal of Fury twist's Echo stacks, and its own defaults where none is set", () => {
		const talents = { unsealedWrath: true, sealOfCommand: true };
		const set = generateWith(talents, { uwTwist: PaladinSealTwist.SealTwistFury, uwSofStacks: 2, uwSorStacks: 2 });
		expect(variable(set, 'Max Seal of Fury Stacks')).toBe('2');
		expect(variable(set, 'Max Seal of Righteousness Stacks')).toBe('2');

		const unset = generateWith(talents, { uwTwist: PaladinSealTwist.SealTwistFury, uwSofStacks: 0, uwSorStacks: 0 });
		expect(variable(unset, 'Max Seal of Fury Stacks')).toBe('1');
		expect(variable(unset, 'Max Seal of Righteousness Stacks')).toBe('3');
	});

	it("sets the Seal of Command twist's Echo stacks, and its own defaults where none is set", () => {
		const talents = { unsealedWrath: true, sealOfCommand: true };
		const set = generateWith(talents, { uwSocStacks: 2, uwSorStacks: 3 });
		expect(variable(set, 'Max Seal of Command Stacks')).toBe('2');
		expect(variable(set, 'Max Seal of Righteousness Stacks')).toBe('3');

		const unset = generateWith(talents, { uwSocStacks: 0, uwSorStacks: 0 });
		expect(variable(unset, 'Max Seal of Command Stacks')).toBe('3');
		expect(variable(unset, 'Max Seal of Righteousness Stacks')).toBe('1');
	});

	it('sets Delay Major CDs in seconds', () => {
		expect(variable(generate({ delayMajorCds: 11 }), 'Delay Major CDs')).toBe('11s');
	});

	it('puts the scheduled cooldowns ahead of the priority list', () => {
		const cooldowns = Cooldowns.create({ cooldowns: [{ id: { rawId: { oneofKind: 'spellId', spellId: 1310905 } }, timings: [10] }] });
		const rotation = generate({}, false, cooldowns);
		expect(rotation.priorityList).toHaveLength(DefaultApl.priorityList.length + 1);
		expect(rotation.priorityList[0].action?.action.oneofKind).toBe('schedule');
	});
});
