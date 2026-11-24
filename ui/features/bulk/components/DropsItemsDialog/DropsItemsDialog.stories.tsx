import { ArmorType, Faction, MobType, RangedWeaponType, Stat, WeaponType } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { PlayerSpecs } from '@sim/player/specs';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsItemListFixture, dropsNamesFixture, dropsTreeFixture } from '@testing/storybook/fixtures/bulk-drops';
import { fn } from 'storybook/test';

import { dropsDefaultFilters } from '../../model/drops/filters';
import { DropsItemsDialogView } from './DropsItemsDialog';

const filters = dropsDefaultFilters(PlayerSpecs.getPlayerClass(PlayerSpecs.DpsWarrior));

const meta = {
	title: 'Features/Bulk/DropsItemsDialog',
	component: DropsItemsDialogView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		list: dropsItemListFixture(),
		grouping: 'slot',
		onGroupingChange: fn(),
		tree: dropsTreeFixture(),
		names: dropsNamesFixture(),
		filters,
		activeFilters: 0,
		filtersMenu: {
			onFiltersChange: fn(),
			armorTypes: [ArmorType.ArmorTypeCloth, ArmorType.ArmorTypeLeather, ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate],
			weaponTypes: [WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword, WeaponType.WeaponTypeMace],
			rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun],
			areaTypes: [],
			mobTypes: [MobType.MobTypeUndead],
			creatureFamilies: [],
			itemStats: [Stat.StatStrength, Stat.StatAgility, Stat.StatStamina],
			canDualWield: true,
			raceFaction: Faction.Alliance,
		},
		onIncludedChange: fn(),
	},
	argTypes: { list: { control: false }, tree: { control: false }, names: { control: false }, filters: { control: false }, filtersMenu: { control: false } },
} satisfies Meta<typeof DropsItemsDialogView>;

export default meta;
type Story = StoryObj<typeof meta>;

// All picked on the rail: every listed item once, the worn one greyed out and the unticked one unchecked.
export const BySlot: Story = {};

// The rail lists the bosses; an item under two of them still shows once under All.
export const BySource: Story = { args: { grouping: 'source' } };

// Plate left unticked hides the gauntlets; the rail still counts them.
export const Filtered: Story = {
	args: { filters: DatabaseFilters.create({ ...filters, armorTypes: [ArmorType.ArmorTypeMail] }), activeFilters: 1 },
};

// A phone's width: the rail is a row of chips that scrolls sideways, the list fills the dialog, and each row's sources
// sit under its name.
export const Phone: Story = { globals: { viewport: { value: 'mobile1', isRotated: false } } };
