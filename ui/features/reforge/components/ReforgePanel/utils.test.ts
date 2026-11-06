import { PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { isIncludedStat } from './utils';

describe('INCLUDED_STATS', () => {
	it('offers expertise as a percent only', () => {
		expect(isIncludedStat(UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent))).toBe(true);
		expect(isIncludedStat(UnitStat.fromStat(Stat.StatExpertiseRating))).toBe(false);
	});
});
