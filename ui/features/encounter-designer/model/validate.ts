import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, DesignTime, EncounterDesign } from '@generated/proto/common';
import { DamageAmpType, DesignBossAbility_Kind } from '@generated/proto/common';

import type { DesignEventRef } from './lanes';
import { firesAtInstants, healthKnots, occurrences } from './lanes';
import { clamp, type TimeWindow } from './time';

type DesignIssueSeverity = 'error' | 'warning' | 'info';
/** A stretch of time the issue's text names comes in fractions of the fight. */
export type DesignIssue = { ref: DesignEventRef; severity: DesignIssueSeverity; key: string; span?: TimeWindow };

const key = (name: string) => `settings_tab.encounter.designer.validation.${name}`;

export const DESIGN_ISSUE_KEYS = {
	endBeforeStart: key('end_before_start'),
	outsideFight: key('outside_fight'),
	overlap: key('overlap'),
	downtimeOverlap: key('downtime_overlap'),
	healthIncreases: key('health_increases'),
	targetOutOfRange: key('target_out_of_range'),
	waveOutOfRange: key('wave_out_of_range'),
	allTargetsDown: key('all_targets_down'),
	tooManyKeyframes: key('too_many_keyframes'),
	ampNotPositive: key('amp_not_positive'),
	tooManyAdds: key('too_many_adds'),
	noAdds: key('no_adds'),
	moveGoesNowhere: key('move_goes_nowhere'),
	executeOrder: key('execute_order'),
	noWeaponDamage: key('no_weapon_damage'),
	noExtraAttacks: key('no_extra_attacks'),
	waveOverPool: key('wave_over_pool'),
	healthStopsShort: key('health_stops_short'),
} as const;

// The sim's health curve holds 16 points in time order and keeps the first 16. Its own start point (0 s, 100%), there
// when the first keyframe is after the pull, comes first and always stays. Its own end point (the fight's end, 0%),
// there when the last keyframe is before the end and above 0%, comes last and is the first to go.
const MAX_HEALTH_POINTS = 16;

// The most adds one wave spawns at once; the sim skips a wave with more.
const MAX_WAVE_ADDS = 64;
// The sim's cap on one repeating event's occurrences per iteration (designMaxOccurrences in sim/core/design_runner.go).
const MAX_OCCURRENCES = 1000;

// The most copies of a wave's adds alive at once, as the sim sizes its pool (maxConcurrentOccurrences in
// sim/core/design_runner.go). A wave's adds and these copies share MAX_WAVE_ADDS slots, and a spawn with too few free is
// skipped.
const concurrentCopies = ({ spawn, repeat, lifetime }: AddWave): number => {
	const interval = repeat?.interval ?? 0;
	if (interval <= 0) return 1;
	const spawnJitter = Math.max(spawn?.jitter ?? 0, 0);
	const occurrences =
		repeat!.count > 0
			? Math.min(repeat!.count + 1, MAX_OCCURRENCES)
			: Math.trunc(clamp((1 - ((spawn?.at ?? 0) - spawnJitter)) / interval, 0, MAX_OCCURRENCES - 1)) + 1;
	if (occurrences <= 1) return 1;
	const spacing = interval - 2 * Math.max(repeat!.jitter, 0);
	if (spacing <= 0) return occurrences;
	return Math.min(Math.trunc((Math.max(lifetime, 0) + 2 * spawnJitter) / spacing) + 1, occurrences);
};

const outside = (from: number, to: number, startJitter: number, endJitter = startJitter) => from - startJitter < 0 || to + endJitter > 1;

const overlaps = (a: DesignDowntime, b: DesignDowntime) => (a.start?.at ?? 0) < (b.end?.at ?? 0) && (b.start?.at ?? 0) < (a.end?.at ?? 0);

