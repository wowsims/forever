import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, HealthKeyframe, PresetTarget } from '@generated/proto/common';
import { DamageAmpType, DesignBossAbility_Kind, DesignRepeat, DesignTargets, DesignTime, EncounterDesign, SpellSchool, Target } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSpellSchool } from '@i18n/localization';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';

import type { DesignEvent, DesignEventKind, DesignEventRef } from '../../../model/design/lanes';
import { eventLabelKey, eventWindow, getEvent } from '../../../model/design/lanes';
import type { TimeWindow } from '../../../model/design/time';
import { fractionToSeconds, secondsToFraction } from '../../../model/design/time';

export const designerText = (name: string, options?: Record<string, string | number>): string => i18n.t(`settings_tab.encounter.designer.${name}`, options);

export const formatSeconds = (seconds: number): string => designerText('seconds', { value: Number(seconds.toFixed(1)) });

export const eventName = (design: EncounterDesign, ref: DesignEventRef): string => {
	const named = ref.kind === 'wave' ? design.addWaves[ref.index]?.name : '';
	return named || designerText('event_name', { event: i18n.t(eventLabelKey(ref.kind)), number: ref.index + 1 });
};

// Reads the times from `window` when one is given, so a bar can describe where a drag has it.
export const eventValueText = (design: EncounterDesign, ref: DesignEventRef, baseSeconds: number, window?: TimeWindow): string => {
	const name = eventName(design, ref);
	const stored = eventWindow(design, ref);
	if (!stored) return name;
	const shown = window ?? ('at' in stored ? { start: stored.at, end: stored.at } : stored);
	const at = formatSeconds(fractionToSeconds(shown.start, baseSeconds));
	if (!('at' in stored)) return designerText('span_value', { name, start: at, end: formatSeconds(fractionToSeconds(shown.end, baseSeconds)) });
	if (ref.kind === 'bossHealth') return designerText('keyframe_value', { name, at, health: (getEvent(design, ref) as HealthKeyframe).healthPercent });
	return designerText('point_value', { name, at });
};

export const editEvent = (design: EncounterDesign, ref: DesignEventRef, edit: (event: DesignEvent) => void): EncounterDesign => {
	const next = EncounterDesign.clone(design);
	const event = getEvent(next, ref);
	if (event) edit(event);
	return next;
};

export interface DesignFieldContext {
	baseSeconds: number;
	targetNames: ReadonlyArray<string>;
	waveNames: ReadonlyArray<string>;
	presetTargets: ReadonlyArray<PresetTarget>;
}

interface FieldBase<E> {
	id: string;
	label: string;
	tooltip?: string;
	show?(event: E): boolean;
}

export type DesignFieldConfig<E> =
	| (FieldBase<E> & {
			type: 'number';
			float: boolean;
			get(event: E, context: DesignFieldContext): number;
			set(event: E, value: number, context: DesignFieldContext): void;
	  })
	| (FieldBase<E> & { type: 'boolean'; get(event: E): boolean; set(event: E, value: boolean): void })
	| (FieldBase<E> & { type: 'text'; get(event: E): string; set(event: E, value: string): void })
	| (FieldBase<E> & {
			type: 'choice';
			options(context: DesignFieldContext): Array<ChoiceOption<number>>;
			get(event: E, context: DesignFieldContext): number;
			set(event: E, value: number, context: DesignFieldContext): void;
	  })
	| (FieldBase<E> & {
			type: 'list';
			options(context: DesignFieldContext): Array<EnumValueConfig>;
			get(event: E): Array<number>;
			set(event: E, value: Array<number>): void;
	  });

const label = (name: string) => designerText(`fields.${name}`);
const tooltip = (name: string) => designerText(`tooltips.${name}`);

