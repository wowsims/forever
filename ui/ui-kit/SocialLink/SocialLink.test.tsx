import i18n from '@i18n/config';
import { SOCIALS } from '@sim/constants/other';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SocialLink } from './SocialLink';

const [discord, , patreon] = SOCIALS;

describe('SocialLink', () => {
	it('renders an external link named by its tooltip, opening a new tab safely', () => {
		render(<SocialLink social={discord} testId="social-discord" />);
		const link = screen.getByRole('link', { name: i18n.t(discord.tooltip) });
		expect(link.getAttribute('href')).toBe(discord.href);
		expect(link.getAttribute('target')).toBe('_blank');
		expect(link.getAttribute('rel')).toBe('noopener noreferrer');
		expect(link.getAttribute('data-testid')).toBe('social-discord');
		const icon = link.querySelector('svg')!;
		expect(icon.getAttribute('data-icon')).toBe('discord');
		expect(icon.getAttribute('data-prefix')).toBe('fab');
	});

	it('renders the label beside the icon only for a social that has one', () => {
		const { unmount } = render(<SocialLink social={patreon} />);
		expect(screen.getByRole('link').textContent).toBe(patreon.label);
		unmount();

		render(<SocialLink social={discord} />);
		expect(screen.getByRole('link').textContent).toBe('');
	});

	it('shows its tooltip on hover and nothing before', () => {
		render(<SocialLink social={discord} />);
		const tooltip = i18n.t(discord.tooltip);
		expect(screen.queryByText(tooltip)).toBeNull();

		fireEvent.mouseEnter(screen.getByRole('link'));
		expect(screen.getByText(tooltip)).toBeTruthy();
	});
});
