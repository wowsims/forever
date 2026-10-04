import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { RotationFloatingActionBarChip } from './RotationFloatingActionBarChip';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationFloatingActionBarChip',
	component: RotationFloatingActionBarChip,
	tags: ['autodocs'],
	args: { label: 'Mortal Strike', shown: true, focused: true, onToggle: fn(), onFocus: fn() },
} satisfies Meta<typeof RotationFloatingActionBarChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Shown: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('switch'));
		await expect(args.onToggle).toHaveBeenCalledOnce();
	},
};

export const Hidden: Story = { args: { shown: false } };

export const OutOfTheTabOrder: Story = { args: { focused: false } };
