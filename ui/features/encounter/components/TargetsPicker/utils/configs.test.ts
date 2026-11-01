import { Target as TargetProto } from '@generated/proto/common';
import { describe, expect, it } from 'vitest';

import { bleedImmuneConfig } from './configs';

describe('bleedImmuneConfig', () => {
	it('shows Immune to Bleeds off by default', () => {
		const target = TargetProto.create();
		const bleedImmune = bleedImmuneConfig({
			idPrefix: 'target-0',
			getTarget: () => target,
			modifyTarget: edit => edit(target),
			storeSubscribe: () => () => () => {},
		});

		expect(bleedImmune.getValue?.(null)).toBe(false);
		bleedImmune.setValue?.(null, true);
		expect(target.bleedImmune).toBe(true);
		expect(bleedImmune.getValue?.(null)).toBe(true);
	});
});
