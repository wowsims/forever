import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { MenuSection } from './MenuSection';

describe('MenuSection', () => {
	it('titles the section', () => {
		render(
			<MenuSection name="Weapon Speed">
				<span>control</span>
			</MenuSection>,
		);
		expect(screen.getByTestId('menu-section-title').textContent).toBe('Weapon Speed');
	});
});
