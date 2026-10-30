import {
	AddWave,
	DamageAmpType,
	DesignBossAbility,
	DesignBossAbility_Kind,
	DesignDamageAmp,
	DesignDowntime,
	DesignMove,
	DesignRepeat,
	DesignTargets,
	DesignTime,
	EncounterDesign,
	HealthKeyframe,
	SpellSchool,
} from '@generated/proto/common';
import i18n from '@i18n/config';
import { Encounter } from '@sim/raid/encounter';

export type DesignLaneKind = 'boss' | 'abilities' | 'waves' | 'moves' | 'amps';
export type DesignEventKind = 'downtime' | 'bossHealth' | 'ability' | 'wave' | 'move' | 'amp';

// Events have no ids; a list and an index identify one.
export type DesignEventRef = { kind: DesignEventKind; index: number };

type DesignEventMap = {
	downtime: DesignDowntime;
	bossHealth: HealthKeyframe;
	ability: DesignBossAbility;
	wave: AddWave;
	move: DesignMove;
	amp: DesignDamageAmp;
};
export type DesignEvent<K extends DesignEventKind = DesignEventKind> = DesignEventMap[K];

// A span, or a point for events that have no duration.
export type DesignEventWindow = { start: number; end: number } | { at: number };

export const DESIGN_LANES: readonly DesignLaneKind[] = ['boss', 'abilities', 'waves', 'moves', 'amps'];

const LANE_OF: Record<DesignEventKind, DesignLaneKind> = {
	downtime: 'boss',
	bossHealth: 'boss',
	ability: 'abilities',
	wave: 'waves',
	move: 'moves',
	amp: 'amps',
};

export const laneOf = (kind: DesignEventKind): DesignLaneKind => LANE_OF[kind];

const EVENT_LABEL_KEYS: Record<DesignEventKind, string> = {
	downtime: 'downtime',
	bossHealth: 'boss_health',
	ability: 'ability',
	wave: 'wave',
	move: 'move',
	amp: 'amp',
};

export const laneLabelKey = (lane: DesignLaneKind): string => `settings_tab.encounter.designer.lanes.${lane}`;
export const eventLabelKey = (kind: DesignEventKind): string => `settings_tab.encounter.designer.events.${EVENT_LABEL_KEYS[kind]}`;

const time = (at: number): DesignTime => ({ at, jitter: 0 });
const noRepeat = (): DesignRepeat => ({ interval: 0, jitter: 0, count: 0 });

const DEFAULT_WIDTH = 0.1;

// A new event starting at `at`, with its window pulled back inside the fight.
export const defaultEvent = <K extends DesignEventKind>(kind: K, at = 0.4): DesignEvent<K> => {
	const start = Math.min(Math.max(0, at), 1 - DEFAULT_WIDTH);
	const end = start + DEFAULT_WIDTH;
	const events: DesignEventMap = {
		downtime: { name: '', targetIndex: 0, start: time(start), end: time(end), clearDebuffs: true },
		bossHealth: { name: '', at: Math.min(Math.max(0, at), 1), healthPercent: 50 },
		ability: {
			name: '',
			kind: DesignBossAbility_Kind.SpellHit,
			targetIndex: 0,
			start: time(start),
			end: time(end),
			repeat: noRepeat(),
			school: SpellSchool.SpellSchoolPhysical,
			minDamage: 0,
			maxDamage: 0,
			castTime: 2,
			weaponPercent: 150,
			procChance: 0,
			extraAttacks: 0,
			icd: 0,
			swingSpeedMultiplier: 1,
		},
		wave: {
			name: '',
			adds: [Encounter.defaultTargetProto()],
			spawn: time(start),
			lifetime: DEFAULT_WIDTH,
			repeat: noRepeat(),
			executeProportion20: 0.2,
			executeProportion35: 0.35,
			killShare: 0.25,
		},
		move: { name: '', at: time(start), yards: 5, seconds: 0, repeat: noRepeat() },
		amp: {
			name: '',
			type: DamageAmpType.DamageAmpTypeCasterBuff,
			percent: 10,
			start: time(start),
			end: time(end),
			repeat: noRepeat(),
			targets: DesignTargets.create(),
		},
	};
	return events[kind] as DesignEvent<K>;
};

