import { CreatureFamily, MobType, PresetTarget, Target as TargetProto } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { describe, expect, it } from 'vitest';

import {
	bleedImmuneConfig,
	booleanConfigs,
	creatureFamilyConfig,
	type EncounterTargetFieldContext,
	mobTypeConfig,
	presetTargetConfig,
	type TargetFieldContext,
} from './configs';

const contextOf = (target: TargetProto): TargetFieldContext => ({
	idPrefix: 'target-0',
	getTarget: () => target,
	modifyTarget: edit => edit(target),
});

describe('bleedImmuneConfig', () => {
	it('shows Immune to Bleeds off by default', () => {
		const target = TargetProto.create();
		const bleedImmune = bleedImmuneConfig(contextOf(target));

		expect(bleedImmune.getValue?.(null)).toBe(false);
		bleedImmune.setValue?.(null, true);
		expect(target.bleedImmune).toBe(true);
		expect(bleedImmune.getValue?.(null)).toBe(true);
	});
});

describe('creature family', () => {
	it('drops the family on a new mob type that does not hold it, and keeps one the new mob type holds', () => {
		const target = TargetProto.create({ mobType: MobType.MobTypeHumanoid });
		const context = contextOf(target);
		creatureFamilyConfig(context, target.mobType).setValue(null, CreatureFamily.CreatureFamilyNaga);
		mobTypeConfig(context).setValue(null, MobType.MobTypeBeast);
		expect(target.creatureFamily).toBe(CreatureFamily.CreatureFamilyUnknown);

		target.creatureFamily = CreatureFamily.CreatureFamilyAquatic;
		mobTypeConfig(context).setValue(null, MobType.MobTypeElemental);
		expect(target.creatureFamily).toBe(CreatureFamily.CreatureFamilyAquatic);
	});
});

describe('booleanConfigs', () => {
	it('enables the dual-wield miss penalty only while dual wield is on', () => {
		const target = TargetProto.create({ dualWield: false });
		const penalty = booleanConfigs(contextOf(target)).find(config => config.id.endsWith('dw-miss-penalty'))!;
		expect(penalty.enableWhen?.(null)).toBe(false);

		target.dualWield = true;
		expect(penalty.enableWhen?.(null)).toBe(true);
	});
});

// Encounters register a preset per raid size and difficulty, all sharing one npc id, which is the field's value and
// all the sim reads the target from.
describe('presetTargetConfig', () => {
	const preset = (path: string, id: number) => PresetTarget.create({ path, target: TargetProto.create({ id }) });
	const options = (presets: Array<PresetTarget>) => {
		const context: EncounterTargetFieldContext = {
			...contextOf(TargetProto.create()),
			encounter: { sim: { db: { getAllPresetTargets: () => presets } } } as unknown as Encounter,
			targetIndex: 0,
		};
		return presetTargetConfig(context).values.map(entry => entry.name);
	};

	it('offers one option per npc id, the first preset registered under it', () => {
		expect(
			options([preset('Molten Core/Lucifron', 12118), preset('Molten Core/Lucifron (alternate)', 12118), preset('x/Ragnaros', 11502)]).slice(1),
		).toEqual(['Molten Core/Lucifron', 'x/Ragnaros']);
	});
});
