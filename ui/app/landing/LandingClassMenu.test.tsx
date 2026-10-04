import { PlayerClasses } from '@sim/player/classes/index';
import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { LandingClassMenu } from './LandingClassMenu';

const specsOf = (playerClass: (typeof PlayerClasses.naturalOrder)[number]) => Object.values(playerClass.specs);
const multiSpec = PlayerClasses.naturalOrder.find(playerClass => specsOf(playerClass).length > 1)!;
const singleSpec = PlayerClasses.naturalOrder.find(playerClass => specsOf(playerClass).length === 1)!;

describe('LandingClassMenu', () => {
	it('opens onto one link per spec', () => {
		render(<LandingClassMenu playerClass={multiSpec} />);
		fireEvent.click(document.querySelector('[data-testid="sim-link-dropdown"] > [data-testid="sim-link"]')!);
		const popup = document.querySelector('.ui-landing-sim-link-popup')!;
		expect(popup.querySelectorAll('a[role="menuitem"]')).toHaveLength(specsOf(multiSpec).length);
	});

	it('keeps the spec links in the document while closed, for a crawler', () => {
		render(<LandingClassMenu playerClass={multiSpec} />);
		expect(document.querySelectorAll('.ui-landing-sim-link-popup a')).toHaveLength(specsOf(multiSpec).length);
	});

	it('links straight to the only spec of a one-spec class, with no menu', () => {
		const { getByTestId } = render(<LandingClassMenu playerClass={singleSpec} />);
		expect(getByTestId('sim-link').getAttribute('href')).toBe(specsOf(singleSpec)[0].simLink);
		expect(getByTestId('sim-link').hasAttribute('rel')).toBe(false);
		expect(document.querySelector('.ui-landing-sim-link-popup')).toBeNull();
	});

	it('uses the generic class crest for the priest, not its spell icon', () => {
		render(<LandingClassMenu playerClass={PlayerClasses.Priest} />);
		// Two priest specs now, so the crest sits on the menu trigger and each spec link carries its own icon.
		const trigger = document.querySelector('[data-testid="sim-link-dropdown"] > [data-testid="sim-link"]')!;
		expect(trigger.querySelector('img')!.getAttribute('src')).toContain('class_priest.jpg');
		const specIcons = [...document.querySelectorAll('.ui-landing-sim-link-popup img')].map(img => img.getAttribute('src'));
		expect(specIcons).toEqual(specsOf(PlayerClasses.Priest).map(spec => spec.getIcon('large')));
		expect(specIcons.some(src => src!.includes('class_priest.jpg'))).toBe(false);
	});
});
