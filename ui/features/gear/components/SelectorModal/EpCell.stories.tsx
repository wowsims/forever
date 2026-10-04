import type { Meta, StoryObj } from '@storybook/react-vite';

import { EpCell } from './EpCell';

const meta = {
	title: 'Features/Gear/SelectorModal/Components/EpCell',
	component: EpCell,
	tags: ['autodocs'],
	args: { ep: 42.4, equippedEP: 38 },
} satisfies Meta<typeof EpCell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Upgrade: Story = {};

export const Downgrade: Story = { args: { ep: 31, equippedEP: 38 } };

export const Equipped: Story = { args: { ep: 38, equippedEP: 38 } };
