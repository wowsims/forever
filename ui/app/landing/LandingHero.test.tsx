import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../assets/locales/en/translation.json';
import { LandingHero } from './LandingHero';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('LandingHero', () => {
	it('heads the page with the expansion name and renders the welcome text without its markup tags', () => {
		render(<LandingHero />);
		expect(screen.getByRole('heading', { level: 1 }).textContent).toBe('Forever Simulations');
		const blurb = document.getElementById('description');
		expect(blurb?.textContent).toMatch(/^WoWSims is a fan-made open-source project/);
		expect(blurb?.textContent).not.toMatch(/<\w+>/);
	});
});
