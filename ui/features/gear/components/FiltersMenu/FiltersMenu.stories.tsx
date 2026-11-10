import { ArmorType, Faction, ItemSlot, RangedWeaponType, WeaponType } from '@generated/proto/common';
import { DatabaseFilters, SourceFilterOption } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { FiltersMenuView } from './FiltersMenu';

const meta = {
	title: 'Features/Gear/FiltersMenu/FiltersMenu',
	component: FiltersMenuView,
	tags: ['autodocs'],
	args: {
		slot: ItemSlot.ItemSlotHead,
		open: true,
		onOpenChange: fn(),
		filters: DatabaseFilters.create({ sources: [SourceFilterOption.SourceCrafting], armorTypes: [ArmorType.ArmorTypePlate], maxIlvl: 80 }),
		onFiltersChange: fn(),
		armorTypes: [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate],
		weaponTypes: [WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword, WeaponType.WeaponTypeMace],
		rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun],
		canDualWield: true,
		raceFaction: Faction.Horde,
	},
	argTypes: { filters: { control: false } },
} satisfies Meta<typeof FiltersMenuView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ArmorSlot: Story = {};

export const WeaponSlot: Story = { args: { slot: ItemSlot.ItemSlotMainHand } };

export const RangedSlot: Story = { args: { slot: ItemSlot.ItemSlotRanged } };

export const Trinket: Story = { args: { slot: ItemSlot.ItemSlotTrinket1 } };