export const eventsOf = <K extends DesignEventKind>(design: EncounterDesign, kind: K): DesignEvent<K>[] => {
	const lists: { [P in DesignEventKind]: DesignEvent<P>[] } = {
		downtime: design.downtimes,
		bossHealth: design.bossHealth,
		ability: design.bossAbilities,
		wave: design.addWaves,
		move: design.moves,
		amp: design.damageAmps,
	};
	return lists[kind] as DesignEvent<K>[];
};

export const getEvent = (design: EncounterDesign, ref: DesignEventRef): DesignEvent | undefined => eventsOf(design, ref.kind)[ref.index];

// Mutates the design, which is meant to be an immer draft.
export const addEvent = (design: EncounterDesign, kind: DesignEventKind, at?: number): DesignEventRef => {
	const list: DesignEvent[] = eventsOf(design, kind);
	list.push(defaultEvent(kind, at));
	return { kind, index: list.length - 1 };
};

export const removeEvent = (design: EncounterDesign, ref: DesignEventRef): void => {
	const list = eventsOf(design, ref.kind);
	if (ref.index >= 0 && ref.index < list.length) list.splice(ref.index, 1);
};

const span = (start: DesignTime | undefined, end: DesignTime | undefined): DesignEventWindow => {
	const from = start?.at ?? 0;
	return { start: from, end: end?.at ?? from };
};

const repeatOf = (event: DesignEvent): DesignRepeat | undefined => ('repeat' in event ? event.repeat : undefined);

// A spell hit or a strike fires at its start and at each repeat that begins by its end; a thrash or swing timing
// window lasts from its start to its end.
export const firesAtInstants = (ability: DesignBossAbility): boolean =>
	ability.kind === DesignBossAbility_Kind.SpellHit || ability.kind === DesignBossAbility_Kind.Strike;

// The player run speed the sim walks a move's yards at (`GetMovementSpeed` in sim/core/movement.go), in yards per second.
export const PLAYER_RUN_SPEED = 7;

// A move's yards take precedence over its seconds, as the sim reads them.
export const moveSeconds = (move: DesignMove): number => (move.yards > 0 ? move.yards / PLAYER_RUN_SPEED : move.seconds);

export const eventWindow = (design: EncounterDesign, ref: DesignEventRef): DesignEventWindow | undefined => {
	const event = getEvent(design, ref);
	if (!event) return undefined;
	switch (ref.kind) {
		case 'bossHealth':
			return { at: (event as HealthKeyframe).at };
		case 'move':
			return { at: (event as DesignMove).at?.at ?? 0 };
		case 'wave': {
			const wave = event as AddWave;
			const start = wave.spawn?.at ?? 0;
			return { start, end: start + wave.lifetime };
		}
		case 'ability': {
			const ability = event as DesignBossAbility;
			return span(ability.start, ability.end);
		}
		default: {
			const timed = event as DesignDowntime | DesignDamageAmp;
			return span(timed.start, timed.end);
		}
	}
};

// Every occurrence a repeat produces, the first included, spaced by the interval with jitter ignored. The count is the
// number of repeats after the first and 0 runs to the end of the fight; the result holds at most maxCount entries. A
// spell hit or a strike is an instant per occurrence and stops at its end, as the sim fires only the occurrences that
// begin by it.
export const occurrences = (design: EncounterDesign, ref: DesignEventRef, maxCount: number): { start: number; end: number }[] => {
	const event = getEvent(design, ref);
	const window = eventWindow(design, ref);
	if (!event || !window || maxCount <= 0) return [];
	const start = 'at' in window ? window.at : window.start;
	const instants = ref.kind === 'ability' && firesAtInstants(event as DesignBossAbility);
	const width = 'at' in window || instants ? 0 : window.end - window.start;
	const last = instants ? ((event as DesignBossAbility).end?.at ?? start) : Infinity;
	const repeat = repeatOf(event);
	const interval = repeat?.interval ?? 0;
	const total = interval > 0 ? (repeat!.count > 0 ? repeat!.count + 1 : Infinity) : 1;
	const result: { start: number; end: number }[] = [];
	for (let i = 0; i < total && result.length < maxCount; i++) {
		const from = start + i * interval;
		if ((i > 0 && from > 1) || from > last) break;
		result.push({ start: from, end: from + width });
	}
	return result;
};

