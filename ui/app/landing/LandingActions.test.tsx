import i18n from '@i18n/config';
import { ADDON_URL, DISCORD_URL, GITHUB_ORG_URL } from '@sim/constants/other';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { LandingActions } from './LandingActions';

describe('LandingActions', () => {
	it('links Discord, GitHub and the exporter addon in a new tab with no opener', () => {
		render(<LandingActions />);
		const links: Array<[string, string]> = [
			[i18n.t('landing.actions.discord'), DISCORD_URL],
			[i18n.t('landing.actions.github'), GITHUB_ORG_URL],
			[i18n.t('landing.actions.addon'), ADDON_URL],
		];

		links.forEach(([name, href]) => {
			const link = screen.getByRole('link', { name });
			expect(link.getAttribute('href')).toBe(href);
			expect(link.getAttribute('target')).toBe('_blank');
			expect(link.getAttribute('rel')).toBe('noopener noreferrer');
		});
	});
});
