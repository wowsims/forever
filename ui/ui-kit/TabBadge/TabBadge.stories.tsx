import type { Meta, StoryObj } from '@storybook/react-vite';

import { TabBadge } from './TabBadge';

const meta = {
	title: 'Kit/TabBadge',
	component: TabBadge,
	tags: ['autodocs'],
	args: { label: 'Beta' },
	render: args => (
		<span className="text-sm font-bold text-white">
			Bulk
			<TabBadge {...args} />
		</span>
	),
} satisfies Meta<typeof TabBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
