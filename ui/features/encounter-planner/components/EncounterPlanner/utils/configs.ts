import type { StringFieldConfig } from '@features/inputs/configs';
import type { AddWave, HealthKeyframe, PlanBossAbility, PlanDamageAmp, PlanDowntime, PlanMove, PresetEncounter, Target } from '@generated/proto/common';
import {
	DamageAmpType,
	EncounterPlan,
	PlanBossAbility_Kind,
	PlanBossAbility_Source,
	PlanRepeat,
	PlanTargets,
	PlanTime,
	PresetTarget,
	SpellSchool,
} from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSpellSchool } from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';

import type { PlanEvent, PlanEventKind, PlanEventRef } from '../../../model/lanes';
import { firesAtInstants, getEvent, isBossCastArea, withAbilityKind } from '../../../model/lanes';
import { fractionToSeconds, secondsToFraction } from '../../../model/time';
import { PLAN_ISSUE_KEYS } from '../../../model/validate';

export const plannerText = (name: string, options?: Record<string, string | number>): string =>
	options ? i18n.t(`settings_tab.encounter.planner.${name}`, options) : i18n.t(`settings_tab.encounter.planner.${name}`);

export const nameConfig = (): StringFieldConfig<Encounter> => ({
	id: 'encounter-planner-name',
	label: plannerText('name'),
	storeField: 'encounter:name',
	getValue: encounter => encounter.getName(),
	setValue: (encounter, name) => encounter.setName(name),
});

export const editEvent = (plan: EncounterPlan, ref: PlanEventRef, edit: (event: PlanEvent) => void): EncounterPlan => {
	const next = EncounterPlan.clone(plan);
	const event = getEvent(next, ref);
	if (event) edit(event);
	return next;
};

export interface PlanFieldContext {
	baseSeconds: number;
	targetNames: ReadonlyArray<string>;
	waveNames: ReadonlyArray<string>;
}

// The event panel's sections, in the order it shows them. A field with no section sits at the top, with the fields that
// say what the event is.
export const PLAN_SECTIONS = ['timing', 'repeat', 'damage', 'effect', 'execute', 'debuffs', 'adds'] as const;
export type PlanSection = (typeof PLAN_SECTIONS)[number];

interface FieldBase<E> {
	id: string;
	label: string;
	tooltip?: string;
	section?: PlanSection;
	show?(event: E): boolean;
}

export type PlanFieldConfig<E> =
	| (FieldBase<E> & {
			type: 'number';
			float: boolean;
			get(event: E, context: PlanFieldContext): number;
			set(event: E, value: number, context: PlanFieldContext): void;
	  })
	| (FieldBase<E> & { type: 'boolean'; get(event: E): boolean; set(event: E, value: boolean): void })
	| (FieldBase<E> & { type: 'text'; get(event: E): string; set(event: E, value: string): void })
	| (FieldBase<E> & {
			type: 'choice';
			options(context: PlanFieldContext): Array<ChoiceOption<number>>;
			get(event: E): number;
			set(event: E, value: number): void;
	  })
	| (FieldBase<E> & {
			type: 'list';
			options(context: PlanFieldContext): Array<EnumValueConfig>;
			get(event: E): Array<number>;
			set(event: E, value: Array<number>): void;
	  });

const label = (name: string) => plannerText(`fields.${name}`);
const tooltip = (name: string) => plannerText(`tooltips.${name}`);

type NumberKey<E> = { [K in keyof E]-?: E[K] extends number ? K : never }[keyof E];

const prop = <E>(key: NumberKey<E>): [(event: E) => number, (event: E, value: number) => void] => [
	event => event[key] as number,
	(event, value) => {
		(event[key] as number) = value;
	},
];

// Stored as a fraction of the fight, shown in seconds.
const seconds = <E>(
	name: string,
	read: (event: E) => number,
	write: (event: E, value: number) => void,
	extra: Partial<FieldBase<E>> = {},
): PlanFieldConfig<E> => ({
	type: 'number',
	float: true,
	id: name,
	label: label(name),
	...extra,
	get: (event, context) => fractionToSeconds(read(event), context.baseSeconds),
	set: (event, value, context) => write(event, secondsToFraction(value, context.baseSeconds)),
});

const plain = <E>(
	name: string,
	read: (event: E) => number,
	write: (event: E, value: number) => void,
	{ float = false, ...extra }: Partial<FieldBase<E>> & { float?: boolean } = {},
): PlanFieldConfig<E> => ({ type: 'number', float, id: name, label: label(name), ...extra, get: read, set: (event, value) => write(event, value) });

