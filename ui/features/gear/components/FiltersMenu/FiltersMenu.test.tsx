import { ArmorType, Faction, ItemSlot, RangedWeaponType, WeaponType } from '@generated/proto/common';
import { DatabaseFilters, FactionFilter, SourceFilterOption } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { storeSource as store } from '@testing/gear';
import { renderWithHost } from '@testing/react';
import { fakeHost } from '@testing/sim/fake_host';
import { act, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { FiltersMenuViewProps } from './FiltersMenu';

vi.mock('@sim/state/subscriptions', async () =>
	(await import('@testing/sim/subscriptions')).mockSubscriptions((await import('@testing/gear')).storeSource.subscribe),
);

const { FiltersMenu, FiltersMenuView } = await import('./FiltersMenu');

interface ClassStub {
	armorTypes?: ArmorType[];
	weaponTypes?: WeaponType[];
	rangedWeaponTypes?: RangedWeaponType[];
	canDualWield?: boolean;
}

const factionLabel = (label: string) => `gear_tab.gear_picker.filters.faction_labels.${label}`;
const factionOptions = () => screen.getAllByRole('option');

describe('FiltersMenu', () => {
	let filters: DatabaseFilters;
	let setFilters: ReturnType<typeof vi.fn>;
	let raceFaction: Faction;
	let rootElem: HTMLElement;

	const setup = (
		{ armorTypes = [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate], weaponTypes = [], rangedWeaponTypes = [], canDualWield = false }: ClassStub = {},
		{ slot = ItemSlot.ItemSlotHead, open = true, onOpenChange = vi.fn() } = {},
	) => {
		// A fresh root per mount: two mounts in one test compare two class shapes, and React must be
		// left to unmount its own portal rather than have the container emptied under it.
		rootElem = document.createElement('div');
		rootElem.className = 'sim-ui';
		document.body.appendChild(rootElem);
		const host = fakeHost({
			rootElem,
			player: {
				sim: {
					getFilters: () => DatabaseFilters.clone(filters),
					setFilters,
				},
				getPlayerClass: () => ({ armorTypes, weaponTypes: weaponTypes.map(weaponType => ({ weaponType })), rangedWeaponTypes }),
				getPlayerSpec: () => ({ canDualWield }),
				getFaction: () => raceFaction,
			},
		});

		render(
			<PortalContainerContext value={rootElem}>
				<SimHostProvider host={host}>
					<FiltersMenu slot={slot} open={open} onOpenChange={onOpenChange} />
				</SimHostProvider>
			</PortalContainerContext>,
		);
		return { onOpenChange };
	};

	const sections = () => Array.from(rootElem.querySelectorAll('[data-testid="menu-section-title"]')).map(node => node.textContent);
	const pickerIds = () => Array.from(rootElem.querySelectorAll<HTMLElement>('[data-testid="filters-menu"] [id^=filter]')).map(node => node.id);

	beforeEach(() => {
		filters = DatabaseFilters.create({});
		raceFaction = Faction.Horde;
		setFilters = vi.fn((next: DatabaseFilters) => {
			filters = next;
		});
	});

	it('hands the class weapon types to the weapon type section of a weapon slot', () => {
		setup({ weaponTypes: [WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword] }, { slot: ItemSlot.ItemSlotMainHand });

		expect(pickerIds()).toContain(`filters-weapon-type-${WeaponType.WeaponTypeSword}`);
	});

	it('hands the class armor and ranged weapon types to their filter sections', () => {
		setup({ armorTypes: [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate] }, { slot: ItemSlot.ItemSlotHead });
		expect(pickerIds()).toContain(`filters-armor-type-${ArmorType.ArmorTypePlate}`);

		setup({ rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun] }, { slot: ItemSlot.ItemSlotRanged });
		expect(sections()).toContain('gear_tab.gear_picker.ranged_weapon_type');
	});

	it('adds the off hand speed and ranged sections only for a spec that can dual wield and a class with ranged weapons', () => {
		setup({ weaponTypes: [WeaponType.WeaponTypeAxe] }, { slot: ItemSlot.ItemSlotMainHand });
		expect(pickerIds()).not.toContain('filters-min-oh-weapon-speed');
		expect(sections()).not.toContain('gear_tab.gear_picker.ranged_weapon_type');

		setup({ weaponTypes: [WeaponType.WeaponTypeAxe], canDualWield: true }, { slot: ItemSlot.ItemSlotMainHand });
		expect(pickerIds()).toEqual(expect.arrayContaining(['filters-min-oh-weapon-speed', 'filters-max-oh-weapon-speed']));
	});

	it('portals into the sim root for the spec theme, writes a source toggle and a faction pick back through the sim, and closes on its own close button', async () => {
		const source = SourceFilterOption.SourceCrafting;
		const { onOpenChange } = setup();

		expect(rootElem.querySelector('[data-testid="filters-menu"]')).not.toBeNull();
		expect(document.body.querySelector(':scope > [data-testid="filters-menu"]')).toBeNull();

		const checkbox = rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!;
		expect(checkbox.checked).toBe(false);

		fireEvent.click(checkbox);
		expect(setFilters).toHaveBeenCalledTimes(1);
		expect(setFilters.mock.calls[0][0].sources).toContain(source);

		act(() => store.notify());
		expect(rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!.checked).toBe(true);

		fireEvent.click(rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!);
		expect(setFilters.mock.calls[1][0].sources).not.toContain(source);

		await userEvent.click(rootElem.querySelector<HTMLElement>('#filters-faction-restriction')!);
		expect(factionOptions().map(option => option.textContent)).toEqual([factionLabel('horde_only'), factionLabel('none')]);

		await userEvent.click(factionOptions()[1]);
		expect(setFilters.mock.lastCall![0].factionFilter).toBe(FactionFilter.FactionFilterAny);

		fireEvent.click(rootElem.querySelector<HTMLButtonElement>('[data-testid="filters-menu"] [data-testid="sim-dialog-close"]')!);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it("names the race's faction by default, follows a race change, and goes back to it from any", async () => {
		setup();
		const trigger = () => rootElem.querySelector<HTMLElement>('#filters-faction-restriction')!;
		expect(trigger().textContent).toContain(factionLabel('horde_only'));

		raceFaction = Faction.Alliance;
		act(() => store.notify());
		expect(trigger().textContent).toContain(factionLabel('alliance_only'));

		await userEvent.click(trigger());
		await userEvent.click(factionOptions()[1]);
		expect(setFilters.mock.lastCall![0].factionFilter).toBe(FactionFilter.FactionFilterAny);

		act(() => store.notify());
		await userEvent.click(trigger());
		await userEvent.click(factionOptions()[0]);
		expect(setFilters.mock.lastCall![0].factionFilter).toBe(FactionFilter.FactionFilterRace);
	});
});

const renderView = (props: Partial<FiltersMenuViewProps> = {}) => {
	const onFiltersChange = vi.fn<FiltersMenuViewProps['onFiltersChange']>();
	renderWithHost(
		<FiltersMenuView
			slot={ItemSlot.ItemSlotHead}
			open
			onOpenChange={vi.fn()}
			filters={DatabaseFilters.create({})}
			onFiltersChange={onFiltersChange}
			armorTypes={[ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate]}
			weaponTypes={[WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword]}
			rangedWeaponTypes={[RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun]}
			canDualWield={false}
			raceFaction={Faction.Horde}
			{...props}
		/>,
	);
	// Applies the last edit to `start`, the way the container applies it to the stored filters.
	const lastEdit = (start = DatabaseFilters.create({})) => {
		const next = DatabaseFilters.clone(start);
		onFiltersChange.mock.lastCall![0](next);
		return next;
	};
	return { onFiltersChange, lastEdit };
};

const viewSections = () => screen.getAllByTestId('menu-section-title').map(node => node.textContent);
const input = (id: string) => document.getElementById(id) as HTMLInputElement;

describe('FiltersMenuView', () => {
	it('gives an armor slot the armor type section after the general, source and raid ones', () => {
		renderView();
		expect(viewSections()).toEqual([
			'gear_tab.gear_picker.filters.general',
			'gear_tab.gear_picker.filters.source',
			'gear_tab.gear_picker.filters.raids',
			'gear_tab.gear_picker.armor_type',
		]);
	});

	it('gives a weapon slot the weapon sections, with off hand speeds for a dual wielder', () => {
		renderView({ slot: ItemSlot.ItemSlotMainHand, canDualWield: true });
		expect(viewSections().slice(3)).toEqual(['gear_tab.gear_picker.weapon_type', 'gear_tab.gear_picker.weapon_speed']);
		expect(input('filters-min-oh-weapon-speed')).not.toBeNull();
	});

	it('gives the ranged slot the ranged sections', () => {
		renderView({ slot: ItemSlot.ItemSlotRanged });
		expect(viewSections().slice(3)).toEqual(['gear_tab.gear_picker.ranged_weapon_type', 'gear_tab.gear_picker.ranged_weapon_speed']);
	});

	// One type is no choice at all, so the section is not worth the space.
	it.each([
		['armor', { armorTypes: [ArmorType.ArmorTypePlate] }],
		['ranged', { slot: ItemSlot.ItemSlotRanged, rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeGun] }],
	] as const)('leaves the %s type section out for a class with a single type', (_kind, props) => {
		renderView(props);
		expect(viewSections()).toEqual(['gear_tab.gear_picker.filters.general', 'gear_tab.gear_picker.filters.source', 'gear_tab.gear_picker.filters.raids']);
	});

	it('shows the filters it is handed', async () => {
		renderView({
			filters: DatabaseFilters.create({
				sources: [SourceFilterOption.SourceCrafting],
				minIlvl: 0,
				maxIlvl: 528,
				factionFilter: FactionFilter.FactionFilterAny,
			}),
		});
		expect(input(`filters-source-${SourceFilterOption.SourceCrafting}`).checked).toBe(true);
		expect(input('filters-min-ilvl').value).toBe('');
		expect(input('filters-max-ilvl').value).toBe('528');
		await userEvent.click(document.getElementById('filters-faction-restriction')!);
		expect(factionOptions().map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'true']);
	});

	it('adds a ticked list entry and removes an unticked one', () => {
		const source = SourceFilterOption.SourceCrafting;
		const { lastEdit } = renderView({ filters: DatabaseFilters.create({ armorTypes: [ArmorType.ArmorTypePlate] }) });

		fireEvent.click(input(`filters-source-${source}`));
		expect(lastEdit().sources).toEqual([source]);

		fireEvent.click(input(`filters-armor-type-${ArmorType.ArmorTypePlate}`));
		expect(lastEdit(DatabaseFilters.create({ armorTypes: [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate] })).armorTypes).toEqual([
			ArmorType.ArmorTypeMail,
		]);
	});

	it('writes a faction pick', async () => {
		const { lastEdit } = renderView({ filters: DatabaseFilters.create({ factionFilter: FactionFilter.FactionFilterAny }) });
		await userEvent.click(document.getElementById('filters-faction-restriction')!);
		await userEvent.click(factionOptions()[0]);
		expect(lastEdit(DatabaseFilters.create({ factionFilter: FactionFilter.FactionFilterAny })).factionFilter).toBe(FactionFilter.FactionFilterRace);
	});
});