// Stored as a fraction of the fight, shown in seconds.
const seconds = <E>(
	name: string,
	read: (event: E) => number,
	write: (event: E, value: number) => void,
	extra: Partial<FieldBase<E>> = {},
): DesignFieldConfig<E> => ({
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
): DesignFieldConfig<E> => ({ type: 'number', float, id: name, label: label(name), ...extra, get: read, set: (event, value) => write(event, value) });

// Stored as 0..1, shown as a percentage.
const percent = <E>(
	name: string,
	read: (event: E) => number,
	write: (event: E, value: number) => void,
	extra: Partial<FieldBase<E>> = {},
): DesignFieldConfig<E> =>
	plain(
		name,
		event => read(event) * 100,
		(event, value) => write(event, value / 100),
		{ float: true, ...extra },
	);

type TimeSlot = 'start' | 'end' | 'spawn' | 'at';
type Timed<S extends TimeSlot> = { [P in S]?: DesignTime };

const timeFields = <S extends TimeSlot, E extends Timed<S>>(slot: S, names: { at: string; jitter: string }): Array<DesignFieldConfig<E>> => {
	const time = (event: Timed<S>): DesignTime => (event[slot] ??= DesignTime.create());
	return [
		seconds<E>(
			names.at,
			event => event[slot]?.at ?? 0,
			(event, value) => (time(event).at = value),
		),
		seconds<E>(
			names.jitter,
			event => event[slot]?.jitter ?? 0,
			(event, value) => (time(event).jitter = value),
			{ tooltip: tooltip('jitter') },
		),
	];
};

const spanFields = <E extends Timed<'start' | 'end'>>(): Array<DesignFieldConfig<E>> => [
	...timeFields<'start', E>('start', { at: 'start', jitter: 'start_jitter' }),
	...timeFields<'end', E>('end', { at: 'end', jitter: 'end_jitter' }),
];

const repeatFields = <E extends { repeat?: DesignRepeat }>(): Array<DesignFieldConfig<E>> => {
	const repeat = (event: E): DesignRepeat => (event.repeat ??= DesignRepeat.create());
	const repeats = (event: E) => (event.repeat?.interval ?? 0) > 0;
	return [
		seconds<E>(
			'repeat_interval',
			event => event.repeat?.interval ?? 0,
			(event, value) => (repeat(event).interval = value),
			{ tooltip: tooltip('repeat_interval') },
		),
		seconds<E>(
			'repeat_jitter',
			event => event.repeat?.jitter ?? 0,
			(event, value) => (repeat(event).jitter = value),
			{ tooltip: tooltip('jitter'), show: repeats },
		),
		plain<E>(
			'repeat_count',
			event => event.repeat?.count ?? 0,
			(event, value) => (repeat(event).count = Math.max(0, Math.round(value))),
			{
				tooltip: tooltip('repeat_count'),
				show: repeats,
			},
		),
	];
};

const targetField = <E extends { targetIndex: number }>(): DesignFieldConfig<E> => ({
	type: 'choice',
	id: 'target',
	label: label('target'),
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
): DesignFieldConfig<E> => ({
	type: 'choice',
	id: name,
	label: label(name),
	...extra,
	options: () => values.map(([value, text]) => ({ value, label: text })),
	get: read,
	set: (event, value) => write(event, value as V),
});

const AMP_TYPES: ReadonlyArray<[DamageAmpType, string]> = [
	[DamageAmpType.DamageAmpTypeCasterBuff, 'caster_buff'],
	[DamageAmpType.DamageAmpTypeEnvironmentBuff, 'environment_buff'],
	[DamageAmpType.DamageAmpTypeTargetDebuff, 'target_debuff'],
];

const ABILITY_KINDS: ReadonlyArray<[DesignBossAbility_Kind, string]> = [
	[DesignBossAbility_Kind.SpellHit, 'spell_hit'],
	[DesignBossAbility_Kind.Strike, 'strike'],
	[DesignBossAbility_Kind.Thrash, 'thrash'],
	[DesignBossAbility_Kind.SwingTiming, 'swing_timing'],
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

const downtimeFields = (): Array<DesignFieldConfig<DesignDowntime>> => [
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

const bossHealthFields = (): Array<DesignFieldConfig<HealthKeyframe>> => [
	seconds(
		'at',
		event => event.at,
		(event, value) => (event.at = value),
	),
	plain(
		'health_percent',
		event => event.healthPercent,
		(event, value) => (event.healthPercent = value),
		{ float: true },
	),
];

const DYNAMIC_ADD_PATH = 'Default/Dynamic Add';

// The preset target a new wave's adds copy: the Dynamic Add, else the first preset that pulls no tank, else the first
// preset. A tank index pulls a tank only when it names one of the raid's tanks, as the sim's tank setup reads it.
export const defaultWaveTemplate = (presetTargets: ReadonlyArray<PresetTarget>, tankCount: number): Target | undefined => {
	const pullsNoTank = ({ target }: PresetTarget) => !!target && (target.tankIndex < 0 || target.tankIndex >= tankCount);
	const preset = presetTargets.find(({ path, target }) => path === DYNAMIC_ADD_PATH && !!target) ?? presetTargets.find(pullsNoTank) ?? presetTargets[0];
	return preset?.target;
};

const waveFields = (): Array<DesignFieldConfig<AddWave>> => [
	{
		type: 'text',
		id: 'name',
		label: label('name'),
		get: event => event.name,
		set: (event, value) => {
			event.name = value;
		},
	},
	{
		type: 'choice',
		id: 'template',
		label: label('template'),
		tooltip: tooltip('template'),
		options: context => context.presetTargets.map((preset, value) => ({ value, label: preset.path })),
		get: (event, context) => context.presetTargets.findIndex(preset => !!event.target && !!preset.target && Target.equals(event.target, preset.target)),
		set: (event, value, context) => {
			const preset = context.presetTargets[value]?.target;
			if (preset) event.target = Target.clone(preset);
		},
	},
	plain(
		'count',
		event => event.count,
		(event, value) => (event.count = Math.max(1, Math.round(value))),
	),
	...timeFields<'spawn', AddWave>('spawn', { at: 'spawn', jitter: 'spawn_jitter' }),
	seconds(
		'lifetime',
		event => event.lifetime,
		(event, value) => (event.lifetime = value),
		{ tooltip: tooltip('lifetime') },
	),
	...repeatFields<AddWave>(),
	percent(
		'execute_20',
		event => event.executeProportion20,
		(event, value) => (event.executeProportion20 = value),
		{ tooltip: tooltip('execute_20') },
	),
	percent(
		'execute_35',
		event => event.executeProportion35,
		(event, value) => (event.executeProportion35 = value),
		{ tooltip: tooltip('execute_35') },
	),
	percent(
		'kill_share',
		event => event.killShare,
		(event, value) => (event.killShare = value),
		{ tooltip: tooltip('kill_share') },
	),
];

const moveFields = (): Array<DesignFieldConfig<DesignMove>> => [
	...timeFields<'at', DesignMove>('at', { at: 'at', jitter: 'at_jitter' }),
	plain(
		'yards',
		event => event.yards,
		(event, value) => (event.yards = value),
		{ float: true },
	),
	plain(
		'seconds',
		event => event.seconds,
		(event, value) => (event.seconds = value),
		{ float: true, tooltip: tooltip('seconds'), show: event => event.yards === 0 },
	),
	...repeatFields<DesignMove>(),
];

const isTargetDebuff = (event: DesignDamageAmp) => event.type === DamageAmpType.DamageAmpTypeTargetDebuff;

const ampFields = (): Array<DesignFieldConfig<DesignDamageAmp>> => [
	enumField<DesignDamageAmp, DamageAmpType>(
		'amp_type',
		AMP_TYPES.map(([value, name]) => [value, designerText(`amp_types.${name}`)]),
		event => event.type,
		(event, value) => (event.type = value),
	),
	plain(
		'percent',
		event => event.percent,
		(event, value) => (event.percent = value),
		{ float: true },
	),
	...spanFields<DesignDamageAmp>(),
	...repeatFields<DesignDamageAmp>(),
	{
		type: 'list',
		id: 'targets',
		label: label('targets'),
		tooltip: tooltip('targets'),
		show: isTargetDebuff,
		options: context => context.targetNames.map((name, value) => ({ name, value })),
		get: event => event.targets?.targetIndices ?? [],
		set: (event, value) => {
			(event.targets ??= DesignTargets.create()).targetIndices = value;
		},
	},
	{
		type: 'list',
		id: 'waves',
		label: label('waves'),
		tooltip: tooltip('waves'),
		show: isTargetDebuff,
		options: context => context.waveNames.map((name, value) => ({ name, value })),
		get: event => event.targets?.waveIndices ?? [],
		set: (event, value) => {
			(event.targets ??= DesignTargets.create()).waveIndices = value;
		},
	},
];

const ofKind =
	(...kinds: Array<DesignBossAbility_Kind>) =>
	(event: DesignBossAbility) =>
		kinds.includes(event.kind);

const abilityFields = (): Array<DesignFieldConfig<DesignBossAbility>> => {
	const spellHit = ofKind(DesignBossAbility_Kind.SpellHit);
	const thrash = ofKind(DesignBossAbility_Kind.Thrash);
	return [
		enumField<DesignBossAbility, DesignBossAbility_Kind>(
			'ability_kind',
			ABILITY_KINDS.map(([value, name]) => [value, designerText(`ability_kinds.${name}`)]),
			event => event.kind,
			(event, value) => (event.kind = value),
		),
		targetField(),
		...spanFields<DesignBossAbility>(),
		...repeatFields<DesignBossAbility>(),
		enumField<DesignBossAbility, SpellSchool>(
			'school',
			SCHOOLS.map(school => [school, translateSpellSchool(school)]),
			event => event.school,
			(event, value) => (event.school = value),
			{ show: spellHit },
		),
		plain(
			'min_damage',
			event => event.minDamage,
			(event, value) => (event.minDamage = value),
			{ show: spellHit },
		),
		plain(
			'max_damage',
			event => event.maxDamage,
			(event, value) => (event.maxDamage = value),
			{ show: spellHit },
		),
		plain(
			'cast_time',
			event => event.castTime,
			(event, value) => (event.castTime = value),
			{ float: true, show: spellHit },
		),
		plain(
			'weapon_percent',
			event => event.weaponPercent,
			(event, value) => (event.weaponPercent = value),
			{
				float: true,
				show: ofKind(DesignBossAbility_Kind.Strike),
			},
		),
		plain(
			'proc_chance',
			event => event.procChance,
			(event, value) => (event.procChance = value),
			{ float: true, show: thrash },
		),
		plain(
			'extra_attacks',
			event => event.extraAttacks,
			(event, value) => (event.extraAttacks = Math.max(0, Math.round(value))),
			{ show: thrash },
		),
		plain(
			'icd',
			event => event.icd,
			(event, value) => (event.icd = value),
			{ float: true, show: thrash },
		),
		plain(
			'swing_speed_multiplier',
			event => event.swingSpeedMultiplier,
			(event, value) => (event.swingSpeedMultiplier = value),
			{
				float: true,
				tooltip: tooltip('swing_speed_multiplier'),
				show: ofKind(DesignBossAbility_Kind.SwingTiming),
			},
		),
	];
};

const FIELDS: { [K in DesignEventKind]: () => Array<DesignFieldConfig<DesignEvent<K>>> } = {
	downtime: downtimeFields,
	bossHealth: bossHealthFields,
	wave: waveFields,
	move: moveFields,
	amp: ampFields,
	ability: abilityFields,
};

export const designFieldConfigs = <K extends DesignEventKind>(kind: K): Array<DesignFieldConfig<DesignEvent<K>>> => FIELDS[kind]();
