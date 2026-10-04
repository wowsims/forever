import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { fn } from 'storybook/test';

import { BulkEnchantTileView } from './BulkEnchantTile';

const meta = {
	title: 'Features/Bulk/BulkEnchants/BulkEnchantTile',
	component: BulkEnchantTileView,
	tags: ['autodocs'],
	args: {
		name: 'Enchant Cloak - Greater Agility',
		effect: 'Agility +3',
		iconUrl: wowheadIconUrl('inv_misc_note_01'),
		href: 'https://www.wowhead.com/classic/spell=20015',
		checked: false,
		onToggle: fn(),
	},
	decorators: [
		Story => (
			<div className="grid max-w-60">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkEnchantTileView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Unchecked: Story = {};

export const Checked: Story = { args: { checked: true } };

/** An enchant with no effect text of its own shows its name alone. */
export const WithoutEffect: Story = { args: { effect: undefined } };

/** A long name and effect truncate on their own line rather than wrap. */
export const LongText: Story = {
	args: { name: 'Enchant Chest - Major Stats and More', effect: 'Healing and Spell Damage +18 / Spell Hit +1%' },
};
