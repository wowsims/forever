import { EncounterDesign, HealthKeyframe, PresetEncounter, PresetTarget, Target as TargetProto } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { Encounter } from '../../../../../sim/raid/encounter';
import type { Sim } from '../../../../../sim/sim';
import { createSimStore } from '../../../../../sim/state/sim_store';
import { encounterChoiceConfig } from './configs';

const presetFor = (id: number, design?: EncounterDesign) =>
	PresetEncounter.create({
		path: `raid/boss-${id}`,
		targets: [PresetTarget.create({ target: TargetProto.create({ id, name: `Boss ${id}` }) })],
		design,
	});

const raidTarget = presetFor(1);
const dynamicAdds = presetFor(2, EncounterDesign.create({ bossHealth: [HealthKeyframe.create({ healthPercent: 0.5 })] }));

const setup = () => {
	const db = { getAllPresetEncounters: () => [raidTarget, dynamicAdds], getZone: () => null, getNpc: () => null };
	const encounter = new Encounter({ store: createSimStore(), db } as unknown as Sim);
	return { encounter, config: encounterChoiceConfig(encounter, [], { id: 'encounter' }) };
};

const addKeyframe = (encounter: Encounter, keyframe: HealthKeyframe) => {
	const design = EncounterDesign.clone(encounter.getDesign());
	design.bossHealth.push(keyframe);
	encounter.setDesign(design);
};

const CUSTOM_ENCOUNTER = 'custom';
const RAID_TARGET = 'preset:0';
const DYNAMIC_ADDS = 'preset:1';

describe('encounterChoiceConfig', () => {
	it('picking a preset with a design sets its targets and design, and shows that preset', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		expect(encounter.getTargets().map(t => t.id)).toEqual([2]);
		expect(EncounterDesign.equals(encounter.getDesign(), dynamicAdds.design!)).toBe(true);
		expect(config.getValue(encounter)).toBe(DYNAMIC_ADDS);
	});

	it('editing the design of a picked preset flips the picker to Custom', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		addKeyframe(encounter, HealthKeyframe.create({ healthPercent: 0.2 }));
		expect(config.getValue(encounter)).toBe(CUSTOM_ENCOUNTER);
	});

	it('a preset without a design matches a plain encounter, and stops matching once a design is added', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, RAID_TARGET);
		expect(encounter.hasDesign()).toBe(false);
		expect(config.getValue(encounter)).toBe(RAID_TARGET);
		addKeyframe(encounter, HealthKeyframe.create());
		expect(config.getValue(encounter)).toBe(CUSTOM_ENCOUNTER);
	});

	it('picking a design-less preset after a design clears the design', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		config.setValue(encounter, RAID_TARGET);
		expect(encounter.hasDesign()).toBe(false);
		expect(config.getValue(encounter)).toBe(RAID_TARGET);
	});
});
