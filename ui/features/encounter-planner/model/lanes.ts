import {
	AddWave,
	DamageAmpType,
	EncounterPlan,
	HealthKeyframe,
	PlanBossAbility,
	PlanBossAbility_Kind,
	PlanDamageAmp,
	PlanDowntime,
	PlanMove,
	PlanRepeat,
	PlanTargets,
	PlanTime,
	SpellSchool,
} from '@generated/proto/common';
import i18n from '@i18n/config';
import { Encounter } from '@sim/raid/encounter';

import { clamp } from './time';

export type PlanLaneKind = 'boss' | 'abilities' | 'waves' | 'moves' | 'amps';
export type PlanEventKind = 'downtime' | 'bossHealth' | 'ability' | 'wave' | 'move' | 'amp';

// Events have no ids; a list and an index identify one.
export type PlanEventRef = { kind: PlanEventKind; index: number };

// What the timeline and the panel have selected: an event of the plan, or an encounter target, which lives on the
// encounter rather than in the plan.
export type PlannerSelection = PlanEventRef | { kind: 'target'; index: number };

export const isEventRef = (selection: PlannerSelection | undefined): selection is PlanEventRef => !!selection && selection.kind !== 'target';

type PlanEventMap = {
	downtime: PlanDowntime;
	bossHealth: HealthKeyframe;
	ability: PlanBossAbility;
	wave: AddWave;
	move: PlanMove;
	amp: PlanDamageAmp;
};
export type PlanEvent<K extends PlanEventKind = PlanEventKind> = PlanEventMap[K];

// A span, or a point for events that have no duration.
type PlanEventWindow = { start: number; end: number } | { at: number };

// The 'boss' lane is one row per encounter target, which holds that target's downtimes, and on the first target the boss health.
export const PLAN_LANES: readonly PlanLaneKind[] = ['boss', 'abilities', 'waves', 'moves', 'amps'];

export const LANE_EVENT_KINDS: Record<PlanLaneKind, ReadonlyArray<PlanEventKind>> = {
	boss: ['downtime', 'bossHealth'],
	abilities: ['ability'],
	waves: ['wave'],
	moves: ['move'],
	amps: ['amp'],
};

export const laneLabelKey = (lane: PlanLaneKind): string => `settings_tab.encounter.planner.lanes.${lane}`;
export const eventLabelKey = (kind: PlanEventKind): string => `settings_tab.encounter.planner.events.${kind === 'bossHealth' ? 'boss_health' : kind}`;

const time = (at: number): PlanTime => ({ at, jitter: 0 });
const noRepeat = (): PlanRepeat => ({ interval: 0, jitter: 0, count: 0 });

const DEFAULT_WIDTH = 0.1;

export const defaultEvent = <K extends PlanEventKind>(kind: K): PlanEvent<K> => {
	const start = 0.4;
	const end = start + DEFAULT_WIDTH;
	const events: PlanEventMap = {
		downtime: { name: '', targetIndex: 0, start: time(start), end: time(end), clearDebuffs: true },
		bossHealth: { name: '', at: start, healthPercent: 50 },
		ability: {
			name: '',
			kind: PlanBossAbility_Kind.SpellHit,
			targetIndex: 0,
			start: time(start),
			end: time(end),
			repeat: noRepeat(),
			school: SpellSchool.SpellSchoolPhysical,
			minDamage: 0,
			maxDamage: 0,
			castTime: 2,
			weaponPercent: 150,
			procChance: 20,
			extraAttacks: 1,
			icd: 0,
			swingSpeedMultiplier: 1,
			// A pulse's and an area hit's own fields, which withAbilityKind fills.
			tickInterval: 0,
			tickJitter: 0,
			resistible: false,
			avoidable: false,
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
			holdDebuffs: false,
			debuffDelay: 0,
			debuffJitter: 0,
		},
		move: { name: '', at: time(start), yards: 5, seconds: 0, repeat: noRepeat() },
		amp: {
			name: '',
			type: DamageAmpType.DamageAmpTypeCasterBuff,
			percent: 10,
			start: time(start),
			end: time(end),
			repeat: noRepeat(),
			targets: PlanTargets.create(),
		},
	};
	return events[kind] as PlanEvent<K>;
};

