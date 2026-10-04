import type { Meta, StoryObj } from '@storybook/react-vite';
import { withTooltipSurface } from '@testing/storybook/decorators';

import { TooltipRow } from './TooltipRow';

const meta = {
	title: 'Features/CharacterStats/Components/TooltipRow',
	component: TooltipRow,
	tags: ['autodocs'],
	args: { label: 'Gear', value: '+180' },
	decorators: [
		Story => (
			<div className="w-56">
				<Story />
			</div>
		),
		withTooltipSurface,
	],
} satisfies Meta<typeof TooltipRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Success: Story = { args: { label: 'Crush', value: '103.00%', className: 'text-success' } };
export const Danger: Story = { args: { label: 'Shear', value: '60.70%', className: 'text-danger' } };
