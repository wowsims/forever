import { Class, ConsumableType, PseudoStat, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { DatabaseFilters } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';
import { createSimStore, patchSlice, type SimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';

const agility = Object.assign(new Array<number>(42).fill(0), { [Stat.StatAgility]: 25 });
const expertise = Object.assign(new Array<number>(42).fill(0), { [Stat.StatExpertiseRating]: 20 });
const CONSUMABLES = [
	Consumable.create({ id: 9187, name: 'Elixir of Greater Agility', type: ConsumableType.ConsumableTypeElixir, lvl: 38, stats: agility }),
	Consumable.create({ id: 9188, name: 'Elixir of Expertise', type: ConsumableType.ConsumableTypeElixir, lvl: 38, stats: expertise }),
];

vi.mock('@sim/proto/database', () => ({ Database: { getSync: () => ({ getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined }) } }));

const { ConsumesFilters } = await import('./ConsumesFilters');

let store: SimStore;
let sim: {
	store: SimStore;
	getFilters: () => DatabaseFilters;
	setFilters: (filters: DatabaseFilters) => void;
	getMinConsumableLvl: (type: ConsumableType) => number;
};

const player = {
	getClass: () => Class.ClassWarrior,
	getSpecConfig: () => ({
		epStats: [Stat.StatStrength, Stat.StatAttackPower, Stat.StatExpertiseRating],
		displayStats: UnitStat.createDisplayStatArray([], [PseudoStat.PseudoStatExpertisePercent]),
	}),
};

const open = () => {
	renderWithHost(<ConsumesFilters />, { sim, player, getStorageKey: (keyPart: string) => `test${keyPart}` });
	fireEvent.click(screen.getByTestId('consumes-filters-button'));
};

const checkbox = (stat: Stat) => screen.getByRole('checkbox', { name: UnitStat.fromStat(stat).getShortName(Class.ClassWarrior) });
const levelField = () => screen.queryByLabelText(i18n.t('settings_tab.consumables.filters.types.elixir'));
const showTab = (labelKey: string) => fireEvent.click(screen.getByRole('tab', { name: i18n.t(labelKey) }));
const shownInSlot = () => sim.getFilters().statFilters?.consumables?.stats.map(stat => UnitStat.fromProto(stat).getKey());

beforeEach(() => {
	store = createSimStore();
	sim = {
		store,
		getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
		setFilters: filters => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
		getMinConsumableLvl: type => store.getState().sim.filters.minConsumableLvl[type] ?? 0,
	};
});

// The test setup loads no locale, so every stat reads as its own key; the rating and percent only share a name in English.
describe('ConsumesFilters', () => {
	beforeAll(() => {
		i18n.addResourceBundle('en', 'translation', english, true, true);
	});
	afterAll(() => {
		i18n.removeResourceBundle('en', 'translation');
	});

	it('opens under its title on the stats tab, with the level fields on the other tab', () => {
		open();

		expect(screen.getByRole('button', { name: i18n.t('settings_tab.consumables.filters.trigger') })).toBe(screen.getByTestId('consumes-filters-button'));
		expect(screen.getByRole('dialog', { name: i18n.t('settings_tab.consumables.filters.title') })).toBeTruthy();
		expect(screen.getAllByRole('tab').map(tab => tab.textContent)).toEqual([
			i18n.t('stat_filters.title'),
			i18n.t('settings_tab.consumables.filters.min_lvl'),
		]);
		expect(screen.getByRole('tab', { selected: true }).textContent).toBe(i18n.t('stat_filters.title'));
		expect(checkbox(Stat.StatStrength)).toBeTruthy();
		expect(levelField()).toBeNull();
	});

	it('shows the level fields in place of the stats on the level tab', async () => {
		open();

		showTab('settings_tab.consumables.filters.min_lvl');

		expect(levelField()).toBeTruthy();
		await waitFor(() => expect(screen.queryAllByRole('checkbox')).toHaveLength(0));
		expect(screen.getByTestId('consumes-filters-reset')).toBeTruthy();
	});

	it('checks the spec stats and a stat only a consumable grants unchecked', () => {
		open();

		expect(checkbox(Stat.StatStrength).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(Stat.StatAgility).getAttribute('aria-checked')).toBe('false');
	});

	it('writes a toggle to the consumables slot and leaves the levels alone', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();

		fireEvent.click(checkbox(Stat.StatAgility));

		expect(shownInSlot()).toContain('StatAgility');
		expect(sim.getFilters().minConsumableLvl).toEqual({ [ConsumableType.ConsumableTypeElixir]: 50 });
	});

	it('unticks the Expertise row for the expertise rating the spec weighs along with its percent', () => {
		open();
		const expertiseRow = () => screen.getByRole('checkbox', { name: 'Expertise' });
		expect(expertiseRow().getAttribute('aria-checked')).toBe('true');

		fireEvent.click(expertiseRow());

		expect(expertiseRow().getAttribute('aria-checked')).toBe('false');
		expect(shownInSlot()).not.toContain('StatExpertiseRating');
		expect(shownInSlot()).not.toContain('PseudoStatExpertisePercent');
	});

	it('offers the level reset only while a minimum level is set', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();
		showTab('settings_tab.consumables.filters.min_lvl');
		const reset = screen.getByTestId('consumes-filters-reset');
		expect(reset.hasAttribute('disabled')).toBe(false);

		fireEvent.click(reset);
		expect(reset.hasAttribute('disabled')).toBe(true);
	});

	it('marks the trigger customised while the stats or a minimum level differ from the defaults', () => {
		const customised = () => screen.getByTestId('consumes-filters-button').hasAttribute('data-customised');
		open();
		expect(customised()).toBe(false);

		fireEvent.click(checkbox(Stat.StatAgility));
		expect(customised()).toBe(true);
		fireEvent.click(screen.getByTestId('consumes-filters-stats-reset'));
		expect(customised()).toBe(false);

		act(() => sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } })));
		expect(customised()).toBe(true);
	});

	it('resets the stats and the levels apart', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();
		fireEvent.click(checkbox(Stat.StatAgility));

		showTab('settings_tab.consumables.filters.min_lvl');
		fireEvent.click(screen.getByTestId('consumes-filters-reset'));
		expect(sim.getFilters().minConsumableLvl).toEqual({});
		expect(shownInSlot()).toContain('StatAgility');

		showTab('stat_filters.title');
		fireEvent.click(screen.getByTestId('consumes-filters-stats-reset'));
		expect(shownInSlot()).toBeUndefined();
	});
});
