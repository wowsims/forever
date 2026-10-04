import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Skeleton } from './Skeleton';

describe('Skeleton', () => {
	it('carries a default testId and takes another', () => {
		const { container, rerender } = render(<Skeleton />);
		expect(container.firstElementChild!.getAttribute('data-testid')).toBe('skeleton');

		rerender(<Skeleton testId="dps-skeleton" />);
		expect(container.firstElementChild!.getAttribute('data-testid')).toBe('dps-skeleton');
	});
});
