import type { AddWave, DesignBossAbility, DesignDamageAmp, DesignDowntime, DesignMove, DesignTime, EncounterDesign } from '@generated/proto/common';
import { DamageAmpType } from '@generated/proto/common';

import type { DesignEventRef } from './lanes';

export type DesignIssueSeverity = 'error' | 'warning' | 'info';
export type DesignIssue = { ref: DesignEventRef; severity: DesignIssueSeverity; key: string };

const key = (name: string) => `settings_tab.encounter.designer.validation.${name}`;

export const DESIGN_ISSUE_KEYS = {
	endBeforeStart: key('endBeforeStart'),
	outsideFight: key('outsideFight'),
	overlap: key('overlap'),
	healthIncreases: key('healthIncreases'),
	targetOutOfRange: key('targetOutOfRange'),
	waveOutOfRange: key('waveOutOfRange'),
	allTargetsDown: key('allTargetsDown'),
} as const;

const outside = (from: number, to: number, startJitter: number, endJitter = startJitter) => from - startJitter < 0 || to + endJitter > 1;

const overlaps = (a: DesignDowntime, b: DesignDowntime) => (a.start?.at ?? 0) < (b.end?.at ?? 0) && (b.start?.at ?? 0) < (a.end?.at ?? 0);

export const validateDesign = (design: EncounterDesign, targetCount: number): DesignIssue[] => {
	const issues: DesignIssue[] = [];
	const add = (ref: DesignEventRef, severity: DesignIssueSeverity, issueKey: string) => issues.push({ ref, severity, key: issueKey });
	const badTarget = (index: number) => index < 0 || index >= targetCount;

	const checkSpan = (ref: DesignEventRef, start: DesignTime | undefined, end: DesignTime | undefined) => {
		const from = start?.at ?? 0;
		const to = end?.at ?? 0;
		if (to <= from) add(ref, 'error', DESIGN_ISSUE_KEYS.endBeforeStart);
		if (outside(from, to, start?.jitter ?? 0, end?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
	};

	design.downtimes.forEach((downtime: DesignDowntime, index) => {
		const ref: DesignEventRef = { kind: 'downtime', index };
		checkSpan(ref, downtime.start, downtime.end);
		if (badTarget(downtime.targetIndex)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
		const clash = design.downtimes.some(
			(other, otherIndex) => otherIndex < index && other.targetIndex === downtime.targetIndex && overlaps(other, downtime),
		);
		if (clash) add(ref, 'error', DESIGN_ISSUE_KEYS.overlap);
	});

	design.bossAbilities.forEach((ability: DesignBossAbility, index) => {
		const ref: DesignEventRef = { kind: 'ability', index };
		checkSpan(ref, ability.start, ability.end);
		if (badTarget(ability.targetIndex)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
	});

	design.damageAmps.forEach((amp: DesignDamageAmp, index) => {
		const ref: DesignEventRef = { kind: 'amp', index };
		checkSpan(ref, amp.start, amp.end);
		if (amp.type !== DamageAmpType.DamageAmpTypeTargetDebuff) return;
		if (amp.targets?.targetIndices.some(badTarget)) add(ref, 'error', DESIGN_ISSUE_KEYS.targetOutOfRange);
		if (amp.targets?.waveIndices.some(wave => wave < 0 || wave >= design.addWaves.length)) add(ref, 'error', DESIGN_ISSUE_KEYS.waveOutOfRange);
	});

	design.addWaves.forEach((wave: AddWave, index) => {
		const ref: DesignEventRef = { kind: 'wave', index };
		const from = wave.spawn?.at ?? 0;
		if (wave.lifetime <= 0) add(ref, 'error', DESIGN_ISSUE_KEYS.endBeforeStart);
		if (outside(from, from + wave.lifetime, wave.spawn?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
	});

	// Moves have no length in fractions, so two at the same instant are the clash.
	design.moves.forEach((move: DesignMove, index) => {
		const ref: DesignEventRef = { kind: 'move', index };
		const at = move.at?.at ?? 0;
		if (outside(at, at, move.at?.jitter ?? 0)) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		if (design.moves.some((other, otherIndex) => otherIndex < index && (other.at?.at ?? 0) === at)) add(ref, 'error', DESIGN_ISSUE_KEYS.overlap);
	});

	design.bossHealth.forEach((frame, index) => {
		const ref: DesignEventRef = { kind: 'bossHealth', index };
		if (frame.at < 0 || frame.at > 1) add(ref, 'warning', DESIGN_ISSUE_KEYS.outsideFight);
		const previous = design.bossHealth
			.filter((other, otherIndex) => otherIndex !== index && (other.at < frame.at || (other.at === frame.at && otherIndex < index)))
			.reduce<typeof frame | undefined>((best, other) => (!best || other.at >= best.at ? other : best), undefined);
		if (previous && frame.healthPercent > previous.healthPercent) add(ref, 'error', DESIGN_ISSUE_KEYS.healthIncreases);
	});

	if (targetCount > 0) {
		const covers = (downtime: DesignDowntime, at: number) => (downtime.start?.at ?? 0) <= at && at < (downtime.end?.at ?? 0);
		const allDownAt = (at: number) =>
			Array.from({ length: targetCount }, (_, target) => target).every(target => design.downtimes.some(d => d.targetIndex === target && covers(d, at)));
		const instants = design.downtimes.map(d => d.start?.at ?? 0).filter(allDownAt);
		design.downtimes.forEach((downtime, index) => {
			if (instants.some(at => covers(downtime, at))) add({ kind: 'downtime', index }, 'info', DESIGN_ISSUE_KEYS.allTargetsDown);
		});
	}

	return issues;
};
