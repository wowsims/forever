import { Player as PlayerProto } from '@generated/proto/api';
import {
	AddWave,
	AreaType,
	Encounter as EncounterProto,
	EncounterPlan,
	HealthKeyframe,
	PlanBossAbility,
	PlanDamageAmp,
	PlanDowntime,
	PlanMove,
	PlanTargets,
	PresetEncounter,
	PresetTarget,
	Target as TargetProto,
} from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { SimSettingCategories } from '../constants/sim_settings';
import type { Sim } from '../sim';
import { applyIndividualSimSettings, individualSimSettingsToProto } from '../state/serialization';
import { createSimStore } from '../state/sim_store';
import { Encounter, remapPlanTargets } from './encounter';

const ZONES: Record<number, { id: number; name: string; areaTypes: Array<AreaType> }> = {
	2677: { id: 2677, name: 'Blackwing Lair', areaTypes: [AreaType.AreaTypeCavernous, AreaType.AreaTypeVolcanic] },
	44: { id: 44, name: 'Redridge Mountains', areaTypes: [AreaType.AreaTypeForestGrassland, AreaType.AreaTypeMountainous] },
	1: { id: 1, name: 'Nowhere', areaTypes: [] },
};
const NPC_ZONES: Record<number, number> = { 11583: 2677, 71543: 1 };
const db = {
	getZone: (id: number) => ZONES[id] ?? null,
	getNpc: (id: number) => (NPC_ZONES[id] ? { id, name: '', zoneId: NPC_ZONES[id] } : null),
};
const makeEncounter = () => new Encounter({ store: createSimStore(), db } as unknown as Sim);

const presetEncounter = () =>
	PresetEncounter.create({
		path: 'raid/preset-boss',
		targets: [
			PresetTarget.create({
				path: 'preset-boss',
				target: TargetProto.create({
					id: 71543,
					name: 'Preset Boss',
					level: 93,
					swingSpeed: 2,
					stats: [3],
				}),
			}),
		],
	});

