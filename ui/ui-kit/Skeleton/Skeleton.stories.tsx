import type { Meta, StoryObj } from '@storybook/react-vite';

import { Skeleton } from './Skeleton';

const meta = {
	title: 'Kit/Skeleton',
	component: Skeleton,
	tags: ['autodocs'],
	render: args => (
		<p className="text-sm">
			Spell Power <Skeleton {...args} />
		</p>
	),
} satisfies Meta<typeof Skeleton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Wide: Story = { args: { className: 'w-32' } };
