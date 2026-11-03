import { Class, Stat } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { createSimStore, patchSlice, type SimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { StatFilterPopover } from '../components/StatFilterPopover';
import { type StatFilterScope, type StatFilterSlot, withSelection } from '../model/stat_filters';
import { useStatFilter } from './useStatFilter';

const strength = UnitStat.fromStat(Stat.StatStrength);
const agility = UnitStat.fromStat(Stat.StatAgility);
const stamina = UnitStat.fromStat(Stat.StatStamina);
const hitRating = UnitStat.fromStat(Stat.StatHitRating);
const armor = UnitStat.fromStat(Stat.StatArmor);

const scope: StatFilterScope = { defaults: [strength, stamina], universe: [strength, agility, stamina] };

let store: SimStore;
let sim: { store: SimStore; getFilters: () => DatabaseFilters; setFilters: (filters: DatabaseFilters) => void };
let latest: ReturnType<typeof useStatFilter>;

const Probe = () => {
	latest = useStatFilter('buffs', scope);
	return <span>{latest.shown.map(stat => stat.getKey()).join(',')}</span>;
};

const mount = () => renderWithHost(<Probe />, { sim, player: { getClass: () => Class.ClassWarrior }, getStorageKey: (keyPart: string) => `test${keyPart}` });

const keys = (stats: ReadonlyArray<UnitStat>) => stats.map(stat => stat.getKey()).join(',');

beforeEach(() => {
	window.localStorage.clear();
	store = createSimStore();
	sim = {
		store,
		getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
		setFilters: filters => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
	};
});

describe('useStatFilter', () => {
	it('shows the scope defaults while the slot is unset', () => {
		const { container } = mount();

		expect(container.textContent).toBe(keys([strength, stamina]));
		expect(latest.customised).toBe(false);
	});

	it('writes a new list through the sim and shows it', () => {
		const { container } = mount();

		act(() => latest.onChange([agility]));

		expect(sim.getFilters().statFilters?.buffs).toBeDefined();
		expect(container.textContent).toBe(keys([agility]));
		expect(latest.customised).toBe(true);
	});

	it('marks the slot customised when it is set to the defaults', () => {
		mount();
		const before = latest.shown;

		act(() => latest.onChange([strength, stamina]));

		expect(latest.customised).toBe(true);
		expect(latest.shown).toBe(before);
	});

	it('clears the slot back to the defaults on reset', () => {
		const { container } = mount();
		act(() => latest.onChange([agility]));

		act(() => latest.onReset());

		expect(sim.getFilters().statFilters?.buffs).toBeUndefined();
		expect(container.textContent).toBe(keys([strength, stamina]));
		expect(latest.customised).toBe(false);
	});

	it('keeps the same shown list through a write to another filter', () => {
		mount();
		const before = latest.shown;

		act(() => sim.setFilters(DatabaseFilters.create({ ...sim.getFilters(), minIlvl: 50 })));

		expect(sim.getFilters().minIlvl).toBe(50);
		expect(latest.shown).toBe(before);
	});

	it('keeps a write to another field that lands between renders', () => {
		mount();
		const { onChange } = latest;

		act(() => sim.setFilters(DatabaseFilters.create({ ...sim.getFilters(), minIlvl: 50 })));
		act(() => onChange([agility]));

		expect(sim.getFilters().minIlvl).toBe(50);
		expect(sim.getFilters().statFilters?.buffs).toBeDefined();
	});
});

describe('the stat filter popover open groups', () => {
	const groupsScope: StatFilterScope = { defaults: [strength, armor], universe: [strength, agility, hitRating, armor] };
	const GroupsProbe = ({ slot }: { slot: StatFilterSlot }) => {
		latest = useStatFilter(slot, groupsScope);
		return <StatFilterPopover {...latest} ariaLabel="Filter test stats" testId="test-filter" />;
	};
	const mountGroups = (slot: StatFilterSlot) =>
		renderWithHost(<GroupsProbe slot={slot} />, {
			sim,
			player: { getClass: () => Class.ClassWarrior },
			getStorageKey: (keyPart: string) => `test${keyPart}`,
		});
	const togglePopover = () =>
		act(async () => {
			fireEvent.click(screen.getByTestId('test-filter-button'));
		});
	const expandedGroups = () =>
		screen
			.queryAllByTestId(/^test-filter-group-[a-z]+-trigger$/)
			.filter(trigger => trigger.getAttribute('aria-expanded') === 'true')
			.map(trigger => trigger.getAttribute('data-testid')!.replace(/^test-filter-group-|-trigger$/g, ''));

	it('opens the groups that hold a shown stat while nothing is stored', async () => {
		mountGroups('ep');

		await togglePopover();

		expect(expandedGroups()).toEqual(['attributes', 'defense']);
	});

	it('opens the groups of a filter that loaded after the hook mounted', async () => {
		mountGroups('ep');
		act(() => sim.setFilters(withSelection(sim.getFilters(), 'ep', [hitRating])));

		await togglePopover();

		expect(expandedGroups()).toEqual(['gear']);
	});

	it('keeps a group open after its last stat is unticked until the popover closes', async () => {
		mountGroups('ep');
		await togglePopover();

		act(() => {
			fireEvent.click(screen.getByRole('checkbox', { name: armor.getShortName(Class.ClassWarrior) }));
		});
		expect(expandedGroups()).toEqual(['attributes', 'defense']);

		await togglePopover();
		await togglePopover();
		expect(expandedGroups()).toEqual(['attributes']);
	});

	it("stores a change under the panel's own key and reads it back on the next mount", async () => {
		const view = mountGroups('ep');
		await togglePopover();

		act(() => {
			fireEvent.click(screen.getByTestId('test-filter-group-defense-trigger'));
		});

		expect(expandedGroups()).toEqual(['attributes']);
		expect(JSON.parse(window.localStorage.getItem('test__epStatFilterGroups')!)).toEqual(['attributes']);
		expect(window.localStorage.getItem('test__consumablesStatFilterGroups')).toBeNull();
		view.unmount();
		mountGroups('ep');
		await togglePopover();
		expect(expandedGroups()).toEqual(['attributes']);
	});

	it('takes the stored groups over the default', async () => {
		window.localStorage.setItem('test__consumablesStatFilterGroups', JSON.stringify([]));
		mountGroups('consumables');

		await togglePopover();

		expect(expandedGroups()).toEqual([]);
	});
});

describe('useStatFilter on unrelated writes', () => {
	it('does not work out the shown stats again for a write to another slice', () => {
		const filter = vi.fn(Array.prototype.filter);
		const universe = Object.assign([strength, agility, stamina], { filter });
		const Watched = () => {
			latest = useStatFilter('buffs', { defaults: scope.defaults, universe: universe as unknown as UnitStat[] });
			return null;
		};
		renderWithHost(<Watched />, { sim, player: { getClass: () => Class.ClassWarrior }, getStorageKey: (keyPart: string) => `test${keyPart}` });
		const before = latest.shown;
		filter.mockClear();

		act(() => patchSlice(store, 'ui', {}));

		expect(filter).not.toHaveBeenCalled();
		expect(latest.shown).toBe(before);
	});
});
