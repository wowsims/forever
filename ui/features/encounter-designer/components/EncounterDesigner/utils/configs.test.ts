import { DesignBossAbility_Kind, EncounterDesign, MobType, PresetTarget, Target } from '@generated/proto/common';
import { eventAt } from '@testing/encounter_design';
import { describe, expect, it, vi } from 'vitest';

import { defaultEvent } from '../../../model/lanes';
import {
	defaultWaveTemplate,
	type DesignFieldConfig,
	designFieldConfigs,
	type DesignFieldContext,
	editEvent,
	eventName,
	eventTooltip,
	eventValueText,
	occurrenceName,
	parseTip,
	serializeTip,
} from './configs';

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

const context: DesignFieldContext = { baseSeconds: 180, targetNames: ['Boss'], waveNames: [], presetTargets: [] };

const field = <E>(fields: Array<DesignFieldConfig<E>>, id: string) => fields.find(entry => entry.id === id)!;

describe('eventName', () => {
	it('uses the name of any kind, set or not, and counts the repeats from 2', () => {
		const design = EncounterDesign.create({ downtimes: [{ ...defaultEvent('downtime'), name: 'Phase 2' }, defaultEvent('downtime')] });
		expect(eventName(design, { kind: 'downtime', index: 0 })).toBe('Phase 2');
		expect(eventName(design, { kind: 'downtime', index: 1 })).toBe('event_name(downtime,2)');
		expect(occurrenceName('Phase 2', 1)).toBe('Phase 2');
		expect(occurrenceName('Phase 2', 3)).toBe('occurrence_name(Phase 2,3)');
	});

	it('uses a wave name and falls back to the kind and position', () => {
		const design = EncounterDesign.create({ addWaves: [{ ...defaultEvent('wave'), name: 'Sons of Flame' }, defaultEvent('wave')] });
		expect(eventName(design, { kind: 'wave', index: 0 })).toBe('Sons of Flame');
		expect(eventName(design, { kind: 'wave', index: 1 })).toBe('event_name(wave,2)');
	});
});

describe('eventValueText', () => {
	const design = EncounterDesign.create({
		downtimes: [eventAt('downtime', 0.5)],
		moves: [eventAt('move', 0.5)],
		bossHealth: [eventAt('bossHealth', 0.5)],
	});

	it('describes a span by its start and end, and a point by its time', () => {
		expect(eventValueText(design, { kind: 'downtime', index: 0 }, 180)).toBe('span_value(event_name(downtime,1),seconds(90),seconds(108))');
		expect(eventValueText(design, { kind: 'move', index: 0 }, 180)).toBe('point_value(event_name(move,1),seconds(90))');
		expect(eventValueText(design, { kind: 'bossHealth', index: 0 }, 180)).toBe('health_value(event_name(boss_health,1),100,50,seconds(0),seconds(90))');
	});

	it('reads a given window over the stored one', () => {
		expect(eventValueText(design, { kind: 'downtime', index: 0 }, 180, { start: 0.1, end: 0.2 })).toBe(
			'span_value(event_name(downtime,1),seconds(18),seconds(36))',
		);
	});
});

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

