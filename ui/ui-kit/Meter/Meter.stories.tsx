import type { Meta, StoryObj } from '@storybook/react-vite';

import { Meter } from './Meter';

const meta = {
	title: 'Kit/Meter',
	component: Meter,
	tags: ['autodocs'],
	args: { value: 3, max: 5 },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof Meter>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ExtraSmall: Story = { args: { size: 'xs' } };

export const GrayTrack: Story = { args: { track: 'gray' } };
