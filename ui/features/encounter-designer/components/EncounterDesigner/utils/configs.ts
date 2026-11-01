import type { StringFieldConfig } from '@features/inputs/configs';
import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, HealthKeyframe, PresetEncounter, Target } from '@generated/proto/common';
import {
	DamageAmpType,
	DesignBossAbility_Kind,
	DesignRepeat,
	DesignTargets,
	DesignTime,
	EncounterDesign,
	PresetTarget,
	SpellSchool,
} from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateSpellSchool } from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';

import type { DesignEvent, DesignEventKind, DesignEventRef } from '../../../model/lanes';
import { getEvent } from '../../../model/lanes';
import { fractionToSeconds, secondsToFraction } from '../../../model/time';

export const designerText = (name: string, options?: Record<string, string | number>): string => i18n.t(`settings_tab.encounter.designer.${name}`, options);

export const nameConfig = (): StringFieldConfig<Encounter> => ({
	id: 'encounter-designer-name',
	label: designerText('name'),
	storeField: 'encounter:name',
	getValue: encounter => encounter.getName(),
	setValue: (encounter, name) => encounter.setName(name),
});

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
			get(event: E): number;
			set(event: E, value: number): void;
	  })
	| (FieldBase<E> & {
			type: 'list';
			options(context: DesignFieldContext): Array<EnumValueConfig>;
			get(event: E): Array<number>;
			set(event: E, value: Array<number>): void;
	  });

const label = (name: string) => designerText(`fields.${name}`);
const tooltip = (name: string) => designerText(`tooltips.${name}`);

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

export const AMP_TYPES: ReadonlyArray<[DamageAmpType, string]> = [
	[DamageAmpType.DamageAmpTypeCasterBuff, 'caster_buff'],
	[DamageAmpType.DamageAmpTypeEnvironmentBuff, 'environment_buff'],
	[DamageAmpType.DamageAmpTypeTargetDebuff, 'target_debuff'],
];

export const ABILITY_KINDS: ReadonlyArray<[DesignBossAbility_Kind, string]> = [
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

const nameField = <E extends { name: string }>(): DesignFieldConfig<E> => ({
	type: 'text',
	id: 'name',
	label: label('name'),
	get: event => event.name,
	set: (event, value) => {
		event.name = value;
	},
});

const downtimeFields = (): Array<DesignFieldConfig<DesignDowntime>> => [
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

const bossHealthFields = (): Array<DesignFieldConfig<HealthKeyframe>> => [
	nameField(),
	seconds('at', ...prop<HealthKeyframe>('at')),
	plain('health_percent', ...prop<HealthKeyframe>('healthPercent'), { float: true }),
];

// The preset targets, and the add each preset encounter's waves copy: a wave's add is no target of the encounter, so the
// targets alone leave it out. An add is listed under its encounter's folder, by its name.
export const designerPresetTargets = (encounters: ReadonlyArray<PresetEncounter>, targets: ReadonlyArray<PresetTarget>): Array<PresetTarget> => {
	const byPath = new Map(targets.map(preset => [preset.path, preset]));
	for (const encounter of encounters) {
		const folder = encounter.path.slice(0, Math.max(0, encounter.path.indexOf('/')));
		for (const add of encounter.design?.addWaves.flatMap(wave => wave.adds) ?? []) {
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

// The wave's adds are not fields: the panel lists them in `DesignerAdds`, each with its template and stats.
const waveFields = (): Array<DesignFieldConfig<AddWave>> => [
	nameField(),
	...timeFields<'spawn', AddWave>('spawn', { at: 'spawn', jitter: 'spawn_jitter' }),
	seconds('lifetime', ...prop<AddWave>('lifetime'), { tooltip: tooltip('lifetime') }),
	...repeatFields<AddWave>(),
	percent('execute_20', ...prop<AddWave>('executeProportion20'), { tooltip: tooltip('execute_20') }),
	percent('execute_35', ...prop<AddWave>('executeProportion35'), { tooltip: tooltip('execute_35') }),
	percent('kill_share', ...prop<AddWave>('killShare'), { tooltip: tooltip('kill_share') }),
];

const moveFields = (): Array<DesignFieldConfig<DesignMove>> => [
	nameField(),
	...timeFields<'at', DesignMove>('at', { at: 'at', jitter: 'at_jitter' }),
	plain('yards', ...prop<DesignMove>('yards'), { float: true }),
	plain('seconds', ...prop<DesignMove>('seconds'), { float: true, tooltip: tooltip('seconds'), show: event => event.yards === 0 }),
	...repeatFields<DesignMove>(),
];

const isTargetDebuff = (event: DesignDamageAmp) => event.type === DamageAmpType.DamageAmpTypeTargetDebuff;

const ampFields = (): Array<DesignFieldConfig<DesignDamageAmp>> => [
	nameField(),
	enumField<DesignDamageAmp, DamageAmpType>(
		'amp_type',
		AMP_TYPES.map(([value, name]) => [value, designerText(`amp_types.${name}`)]),
		event => event.type,
		(event, value) => (event.type = value),
	),
	plain('percent', ...prop<DesignDamageAmp>('percent'), { float: true }),
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
		nameField(),
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
		plain('min_damage', ...prop<DesignBossAbility>('minDamage'), { show: spellHit }),
		plain('max_damage', ...prop<DesignBossAbility>('maxDamage'), { show: spellHit }),
		plain('cast_time', ...prop<DesignBossAbility>('castTime'), { float: true, show: spellHit }),
		plain('weapon_percent', ...prop<DesignBossAbility>('weaponPercent'), { float: true, show: ofKind(DesignBossAbility_Kind.Strike) }),
		plain('proc_chance', ...prop<DesignBossAbility>('procChance'), { float: true, show: thrash }),
		plain(
			'extra_attacks',
			event => event.extraAttacks,
			(event, value) => (event.extraAttacks = Math.max(0, Math.round(value))),
			{ show: thrash },
		),
		plain('icd', ...prop<DesignBossAbility>('icd'), { float: true, show: thrash }),
		plain('swing_speed_multiplier', ...prop<DesignBossAbility>('swingSpeedMultiplier'), {
			float: true,
			tooltip: tooltip('swing_speed_multiplier'),
			show: ofKind(DesignBossAbility_Kind.SwingTiming),
		}),
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
