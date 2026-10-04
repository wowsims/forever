import { describe, expect, it } from 'vitest';

import type { PickerLayout } from './input';

describe('PickerLayout', () => {
	it('rejects icon, which only the icon pickers set for themselves', () => {
		// @ts-expect-error icon is excluded from PickerLayout, unlike FieldLayout.
		const layout: PickerLayout = 'icon';
		expect(layout).toBe('icon');
	});
});
