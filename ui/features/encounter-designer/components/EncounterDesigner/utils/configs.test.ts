import { DesignBossAbility_Kind, EncounterDesign, PresetTarget, Target } from '@generated/proto/common';
import { eventAt } from '@testing/encounter_design';
import { describe, expect, it, vi } from 'vitest';

import { defaultEvent } from '../../../model/lanes';
import { defaultWaveTemplate, type DesignFieldConfig, designFieldConfigs, type DesignFieldContext, editEvent } from './configs';

// Renders a string as its last key segment and its options, so a test reads which string and values were picked.
vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string, options?: Record<string, unknown>) => {
			const name = key.split('.').pop();
			return options ? `${name}(${Object.values(options).join(',')})` : name;
		},
	},
}));
vi.mock('@i18n/localization', () => ({ translateSpellSchool: (school: number) => String(school), translateMobType: (type: number) => `mob${type}` }));

const context: DesignFieldContext = { baseSeconds: 180, targetNames: ['Boss'], waveNames: [] };

const field = <E>(fields: Array<DesignFieldConfig<E>>, id: string) => fields.find(entry => entry.id === id)!;

describe('editEvent', () => {
	it('edits a copy and leaves the design it was given alone', () => {
		const design = EncounterDesign.create({ moves: [eventAt('move', 0.2)] });
		const next = editEvent(design, { kind: 'move', index: 0 }, event => {
			(event as (typeof design.moves)[number]).yards = 30;
		});
		expect(next.moves[0].yards).toBe(30);
		expect(design.moves[0].yards).toBe(5);
	});
});

describe('designFieldConfigs', () => {
	it('shows fight fractions in seconds and writes them back as fractions', () => {
		const wave = eventAt('wave', 0.2);
		const lifetime = field(designFieldConfigs('wave'), 'lifetime');
		if (lifetime.type !== 'number') throw new Error('lifetime is a number field');
		expect(lifetime.get(wave, context)).toBeCloseTo(18);
		lifetime.set(wave, 36, context);
		expect(wave.lifetime).toBeCloseTo(0.2);
	});

	it('shows kill share as a percentage of a 0..1 value', () => {
		const wave = defaultEvent('wave');
		const killShare = field(designFieldConfigs('wave'), 'kill_share');
		if (killShare.type !== 'number') throw new Error('kill share is a number field');
		killShare.set(wave, 40, context);
		expect(wave.killShare).toBeCloseTo(0.4);
	});

	it('shows the ability fields only for the kinds that read them', () => {
		const shown = (kind: DesignBossAbility_Kind) => {
			const ability = { ...defaultEvent('ability'), kind };
			return designFieldConfigs('ability')
				.filter(entry => entry.show?.(ability) ?? true)
				.map(entry => entry.id);
		};
		expect(shown(DesignBossAbility_Kind.SpellHit)).toEqual(expect.arrayContaining(['school', 'min_damage', 'cast_time']));
		expect(shown(DesignBossAbility_Kind.Strike)).not.toContain('school');
		expect(shown(DesignBossAbility_Kind.Thrash)).toEqual(expect.arrayContaining(['proc_chance', 'extra_attacks', 'icd']));
		expect(shown(DesignBossAbility_Kind.SwingTiming)).toContain('swing_speed_multiplier');
		expect(shown(DesignBossAbility_Kind.SwingTiming)).not.toContain('weapon_percent');
	});
});

describe('defaultWaveTemplate', () => {
	const preset = (path: string, tankIndex: number) => PresetTarget.create({ path, target: Target.create({ name: path, tankIndex }) });
	const boss = preset('Default/Boss', 0);
	const unTanked = preset('Default/Movement', -1);
	const dynamicAdd = preset('Default/Dynamic Add', 1);

	it('prefers the Dynamic Add, then the first preset that pulls no tank, then the first preset', () => {
		expect(defaultWaveTemplate([boss, unTanked, dynamicAdd], 1)?.name).toBe('Default/Dynamic Add');
		expect(defaultWaveTemplate([boss, unTanked], 1)?.name).toBe('Default/Movement');
		expect(defaultWaveTemplate([boss], 1)?.name).toBe('Default/Boss');
		expect(defaultWaveTemplate([], 1)).toBeUndefined();
	});

	it('counts a tank index past the raid tanks as pulling none', () => {
		expect(defaultWaveTemplate([boss, preset('Default/Second', 1)], 1)?.name).toBe('Default/Second');
		expect(defaultWaveTemplate([boss, unTanked], 0)?.name).toBe('Default/Boss');
	});
});