// Stored as 0..1, shown as a percentage.
const percent = <E>(
	name: string,
	read: (event: E) => number,
	write: (event: E, value: number) => void,
	extra: Partial<FieldBase<E>> = {},
): PlanFieldConfig<E> =>
	plain(
		name,
		event => read(event) * 100,
		(event, value) => write(event, value / 100),
		{ float: true, ...extra },
	);

type TimeSlot = 'start' | 'end' | 'spawn' | 'at';
type Timed<S extends TimeSlot> = { [P in S]?: PlanTime };

const timeFields = <S extends TimeSlot, E extends Timed<S>>(
	slot: S,
	names: { at: string; jitter: string },
	extra: Partial<FieldBase<E>> = {},
): Array<PlanFieldConfig<E>> => {
	const time = (event: Timed<S>): PlanTime => (event[slot] ??= PlanTime.create());
	return [
		seconds<E>(
			names.at,
			event => event[slot]?.at ?? 0,
			(event, value) => (time(event).at = value),
			extra,
		),
		seconds<E>(
			names.jitter,
			event => event[slot]?.jitter ?? 0,
			(event, value) => (time(event).jitter = value),
			{ ...extra, tooltip: tooltip('jitter') },
		),
	];
};

const spanFields = <E extends Timed<'start' | 'end'>>(extra: Partial<FieldBase<E>> = {}): Array<PlanFieldConfig<E>> => [
	...timeFields<'start', E>('start', { at: 'start', jitter: 'start_jitter' }, extra),
	...timeFields<'end', E>('end', { at: 'end', jitter: 'end_jitter' }, extra),
];

const repeatFields = <E extends { repeat?: PlanRepeat }>(extra: Partial<FieldBase<E>> = {}): Array<PlanFieldConfig<E>> => {
	const repeat = (event: E): PlanRepeat => (event.repeat ??= PlanRepeat.create());
	const repeats = (event: E) => (event.repeat?.interval ?? 0) > 0;
	return [
		seconds<E>(
			'repeat_interval',
			event => event.repeat?.interval ?? 0,
			(event, value) => (repeat(event).interval = value),
			{ ...extra, tooltip: tooltip('repeat_interval') },
		),
		seconds<E>(
			'repeat_jitter',
			event => event.repeat?.jitter ?? 0,
			(event, value) => (repeat(event).jitter = value),
			{ ...extra, tooltip: tooltip('jitter'), show: repeats },
		),
		plain<E>(
			'repeat_count',
			event => event.repeat?.count ?? 0,
			(event, value) => (repeat(event).count = Math.max(0, Math.round(value))),
			{
				...extra,
				tooltip: tooltip('repeat_count'),
				show: repeats,
			},
		),
	];
};

const targetField = <E extends { targetIndex: number }>(extra: Partial<FieldBase<E>> = {}): PlanFieldConfig<E> => ({
	type: 'choice',
	id: 'target',
	label: label('target'),
	...extra,
	options: context => context.targetNames.map((name, value) => ({ value, label: name })),
	get: event => event.targetIndex,
	set: (event, value) => {
		event.targetIndex = value;
	},
});

const enumField = <E, V extends number>(
	name: string,
	values: ReadonlyArray<[V, string]>,
	read: (event: E) => V,
	write: (event: E, value: V) => void,
	extra: Partial<FieldBase<E>> = {},
): PlanFieldConfig<E> => ({
	type: 'choice',
	id: name,
	label: label(name),
	...extra,
	options: () => values.map(([value, text]) => ({ value, label: text })),
	get: read,
	set: (event, value) => write(event, value as V),
});

export const AMP_TYPES: ReadonlyArray<[DamageAmpType, string]> = [
	[DamageAmpType.DamageAmpTypeCasterBuff, 'caster_buff'],
	[DamageAmpType.DamageAmpTypeEnvironmentBuff, 'environment_buff'],
	[DamageAmpType.DamageAmpTypeTargetDebuff, 'target_debuff'],
];

export const ABILITY_KINDS: ReadonlyArray<[PlanBossAbility_Kind, string]> = [
	[PlanBossAbility_Kind.SpellHit, 'spell_hit'],
	[PlanBossAbility_Kind.Strike, 'strike'],
	[PlanBossAbility_Kind.Thrash, 'thrash'],
	[PlanBossAbility_Kind.SwingTiming, 'swing_timing'],
	[PlanBossAbility_Kind.Pulse, 'pulse'],
	[PlanBossAbility_Kind.AreaHit, 'area_hit'],
];

