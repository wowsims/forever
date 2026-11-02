import { PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { INCLUDED_STATS, isIncludedStat } from './utils';

describe('INCLUDED_STATS', () => {
	it('offers expertise as a percent only', () => {
		expect(isIncludedStat(UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent))).toBe(true);
		expect(isIncludedStat(UnitStat.fromStat(Stat.StatExpertiseRating))).toBe(false);
	});

	it('keeps dodge and parry rating, which have no percent in the list', () => {
		expect(isIncludedStat(UnitStat.fromStat(Stat.StatDodgeRating))).toBe(true);
		expect(isIncludedStat(UnitStat.fromStat(Stat.StatParryRating))).toBe(true);
		expect(INCLUDED_STATS.some(stat => stat.equalsPseudoStat(PseudoStat.PseudoStatDodgePercent))).toBe(false);
	});
});
