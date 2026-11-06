import { EncounterPlan, HealthKeyframe, PresetEncounter, PresetTarget, Target as TargetProto } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { Encounter } from '../../../../../sim/raid/encounter';
import type { Sim } from '../../../../../sim/sim';
import { createSimStore } from '../../../../../sim/state/sim_store';
import { encounterChoiceConfig } from './configs';

const presetFor = (id: number, plan?: EncounterPlan) =>
	PresetEncounter.create({
		path: `raid/boss-${id}`,
		targets: [PresetTarget.create({ target: TargetProto.create({ id, name: `Boss ${id}` }) })],
		plan,
	});

const raidTarget = presetFor(1);
const dynamicAdds = presetFor(2, EncounterPlan.create({ bossHealth: [HealthKeyframe.create({ healthPercent: 0.5 })] }));

const setup = () => {
	const db = { getAllPresetEncounters: () => [raidTarget, dynamicAdds], getZone: () => null, getNpc: () => null };
	const encounter = new Encounter({ store: createSimStore(), db } as unknown as Sim);
	return { encounter, config: encounterChoiceConfig(encounter, [], { id: 'encounter' }) };
};

const addKeyframe = (encounter: Encounter, keyframe: HealthKeyframe) => {
	const plan = EncounterPlan.clone(encounter.getPlan());
	plan.bossHealth.push(keyframe);
	encounter.setPlan(plan);
};

const CUSTOM_ENCOUNTER = 'custom';
const RAID_TARGET = 'preset:0';
const DYNAMIC_ADDS = 'preset:1';

describe('encounterChoiceConfig', () => {
	it('picking a preset with a plan sets its targets and plan, and shows that preset', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		expect(encounter.getTargets().map(t => t.id)).toEqual([2]);
		expect(EncounterPlan.equals(encounter.getPlan(), dynamicAdds.plan!)).toBe(true);
		expect(config.getValue(encounter)).toBe(DYNAMIC_ADDS);
	});

	it('editing the plan of a picked preset flips the picker to Custom', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		addKeyframe(encounter, HealthKeyframe.create({ healthPercent: 0.2 }));
		expect(config.getValue(encounter)).toBe(CUSTOM_ENCOUNTER);
	});

	it('a preset without a plan matches a plain encounter, and stops matching once a plan is added', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, RAID_TARGET);
		expect(encounter.hasPlan()).toBe(false);
		expect(config.getValue(encounter)).toBe(RAID_TARGET);
		addKeyframe(encounter, HealthKeyframe.create());
		expect(config.getValue(encounter)).toBe(CUSTOM_ENCOUNTER);
	});

	it('picking a preset without a plan after one with a plan clears the plan', () => {
		const { encounter, config } = setup();
		config.setValue(encounter, DYNAMIC_ADDS);
		config.setValue(encounter, RAID_TARGET);
		expect(encounter.hasPlan()).toBe(false);
		expect(config.getValue(encounter)).toBe(RAID_TARGET);
	});
});