export const validateDesign = (design: EncounterDesign, targetCount: number): DesignIssue[] => {
	const issues: DesignIssue[] = [];
	const add = (ref: DesignEventRef, severity: DesignIssueSeverity, issueKey: string, span?: TimeWindow) =>
		issues.push({ ref, severity, key: issueKey, ...(span ? { span } : {}) });
	const badTarget = (index: number) => index < 0 || index >= targetCount;

	const checkSpan = (ref: DesignEventRef, start: DesignTime | undefined, end: DesignTime | undefined) => {
		const from = start?.at ?? 0;
		const to = end?.at ?? 0;
		if (to <= from) add(ref, 'error', DESIGN_ISSUE_KEYS.endBeforeStart);
		if (outside(from, to, start?.jitter ?? 0, end?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
	};

	// The sim keeps a target down until the last of its overlapping downtimes ends.
	design.downtimes.forEach((downtime: DesignDowntime, index) => {
		const ref: DesignEventRef = { kind: 'downtime', index };
		checkSpan(ref, downtime.start, downtime.end);
		if (badTarget(downtime.targetIndex)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
		const clash = design.downtimes.some(
			(other, otherIndex) => otherIndex < index && other.targetIndex === downtime.targetIndex && overlaps(other, downtime),
		);
		if (clash) add(ref, 'warning', DESIGN_ISSUE_KEYS.downtimeOverlap);
	});

	// A spell hit or a strike fires at its start when its end is not before it, so only its start must be in the fight.
	design.bossAbilities.forEach((ability: DesignBossAbility, index) => {
		const ref: DesignEventRef = { kind: 'ability', index };
		if (firesAtInstants(ability)) {
			const from = ability.start?.at ?? 0;
			if ((ability.end?.at ?? 0) < from) add(ref, 'error', DESIGN_ISSUE_KEYS.endBeforeStart);
			if (outside(from, from, ability.start?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		} else {
			checkSpan(ref, ability.start, ability.end);
		}
		if (badTarget(ability.targetIndex)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
		if (ability.kind === DesignBossAbility_Kind.Strike && ability.weaponPercent <= 0) add(ref, 'error', DESIGN_ISSUE_KEYS.noWeaponDamage);
		if (ability.kind === DesignBossAbility_Kind.Thrash && (ability.procChance <= 0 || ability.extraAttacks <= 0))
			add(ref, 'error', DESIGN_ISSUE_KEYS.noExtraAttacks);
	});

	design.damageAmps.forEach((amp: DesignDamageAmp, index) => {
		const ref: DesignEventRef = { kind: 'amp', index };
		checkSpan(ref, amp.start, amp.end);
		if (amp.percent <= 0) add(ref, 'error', DESIGN_ISSUE_KEYS.ampNotPositive);
		if (amp.type !== DamageAmpType.DamageAmpTypeTargetDebuff) return;
		if (amp.targets?.targetIndices.some(badTarget)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
		if (amp.targets?.waveIndices.some(wave => wave < 0 || wave >= design.addWaves.length)) add(ref, 'error', DESIGN_ISSUE_KEYS.waveOutOfRange);
	});

	// An add whose lifetime runs past the fight's end lives to the end, so only the spawn must be in the fight.
	design.addWaves.forEach((wave: AddWave, index) => {
		const ref: DesignEventRef = { kind: 'wave', index };
		const from = wave.spawn?.at ?? 0;
		if (wave.lifetime <= 0) add(ref, 'error', DESIGN_ISSUE_KEYS.endBeforeStart);
		if (from >= 1 || outside(from, from, wave.spawn?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		if (wave.adds.length === 0) add(ref, 'error', DESIGN_ISSUE_KEYS.noAdds);
		if (wave.adds.length > MAX_WAVE_ADDS) add(ref, 'error', DESIGN_ISSUE_KEYS.tooManyAdds);
		else if (wave.adds.length * concurrentCopies(wave) > MAX_WAVE_ADDS) add(ref, 'warning', DESIGN_ISSUE_KEYS.waveOverPool);
		if (wave.executeProportion20 > wave.executeProportion35) add(ref, 'warning', DESIGN_ISSUE_KEYS.executeOrder);
	});

	// Moves have no length in fractions, so two at the same instant are the clash.
	design.moves.forEach((move: DesignMove, index) => {
		const ref: DesignEventRef = { kind: 'move', index };
		const at = move.at?.at ?? 0;
		if (outside(at, at, move.at?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		if (design.moves.some((other, otherIndex) => otherIndex < index && (other.at?.at ?? 0) === at)) add(ref, 'error', DESIGN_ISSUE_KEYS.overlap);
		if (move.yards <= 0 && move.seconds <= 0) add(ref, 'error', DESIGN_ISSUE_KEYS.moveGoesNowhere);
	});

	design.bossHealth.forEach((frame, index) => {
		const ref: DesignEventRef = { kind: 'bossHealth', index };
		if (frame.at < 0 || frame.at > 1) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		const previous = design.bossHealth
			.filter((other, otherIndex) => otherIndex !== index && (other.at < frame.at || (other.at === frame.at && otherIndex < index)))
			.reduce<typeof frame | undefined>((best, other) => (!best || other.at >= best.at ? other : best), undefined);
		if (previous && frame.healthPercent > previous.healthPercent) add(ref, 'error', DESIGN_ISSUE_KEYS.healthIncreases);
	});
	const knots = healthKnots(design);
	const keptKnots = MAX_HEALTH_POINTS - (knots[0]?.at > 0 ? 1 : 0);
	knots.slice(keptKnots).forEach(({ index }) => add({ kind: 'bossHealth', index }, 'error', DESIGN_ISSUE_KEYS.tooManyKeyframes));
	// The curve is full here, so the sim's own end point does not fit after this keyframe.
	const lastKept = knots[keptKnots - 1];
	if (lastKept && lastKept.at < 1 && lastKept.percent > 0) add({ kind: 'bossHealth', index: lastKept.index }, 'warning', DESIGN_ISSUE_KEYS.healthStopsShort);

	// The player idles where every target is down and no add is up. An add is up for each occurrence of a wave the sim
	// spawns, from its spawn for its lifetime, with jitter ignored as the timeline draws it.
	if (targetCount > 0) {
		const covers = (downtime: DesignDowntime, at: number) => (downtime.start?.at ?? 0) <= at && at < (downtime.end?.at ?? 0);
		const addLives = design.addWaves
			.map((wave, index) => ({ wave, index }))
			.filter(({ wave }) => wave.adds.length > 0 && wave.adds.length <= MAX_WAVE_ADDS && wave.lifetime > 0)
			.flatMap(({ index }) => occurrences(design, { kind: 'wave', index }, MAX_OCCURRENCES));
		const idleAt = (at: number) =>
			Array.from({ length: targetCount }, (_, target) => target).every(target => design.downtimes.some(d => d.targetIndex === target && covers(d, at))) &&
			!addLives.some(life => life.start <= at && at < life.end);
		const edges = [...design.downtimes.flatMap(d => [d.start?.at ?? 0, d.end?.at ?? 0]), ...addLives.flatMap(life => [life.start, life.end])];

		design.downtimes.forEach((downtime, index) => {
			const from = downtime.start?.at ?? 0;
			const to = downtime.end?.at ?? 0;
			const cuts = [...new Set([from, to, ...edges.filter(at => from < at && at < to)])].sort((a, b) => a - b);
			const stretches: TimeWindow[] = [];
			// Between two neighbouring edges nothing changes, so the midpoint stands for the whole piece.
			cuts.slice(1).forEach((end, i) => {
				const start = cuts[i];
				if (end <= start || !idleAt((start + end) / 2)) return;
				const last = stretches.at(-1);
				if (last?.end === start) last.end = end;
				else stretches.push({ start, end });
			});
			stretches.forEach(span => add({ kind: 'downtime', index }, 'info', DESIGN_ISSUE_KEYS.allTargetsDown, span));
		});
	}

	return issues;
};
