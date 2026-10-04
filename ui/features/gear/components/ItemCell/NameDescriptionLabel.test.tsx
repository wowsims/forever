import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { NameDescriptionLabel } from './NameDescriptionLabel';

describe('NameDescriptionLabel', () => {
	it.each(['Heroic', ''])('wraps "%s" in parentheses', nameDescription => {
		const { container } = render(<NameDescriptionLabel nameDescription={nameDescription} />);

		expect(container.textContent).toBe(`(${nameDescription})`);
	});
});
