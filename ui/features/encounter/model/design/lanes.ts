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
		downtime: { targetIndex: 0, start: time(start), end: time(end), clearDebuffs: true },
		bossHealth: { at: Math.min(Math.max(0, at), 1), healthPercent: 50 },
		ability: {
			kind: DesignBossAbility_Kind.SpellHit,
			targetIndex: 0,
			start: time(start),
			end: time(end),
			repeat: noRepeat(),
			school: SpellSchool.SpellSchoolPhysical,
			minDamage: 0,
			maxDamage: 0,
			castTime: 2,
			weaponPercent: 0,
			procChance: 0,
			extraAttacks: 0,
			icd: 0,
			swingSpeedMultiplier: 1,
		},
		wave: {
			name: '',
			count: 1,
			spawn: time(start),
			lifetime: DEFAULT_WIDTH,
			repeat: noRepeat(),
			executeProportion20: 0.2,
			executeProportion35: 0.35,
			killShare: 0.25,
		},
		move: { at: time(start), yards: 5, seconds: 0, repeat: noRepeat() },
		amp: {
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

// A spell hit or a strike happens at an instant, and its end only cuts its repeats; a thrash or swing timing window
// lasts from its start to its end.
export const isPointAbility = (ability: DesignBossAbility): boolean =>
	ability.kind === DesignBossAbility_Kind.SpellHit || ability.kind === DesignBossAbility_Kind.Strike;

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
			return isPointAbility(ability) ? { at: ability.start?.at ?? 0 } : span(ability.start, ability.end);
		}
		default: {
			const timed = event as DesignDowntime | DesignDamageAmp;
			return span(timed.start, timed.end);
		}
	}
};

// Every occurrence a repeat produces, the first included, spaced by the interval with jitter ignored. The count is the
// number of repeats after the first and 0 runs to the end of the fight; the result holds at most maxCount entries. A
// spell hit or a strike stops at its end, as the sim fires only the occurrences that begin by it.
export const occurrences = (design: EncounterDesign, ref: DesignEventRef, maxCount: number): { start: number; end: number }[] => {
	const event = getEvent(design, ref);
	const window = eventWindow(design, ref);
	if (!event || !window || maxCount <= 0) return [];
	const start = 'at' in window ? window.at : window.start;
	const width = 'at' in window ? 0 : window.end - window.start;
	const last = ref.kind === 'ability' && isPointAbility(event as DesignBossAbility) ? ((event as DesignBossAbility).end?.at ?? start) : Infinity;
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
