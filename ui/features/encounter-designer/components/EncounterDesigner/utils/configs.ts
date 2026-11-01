import type { StringFieldConfig } from '@features/inputs/configs';
import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, HealthKeyframe, PresetEncounter } from '@generated/proto/common';
import {
	DamageAmpType,
	DesignBossAbility_Kind,
	DesignRepeat,
	DesignTargets,
	DesignTime,
	EncounterDesign,
	PresetTarget,
	SpellSchool,
	Target,
} from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateMobType, translateSpellSchool } from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';
import type { ChoiceOption } from '@ui-kit/ChoiceField';
import type { EnumValueConfig } from '@ui-kit/MultiComboBox';

import type { DesignEvent, DesignEventKind, DesignEventRef } from '../../../model/lanes';
import { eventLabelKey, eventWindow, firesAtInstants, getEvent, healthSegments, moveSeconds } from '../../../model/lanes';
import type { TimeWindow } from '../../../model/time';
import { fractionToSeconds, secondsToFraction } from '../../../model/time';

export const designerText = (name: string, options?: Record<string, string | number>): string => i18n.t(`settings_tab.encounter.designer.${name}`, options);

export const nameConfig = (): StringFieldConfig<Encounter> => ({
	id: 'encounter-designer-name',
	label: designerText('name'),
	storeField: 'encounter:name',
	getValue: encounter => encounter.getName(),
	setValue: (encounter, name) => encounter.setName(name),
});

// A health percent as read, to a tenth: a hold across a downtime can start at a third of the way down a line.
export const formatPercent = (percent: number): number => Number(percent.toFixed(1));

export const formatSeconds = (seconds: number): string => designerText('seconds', { value: Number(seconds.toFixed(1)) });

export const eventName = (design: EncounterDesign, ref: DesignEventRef): string => {
	const named = getEvent(design, ref)?.name;
	return named || designerText('event_name', { event: i18n.t(eventLabelKey(ref.kind)), number: ref.index + 1 });
};

// The first occurrence keeps the bare name; the repeats after it count from 2.
export const occurrenceName = (name: string, occurrence: number): string =>
	occurrence > 1 ? designerText('occurrence_name', { name, number: occurrence }) : name;

// Reads the times from `window` when one is given, so a bar can describe where a drag has it.
export const eventValueText = (design: EncounterDesign, ref: DesignEventRef, baseSeconds: number, window?: TimeWindow): string => {
	const name = eventName(design, ref);
	const stored = eventWindow(design, ref);
	if (!stored) return name;
	if (ref.kind === 'bossHealth') {
		const segment = healthSegments(design).find(found => found.index === ref.index);
		const start = formatSeconds(fractionToSeconds(window?.start ?? segment?.start ?? 0, baseSeconds));
		const end = formatSeconds(fractionToSeconds(window?.end ?? segment?.end ?? 0, baseSeconds));
		return designerText('health_value', {
			name,
			from: formatPercent(segment?.fromPercent ?? 100),
			to: formatPercent(segment?.toPercent ?? 100),
			start,
			end,
		});
	}
	const shown = window ?? ('at' in stored ? { start: stored.at, end: stored.at } : stored);
	const at = formatSeconds(fractionToSeconds(shown.start, baseSeconds));
	if (!('at' in stored) || (ref.kind === 'move' && window))
		return designerText('span_value', { name, start: at, end: formatSeconds(fractionToSeconds(shown.end, baseSeconds)) });
	return designerText('point_value', { name, at });
};

// A hover tooltip: the title on the first line, then one line per rule.
export interface EventTip {
	title: string;
	lines: ReadonlyArray<string>;
}

// A tooltip anchor carries its text in one attribute, a line each; `parseTip` reads it back for the Tooltip to render.
export const serializeTip = (tip: EventTip): string => [tip.title, ...tip.lines].join('\n');
export const parseTip = (text: string | null | undefined): EventTip | undefined => {
	if (!text) return undefined;
	const [title, ...lines] = text.split('\n');
	return { title, lines };
};

// "Adds #2 — Add wave 'Adds'": a repeat of an event leads its title with the occurrence.
export const occurrenceTip = (tip: EventTip, label: string): EventTip => ({ ...tip, title: designerText('tip.occurrence_title', { label, title: tip.title }) });

// "3 adds (Lvl 60 Beast ×2, Lvl 61 Demon)": the adds grouped by level and mob type, in the order they first appear.
const addsSummary = (adds: ReadonlyArray<Target>): string => {
	const groups = new Map<string, { level: number; type: string; count: number }>();
	for (const add of adds) {
		const key = `${add.level}:${add.mobType}`;
		const group = groups.get(key) ?? { level: add.level, type: translateMobType(add.mobType), count: 0 };
		group.count++;
		groups.set(key, group);
	}
	const listed = [...groups.values()].map(group =>
		designerText(group.count > 1 ? 'tip.add_group_count' : 'tip.add_group', { level: group.level, type: group.type, count: group.count }),
	);
	return designerText('tip.adds', { count: adds.length, groups: listed.join(', ') });
};

