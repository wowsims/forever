import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, EquipmentSpec, Spec } from '@generated/proto/common';
import { Warlock_Rotation, WarlockRotation } from '@generated/proto/warlock';
import type { Player } from '@sim/player/player';
import { describe, expect, it } from 'vitest';

import AffWrackBuildJson from './builds/aff_wrack.build.json';
import DemoFireBrandBuildJson from './builds/demo_fire_brand.build.json';
import DemoShadowBrandBuildJson from './builds/demo_shadow_brand.build.json';
import DestroBuildJson from './builds/destro.build.json';
import SmRuinBuildJson from './builds/sm_ruin.build.json';
import AffWrackGearJson from './gear_sets/aff_wrack.gear.json';
import DemoFireBrandGearJson from './gear_sets/demo_fire_brand.gear.json';
import DemoShadowBrandGearJson from './gear_sets/demo_shadow_brand.gear.json';
import DestroGearJson from './gear_sets/destro.gear.json';
import SmRuinGearJson from './gear_sets/sm_ruin.gear.json';
import spec from './spec';

// The generator reads nothing from the player.
const player = {} as Player<Spec.SpecWarlock>;

const generate = (simple: Partial<Warlock_Rotation>) =>
	APLRotation.toJson(
		spec.simpleRotation!(player, Warlock_Rotation.create({ rotation: WarlockRotation.WarlockRotationDemoFireBrand, ...simple }), Cooldowns.create()),
	);

const gearOf = (json: unknown) => EquipmentSpec.fromJson(json as any);

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
	])('runs the build rotation of choice %i as the build has it', (rotation, build) => {
		expect(generate({ rotation })).toEqual(buildRotation(build));
	});

	it.each([
		['demo_fire_brand', DemoFireBrandBuildJson],
		['demo_shadow_brand', DemoShadowBrandBuildJson],
		['destro', DestroBuildJson],
		['sm_ruin', SmRuinBuildJson],
		['aff_wrack', AffWrackBuildJson],
	])('%s casts only the trainer ranks', (_name, build) => {
		expect([25307, 25309, 25311].map(id => count(buildRotation(build), id))).toEqual([0, 0, 0]);
	});
});

describe('Warlock builds', () => {
	it.each([
		['demo_fire_brand', DemoFireBrandBuildJson, DemoFireBrandGearJson],
		['demo_shadow_brand', DemoShadowBrandBuildJson, DemoShadowBrandGearJson],
		['destro', DestroBuildJson, DestroGearJson],
		['sm_ruin', SmRuinBuildJson, SmRuinGearJson],
		['aff_wrack', AffWrackBuildJson, AffWrackGearJson],
	])('%s holds the gear of its gear set', (_name, build, gear) => {
		expect(gearOf(build.player.equipment)).toEqual(gearOf(gear));
	});
});
