import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translatePlayerSpec } from '@i18n/localization';
import { PlayerClasses } from '@sim/player/classes/index';
import { PlayerSpecs } from '@sim/player/specs/index';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SimTitleDropdown } from './SimTitleDropdown';

const mount = () => render(<SimTitleDropdown currentSpec={PlayerSpecs.fromProto(Spec.SpecDpsWarrior)} />);

const open = (element: HTMLElement) => act(() => void fireEvent.click(element));

const classRows = () => within(screen.getAllByTestId('sim-title-popup')[0]).getAllByTestId('sim-link');

const openRoot = () => {
	mount();
	open(within(screen.getByTestId('sim-link-dropdown')).getByTestId('sim-link'));
};

describe('SimTitleDropdown', () => {
	it('lists every class as a submenu menuitem', () => {
		openRoot();

		const rows = classRows();
		expect(rows).toHaveLength(PlayerClasses.naturalOrder.length);
		expect(rows.map(row => row.getAttribute('role'))).toEqual(rows.map(() => 'menuitem'));
		expect(rows.map(row => row.getAttribute('aria-haspopup'))).toEqual(rows.map(() => 'menu'));
		expect(within(screen.getAllByTestId('sim-title-popup')[0]).getAllByRole('menuitem')).toHaveLength(PlayerClasses.naturalOrder.length);
	});

	it('opens a class row onto that class’s spec links', () => {
		openRoot();
		const warrior = PlayerClasses.Warrior;
		const row = classRows()[PlayerClasses.naturalOrder.indexOf(warrior)];

		open(row);

		expect(row.getAttribute('aria-expanded')).toBe('true');
		const links = within(screen.getAllByTestId('sim-title-popup')[1]).getAllByRole('menuitem') as Array<HTMLAnchorElement>;
		expect(links.map(link => new URL(link.href).pathname)).toEqual(Object.values(warrior.specs).map(spec => spec.simLink));
	});

	it('titles a one-sim class’s link All Specializations, and a spec’s link by its spec', () => {
		openRoot();
		const specTitles = (playerClass: (typeof PlayerClasses.naturalOrder)[number]) => {
			open(classRows()[PlayerClasses.naturalOrder.indexOf(playerClass)]);
			const popup = screen.getAllByTestId('sim-title-popup').at(-1)!;
			return within(popup)
				.getAllByTestId('sim-link-title')
				.map(title => title.textContent);
		};

		expect(specTitles(PlayerClasses.Hunter)).toEqual([i18n.t('landing.home.allSpecs')]);
		expect(specTitles(PlayerClasses.Warrior)).toEqual(Object.values(PlayerClasses.Warrior.specs).map(spec => translatePlayerSpec(spec)));
	});
});
