import i18n from '@i18n/config';
import { translatePlayerSpec, translateStatus } from '@i18n/localization';
import { LaunchStatus } from '@sim/constants/other';
import { PlayerClasses } from '@sim/player/classes/index';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { classLaunchStatus, SPEC_THEME_CLASS } from './landing_classes';
import { LandingClassCard } from './LandingClassCard';

const specsOf = (playerClass: (typeof PlayerClasses.naturalOrder)[number]) => Object.values(playerClass.specs);
const multiSpec = PlayerClasses.naturalOrder.find(playerClass => specsOf(playerClass).length > 1)!;
const singleSpec = PlayerClasses.naturalOrder.find(playerClass => specsOf(playerClass).length === 1)!;

describe('LandingClassCard', () => {
	it('links every spec of a multi-spec class from the card', () => {
		render(<LandingClassCard playerClass={multiSpec} index={0} />);
		const links = screen.getAllByTestId('spec-link');
		expect(links.map(link => link.getAttribute('href'))).toEqual(specsOf(multiSpec).map(spec => spec.simLink));
		expect(links.map(link => link.textContent)).toEqual(specsOf(multiSpec).map(spec => translatePlayerSpec(spec)));
		expect(screen.getByTestId('class-card').tagName).not.toBe('A');
	});

	it('gives a one-sim class a single All Specializations link with the class icon, and keeps the card itself a plain element', () => {
		render(<LandingClassCard playerClass={singleSpec} index={0} />);
		const [link] = screen.getAllByTestId('spec-link');
		expect(screen.getAllByTestId('spec-link')).toHaveLength(1);
		expect(link.getAttribute('href')).toBe(specsOf(singleSpec)[0].simLink);
		expect(link.hasAttribute('rel')).toBe(false);
		expect(link.textContent).toBe(i18n.t('landing.home.allSpecs'));
		expect(link.querySelector('img')!.getAttribute('src')).toBe(singleSpec.getIcon('large'));
		expect(screen.getByTestId('class-card').tagName).not.toBe('A');
	});

	it("covers the card with the background of one of the class's own specs", () => {
		render(<LandingClassCard playerClass={multiSpec} index={0} />);
		const cover = screen.getByTestId('class-card').querySelector('[aria-hidden="true"]')!;
		const own = specsOf(multiSpec).map(spec => SPEC_THEME_CLASS[spec.specID as keyof typeof SPEC_THEME_CLASS]);
		expect([...cover.classList].filter(name => name.endsWith('-sim-ui'))).toEqual([expect.toBeOneOf(own)]);
	});

	it('tints the card by its class', () => {
		render(<LandingClassCard playerClass={multiSpec} index={0} />);
		expect(screen.getByTestId('class-card').dataset.class).toBe(PlayerClasses.getCssScheme(multiSpec));
	});

	it('shows the class status under its name unless the sim has launched', () => {
		render(<LandingClassCard playerClass={multiSpec} index={0} />);
		const badge = screen.queryByTestId('launch-status-label');
		if (classLaunchStatus(multiSpec) === LaunchStatus.Launched) expect(badge).toBeNull();
		else expect(badge!.textContent).toBe(translateStatus(classLaunchStatus(multiSpec)));
	});

	it('uses the generic class crest for the priest, and each spec link its own icon', () => {
		render(<LandingClassCard playerClass={PlayerClasses.Priest} index={0} />);
		const card = screen.getByTestId('class-card');
		const [crest, ...specIcons] = [...card.querySelectorAll('img')].map(img => img.getAttribute('src'));
		expect(crest).toContain('class_priest.jpg');
		expect(specIcons).toEqual(specsOf(PlayerClasses.Priest).map(spec => spec.getIcon('large')));
	});
});
