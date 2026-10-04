import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SimLinkContent } from './SimLinkContent';

describe('SimLinkContent', () => {
	it('renders the icon and the title, and no label or status it was not given', () => {
		render(<SimLinkContent iconPath="/mage.png" title="Mage" />);

		expect(screen.getByTestId('sim-link-content')).not.toBeNull();
		expect(screen.getByTestId('sim-link-icon').getAttribute('src')).toBe('/mage.png');
		expect(screen.getByTestId('sim-link-icon').getAttribute('alt')).toBe('');
		expect(screen.getByTestId('sim-link-title').textContent).toBe('Mage');
		expect(screen.queryByTestId('sim-link-label')).toBeNull();
		expect(screen.queryByTestId('launch-status-label')).toBeNull();
	});

	it('renders the label above the title and the status below it', () => {
		render(<SimLinkContent iconPath="/mage.png" label="Mage" title="Fire" status="Phase 1 · Beta" />);

		const texts = [...screen.getByTestId('sim-link-title').parentElement!.children].map(child => child.getAttribute('data-testid'));
		expect(texts).toEqual(['sim-link-label', 'sim-link-title', 'launch-status-label']);
		expect(screen.getByTestId('launch-status-label').textContent).toBe('Phase 1 · Beta');
	});

	it('renders no status for an empty one', () => {
		render(<SimLinkContent iconPath="/mage.png" title="Mage" status="" />);

		expect(screen.queryByTestId('launch-status-label')).toBeNull();
	});

	it('takes its testId', () => {
		render(<SimLinkContent iconPath="/mage.png" title="Fire" testId="fire-link" />);

		expect(screen.getByTestId('fire-link')).not.toBeNull();
		expect(screen.queryByTestId('sim-link-content')).toBeNull();
	});
});
