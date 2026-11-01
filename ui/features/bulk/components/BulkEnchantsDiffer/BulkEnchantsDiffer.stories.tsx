import type { Meta, StoryObj } from '@storybook/react-vite';

import { BulkEnchantsDiffer } from './BulkEnchantsDiffer';

const meta = {
	title: 'Features/Bulk/BulkEnchantsDiffer',
	component: BulkEnchantsDiffer,
	tags: ['autodocs'],
	args: { enchants: ['Back: Enchant Cloak - Greater Agility', 'Chest: Enchant Chest - Greater Stats'] },
} satisfies Meta<typeof BulkEnchantsDiffer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const One: Story = { args: { enchants: ['Main Hand: Enchant Weapon - Crusader'] } };