const ABILITY_SOURCES: ReadonlyArray<[PlanBossAbility_Source, string]> = [
	[PlanBossAbility_Source.Environment, 'environment'],
	[PlanBossAbility_Source.Boss, 'boss'],
];

const SCHOOLS: ReadonlyArray<SpellSchool> = [
	SpellSchool.SpellSchoolPhysical,
	SpellSchool.SpellSchoolArcane,
	SpellSchool.SpellSchoolFire,
	SpellSchool.SpellSchoolFrost,
	SpellSchool.SpellSchoolHoly,
	SpellSchool.SpellSchoolNature,
	SpellSchool.SpellSchoolShadow,
];

const nameField = <E extends { name: string }>(): PlanFieldConfig<E> => ({
	type: 'text',
	id: 'name',
	label: label('name'),
	get: event => event.name,
	set: (event, value) => {
		event.name = value;
	},
});

const downtimeFields = (): Array<PlanFieldConfig<PlanDowntime>> => [
	nameField(),
	targetField(),
	...spanFields(),
	{
		type: 'boolean',
		id: 'clear_debuffs',
		label: label('clear_debuffs'),
		tooltip: tooltip('clear_debuffs'),
		get: event => event.clearDebuffs,
		set: (event, value) => {
			event.clearDebuffs = value;
		},
	},
];

const bossHealthFields = (): Array<PlanFieldConfig<HealthKeyframe>> => [
	nameField(),
	seconds('at', ...prop<HealthKeyframe>('at')),
	plain('health_percent', ...prop<HealthKeyframe>('healthPercent'), { float: true }),
];

// The preset targets, and the add each preset encounter's waves copy: a wave's add is no target of the encounter, so the
// targets alone leave it out. An add is listed under its encounter's folder, by its name.
export const plannerPresetTargets = (encounters: ReadonlyArray<PresetEncounter>, targets: ReadonlyArray<PresetTarget>): Array<PresetTarget> => {
	const byPath = new Map(targets.map(preset => [preset.path, preset]));
	for (const encounter of encounters) {
		const folder = encounter.path.slice(0, Math.max(0, encounter.path.indexOf('/')));
		for (const add of encounter.plan?.addWaves.flatMap(wave => wave.adds) ?? []) {
			const path = `${folder}/${add.name}`;
			if (!byPath.has(path)) byPath.set(path, PresetTarget.create({ path, target: add }));
		}
	}
	return [...byPath.values()];
};

const DYNAMIC_ADD_PATH = 'Default/Dynamic Add';

// The preset target a new wave's adds copy: the Dynamic Add, else the first preset that pulls no tank, else the first
// preset. A tank index pulls a tank only when it names one of the raid's tanks, as the sim's tank setup reads it.
export const defaultWaveTemplate = (presetTargets: ReadonlyArray<PresetTarget>, tankCount: number): Target | undefined => {
	const pullsNoTank = ({ target }: PresetTarget) => !!target && (target.tankIndex < 0 || target.tankIndex >= tankCount);
	const preset = presetTargets.find(({ path, target }) => path === DYNAMIC_ADD_PATH && !!target) ?? presetTargets.find(pullsNoTank) ?? presetTargets[0];
	return preset?.target;
};

const appliesDebuffs = (event: AddWave) => !event.holdDebuffs;

// The wave's adds are not fields: the panel lists them in `PlannerAdds`, each with its template and stats.
const waveFields = (): Array<PlanFieldConfig<AddWave>> => [
	nameField(),
	...timeFields<'spawn', AddWave>('spawn', { at: 'spawn', jitter: 'spawn_jitter' }, { section: 'timing' }),
	seconds('lifetime', ...prop<AddWave>('lifetime'), { tooltip: tooltip('lifetime'), section: 'timing' }),
	...repeatFields<AddWave>({ section: 'repeat' }),
	percent('execute_20', ...prop<AddWave>('executeProportion20'), { tooltip: tooltip('execute_20'), section: 'execute' }),
	percent('execute_35', ...prop<AddWave>('executeProportion35'), { tooltip: tooltip('execute_35'), section: 'execute' }),
	percent('kill_share', ...prop<AddWave>('killShare'), { tooltip: tooltip('kill_share'), section: 'execute' }),
	{
		type: 'boolean',
		id: 'apply_debuffs',
		label: label('apply_debuffs'),
		tooltip: tooltip('apply_debuffs'),
		section: 'debuffs',
		get: appliesDebuffs,
		set: (event, value) => {
			event.holdDebuffs = !value;
		},
	},
	plain('debuff_delay', ...prop<AddWave>('debuffDelay'), { float: true, tooltip: tooltip('debuff_delay'), section: 'debuffs', show: appliesDebuffs }),
	plain('debuff_jitter', ...prop<AddWave>('debuffJitter'), { float: true, tooltip: tooltip('debuff_jitter'), section: 'debuffs', show: appliesDebuffs }),
];