export const eventsOf = <K extends PlanEventKind>(plan: EncounterPlan, kind: K): PlanEvent<K>[] => {
	const lists: { [P in PlanEventKind]: PlanEvent<P>[] } = {
		downtime: plan.downtimes,
		bossHealth: plan.bossHealth,
		ability: plan.bossAbilities,
		wave: plan.addWaves,
		move: plan.moves,
		amp: plan.damageAmps,
	};
	return lists[kind] as PlanEvent<K>[];
};

export const getEvent = (plan: EncounterPlan, ref: PlanEventRef): PlanEvent | undefined => eventsOf(plan, ref.kind)[ref.index];

// A downtime takes the target at targetIndex. Mutates the plan.
export const addEvent = (plan: EncounterPlan, kind: PlanEventKind, targetIndex = 0): PlanEventRef => {
	const list: PlanEvent[] = eventsOf(plan, kind);
	const event = defaultEvent(kind);
	if (kind === 'downtime') (event as PlanDowntime).targetIndex = targetIndex;
	list.push(event);
	return { kind, index: list.length - 1 };
};

export const removeEvent = (plan: EncounterPlan, ref: PlanEventRef): void => {
	const list = eventsOf(plan, ref.kind);
	if (ref.index < 0 || ref.index >= list.length) return;
	list.splice(ref.index, 1);
	if (ref.kind !== 'wave') return;
	// Amps name waves by index: a later wave moves down one, and a removed one becomes -1 so validation flags it.
	for (const amp of plan.damageAmps) {
		if (amp.targets) amp.targets.waveIndices = amp.targets.waveIndices.map(wave => (wave === ref.index ? -1 : wave > ref.index ? wave - 1 : wave));
	}
};

// A spell hit, a strike or an area hit fires at its start and at each repeat that begins by its end; a thrash, swing
// timing or pulse window lasts from its start to its end.
export const firesAtInstants = (ability: PlanBossAbility): boolean =>
	ability.kind === PlanBossAbility_Kind.SpellHit || ability.kind === PlanBossAbility_Kind.Strike || ability.kind === PlanBossAbility_Kind.AreaHit;

// A pulse or an area hit: area damage from no target, which hits the player and every pet and ignores the target index.
export const isAreaAbility = (ability: PlanBossAbility): boolean =>
	ability.kind === PlanBossAbility_Kind.Pulse || ability.kind === PlanBossAbility_Kind.AreaHit;

const AREA_DAMAGE_DEFAULTS: Partial<Record<PlanBossAbility_Kind, { min: number; max: number }>> = {
	[PlanBossAbility_Kind.Pulse]: { min: 500, max: 700 },
	[PlanBossAbility_Kind.AreaHit]: { min: 1500, max: 2000 },
};

// Sets the ability's kind. A switch to a pulse or an area hit fills the damage, as fire, and the tick interval and jitter
// (1 s and 0.2 s) where they are 0, and turns Resistible on when it comes from a kind without that toggle. Mutates the
// ability.
export const withAbilityKind = (ability: PlanBossAbility, kind: PlanBossAbility_Kind): void => {
	const wasArea = isAreaAbility(ability);
	ability.kind = kind;
	const damage = AREA_DAMAGE_DEFAULTS[kind];
	if (!damage) return;
	if (ability.maxDamage <= 0) {
		ability.minDamage = damage.min;
		ability.maxDamage = damage.max;
		ability.school = SpellSchool.SpellSchoolFire;
	}
	if (kind === PlanBossAbility_Kind.Pulse && ability.tickInterval <= 0) {
		ability.tickInterval = 1;
		ability.tickJitter = 0.2;
	}
	if (!wasArea) ability.resistible = true;
};

// The tick times of one pulse window, in fractions of the fight: one at its start, then one every interval before its
// end, so a window shorter than one interval ticks once. Jitter is ignored, as the timeline draws it.
export const pulseTicks = (window: { start: number; end: number }, intervalSeconds: number, baseSeconds: number, maxCount: number): Array<number> => {
	const interval = baseSeconds > 0 ? intervalSeconds / baseSeconds : 0;
	if (interval <= 0 || window.end <= window.start) return [];
	const ticks: Array<number> = [];
	for (let at = window.start; at < window.end && ticks.length < maxCount; at = window.start + ticks.length * interval) ticks.push(at);
	return ticks;
};