export interface HealthSegment {
	/** The keyframe the segment ends at. */
	index: number;
	/** The keyframe the segment starts at, or undefined when the sim adds the pull's 100% start point. */
	previous: number | undefined;
	fromPercent: number;
	toPercent: number;
	start: number;
	end: number;
}

const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

// The curve's knots as the sim builds them (`bossHealthKnots` in sim/core/design_runner.go): clamped, in time order
// (stable), never rising.
const healthKnots = (design: EncounterDesign): Array<{ index: number; at: number; percent: number }> => {
	const knots = design.bossHealth
		.map((keyframe, index) => ({ index, at: clamp(keyframe.at, 0, 1), percent: clamp(keyframe.healthPercent, 0, 100) }))
		.sort((a, b) => a.at - b.at || a.index - b.index);
	for (let i = 1; i < knots.length; i++) knots[i].percent = Math.min(knots[i].percent, knots[i - 1].percent);
	return knots;
};

// One segment per keyframe, from the keyframe before it in time (or the implied 0 s, 100% start) to the keyframe.
export const healthSegments = (design: EncounterDesign): Array<HealthSegment> =>
	healthKnots(design).map((knot, i, knots) => ({
		index: knot.index,
		previous: i > 0 ? knots[i - 1].index : undefined,
		fromPercent: i > 0 ? knots[i - 1].percent : 100,
		toPercent: knot.percent,
		start: i > 0 ? knots[i - 1].at : 0,
		end: knot.at,
	}));

// The segment the sim appends when the last keyframe is before the fight's end and above 0%: it falls to 0% at the end.
export const healthEndSegment = (design: EncounterDesign): { start: number; end: number; fromPercent: number } | undefined => {
	const last = healthKnots(design).at(-1);
	return last && last.at < 1 && last.percent > 0 ? { start: last.at, end: 1, fromPercent: last.percent } : undefined;
};

const shiftTime = (time: DesignTime | undefined, delta: number) => {
	if (time) time.at += delta;
};

// Appends a copy of the event to its list, shifted by deltaFraction but kept inside the fight, and returns the copy's
// ref. A named event's copy is named "<name> (copy)". Mutates the design, which is meant to be an immer draft.
export const duplicateEvent = (design: EncounterDesign, ref: DesignEventRef, deltaFraction: number): DesignEventRef => {
	const source = getEvent(design, ref);
	if (!source) return ref;
	const copies: { [K in DesignEventKind]: (event: DesignEvent<K>) => DesignEvent<K> } = {
		downtime: event => DesignDowntime.clone(event),
		bossHealth: event => HealthKeyframe.clone(event),
		ability: event => DesignBossAbility.clone(event),
		wave: event => AddWave.clone(event),
		move: event => DesignMove.clone(event),
		amp: event => DesignDamageAmp.clone(event),
	};
	const copy = copies[ref.kind](source as never) as DesignEvent;
	const window = eventWindow(design, ref)!;
	const [from, to] = 'at' in window ? [window.at, window.at] : [window.start, ref.kind === 'wave' ? window.start : window.end];
	const delta = clamp(deltaFraction, -from, 1 - to);
	switch (ref.kind) {
		case 'bossHealth':
			(copy as HealthKeyframe).at += delta;
			break;
		case 'move':
			shiftTime((copy as DesignMove).at, delta);
			break;
		case 'wave': {
			const wave = copy as AddWave;
			shiftTime(wave.spawn, delta);
			break;
		}
		default: {
			const timed = copy as DesignDowntime | DesignDamageAmp | DesignBossAbility;
			shiftTime(timed.start, delta);
			shiftTime(timed.end, delta);
		}
	}
	if (copy.name) copy.name = i18n.t('settings_tab.encounter.designer.copy_name', { name: copy.name });
	const list: DesignEvent[] = eventsOf(design, ref.kind);
	list.push(copy);
	return { kind: ref.kind, index: list.length - 1 };
};