const moveFields = (): Array<PlanFieldConfig<PlanMove>> => [
	nameField(),
	...timeFields<'at', PlanMove>('at', { at: 'at', jitter: 'at_jitter' }),
	plain('yards', ...prop<PlanMove>('yards'), { float: true }),
	plain('seconds', ...prop<PlanMove>('seconds'), { float: true, tooltip: tooltip('seconds'), show: event => event.yards === 0 }),
	...repeatFields<PlanMove>(),
];

const isTargetDebuff = (event: PlanDamageAmp) => event.type === DamageAmpType.DamageAmpTypeTargetDebuff;

const ampFields = (): Array<PlanFieldConfig<PlanDamageAmp>> => [
	nameField(),
	enumField<PlanDamageAmp, DamageAmpType>(
		'amp_type',
		AMP_TYPES.map(([value, name]) => [value, plannerText(`amp_types.${name}`)]),
		event => event.type,
		(event, value) => (event.type = value),
	),
	plain('percent', ...prop<PlanDamageAmp>('percent'), { float: true, section: 'effect' }),
	...spanFields<PlanDamageAmp>({ section: 'timing' }),
	...repeatFields<PlanDamageAmp>({ section: 'repeat' }),
	{
		type: 'list',
		id: 'targets',
		label: label('targets'),
		tooltip: tooltip('targets'),
		section: 'effect',
		show: isTargetDebuff,
		options: context => context.targetNames.map((name, value) => ({ name, value })),
		get: event => event.targets?.targetIndices ?? [],
		set: (event, value) => {
			(event.targets ??= PlanTargets.create()).targetIndices = value;
		},
	},
	{
		type: 'list',
		id: 'waves',
		label: label('waves'),
		tooltip: tooltip('waves'),
		section: 'effect',
		show: isTargetDebuff,
		options: context => context.waveNames.map((name, value) => ({ name, value })),
		get: event => event.targets?.waveIndices ?? [],
		set: (event, value) => {
			(event.targets ??= PlanTargets.create()).waveIndices = value;
		},
	},
];

const ofKind =
	(...kinds: Array<PlanBossAbility_Kind>) =>
	(event: PlanBossAbility) =>
		kinds.includes(event.kind);

