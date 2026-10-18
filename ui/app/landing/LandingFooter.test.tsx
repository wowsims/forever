import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import english from '../../../assets/locales/en/translation.json';
import { LandingFooter } from './LandingFooter';

beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('LandingFooter', () => {
	it('runs the copyright through the current year', () => {
		render(<LandingFooter />);
		expect(screen.getByRole('contentinfo').textContent).toContain(`© 2021-${new Date().getFullYear()} WoWSims team`);
	});
});
