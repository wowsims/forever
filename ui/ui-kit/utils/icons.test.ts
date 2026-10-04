import { describe, expect, it } from 'vitest';

import { resolvedIconStyle } from './icons';

describe('resolvedIconStyle', () => {
	it('paints the icon as a quoted background image', () => {
		expect(resolvedIconStyle({ iconUrl: 'https://example.com/kings.jpg', name: 'Kings' })).toEqual({
			backgroundImage: "url('https://example.com/kings.jpg')",
		});
	});

	it('leaves the style off without an icon', () => {
		expect(resolvedIconStyle(undefined)).toBeUndefined();
	});
});
