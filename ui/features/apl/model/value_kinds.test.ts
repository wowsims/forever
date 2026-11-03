import { describe, expect, it, vi } from 'vitest';

import { valueKinds } from './value_kinds';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

describe('value kinds', () => {
	it('gives the execute phase an optional target', () => {
		expect(valueKinds.isExecutePhase.fields).toContainEqual({ type: 'unit', field: 'targetUnit', unitSet: 'targets' });
	});
});
