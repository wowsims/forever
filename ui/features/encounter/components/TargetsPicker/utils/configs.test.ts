import { Target as TargetProto } from '@generated/proto/common';
import type { Encounter } from '@sim/raid/encounter';
import { describe, expect, it } from 'vitest';

import { bleedImmuneConfig } from './configs';

describe('bleedImmuneConfig', () => {
	it('shows Immune to Bleeds off by default', () => {
		const target = TargetProto.create();
		const encounter = {
			modifyTarget: (_: number, modify: (target: TargetProto) => void) => modify(target),
		} as unknown as Encounter;
		const bleedImmune = bleedImmuneConfig({ encounter, targetIndex: 0, getTarget: () => target });

		expect(bleedImmune.getValue?.(null)).toBe(false);
		bleedImmune.setValue?.(null, true);
		expect(target.bleedImmune).toBe(true);
		expect(bleedImmune.getValue?.(null)).toBe(true);
	});
});
