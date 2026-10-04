import type { Meta, StoryObj } from '@storybook/react-vite';

import { ItemNoticeIconView } from './ItemNoticeIcon';

const meta = {
	title: 'Features/Gear/GearPicker/Components/ItemNoticeIcon',
	component: ItemNoticeIconView,
	tags: ['autodocs'],
	args: { level: 'warning', children: <p>The following enchant effect (on-use or proc) is not implemented!</p> },
	argTypes: { children: { control: false } },
} satisfies Meta<typeof ItemNoticeIconView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Warning: Story = {};

export const Info: Story = { args: { level: 'info', children: <p>Only drops in Forest and Grassland areas.</p> } };