describe('Encounter', () => {
	it.each([
		['applyPreset', (encounter: Encounter, preset: PresetEncounter) => encounter.applyPreset(preset)],
		['applyPresetTarget', (encounter: Encounter, preset: PresetEncounter) => encounter.applyPresetTarget(preset.targets[0], 0)],
	] as const)('%s stores a copy of the preset target, down to its stats', (_name, apply) => {
		const preset = presetEncounter();
		const dbTarget = preset.targets[0].target!;
		const encounter = makeEncounter();

		apply(encounter, preset);

		const stored = encounter.getTarget(0)!;
		expect(stored).not.toBe(dbTarget);
		expect(stored.stats).not.toBe(dbTarget.stats);
		expect(TargetProto.equals(stored, dbTarget)).toBe(true);
	});

	describe('area types', () => {
		it('holds a set: duplicates and Unknown drop, the order is the enum order', () => {
			const encounter = makeEncounter();

			encounter.setAreaTypes([AreaType.AreaTypeHaunted, AreaType.AreaTypeUnknown, AreaType.AreaTypeForestGrassland, AreaType.AreaTypeHaunted]);

			expect(encounter.getAreaTypes()).toEqual([AreaType.AreaTypeForestGrassland, AreaType.AreaTypeHaunted]);
		});

		it('takes its set from a zone, and forgets the zone once the set is edited by hand', () => {
			const encounter = makeEncounter();

			encounter.setZone(2677);
			expect(encounter.getZoneId()).toBe(2677);
			expect(encounter.getAreaTypes()).toEqual([AreaType.AreaTypeCavernous, AreaType.AreaTypeVolcanic]);

			encounter.setAreaTypes([...encounter.getAreaTypes(), AreaType.AreaTypeSnowy]);
			expect(encounter.getZoneId()).toBe(0);
			expect(encounter.getAreaTypes()).toEqual([AreaType.AreaTypeSnowy, AreaType.AreaTypeCavernous, AreaType.AreaTypeVolcanic]);

			encounter.setZone(0);
			expect(encounter.getAreaTypes()).toEqual([AreaType.AreaTypeSnowy, AreaType.AreaTypeCavernous, AreaType.AreaTypeVolcanic]);
			encounter.setZone(999999);
			expect(encounter.getZoneId()).toBe(0);
		});

		it('fills the zone in from a preset boss whose zone the database knows, and leaves it for one it does not', () => {
			const encounter = makeEncounter();
			encounter.setZone(44);

			encounter.applyPreset(PresetEncounter.create({ targets: [PresetTarget.create({ target: TargetProto.create({ id: 11583 }) })] }));
			expect(encounter.getZoneId()).toBe(2677);
			expect(encounter.getAreaTypes()).toEqual([AreaType.AreaTypeCavernous, AreaType.AreaTypeVolcanic]);

			encounter.applyPreset(presetEncounter());
			expect(encounter.getZoneId()).toBe(2677);
		});

		it('survives the proto round trip and clears on a proto without any', () => {
			const encounter = makeEncounter();
			encounter.setAreaTypes([AreaType.AreaTypeDesert, AreaType.AreaTypeCavernous]);

			const proto = encounter.toProto();
			expect(proto.areaTypes).toEqual([AreaType.AreaTypeDesert, AreaType.AreaTypeCavernous]);
			expect(proto.zoneId).toBe(0);

			const other = makeEncounter();
			other.fromProto(proto);
			expect(other.getAreaTypes()).toEqual([AreaType.AreaTypeDesert, AreaType.AreaTypeCavernous]);

			encounter.setZone(44);
			other.fromProto(encounter.toProto());
			expect(other.getZoneId()).toBe(44);
			expect(other.getAreaTypes()).toEqual([AreaType.AreaTypeForestGrassland, AreaType.AreaTypeMountainous]);

			other.fromProto(EncounterProto.create({ targets: [Encounter.defaultTargetProto()] }));
			expect(other.getAreaTypes()).toEqual([]);
		});
	});

	it('leaves the database preset untouched when a writer edits the applied target in place', () => {
		const preset = presetEncounter();
		const dbTarget = preset.targets[0].target!;
		const encounter = makeEncounter();

		encounter.applyPreset(preset);
		const stored = encounter.getTarget(0)!;
		stored.level = 88;
		stored.stats[0] = 9;

		expect(dbTarget.level).toBe(93);
		expect(dbTarget.stats[0]).toBe(3);
	});

	describe('name', () => {
		it('survives the proto, binary and JSON round trips, and clears on a proto without one', () => {
			const encounter = makeEncounter();
			encounter.setName('Phase two');
			const other = makeEncounter();

			other.fromProto(encounter.toProto());
			expect(other.getName()).toBe('Phase two');
			other.fromProto(EncounterProto.create({ targets: [Encounter.defaultTargetProto()] }));
			expect(other.getName()).toBe('');

			other.fromProto(EncounterProto.fromBinary(EncounterProto.toBinary(encounter.toProto())));
			expect(other.getName()).toBe('Phase two');
			other.applyDefaults();
			other.fromProto(EncounterProto.fromJsonString(EncounterProto.toJsonString(encounter.toProto())));
			expect(other.getName()).toBe('Phase two');
		});

		it('keeps a named encounter from matching a preset, and clears the name and the My encounters mark when a preset is applied', () => {
			const preset = presetEncounter();
			const encounter = makeEncounter();
			encounter.applyPreset(preset);
			expect(encounter.matchesPreset(preset)).toBe(true);

			encounter.setName('Mine');
			expect(encounter.matchesPreset(preset)).toBe(false);

			encounter.setMyEncounter(true);
			encounter.applyPreset(preset);
			expect(encounter.getName()).toBe('');
			expect(encounter.isMyEncounter()).toBe(false);
			expect(encounter.matchesPreset(preset)).toBe(true);
		});

		it('carries the My encounters mark through the proto, and loses it on a proto without one', () => {
			const encounter = makeEncounter();
			encounter.setName('Mine');
			encounter.setMyEncounter(true);
			const other = makeEncounter();

			other.fromProto(EncounterProto.fromJsonString(EncounterProto.toJsonString(encounter.toProto())));
			expect(other.isMyEncounter()).toBe(true);
			other.fromProto(EncounterProto.create({ name: 'Mine', targets: [Encounter.defaultTargetProto()] }));
			expect(other.isMyEncounter()).toBe(false);
		});
	});

	describe('plan', () => {
		const withThreeTargets = () => {
			const encounter = makeEncounter();
			encounter.setTargets([1, 2, 3].map(id => TargetProto.create({ id })));
			encounter.setPlan(
				EncounterPlan.create({
					downtimes: [PlanDowntime.create({ targetIndex: 0 }), PlanDowntime.create({ targetIndex: 2 })],
					bossAbilities: [PlanBossAbility.create({ targetIndex: 1 })],
					damageAmps: [PlanDamageAmp.create({ targets: PlanTargets.create({ targetIndices: [1, 2] }) })],
				}),
			);
			return encounter;
		};
		const indices = (encounter: Encounter) => {
			const plan = encounter.getPlan();
			return {
				downtimes: plan.downtimes.map(d => d.targetIndex),
				abilities: plan.bossAbilities.map(a => a.targetIndex),
				amps: plan.damageAmps.map(a => a.targets?.targetIndices),
			};
		};

		it('starts plain and counts any non-empty list as a plan', () => {
			const encounter = makeEncounter();
			expect(encounter.hasPlan()).toBe(false);

			const lanes: Array<Partial<EncounterPlan>> = [
				{ downtimes: [PlanDowntime.create()] },
				{ bossHealth: [HealthKeyframe.create()] },
				{ addWaves: [AddWave.create()] },
				{ moves: [PlanMove.create()] },
				{ damageAmps: [PlanDamageAmp.create()] },
				{ bossAbilities: [PlanBossAbility.create()] },
			];
			const read = lanes.map(lane => {
				encounter.setPlan(EncounterPlan.create(lane));
				return encounter.hasPlan();
			});
			expect(read).toEqual([true, true, true, true, true, true]);

			encounter.setPlan(EncounterPlan.create());
			expect(encounter.hasPlan()).toBe(false);
		});

		it('survives the proto round trip and the binary form links use, and clears on a proto without one', () => {
			const encounter = withThreeTargets();
			const other = makeEncounter();

			other.fromProto(encounter.toProto());
			expect(EncounterPlan.equals(other.getPlan(), encounter.getPlan())).toBe(true);
			expect(other.getPlan()).not.toBe(encounter.getPlan());

			// Links and saved entries carry the encounter as binary.
			const third = makeEncounter();
			third.fromProto(EncounterProto.fromBinary(EncounterProto.toBinary(encounter.toProto())));
			expect(EncounterPlan.equals(third.getPlan(), encounter.getPlan())).toBe(true);

			other.fromProto(EncounterProto.create({ targets: [Encounter.defaultTargetProto()] }));
			expect(other.hasPlan()).toBe(false);
		});

		it('keeps the name of every event kind through the proto and binary round trip', () => {
			const encounter = makeEncounter();
			encounter.setPlan(
				EncounterPlan.create({
					downtimes: [PlanDowntime.create({ name: 'Phase 2' })],
					bossHealth: [HealthKeyframe.create({ name: 'Enrage' })],
					bossAbilities: [PlanBossAbility.create({ name: 'Cleave' })],
					damageAmps: [PlanDamageAmp.create({ name: 'Buff' })],
					moves: [PlanMove.create({ name: 'Run' })],
				}),
			);
			const other = makeEncounter();
			other.fromProto(EncounterProto.fromBinary(EncounterProto.toBinary(encounter.toProto())));
			const plan = other.getPlan();
			expect([plan.downtimes[0].name, plan.bossHealth[0].name, plan.bossAbilities[0].name, plan.damageAmps[0].name, plan.moves[0].name]).toEqual([
				'Phase 2',
				'Enrage',
				'Cleave',
				'Buff',
				'Run',
			]);
		});

		it('leaves a plain encounter without a plan message, so it equals a preset proto that has none', () => {
			const plain = makeEncounter().toProto();
			expect(plain.plan).toBeUndefined();
			expect(EncounterProto.equals(plain, EncounterProto.create({ ...plain, plan: undefined }))).toBe(true);
		});

		it('stores a copy of the plan it is given', () => {
			const encounter = makeEncounter();
			const plan = EncounterPlan.create({ bossHealth: [HealthKeyframe.create()] });
			encounter.setPlan(plan);
			plan.bossHealth.length = 0;
			expect(encounter.getPlan().bossHealth).toHaveLength(1);
		});

		it('leaves an empty damage amp target list alone on reorder', () => {
			const encounter = withThreeTargets();
			encounter.setPlan(EncounterPlan.create({ damageAmps: [PlanDamageAmp.create({ targets: PlanTargets.create() })] }));
			const [a, b, c] = encounter.getTargets();
			encounter.setTargets([c, b, a]);
			expect(encounter.getPlan().damageAmps[0].targets?.targetIndices).toEqual([]);
		});

		it('replaces the plan with the preset plan on applyPreset, and clears it for a preset without one', () => {
			const encounter = withThreeTargets();
			const plan = EncounterPlan.create({ bossHealth: [HealthKeyframe.create({ healthPercent: 0.5 })] });

			encounter.applyPreset(PresetEncounter.create({ targets: [PresetTarget.create({ target: TargetProto.create({ id: 5 }) })], plan }));
			expect(EncounterPlan.equals(encounter.getPlan(), plan)).toBe(true);
			expect(encounter.getPlan()).not.toBe(plan);

			encounter.applyPreset(presetEncounter());
			expect(encounter.hasPlan()).toBe(false);
		});

		it('round-trips through the settings serialization of the Encounter category', () => {
			const encounter = withThreeTargets();
			const player = { toProto: () => PlayerProto.create(), fromProto: () => {}, getPlayerSpec: () => ({}) };
			const categories = [SimSettingCategories.Encounter];

			const settings = individualSimSettingsToProto({ player, sim: { encounter } } as never, categories);
			const loaded = makeEncounter();
			applyIndividualSimSettings({ player, sim: { raid: {}, encounter: loaded } } as never, settings, categories);

			expect(EncounterPlan.equals(loaded.getPlan(), encounter.getPlan())).toBe(true);
		});

		it('setPlan writes nothing for an equal plan', () => {
			const encounter = withThreeTargets();
			let writes = 0;
			encounter.sim.store.subscribe(() => writes++);

			encounter.setPlan(EncounterPlan.clone(encounter.getPlan()));
			expect(writes).toBe(0);

			encounter.setPlan(EncounterPlan.create());
			expect(writes).toBe(1);
		});

		it('follows a reordered target by identity', () => {
			const encounter = withThreeTargets();
			const [a, b, c] = encounter.getTargets();

			encounter.setTargets([c, a, b]);

			expect(indices(encounter)).toEqual({ downtimes: [1, 0], abilities: [2], amps: [[2, 0]] });
		});

		it('sets the index to -1 for a removed target', () => {
			const encounter = withThreeTargets();
			const [a, , c] = encounter.getTargets();

			encounter.setTargets([a, c]);

			expect(indices(encounter)).toEqual({ downtimes: [0, 1], abilities: [-1], amps: [[-1, 1]] });
		});

		it('keeps the index of a target replaced by a new object in place, and the plan itself when nothing moves', () => {
			const encounter = withThreeTargets();
			const plan = encounter.getPlan();
			const [a, b, c] = encounter.getTargets();

			expect(remapPlanTargets(plan, [a, b, c], [a, b, c, TargetProto.create()])).toBe(plan);
			expect(remapPlanTargets(plan, [a, b, c], [a, TargetProto.clone(b), c])).toBe(plan);

			encounter.setTargets([TargetProto.clone(a), c, b]);
			expect(indices(encounter)).toEqual({ downtimes: [0, 1], abilities: [2], amps: [[2, 1]] });
		});
	});
});
