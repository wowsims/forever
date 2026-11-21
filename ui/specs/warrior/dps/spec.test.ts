import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, HandType, Spec } from '@generated/proto/common';
import { DpsWarrior_Rotation, DpsWarriorSpec } from '@generated/proto/warrior';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import * as Presets from './presets';
import spec from './spec';

// A player whose main hand holds a weapon of the given hand type.
const wielding = (handType: HandType) => ({ getEquippedItem: () => ({ item: { handType } }) }) as unknown as Player<Spec.SpecDpsWarrior>;

const generate = (handType: HandType, simple: Partial<DpsWarrior_Rotation>) =>
	APLRotation.toJson(spec.simpleRotation!(wielding(handType), DpsWarrior_Rotation.create(simple), Cooldowns.create()));

// The rotation without the option toggles' hides and the simple settings.
const withoutToggles = (json: unknown): unknown => {
	if (Array.isArray(json)) return json.map(withoutToggles);
	if (json === null || typeof json !== 'object') return json;
	return Object.fromEntries(
		Object.entries(json)
			.filter(([key]) => key !== 'hide' && key !== 'simple')
			.map(([key, value]) => [key, withoutToggles(value)]),
	);
};

// Every spell id the rotation names, as `id` or `id tag N`.
const spellIds = (json: unknown): Array<string> => {
	if (Array.isArray(json)) return json.flatMap(spellIds);
	if (json === null || typeof json !== 'object') return [];
	const { spellId, tag } = json as { spellId?: unknown; tag?: number };
	if (typeof spellId === 'number') return [tag ? `${spellId} tag ${tag}` : `${spellId}`];
	return Object.values(json).flatMap(spellIds);
};

const count = (json: unknown, id: string) => spellIds(json).filter(spellId => spellId === id).length;

describe.each([
	['Fury', DpsWarriorSpec.DpsWarriorSpecFury, HandType.HandTypeOneHand, Presets.FURY_DEFAULT_ROTATION],
	['Fury with a two-hander', DpsWarriorSpec.DpsWarriorSpecFury, HandType.HandTypeTwoHand, Presets.FURY_2H_ROTATION],
	['Arms', DpsWarriorSpec.DpsWarriorSpecArms, HandType.HandTypeTwoHand, Presets.ARMS_DEFAULT_ROTATION],
])('%s simple rotation', (_, warriorSpec, handType, preset) => {
	it('runs the preset rotation when both toggles are on', () => {
		const book = generate(handType, { spec: warriorSpec, useHeroicStrikeRank9: true, useBattleShoutRank7: true });
		expect(withoutToggles(book)).toEqual(withoutToggles(APLRotation.toJson(preset.rotation.rotation!)));
	});

	it('names the spells the preset names when both toggles are on', () => {
		const book = generate(handType, { spec: warriorSpec, useHeroicStrikeRank9: true, useBattleShoutRank7: true });
		expect(spellIds(book)).toEqual(spellIds(APLRotation.toJson(preset.rotation.rotation!)));
	});

	it('casts the trainer ranks by default', () => {
		const rotation = generate(handType, { spec: warriorSpec });
		expect(count(rotation, '11567 tag 1')).toBeGreaterThan(0);
		expect(count(rotation, '11551')).toBeGreaterThan(0);
		expect(count(rotation, '25286 tag 1') + count(rotation, '25289')).toBe(0);
	});

	it('casts the class book ranks when both toggles are on, and keeps both raid Battle Shout ranks', () => {
		const trainer = generate(handType, { spec: warriorSpec });
		const book = generate(handType, { spec: warriorSpec, useHeroicStrikeRank9: true, useBattleShoutRank7: true });
		expect(count(book, '11567 tag 1') + count(book, '11551')).toBe(0);
		expect(count(book, '25286 tag 1')).toBe(count(trainer, '11567 tag 1'));
		expect(count(book, '25289')).toBe(count(trainer, '11551'));
		for (const raid of ['11551 tag -1', '25289 tag -1']) {
			expect(count(trainer, raid)).toBe(1);
			expect(count(book, raid)).toBe(1);
		}
	});

	it.each([
		['useHeroicStrikeRank9', '11567 tag 1', '25286 tag 1'],
		['useBattleShoutRank7', '11551', '25289'],
	] as const)('%s moves only its own spell', (field, trainerId, bookId) => {
		const trainer = generate(handType, { spec: warriorSpec });
		const book = generate(handType, { spec: warriorSpec, [field]: true });
		expect(spellIds(book).map(id => (id === bookId ? trainerId : id))).toEqual(spellIds(trainer));
	});
});
