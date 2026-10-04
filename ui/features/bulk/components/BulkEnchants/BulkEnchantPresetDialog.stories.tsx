import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { fn } from 'storybook/test';

import { BulkEnchantPresetDialogView } from './BulkEnchantPresetDialog';
import { BulkEnchantTileView } from './BulkEnchantTile';

const tiles = (enchants: Array<[string, string | undefined, boolean]>) =>
	enchants.map(([name, effect, checked]) => (
		<BulkEnchantTileView key={name} name={name} effect={effect} iconUrl={wowheadIconUrl('inv_misc_note_01')} checked={checked} onToggle={fn()} />
	));

const meta = {
	title: 'Features/Bulk/BulkEnchants/BulkEnchantPresetDialog',
	component: BulkEnchantPresetDialogView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		title: 'Edit enchant preset',
		name: 'Undead',
		onNameChange: fn(),
		nameTaken: false,
		canSave: true,
		onSave: fn(),
		groups: [
			{
				slot: ItemSlot.ItemSlotBack,
				label: 'Back',
				iconUrl: getEmptySlotIconUrl(ItemSlot.ItemSlotBack),
				checkedCount: 1,
				tiles: tiles([
					['Enchant Cloak - Greater Agility', 'Agility +3', true],
					['Enchant Cloak - Superior Defense', 'Armor +70', false],
					['Enchant Cloak - Greater Fire Resistance', '+15 Fire Resistance', false],
				]),
			},
			{
				slot: ItemSlot.ItemSlotMainHand,
				label: 'Main Hand',
				iconUrl: getEmptySlotIconUrl(ItemSlot.ItemSlotMainHand),
				checkedCount: 2,
				tiles: tiles([
					['Enchant Weapon - Crusader', undefined, true],
					['Enchant Weapon - Unholy Weapon', undefined, true],
					['Enchant Weapon - Agility', 'Agility +15', false],
				]),
			},
			{
				slot: ItemSlot.ItemSlotRanged,
				label: 'Ranged',
				iconUrl: getEmptySlotIconUrl(ItemSlot.ItemSlotRanged),
				checkedCount: 0,
				tiles: tiles([['Accurate Scope', 'Ranged Hit +3', false]]),
			},
		],
		openSlots: [String(ItemSlot.ItemSlotBack), String(ItemSlot.ItemSlotMainHand)],
		onOpenSlotsChange: fn(),
		search: '',
		onSearchChange: fn(),
		showAll: false,
		onShowAllChange: fn(),
	},
	argTypes: { groups: { control: false } },
} satisfies Meta<typeof BulkEnchantPresetDialogView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** A name another preset already has blocks the save button. */
export const NameTaken: Story = { args: { name: 'Beasts', nameTaken: true, canSave: false } };

export const NewPreset: Story = { args: { title: 'New enchant preset', name: '', canSave: false, openSlots: [] } };
