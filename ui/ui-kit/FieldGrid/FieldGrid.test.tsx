import { render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it } from 'vitest';

import { FieldGrid } from './FieldGrid';

describe('FieldGrid', () => {
	it('renders its children and forwards native attributes and the ref', () => {
		const ref = createRef<HTMLDivElement>();
		render(
			<FieldGrid ref={ref} testId="group" role="group" aria-label="Consumes">
				<span>Flask</span>
			</FieldGrid>,
		);
		const group = screen.getByRole('group', { name: 'Consumes' });
		expect(ref.current).toBe(group);
		expect(group.textContent).toBe('Flask');
	});

	it('lets an explicit data-testid stand in for testId', () => {
		render(<FieldGrid data-testid="consumes-food" />);
		expect(screen.getByTestId('consumes-food')).toBeTruthy();
	});
});