describe('eventTooltip', () => {
	const tip = (design: EncounterDesign, kind: Parameters<typeof eventName>[1]['kind']) => eventTooltip(design, { kind, index: 0 }, 180, ['Boss']);

	it("groups a wave's adds by level and mob type, counting a group of more than one", () => {
		const adds = [
			Target.create({ level: 60, mobType: MobType.MobTypeBeast }),
			Target.create({ level: 61, mobType: MobType.MobTypeDemon }),
			Target.create({ level: 60, mobType: MobType.MobTypeBeast }),
		];
		expect(tip(EncounterDesign.create({ addWaves: [{ ...defaultEvent('wave'), adds }] }), 'wave').lines[0]).toBe(
			`adds(3,add_group_count(60,mob${MobType.MobTypeBeast},2), add_group(61,mob${MobType.MobTypeDemon},1))`,
		);
	});

	it('describes a spell hit by what it deals, when it fires and how often', () => {
		const ability = {
			...eventAt('ability', 0.25),
			minDamage: 3000,
			maxDamage: 5000,
			castTime: 2,
			repeat: { interval: 25 / 180, jitter: 0, count: 0 },
		};
		ability.end!.at = 1;
		expect(tip(EncounterDesign.create({ bossAbilities: [ability] }), 'ability')).toEqual({
			title: 'spell_hit',
			lines: ['school_damage(0,3000,5000)', 'cast(seconds(2))', 'every_from(seconds(25),seconds(45),seconds(180))'],
		});
	});

	it('describes a window ability by its range', () => {
		const thrash = { ...eventAt('ability', 0.2), kind: DesignBossAbility_Kind.Thrash, procChance: 20 };
		expect(tip(EncounterDesign.create({ bossAbilities: [thrash] }), 'ability')).toEqual({
			title: 'thrash',
			lines: ['proc(20)', 'range(seconds(36),seconds(54))'],
		});
	});

	it('describes a wave, a move, a downtime, an amp and a boss health segment', () => {
		const wave = {
			...eventAt('wave', 0.1),
			adds: [Target.create({ level: 60 })],
			name: 'Add',
			spawn: { at: 10 / 180, jitter: 2 / 180 },
			lifetime: 20 / 180,
		};
		expect(tip(EncounterDesign.create({ addWaves: [wave] }), 'wave')).toEqual({
			title: 'wave(Add)',
			lines: ['adds(1,add_group(60,mob0,1))', 'spawns_jitter(seconds(10),seconds(2))', 'lives(seconds(20))', 'kill_share(25)'],
		});
		const move = { ...defaultEvent('move'), yards: 5, repeat: { interval: 10 / 180, jitter: 0, count: 0 } };
		expect(tip(EncounterDesign.create({ moves: [move] }), 'move')).toEqual({
			title: 'move',
			lines: ['at(seconds(72))', 'move_yards(5,seconds(0.7))', 'every(seconds(10))'],
		});
		expect(tip(EncounterDesign.create({ downtimes: [eventAt('downtime', 0.5)] }), 'downtime')).toEqual({
			title: 'downtime',
			lines: ['Boss', 'range(seconds(90),seconds(108))', 'clears_debuffs'],
		});
		expect(tip(EncounterDesign.create({ damageAmps: [eventAt('amp', 0.2)] }), 'amp')).toEqual({
			title: 'amp',
			lines: ['amp_value(caster_buff,10)', 'range(seconds(36),seconds(54))'],
		});
		const health = EncounterDesign.create({ bossHealth: [{ ...eventAt('bossHealth', 0.5), healthPercent: 65 }] });
		expect(tip(health, 'bossHealth')).toEqual({ title: 'event_name(boss_health,1)', lines: ['health(100,65)', 'range(seconds(0),seconds(90))'] });
	});

	it('reads a given window over the stored one', () => {
		const downtime = EncounterDesign.create({ downtimes: [eventAt('downtime', 0.5)] });
		expect(eventTooltip(downtime, { kind: 'downtime', index: 0 }, 180, ['Boss'], { start: 0.1, end: 0.2 }).lines).toContain(
			'range(seconds(18),seconds(36))',
		);
	});
});

describe('serializeTip', () => {
	it('keeps the title and each rule on its own line through the attribute', () => {
		const tip = { title: 'Adds #2 — Wave', lines: ['Lives 10s', 'Every 30s'] };
		expect(serializeTip(tip)).toBe('Adds #2 — Wave\nLives 10s\nEvery 30s');
		expect(parseTip(serializeTip(tip))).toEqual(tip);
		expect(parseTip(undefined)).toBeUndefined();
	});
});
