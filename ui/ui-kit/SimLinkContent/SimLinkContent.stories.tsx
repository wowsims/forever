import type { Meta, StoryObj } from '@storybook/react-vite';

import { SimLinkContent } from './SimLinkContent';

const meta = {
	title: 'Kit/SimLinkContent',
	component: SimLinkContent,
	tags: ['autodocs'],
	args: {
		iconPath: 'https://wow.zamimg.com/images/wow/icons/large/classicon_mage.jpg',
		iconClassName: 'border-border',
		title: 'Mage',
	},
} satisfies Meta<typeof SimLinkContent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithLabel: Story = { args: { label: 'Mage', title: 'Fire' } };
export const WithStatus: Story = { args: { label: 'Mage', title: 'Fire', status: 'Phase 1 · Beta' } };