const abilityFields = (): Array<PlanFieldConfig<PlanBossAbility>> => {
	const spellHit = ofKind(PlanBossAbility_Kind.SpellHit);
	const thrash = ofKind(PlanBossAbility_Kind.Thrash);
	const pulse = ofKind(PlanBossAbility_Kind.Pulse);
	const area = ofKind(PlanBossAbility_Kind.Pulse, PlanBossAbility_Kind.AreaHit);
	const casts = (event: PlanBossAbility) => spellHit(event) || isBossCastArea(event);
	const damage = ofKind(PlanBossAbility_Kind.SpellHit, PlanBossAbility_Kind.Pulse, PlanBossAbility_Kind.AreaHit);
	const toggle = (name: 'resistible' | 'avoidable'): PlanFieldConfig<PlanBossAbility> => ({
		type: 'boolean',
		id: name,
		label: label(name),
		tooltip: tooltip(name),
		section: 'damage',
		show: area,
		get: event => event[name],
		set: (event, value) => {
			event[name] = value;
		},
	});
	return [
		nameField(),
		enumField<PlanBossAbility, PlanBossAbility_Kind>(
			'ability_kind',
			ABILITY_KINDS.map(([value, name]) => [value, plannerText(`ability_kinds.${name}`)]),
			event => event.kind,
			withAbilityKind,
		),
		enumField<PlanBossAbility, PlanBossAbility_Source>(
			'source',
			ABILITY_SOURCES.map(([value, name]) => [value, plannerText(`ability_sources.${name}`)]),
			event => event.source,
			(event, value) => (event.source = value),
			{ tooltip: tooltip('source'), show: area },
		),
		targetField({ show: event => !area(event) || isBossCastArea(event) }),
		...timeFields<'start', PlanBossAbility>('start', { at: 'start', jitter: 'start_jitter' }, { section: 'timing' }),
		...timeFields<'end', PlanBossAbility>('end', { at: 'end', jitter: 'end_jitter' }, { section: 'timing', show: event => !firesAtInstants(event) }),
		...repeatFields<PlanBossAbility>({ section: 'repeat' }),
		enumField<PlanBossAbility, SpellSchool>(
			'school',
			SCHOOLS.map(school => [school, translateSpellSchool(school)]),
			event => event.school,
			(event, value) => (event.school = value),
			{ section: 'damage', show: damage },
		),
		plain('min_damage', ...prop<PlanBossAbility>('minDamage'), { section: 'damage', show: damage }),
		plain('max_damage', ...prop<PlanBossAbility>('maxDamage'), { section: 'damage', show: damage }),
		plain('tick_interval', ...prop<PlanBossAbility>('tickInterval'), { float: true, tooltip: tooltip('tick_interval'), section: 'damage', show: pulse }),
		plain('tick_jitter', ...prop<PlanBossAbility>('tickJitter'), { float: true, tooltip: tooltip('tick_jitter'), section: 'damage', show: pulse }),
		toggle('resistible'),
		toggle('avoidable'),
		plain('cast_time', ...prop<PlanBossAbility>('castTime'), { float: true, section: 'timing', show: casts }),
		{
			type: 'boolean',
			id: 'stops_swing',
			label: label('stops_swing'),
			tooltip: tooltip('stops_swing'),
			section: 'timing',
			show: casts,
			get: event => !event.keepsSwingTimer,
			set: (event, value) => {
				event.keepsSwingTimer = !value;
			},
		},
		plain('weapon_percent', ...prop<PlanBossAbility>('weaponPercent'), { float: true, section: 'damage', show: ofKind(PlanBossAbility_Kind.Strike) }),
		plain('proc_chance', ...prop<PlanBossAbility>('procChance'), { float: true, section: 'effect', show: thrash }),
		plain(
			'extra_attacks',
			event => event.extraAttacks,
			(event, value) => (event.extraAttacks = Math.max(0, Math.round(value))),
			{ section: 'effect', show: thrash },
		),
		plain('icd', ...prop<PlanBossAbility>('icd'), { float: true, section: 'effect', show: thrash }),
		plain('swing_speed_multiplier', ...prop<PlanBossAbility>('swingSpeedMultiplier'), {
			float: true,
			tooltip: tooltip('swing_speed_multiplier'),
			section: 'effect',
			show: ofKind(PlanBossAbility_Kind.SwingTiming),
		}),
	];
};

// The section of the field each issue is about. An issue about a field at the top, or on a kind the panel shows
// without sections, opens none.
const ISSUE_SECTIONS: Record<keyof typeof PLAN_ISSUE_KEYS, PlanSection | undefined> = {
	endBeforeStart: 'timing',
	outsideFight: 'timing',
	overlap: undefined,
	downtimeOverlap: undefined,
	healthIncreases: undefined,
	targetOutOfRange: undefined,
	waveOutOfRange: 'effect',
	allTargetsDown: undefined,
	tooManyKeyframes: undefined,
	ampNotPositive: 'effect',
	tooManyAdds: 'adds',
	noAdds: 'adds',
	moveGoesNowhere: undefined,
	executeOrder: 'execute',
	noWeaponDamage: 'damage',
	noExtraAttacks: 'effect',
	noDamage: 'damage',
	minAboveMax: 'damage',
	noTickInterval: 'damage',
	// It only fires on a wave that repeats: one occurrence over the slots is too many adds.
	waveOverPool: 'repeat',
	healthStopsShort: undefined,
};

export const issueSection = (kind: PlanEventKind, issueKey: string): PlanSection | undefined => {
	const name = (Object.keys(PLAN_ISSUE_KEYS) as Array<keyof typeof PLAN_ISSUE_KEYS>).find(entry => PLAN_ISSUE_KEYS[entry] === issueKey);
	if (!name) return undefined;
	// A damage amp's target picker sits in its Effect section; every other target field sits at the top.
	if (name === 'targetOutOfRange' && kind === 'amp') return 'effect';
	return ISSUE_SECTIONS[name];
};

const FIELDS: { [K in PlanEventKind]: () => Array<PlanFieldConfig<PlanEvent<K>>> } = {
	downtime: downtimeFields,
	bossHealth: bossHealthFields,
	wave: waveFields,
	move: moveFields,
	amp: ampFields,
	ability: abilityFields,
};

export const planFieldConfigs = <K extends PlanEventKind>(kind: K): Array<PlanFieldConfig<PlanEvent<K>>> => FIELDS[kind]();
