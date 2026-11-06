import { EncounterPlan, PlanBossAbility_Kind, PresetTarget, Target } from '@generated/proto/common';
import { eventAt } from '@testing/encounter_plan';
import { describe, expect, it, vi } from 'vitest';

import { defaultEvent } from '../../../model/lanes';
import { defaultWaveTemplate, editEvent, type PlanFieldConfig, planFieldConfigs, type PlanFieldContext } from './configs';

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

const context: PlanFieldContext = { baseSeconds: 180, targetNames: ['Boss'], waveNames: [] };

const field = <E>(fields: Array<PlanFieldConfig<E>>, id: string) => fields.find(entry => entry.id === id)!;

describe('editEvent', () => {
	it('edits a copy and leaves the plan it was given alone', () => {
		const plan = EncounterPlan.create({ moves: [eventAt('move', 0.2)] });
		const next = editEvent(plan, { kind: 'move', index: 0 }, event => {
			(event as (typeof plan.moves)[number]).yards = 30;
		});
		expect(next.moves[0].yards).toBe(30);
		expect(plan.moves[0].yards).toBe(5);
	});
});

describe('planFieldConfigs', () => {
	it('shows fight fractions in seconds and writes them back as fractions', () => {
		const wave = eventAt('wave', 0.2);
		const lifetime = field(planFieldConfigs('wave'), 'lifetime');
		if (lifetime.type !== 'number') throw new Error('lifetime is a number field');
		expect(lifetime.get(wave, context)).toBeCloseTo(18);
		lifetime.set(wave, 36, context);
		expect(wave.lifetime).toBeCloseTo(0.2);
	});

	it('shows kill share as a percentage of a 0..1 value', () => {
		const wave = defaultEvent('wave');
		const killShare = field(planFieldConfigs('wave'), 'kill_share');
		if (killShare.type !== 'number') throw new Error('kill share is a number field');
		killShare.set(wave, 40, context);
		expect(wave.killShare).toBeCloseTo(0.4);
	});

	it('shows the ability fields only for the kinds that read them', () => {
		const shown = (kind: PlanBossAbility_Kind) => {
			const ability = { ...defaultEvent('ability'), kind };
			return planFieldConfigs('ability')
				.filter(entry => entry.show?.(ability) ?? true)
				.map(entry => entry.id);
		};
		expect(shown(PlanBossAbility_Kind.SpellHit)).toEqual(expect.arrayContaining(['school', 'min_damage', 'cast_time']));
		expect(shown(PlanBossAbility_Kind.Strike)).not.toContain('school');
		expect(shown(PlanBossAbility_Kind.Thrash)).toEqual(expect.arrayContaining(['proc_chance', 'extra_attacks', 'icd']));
		expect(shown(PlanBossAbility_Kind.SwingTiming)).toContain('swing_speed_multiplier');
		expect(shown(PlanBossAbility_Kind.SwingTiming)).not.toContain('weapon_percent');
		const area = ['school', 'min_damage', 'max_damage', 'resistible', 'avoidable'];
		expect(shown(PlanBossAbility_Kind.Pulse)).toEqual(expect.arrayContaining([...area, 'tick_interval', 'tick_jitter']));
		expect(shown(PlanBossAbility_Kind.AreaHit)).toEqual(expect.arrayContaining(area));
		expect(shown(PlanBossAbility_Kind.AreaHit)).not.toContain('tick_interval');
		for (const kind of [PlanBossAbility_Kind.Pulse, PlanBossAbility_Kind.AreaHit]) expect(shown(kind)).not.toContain('target');
		expect(shown(PlanBossAbility_Kind.SpellHit)).not.toContain('resistible');
	});

	it('fills a working pulse when the kind field picks one', () => {
		const ability = defaultEvent('ability');
		const kind = field(planFieldConfigs('ability'), 'ability_kind');
		if (kind.type !== 'choice') throw new Error('the kind is a choice field');
		kind.set(ability, PlanBossAbility_Kind.Pulse);
		expect([ability.kind, ability.maxDamage, ability.tickInterval, ability.resistible]).toEqual([PlanBossAbility_Kind.Pulse, 700, 1, true]);
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
