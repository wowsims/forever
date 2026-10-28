import { Player as PlayerProto } from '@generated/proto/api';
import {
	AreaType,
	DesignBossAbility,
	DesignDamageAmp,
	DesignDowntime,
	DesignMove,
	DesignTargets,
	Encounter as EncounterProto,
	EncounterDesign,
	HealthKeyframe,
	InputType,
	PresetEncounter,
	PresetTarget,
	Target as TargetProto,
	TargetInput,
} from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { SimSettingCategories } from '../constants/sim_settings';
import type { Sim } from '../sim';
import { applyIndividualSimSettings, individualSimSettingsToProto } from '../state/serialization';
import { createSimStore } from '../state/sim_store';
import { Encounter, remapDesignTargets } from './encounter';

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
					targetInputs: [TargetInput.create({ label: 'Adds', inputType: InputType.Number, numberValue: 3 })],
				}),
			}),
		],
	});

describe('Encounter', () => {
	it.each([
		['applyPreset', (encounter: Encounter, preset: PresetEncounter) => encounter.applyPreset(preset)],
		['applyPresetTarget', (encounter: Encounter, preset: PresetEncounter) => encounter.applyPresetTarget(preset.targets[0], 0)],
	] as const)('%s stores a copy of the preset target, down to the nested target inputs', (_name, apply) => {
		const preset = presetEncounter();
		const dbTarget = preset.targets[0].target!;
		const encounter = makeEncounter();

		apply(encounter, preset);

		const stored = encounter.getTarget(0)!;
		expect(stored).not.toBe(dbTarget);
		expect(stored.targetInputs[0]).not.toBe(dbTarget.targetInputs[0]);
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
		stored.targetInputs[0].numberValue = 9;

		expect(dbTarget.level).toBe(93);
		expect(dbTarget.targetInputs[0].numberValue).toBe(3);
	});

	describe('design', () => {
		const withThreeTargets = () => {
			const encounter = makeEncounter();
			encounter.setTargets([1, 2, 3].map(id => TargetProto.create({ id })));
			encounter.setDesign(
				EncounterDesign.create({
					downtimes: [DesignDowntime.create({ targetIndex: 0 }), DesignDowntime.create({ targetIndex: 2 })],
					bossAbilities: [DesignBossAbility.create({ targetIndex: 1 })],
					damageAmps: [DesignDamageAmp.create({ targets: DesignTargets.create({ targetIndices: [1, 2] }) })],
				}),
			);
			return encounter;
		};
		const indices = (encounter: Encounter) => {
			const design = encounter.getDesign();
			return {
				downtimes: design.downtimes.map(d => d.targetIndex),
				abilities: design.bossAbilities.map(a => a.targetIndex),
				amps: design.damageAmps.map(a => a.targets?.targetIndices),
			};
		};

		it('starts plain and counts any non-empty list as a design', () => {
			const encounter = makeEncounter();
			expect(encounter.hasDesign()).toBe(false);

			encounter.modifyDesign(d => d.bossHealth.push(HealthKeyframe.create()));
			expect(encounter.hasDesign()).toBe(true);
		});

		it('survives the proto round trip and the binary form links use, and clears on a proto without one', () => {
			const encounter = withThreeTargets();
			const other = makeEncounter();

			other.fromProto(encounter.toProto());
			expect(EncounterDesign.equals(other.getDesign(), encounter.getDesign())).toBe(true);
			expect(other.getDesign()).not.toBe(encounter.getDesign());

			// Links and saved entries carry the encounter as binary.
			const third = makeEncounter();
			third.fromProto(EncounterProto.fromBinary(EncounterProto.toBinary(encounter.toProto())));
			expect(EncounterDesign.equals(third.getDesign(), encounter.getDesign())).toBe(true);

			other.fromProto(EncounterProto.create({ targets: [Encounter.defaultTargetProto()] }));
			expect(other.hasDesign()).toBe(false);
		});

		it('keeps the name of every event kind through the proto and binary round trip', () => {
			const encounter = makeEncounter();
			encounter.setDesign(
				EncounterDesign.create({
					downtimes: [DesignDowntime.create({ name: 'Phase 2' })],
					bossHealth: [HealthKeyframe.create({ name: 'Enrage' })],
					bossAbilities: [DesignBossAbility.create({ name: 'Cleave' })],
					damageAmps: [DesignDamageAmp.create({ name: 'Buff' })],
					moves: [DesignMove.create({ name: 'Run' })],
				}),
			);
			const other = makeEncounter();
			other.fromProto(EncounterProto.fromBinary(EncounterProto.toBinary(encounter.toProto())));
			const design = other.getDesign();
			expect([
				design.downtimes[0].name,
				design.bossHealth[0].name,
				design.bossAbilities[0].name,
				design.damageAmps[0].name,
				design.moves[0].name,
			]).toEqual(['Phase 2', 'Enrage', 'Cleave', 'Buff', 'Run']);
		});

		it('leaves a plain encounter without a design message, so it equals a preset proto that has none', () => {
			const plain = makeEncounter().toProto();
			expect(plain.design).toBeUndefined();
			expect(EncounterProto.equals(plain, EncounterProto.create({ ...plain, design: undefined }))).toBe(true);
		});

		it('stores a copy of the design it is given', () => {
			const encounter = makeEncounter();
			const design = EncounterDesign.create({ bossHealth: [HealthKeyframe.create()] });
			encounter.setDesign(design);
			design.bossHealth.length = 0;
			expect(encounter.getDesign().bossHealth).toHaveLength(1);
		});

		it('leaves an empty damage amp target list alone on reorder', () => {
			const encounter = withThreeTargets();
			encounter.modifyDesign(d => d.damageAmps.push(DesignDamageAmp.create({ targets: DesignTargets.create() })));
			const [a, b, c] = encounter.getTargets();
			encounter.setTargets([c, b, a]);
			expect(encounter.getDesign().damageAmps[1].targets?.targetIndices).toEqual([]);
		});

		it('replaces the design with the preset design on applyPreset, and clears it for a preset without one', () => {
			const encounter = withThreeTargets();
			const design = EncounterDesign.create({ bossHealth: [HealthKeyframe.create({ healthPercent: 0.5 })] });

			encounter.applyPreset(PresetEncounter.create({ targets: [PresetTarget.create({ target: TargetProto.create({ id: 5 }) })], design }));
			expect(EncounterDesign.equals(encounter.getDesign(), design)).toBe(true);
			expect(encounter.getDesign()).not.toBe(design);

			encounter.applyPreset(presetEncounter());
			expect(encounter.hasDesign()).toBe(false);
		});

		it('round-trips through the settings serialization of the Encounter category', () => {
			const encounter = withThreeTargets();
			const player = { toProto: () => PlayerProto.create(), fromProto: () => {}, getPlayerSpec: () => ({}) };
			const categories = [SimSettingCategories.Encounter];

			const settings = individualSimSettingsToProto({ player, sim: { encounter } } as never, categories);
			const loaded = makeEncounter();
			applyIndividualSimSettings({ player, sim: { raid: {}, encounter: loaded } } as never, settings, categories);

			expect(EncounterDesign.equals(loaded.getDesign(), encounter.getDesign())).toBe(true);
		});

		it('setDesign writes nothing for an equal design', () => {
			const encounter = withThreeTargets();
			let writes = 0;
			encounter.sim.store.subscribe(() => writes++);

			encounter.setDesign(EncounterDesign.clone(encounter.getDesign()));
			expect(writes).toBe(0);

			encounter.setDesign(EncounterDesign.create());
			expect(writes).toBe(1);
		});

		it('modifyDesign writes once and replaces the stored design', () => {
			const encounter = withThreeTargets();
			const before = encounter.getDesign();
			let writes = 0;
			encounter.sim.store.subscribe(() => writes++);

			encounter.modifyDesign(d => {
				d.downtimes.push(DesignDowntime.create({ targetIndex: 1 }));
				d.bossHealth.push(HealthKeyframe.create());
			});

			expect(writes).toBe(1);
			expect(encounter.getDesign()).not.toBe(before);
			expect(before.downtimes).toHaveLength(2);
			expect(encounter.getDesign().downtimes).toHaveLength(3);
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

		it('keeps the index of a target replaced by a new object in place, and the design itself when nothing moves', () => {
			const encounter = withThreeTargets();
			const design = encounter.getDesign();
			const [a, b, c] = encounter.getTargets();

			expect(remapDesignTargets(design, [a, b, c], [a, b, c, TargetProto.create()])).toBe(design);
			expect(remapDesignTargets(design, [a, b, c], [a, TargetProto.clone(b), c])).toBe(design);

			encounter.setTargets([TargetProto.clone(a), c, b]);
			expect(indices(encounter)).toEqual({ downtimes: [0, 1], abilities: [2], amps: [[2, 1]] });
		});
	});
});