// The player run speed the sim walks a move's yards at (`GetMovementSpeed` in sim/core/movement.go), in yards per second.
export const PLAYER_RUN_SPEED = 7;

// A move's yards take precedence over its seconds, as the sim reads them.
export const moveSeconds = (move: PlanMove): number => (move.yards > 0 ? move.yards / PLAYER_RUN_SPEED : move.seconds);

export const eventWindow = (plan: EncounterPlan, ref: PlanEventRef): PlanEventWindow | undefined => {
	const event = getEvent(plan, ref);
	if (!event) return undefined;
	switch (ref.kind) {
		case 'bossHealth':
			return { at: (event as HealthKeyframe).at };
		case 'move':
			return { at: (event as PlanMove).at?.at ?? 0 };
		case 'wave': {
			const wave = event as AddWave;
			const start = wave.spawn?.at ?? 0;
			return { start, end: start + wave.lifetime };
		}
		default: {
			const timed = event as PlanDowntime | PlanDamageAmp | PlanBossAbility;
			const from = timed.start?.at ?? 0;
			return { start: from, end: timed.end?.at ?? from };
		}
	}
};

// Every occurrence a repeat produces, the first included, spaced by the interval with jitter ignored. The count is the
// number of repeats after the first and 0 runs to the end of the fight; the result holds at most maxCount entries. A
// spell hit or a strike is an instant per occurrence and stops at its end, as the sim fires only the occurrences that
// begin by it.
export const occurrences = (plan: EncounterPlan, ref: PlanEventRef, maxCount: number): { start: number; end: number }[] => {
	const event = getEvent(plan, ref);
	const window = eventWindow(plan, ref);
	if (!event || !window || maxCount <= 0) return [];
	const start = 'at' in window ? window.at : window.start;
	const instants = ref.kind === 'ability' && firesAtInstants(event as PlanBossAbility);
	const width = 'at' in window || instants ? 0 : window.end - window.start;
	const last = instants ? ((event as PlanBossAbility).end?.at ?? start) : Infinity;
	const repeat = 'repeat' in event ? event.repeat : undefined;
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

// The curve's knots as the sim builds them (`bossHealthKnots` in sim/core/plan_runner.go): clamped, in time order
// (stable), never rising.
export const healthKnots = (plan: EncounterPlan): Array<{ index: number; at: number; percent: number }> => {
	const knots = plan.bossHealth
		.map((keyframe, index) => ({ index, at: clamp(keyframe.at, 0, 1), percent: clamp(keyframe.healthPercent, 0, 100) }))
		.sort((a, b) => a.at - b.at || a.index - b.index);
	for (let i = 1; i < knots.length; i++) knots[i].percent = Math.min(knots[i].percent, knots[i - 1].percent);
	return knots;
};

// One segment per keyframe, from the keyframe before it in time (or the implied 0 s, 100% start) to the keyframe.
export const healthSegments = (plan: EncounterPlan): Array<HealthSegment> =>
	healthKnots(plan).map((knot, i, knots) => ({
		index: knot.index,
		previous: i > 0 ? knots[i - 1].index : undefined,
		fromPercent: i > 0 ? knots[i - 1].percent : 100,
		toPercent: knot.percent,
		start: i > 0 ? knots[i - 1].at : 0,
		end: knot.at,
	}));

// The segment the sim appends when the last keyframe is before the fight's end and above 0%: it falls to 0% at the end.
export const healthEndSegment = (plan: EncounterPlan): { start: number; end: number; fromPercent: number } | undefined => {
	const last = healthKnots(plan).at(-1);
	return last && last.at < 1 && last.percent > 0 ? { start: last.at, end: 1, fromPercent: last.percent } : undefined;
};

const shiftTime = (time: PlanTime | undefined, delta: number) => {
	if (time) time.at += delta;
};

// Appends a copy of the event to its list, shifted by deltaFraction but kept inside the fight, and returns the copy's
// ref. A named event's copy is named "<name> (copy)". Mutates the plan.
export const duplicateEvent = (plan: EncounterPlan, ref: PlanEventRef, deltaFraction: number): PlanEventRef => {
	const copy = getEvent(EncounterPlan.clone(plan), ref);
	if (!copy) return ref;
	const window = eventWindow(plan, ref)!;
	const [from, to] = 'at' in window ? [window.at, window.at] : [window.start, ref.kind === 'wave' ? window.start : window.end];
	const delta = clamp(deltaFraction, -from, 1 - to);
	switch (ref.kind) {
		case 'bossHealth':
			(copy as HealthKeyframe).at += delta;
			break;
		case 'move':
			shiftTime((copy as PlanMove).at, delta);
			break;
		case 'wave':
			shiftTime((copy as AddWave).spawn, delta);
			break;
		default: {
			const timed = copy as PlanDowntime | PlanDamageAmp | PlanBossAbility;
			shiftTime(timed.start, delta);
			shiftTime(timed.end, delta);
		}
	}
	if (copy.name) copy.name = i18n.t('settings_tab.encounter.planner.copy_name', { name: copy.name });
	const list: PlanEvent[] = eventsOf(plan, ref.kind);
	list.push(copy);
	return { kind: ref.kind, index: list.length - 1 };
};

/** The longest flat stretch `holdSegment` cuts off a segment's start. */
const HOLD_SECONDS = 10;

// Splits a boss health segment at its start into a flat part, held at the segment's start percent, and the rest, by a
// keyframe at most HOLD_SECONDS in, or half the segment. Mutates the plan. Returns the new keyframe's ref, or undefined
// for a segment with no length to split.
export const holdSegment = (plan: EncounterPlan, ref: PlanEventRef, baseSeconds: number): PlanEventRef | undefined => {
	const segment = healthSegments(plan).find(found => found.index === ref.index);
	if (!segment || segment.end <= segment.start) return undefined;
	const at = segment.start + Math.min(HOLD_SECONDS / baseSeconds, (segment.end - segment.start) / 2);
	plan.bossHealth.push(HealthKeyframe.create({ at, healthPercent: segment.fromPercent }));
	return { kind: 'bossHealth', index: plan.bossHealth.length - 1 };
};

// The boss health the curve gives at a fraction of the fight: the sim's knots plus its own start (0 s, 100%) and end
// (the fight's end, 0%) points, and the line between the two around it. 100% without a keyframe.
export const healthAt = (plan: EncounterPlan, at: number): number => {
	const knots: Array<{ at: number; percent: number }> = healthKnots(plan);
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

// The keyframe a hold across the downtime would cross, if there is one: inside its window, or on its end and lower than
// the health held, which the curve then falls to by the end.
export const holdBlocker = (plan: EncounterPlan, downtimeIndex: number): PlanEventRef | undefined => {
	const downtime = plan.downtimes[downtimeIndex];
	if (!downtime) return undefined;
	const [from, to] = [downtime.start?.at ?? 0, downtime.end?.at ?? 0];
	const held = healthAt(plan, from);
	const inside = healthKnots(plan).find(knot => knot.at > from && (knot.at < to || (knot.at === to && knot.percent < held)));
	return inside && { kind: 'bossHealth', index: inside.index };
};

// Holds the boss's health at what the curve gives at the downtime's start, by keyframes at its start and its end.
// Mutates the plan. Returns the two new refs, none for a hold a keyframe blocks or a downtime of another target than
// the boss.
export const holdDuringDowntime = (plan: EncounterPlan, downtimeIndex: number): Array<PlanEventRef> => {
	const downtime = plan.downtimes[downtimeIndex];
	if (!downtime || downtime.targetIndex !== 0 || holdBlocker(plan, downtimeIndex)) return [];
	const [from, to] = [downtime.start?.at ?? 0, downtime.end?.at ?? 0];
	const healthPercent = healthAt(plan, from);
	plan.bossHealth.push(HealthKeyframe.create({ at: from, healthPercent }), HealthKeyframe.create({ at: to, healthPercent }));
	const last = plan.bossHealth.length - 1;
	return [
		{ kind: 'bossHealth', index: last - 1 },
		{ kind: 'bossHealth', index: last },
	];
};
