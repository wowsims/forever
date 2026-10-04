import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { PlayerSpecs } from '@sim/player/specs';
import { UnitStat } from '@sim/proto/stats';
import type { SpecDefinition } from '@sim/spec_config';
import { describe, expect, it } from 'vitest';

const specs = Object.entries(import.meta.glob<SpecDefinition<any>>('./*/*/spec.ts', { eager: true, import: 'default' }));

const RANGED_ATTACK_STATS = [
	UnitStat.fromStat(Stat.StatRangedAttackPower),
	...[
		PseudoStat.PseudoStatRangedHitPercent,
		PseudoStat.PseudoStatRangedCritPercent,
		PseudoStat.PseudoStatRangedHastePercent,
		PseudoStat.PseudoStatRangedSpeedMultiplier,
	].map(pseudoStat => UnitStat.fromPseudoStat(pseudoStat)),
];

describe('displayStats', () => {
	it.each(specs.filter(([, def]) => PlayerSpecs.fromProto(def.spec).classID !== Class.ClassHunter))('%s lists no ranged attack stat', (_file, def) => {
		const ranged = def.displayStats.filter(stat => RANGED_ATTACK_STATS.some(rangedStat => rangedStat.equals(stat)));
		expect(ranged.map(stat => stat.getKey())).toEqual([]);
	});
});
