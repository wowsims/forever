import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, Spec } from '@generated/proto/common';
import { Warlock_Rotation, WarlockRotation } from '@generated/proto/warlock';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import AffWrackBuildJson from './builds/aff_wrack.build.json';
import DemoFireBrandBuildJson from './builds/demo_fire_brand.build.json';
import DemoShadowBrandBuildJson from './builds/demo_shadow_brand.build.json';
import DestroBuildJson from './builds/destro.build.json';
import SmRuinBuildJson from './builds/sm_ruin.build.json';
import spec from './spec';

// The generator reads nothing from the player.
const player = {} as Player<Spec.SpecWarlock>;

const generate = (simple: Partial<Warlock_Rotation>) =>
	APLRotation.toJson(
		spec.simpleRotation!(player, Warlock_Rotation.create({ rotation: WarlockRotation.WarlockRotationDemoFireBrand, ...simple }), Cooldowns.create()),
	);

const buildRotation = (json: { player: { rotation: unknown } }) => APLRotation.toJson(APLRotation.fromJson(json.player.rotation as any));

const spellIds = (json: unknown): Array<number> => {
	if (Array.isArray(json)) return json.flatMap(spellIds);
	if (json === null || typeof json !== 'object') return [];
	return Object.entries(json).flatMap(([key, value]) => (key === 'spellId' && typeof value === 'number' ? [value] : spellIds(value)));
};

const count = (json: unknown, id: number) => spellIds(json).filter(spellId => spellId === id).length;

describe('Warlock simple rotation', () => {
	it.each([
		[WarlockRotation.WarlockRotationNone, DemoFireBrandBuildJson],
		[WarlockRotation.WarlockRotationDemoFireBrand, DemoFireBrandBuildJson],
		[WarlockRotation.WarlockRotationDemoShadowBrand, DemoShadowBrandBuildJson],
		[WarlockRotation.WarlockRotationDestro, DestroBuildJson],
		[WarlockRotation.WarlockRotationSmRuin, SmRuinBuildJson],
		[WarlockRotation.WarlockRotationAffWrack, AffWrackBuildJson],
	])('runs the build rotation of choice %i as the build has it with every book rank chosen', (rotation, build) => {
		expect(generate({ rotation, useShadowBoltRank10: true, useImmolateRank8: true, useCorruptionRank7: true })).toEqual(buildRotation(build));
	});

	it.each([
		['useShadowBoltRank10', 25307, 11661],
		['useImmolateRank8', 25309, 11668],
		['useCorruptionRank7', 25311, 11672],
	] as const)('%s left false casts the trainer rank in place of book rank %i', (field, bookId, trainerId) => {
		const build = buildRotation(DemoShadowBrandBuildJson);
		const all = { useShadowBoltRank10: true, useImmolateRank8: true, useCorruptionRank7: true };
		const rotation = generate({ ...all, rotation: WarlockRotation.WarlockRotationDemoShadowBrand, [field]: false });
		expect(count(build, bookId)).toBeGreaterThan(0);
		expect(count(rotation, bookId)).toBe(0);
		expect(count(rotation, trainerId)).toBe(count(build, bookId));
		expect(spellIds(rotation).filter(id => id !== trainerId)).toEqual(spellIds(build).filter(id => id !== bookId));
	});

	it('casts only the trainer ranks by default', () => {
		const rotation = generate({ rotation: WarlockRotation.WarlockRotationDemoShadowBrand });
		expect([25307, 25309, 25311].map(id => count(rotation, id))).toEqual([0, 0, 0]);
	});
});