/** The longest flat stretch `holdSegment` cuts off a segment's start. */
const HOLD_SECONDS = 10;

// Splits a boss health segment at its start into a flat part, held at the segment's start percent, and the rest, by a
// keyframe at most HOLD_SECONDS in, or half the segment. Mutates the design, which is meant to be an immer draft.
// Returns the new keyframe's ref, or undefined for a segment with no length to split.
export const holdSegment = (design: EncounterDesign, ref: DesignEventRef, baseSeconds: number): DesignEventRef | undefined => {
	const segment = healthSegments(design).find(found => found.index === ref.index);
	if (!segment || segment.end <= segment.start) return undefined;
	const at = segment.start + Math.min(HOLD_SECONDS / baseSeconds, (segment.end - segment.start) / 2);
	design.bossHealth.push(HealthKeyframe.create({ at, healthPercent: segment.fromPercent }));
	return { kind: 'bossHealth', index: design.bossHealth.length - 1 };
};

// The boss health the curve gives at a fraction of the fight: the sim's knots plus its own start (0 s, 100%) and end
// (the fight's end, 0%) points, and the line between the two around it. 100% without a keyframe.
export const healthAt = (design: EncounterDesign, at: number): number => {
	const knots: Array<{ at: number; percent: number }> = healthKnots(design);
	if (knots.length === 0) return 100;
	if (knots[0].at > 0) knots.unshift({ at: 0, percent: 100 });
	const last = knots[knots.length - 1];
	if (last.at < 1 && last.percent > 0) knots.push({ at: 1, percent: 0 });
	if (at <= knots[0].at) return knots[0].percent;
	for (let i = 1; i < knots.length; i++) {
		const after = knots[i];
		if (at > after.at) continue;
		const before = knots[i - 1];
		return after.at === before.at ? after.percent : before.percent + ((at - before.at) / (after.at - before.at)) * (after.percent - before.percent);
	}
	return knots[knots.length - 1].percent;
};

// The keyframe strictly inside the downtime's window, which a hold across it would cross, if there is one.
export const holdBlocker = (design: EncounterDesign, downtimeIndex: number): DesignEventRef | undefined => {
	const downtime = design.downtimes[downtimeIndex];
	if (!downtime) return undefined;
	const [from, to] = [downtime.start?.at ?? 0, downtime.end?.at ?? 0];
	const inside = healthKnots(design).find(knot => knot.at > from && knot.at < to);
	return inside && { kind: 'bossHealth', index: inside.index };
};

// Holds the boss's health at what the curve gives at the downtime's start, by keyframes at its start and its end.
// Mutates the design, which is meant to be an immer draft. Returns the two new refs, none for a hold a keyframe blocks
// or a downtime of another target than the boss.
export const holdDuringDowntime = (design: EncounterDesign, downtimeIndex: number): Array<DesignEventRef> => {
	const downtime = design.downtimes[downtimeIndex];
	if (!downtime || downtime.targetIndex !== 0 || holdBlocker(design, downtimeIndex)) return [];
	const [from, to] = [downtime.start?.at ?? 0, downtime.end?.at ?? 0];
	const healthPercent = healthAt(design, from);
	design.bossHealth.push(HealthKeyframe.create({ at: from, healthPercent }), HealthKeyframe.create({ at: to, healthPercent }));
	const last = design.bossHealth.length - 1;
	return [
		{ kind: 'bossHealth', index: last - 1 },
		{ kind: 'bossHealth', index: last },
	];
};
