import { AreaType, Encounter as EncounterProto, InputType, PresetEncounter, PresetTarget, Target as TargetProto, TargetInput } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import type { Sim } from '../sim';
import { createSimStore } from '../state/sim_store';
import { Encounter } from './encounter';

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
});
