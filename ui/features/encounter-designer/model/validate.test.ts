import { AddWave, DamageAmpType, DesignBossAbility_Kind, EncounterDesign, HealthKeyframe, Target } from '@generated/proto/common';
import { addEventAt } from '@testing/encounter_design';
import { describe, expect, it } from 'vitest';

import { addEvent, removeEvent } from './lanes';
import { DESIGN_ISSUE_KEYS, validateDesign } from './validate';

const keys = (design: EncounterDesign, targetCount = 1) =>
	validateDesign(design, targetCount).map(issue => `${issue.ref.kind}${issue.ref.index}:${issue.severity}:${issue.key}`);

const downtime = (design: EncounterDesign, start: number, end: number, targetIndex = 0) => {
	addEvent(design, 'downtime');
	const added = design.downtimes[design.downtimes.length - 1];
	added.start!.at = start;
	added.end!.at = end;
	added.targetIndex = targetIndex;
};

describe('validateDesign', () => {
	it('accepts an empty design and the defaults', () => {
		const design = EncounterDesign.create();
		expect(validateDesign(design, 1)).toEqual([]);
		for (const kind of ['bossHealth', 'ability', 'wave', 'move', 'amp'] as const) addEvent(design, kind);
		expect(validateDesign(design, 1)).toEqual([]);
	});

	it('flags an end at or before the start as an error', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.5, 0.5);
		expect(keys(design)).toContain(`downtime0:error:${DESIGN_ISSUE_KEYS.endBeforeStart}`);
	});

	it('warns when the window or its jitter leaves the fight', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.2, 0.4);
		design.downtimes[0].start!.jitter = 0.25;
		expect(keys(design)).toContain(`downtime0:warning:${DESIGN_ISSUE_KEYS.outsideFight}`);
		addEventAt(design, 'wave', 0.9);
		design.addWaves[0].spawn!.jitter = 0.15;
		expect(keys(design)).toContain(`wave0:warning:${DESIGN_ISSUE_KEYS.outsideFight}`);
	});

	it('lets a wave live past the fight end, which only its spawn must stay inside', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave');
		Object.assign(design.addWaves[0], { spawn: { at: 0.95, jitter: 0 }, lifetime: 0.2 });
		expect(keys(design)).toEqual([]);
	});

	it('warns about overlapping downtimes on one target but not on different targets', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.2, 0.4);
		downtime(design, 0.3, 0.5);
		expect(keys(design, 2)).toEqual([`downtime1:warning:${DESIGN_ISSUE_KEYS.downtimeOverlap}`]);
		design.downtimes[1].targetIndex = 1;
		expect(keys(design, 2).filter(k => k.endsWith(DESIGN_ISSUE_KEYS.downtimeOverlap))).toEqual([]);
	});

	it('flags what the sim skips: an amp of 0% or less, a wave of more than 64 adds, a move of no yards and no seconds', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'amp');
		design.damageAmps[0].percent = 0;
		addEvent(design, 'wave');
		design.addWaves[0].adds = Array.from({ length: 65 }, () => Target.create());
		addEvent(design, 'move');
		design.moves[0].yards = 0;
		expect(keys(design)).toEqual([
			`amp0:error:${DESIGN_ISSUE_KEYS.ampNotPositive}`,
			`wave0:error:${DESIGN_ISSUE_KEYS.tooManyAdds}`,
			`move0:error:${DESIGN_ISSUE_KEYS.moveGoesNowhere}`,
		]);
		design.damageAmps[0].percent = -10;
		design.addWaves[0].adds = Array.from({ length: 64 }, () => Target.create());
		design.moves[0].seconds = 2;
		expect(keys(design)).toEqual([`amp0:error:${DESIGN_ISSUE_KEYS.ampNotPositive}`]);
	});

	it('flags a wave that spawns no adds', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave');
		design.addWaves[0].adds = [];
		expect(keys(design)).toEqual([`wave0:error:${DESIGN_ISSUE_KEYS.noAdds}`]);
	});

	it('warns about a wave that spends more of its life below 20% than below 35%', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave');
		Object.assign(design.addWaves[0], { executeProportion20: 0.4, executeProportion35: 0.3 });
		expect(keys(design)).toEqual([`wave0:warning:${DESIGN_ISSUE_KEYS.executeOrder}`]);
	});

	it('lets a spell hit or a strike end at its start, while a window must end after it', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'ability');
		const ability = design.bossAbilities[0];
		ability.end!.at = ability.start!.at;
		expect(keys(design)).toEqual([]);
		ability.end!.at = ability.start!.at - 0.1;
		expect(keys(design)).toEqual([`ability0:error:${DESIGN_ISSUE_KEYS.endBeforeStart}`]);
		ability.kind = DesignBossAbility_Kind.Thrash;
		ability.end!.at = ability.start!.at;
		expect(keys(design)).toEqual([`ability0:error:${DESIGN_ISSUE_KEYS.endBeforeStart}`]);
	});

	it('flags two moves at the same instant', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'move', 0.3);
		addEventAt(design, 'move', 0.3);
		expect(keys(design)).toEqual([`move1:error:${DESIGN_ISSUE_KEYS.overlap}`]);
	});

	it('flags boss health that rises over time, in time order not list order', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'bossHealth', 0.5);
		design.bossHealth[0].healthPercent = 40;
		addEventAt(design, 'bossHealth', 0.2);
		design.bossHealth[1].healthPercent = 80;
		expect(keys(design)).toEqual([]);
		design.bossHealth[1].healthPercent = 30;
		expect(keys(design)).toEqual([`bossHealth0:error:${DESIGN_ISSUE_KEYS.healthIncreases}`]);
	});

	it('flags the boss health keyframes the sim drops past its 16 curve points, and warns when its own end point is what falls off', () => {
		const curve = (design: EncounterDesign) =>
			keys(design).filter(k => k.endsWith(DESIGN_ISSUE_KEYS.tooManyKeyframes) || k.endsWith(DESIGN_ISSUE_KEYS.healthStopsShort));
		const dropped = (index: number) => `bossHealth${index}:error:${DESIGN_ISSUE_KEYS.tooManyKeyframes}`;
		const stopsShort = (index: number) => `bossHealth${index}:warning:${DESIGN_ISSUE_KEYS.healthStopsShort}`;
		// Listed latest first, so list order and time order differ.
		const design = EncounterDesign.create();
		for (let step = 14; step >= 1; step--) {
			addEvent(design, 'bossHealth');
			Object.assign(design.bossHealth[design.bossHealth.length - 1], { at: step / 20, healthPercent: 100 - step * 5 });
		}
		// 14 keyframes and the sim's own start and end points fill the 16.
		expect(curve(design)).toEqual([]);

		design.bossHealth.unshift(HealthKeyframe.create({ at: 15 / 20, healthPercent: 25 }));
		expect(curve(design)).toEqual([stopsShort(0)]);

		design.bossHealth.unshift(HealthKeyframe.create({ at: 16 / 20, healthPercent: 20 }));
		expect(curve(design)).toEqual([dropped(0), stopsShort(1)]);

		// A first keyframe at the pull leaves the sim no start point to add.
		design.bossHealth[15].at = 0;
		expect(curve(design)).toEqual([stopsShort(0)]);

		// A last keyframe at 0% needs no end point.
		design.bossHealth[0].healthPercent = 0;
		expect(curve(design)).toEqual([]);

		addEventAt(design, 'bossHealth', 0.95);
		design.bossHealth[16].healthPercent = 0;
		expect(curve(design)).toEqual([dropped(16)]);
	});

	it('flags the abilities the sim skips for doing nothing, and accepts a new one of any kind', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'ability');
		const ability = design.bossAbilities[0];
		for (const kind of [
			DesignBossAbility_Kind.SpellHit,
			DesignBossAbility_Kind.Strike,
			DesignBossAbility_Kind.Thrash,
			DesignBossAbility_Kind.SwingTiming,
		]) {
			ability.kind = kind;
			expect(keys(design)).toEqual([]);
		}

		ability.kind = DesignBossAbility_Kind.Strike;
		ability.weaponPercent = 0;
		expect(keys(design)).toEqual([`ability0:error:${DESIGN_ISSUE_KEYS.noWeaponDamage}`]);

		ability.kind = DesignBossAbility_Kind.Thrash;
		expect(keys(design)).toEqual([]);
		ability.procChance = 0;
		expect(keys(design)).toEqual([`ability0:error:${DESIGN_ISSUE_KEYS.noExtraAttacks}`]);
		ability.procChance = 20;
		ability.extraAttacks = 0;
		expect(keys(design)).toEqual([`ability0:error:${DESIGN_ISSUE_KEYS.noExtraAttacks}`]);
	});

	it('warns about a wave at the very end of the fight, which never spawns', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'wave', 1);
		expect(keys(design)).toEqual([`wave0:warning:${DESIGN_ISSUE_KEYS.outsideFight}`]);
	});

	it("warns about a wave whose copies alive at once need more than the sim's 64 slots", () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave');
		// A spawn every 0.1 for a life of 0.25 keeps three copies up.
		Object.assign(design.addWaves[0], {
			adds: Array.from({ length: 21 }, () => Target.create()),
			spawn: { at: 0, jitter: 0 },
			lifetime: 0.25,
			repeat: { interval: 0.1, jitter: 0, count: 0 },
		});
		expect(keys(design)).toEqual([]);

		design.addWaves[0].adds.push(Target.create());
		expect(keys(design)).toEqual([`wave0:warning:${DESIGN_ISSUE_KEYS.waveOverPool}`]);

		// Two occurrences are all a count of 1 has, however long they live.
		design.addWaves[0].repeat!.count = 1;
		expect(keys(design)).toEqual([]);

		// Jitter on the repeat brings the spawns closer, so a fourth copy overlaps.
		design.addWaves[0].repeat!.count = 0;
		design.addWaves[0].adds.pop();
		design.addWaves[0].repeat!.jitter = 0.01;
		expect(keys(design)).toEqual([`wave0:warning:${DESIGN_ISSUE_KEYS.waveOverPool}`]);
	});

	it('flags a debuff amp whose wave was removed, and keeps the wave after it', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave');
		addEvent(design, 'wave');
		addEvent(design, 'amp');
		design.damageAmps[0].type = DamageAmpType.DamageAmpTypeTargetDebuff;
		design.damageAmps[0].targets!.waveIndices = [1];
		removeEvent(design, { kind: 'wave', index: 0 });
		expect(keys(design)).toEqual([]);
		removeEvent(design, { kind: 'wave', index: 0 });
		expect(keys(design)).toEqual([`amp0:error:${DESIGN_ISSUE_KEYS.waveOutOfRange}`]);
	});

	it('flags targets out of range, including -1', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.2, 0.4, 2);
		addEvent(design, 'ability');
		design.bossAbilities[0].targetIndex = -1;
		addEvent(design, 'amp');
		design.damageAmps[0].targets!.targetIndices = [0, 3];
		expect(keys(design, 2)).not.toContain(`amp0:error:${DESIGN_ISSUE_KEYS.targetOutOfRange}`);
		design.damageAmps[0].type = DamageAmpType.DamageAmpTypeTargetDebuff;
		design.damageAmps[0].targets!.waveIndices = [0];
		const result = keys(design, 2);
		expect(result).toContain(`downtime0:error:${DESIGN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`ability0:error:${DESIGN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`amp0:error:${DESIGN_ISSUE_KEYS.targetOutOfRange}`);
		expect(result).toContain(`amp0:error:${DESIGN_ISSUE_KEYS.waveOutOfRange}`);
	});

	it('reports info when every target is down at the same instant', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.2, 0.4, 0);
		downtime(design, 0.3, 0.5, 1);
		const result = keys(design, 2);
		expect(result).toContain(`downtime0:info:${DESIGN_ISSUE_KEYS.allTargetsDown}`);
		expect(result).toContain(`downtime1:info:${DESIGN_ISSUE_KEYS.allTargetsDown}`);
		expect(idleSpans(design, 2)).toEqual([
			[0, 0.3, 0.4],
			[1, 0.3, 0.4],
		]);
		design.downtimes[1].start!.at = 0.45;
		expect(keys(design, 2).filter(k => k.includes(':info:'))).toEqual([]);
	});

	const idleSpans = (design: EncounterDesign, targetCount = 1) =>
		validateDesign(design, targetCount)
			.filter(issue => issue.key === DESIGN_ISSUE_KEYS.allTargetsDown)
			.map(issue => [issue.ref.index, issue.span!.start, issue.span!.end]);

	const wave = (design: EncounterDesign, at: number, lifetime: number, { adds = 1, interval = 0 } = {}) =>
		design.addWaves.push(
			AddWave.create({
				adds: Array.from({ length: adds }, () => Target.create()),
				spawn: { at, jitter: 0 },
				lifetime,
				repeat: { interval, count: 0, jitter: 0 },
			}),
		);

	it('reports no idle stretch where an add is up for the whole downtime', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.4, 0.5);
		wave(design, 0.35, 0.2);
		expect(idleSpans(design)).toEqual([]);
	});

	it('names the part of a downtime no add covers, and each part when an add splits it', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.4, 0.5);
		wave(design, 0.45, 0.1);
		expect(idleSpans(design)).toEqual([[0, 0.4, 0.45]]);

		design.addWaves[0].spawn!.at = 0.4375;
		design.addWaves[0].lifetime = 0.03125;
		expect(idleSpans(design)).toEqual([
			[0, 0.4, 0.4375],
			[0, 0.46875, 0.5],
		]);
	});

	it("counts each occurrence of a repeating wave as an add's life", () => {
		const design = EncounterDesign.create();
		downtime(design, 0.375, 0.5);
		wave(design, 0.125, 0.0625, { interval: 0.25 });
		expect(idleSpans(design)).toEqual([[0, 0.4375, 0.5]]);
	});

	it('counts no add of a wave the sim skips', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.4, 0.5);
		wave(design, 0.4, 0.1, { adds: 0 });
		wave(design, 0.4, 0.1, { adds: 65 });
		wave(design, 0.4, 0, {});
		expect(idleSpans(design)).toEqual([[0, 0.4, 0.5]]);
	});

	it('flags a wave with no lifetime', () => {
		const design = EncounterDesign.create();
		addEventAt(design, 'wave', 0.2);
		design.addWaves[0].lifetime = 0;
		expect(keys(design)).toEqual([`wave0:error:${DESIGN_ISSUE_KEYS.endBeforeStart}`]);
	});

	it('applies each time its own jitter to the outside-fight check', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.4, 0.9);
		design.downtimes[0].start!.jitter = 0.3;
		expect(keys(design).filter(k => k.includes(':warning:'))).toEqual([]);
		design.downtimes[0].end!.jitter = 0.2;
		expect(keys(design).filter(k => k.includes(':warning:'))).toEqual([`downtime0:warning:${DESIGN_ISSUE_KEYS.outsideFight}`]);
	});
});