export const eventTooltip = (
	design: EncounterDesign,
	ref: DesignEventRef,
	baseSeconds: number,
	targetNames: ReadonlyArray<string>,
	window?: TimeWindow,
): EventTip => {
	const event = getEvent(design, ref);
	if (!event) return { title: '', lines: [] };
	const tip = (key: string, options?: Record<string, string | number>) => designerText(`tip.${key}`, options);
	const sec = (fraction: number) => formatSeconds(fractionToSeconds(fraction, baseSeconds));
	const found = eventWindow(design, ref);
	const stored = !found ? { start: 0, end: 0 } : 'at' in found ? { start: found.at, end: found.at } : found;
	const shown = window ?? stored;
	const range = tip('range', { start: sec(shown.start), end: sec(shown.end) });
	const repeat = 'repeat' in event && event.repeat && event.repeat.interval > 0 ? sec(event.repeat.interval) : undefined;
	const every = repeat ? [tip('every', { interval: repeat })] : [];
	const built = (title: string, lines: ReadonlyArray<string | undefined>): EventTip => ({ title, lines: lines.filter((line): line is string => !!line) });
	switch (ref.kind) {
		case 'bossHealth': {
			const segment = healthSegments(design).find(found => found.index === ref.index);
			const covered = window ?? { start: segment?.start ?? 0, end: segment?.end ?? 0 };
			const [from, to] = [segment?.fromPercent ?? 100, segment?.toPercent ?? 100];
			const curve =
				from === to ? tip('health_hold', { percent: formatPercent(from) }) : tip('health', { from: formatPercent(from), to: formatPercent(to) });
			return built(eventName(design, ref), [curve, tip('range', { start: sec(covered.start), end: sec(covered.end) })]);
		}
		case 'downtime': {
			const downtime = event as DesignDowntime;
			const target = targetNames[downtime.targetIndex] ?? designerText('target_name', { number: downtime.targetIndex + 1 });
			return built(tip('downtime'), [target, range, downtime.clearDebuffs ? tip('clears_debuffs') : undefined]);
		}
		case 'ability': {
			const ability = event as DesignBossAbility;
			const kind = designerText(`ability_kinds.${ABILITY_KINDS.find(([value]) => value === ability.kind)?.[1]}`);
			const details: Array<string> = [];
			switch (ability.kind) {
				case DesignBossAbility_Kind.SpellHit:
					details.push(
						tip('school_damage', { school: translateSpellSchool(ability.school), min: ability.minDamage, max: ability.maxDamage }),
						tip('cast', { time: formatSeconds(ability.castTime) }),
					);
					break;
				case DesignBossAbility_Kind.Strike:
					details.push(tip('weapon', { percent: ability.weaponPercent }));
					break;
				case DesignBossAbility_Kind.Thrash:
					details.push(tip('proc', { chance: ability.procChance }));
					if (ability.extraAttacks > 0) details.push(tip('extra_attacks', { count: ability.extraAttacks }));
					if (ability.icd > 0) details.push(tip('icd', { time: formatSeconds(ability.icd) }));
					break;
				default:
					details.push(tip('swing_speed', { value: ability.swingSpeedMultiplier }));
			}
			const when = firesAtInstants(ability)
				? [repeat ? tip('every_from', { interval: repeat, start: sec(shown.start), end: sec(shown.end) }) : tip('at', { at: sec(shown.start) })]
				: [range, ...every];
			return built(kind, [...details, ...when]);
		}
		case 'wave': {
			const wave = event as AddWave;
			const jitter = wave.spawn?.jitter ?? 0;
			const spawn = jitter > 0 ? tip('spawns_jitter', { at: sec(shown.start), jitter: sec(jitter) }) : tip('spawns', { at: sec(shown.start) });
			return built(tip('wave', { name: eventName(design, ref) }), [
				addsSummary(wave.adds),
				spawn,
				tip('lives', { time: sec(shown.end - shown.start) }),
				...every,
				tip('kill_share', { percent: Number((wave.killShare * 100).toFixed(1)) }),
			]);
		}
		case 'move': {
			const move = event as DesignMove;
			const time = formatSeconds(Number(moveSeconds(move).toFixed(2)));
			return built(tip('move'), [tip('at', { at: sec(shown.start) }), move.yards > 0 ? tip('move_yards', { yards: move.yards, time }) : time, ...every]);
		}
		default: {
			const amp = event as DesignDamageAmp;
			const type = designerText(`amp_types.${AMP_TYPES.find(([value]) => value === amp.type)?.[1]}`);
			return built(tip('amp'), [tip('amp_value', { type, percent: amp.percent }), range, ...every]);
		}
	}
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
	nameField(),
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
	nameField(),
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
