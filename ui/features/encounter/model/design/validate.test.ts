import { DamageAmpType, EncounterDesign } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { addEvent } from './lanes';
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
		addEvent(design, 'wave', 0.95);
		design.addWaves[0].spawn!.at = 0.95;
		expect(keys(design)).toContain(`wave0:warning:${DESIGN_ISSUE_KEYS.outsideFight}`);
	});

	it('flags overlapping downtimes on one target but not on different targets', () => {
		const design = EncounterDesign.create();
		downtime(design, 0.2, 0.4);
		downtime(design, 0.3, 0.5);
		expect(keys(design, 2)).toContain(`downtime1:error:${DESIGN_ISSUE_KEYS.overlap}`);
		design.downtimes[1].targetIndex = 1;
		expect(keys(design, 2).filter(k => k.endsWith(DESIGN_ISSUE_KEYS.overlap))).toEqual([]);
	});

	it('flags two moves at the same instant', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'move', 0.3);
		addEvent(design, 'move', 0.3);
		expect(keys(design)).toEqual([`move1:error:${DESIGN_ISSUE_KEYS.overlap}`]);
	});

	it('flags boss health that rises over time, in time order not list order', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'bossHealth', 0.5);
		design.bossHealth[0].healthPercent = 40;
		addEvent(design, 'bossHealth', 0.2);
		design.bossHealth[1].healthPercent = 80;
		expect(keys(design)).toEqual([]);
		design.bossHealth[1].healthPercent = 30;
		expect(keys(design)).toEqual([`bossHealth0:error:${DESIGN_ISSUE_KEYS.healthIncreases}`]);
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
		design.downtimes[1].start!.at = 0.45;
		expect(keys(design, 2).filter(k => k.includes(':info:'))).toEqual([]);
	});

	it('flags a wave with no lifetime', () => {
		const design = EncounterDesign.create();
		addEvent(design, 'wave', 0.2);
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
