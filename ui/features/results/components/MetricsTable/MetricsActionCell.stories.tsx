import type { Meta, StoryObj } from '@storybook/react-vite';
import { FLURRY, MORTAL_STRIKE } from '@testing/storybook/fixtures/results-views';
import { expect, fn, userEvent, within } from 'storybook/test';

import { MetricsActionCell } from './MetricsActionCell';

const meta = {
	title: 'Features/Results/MetricsTable/Components/MetricsActionCell',
	component: MetricsActionCell,
	tags: ['autodocs'],
	args: { name: MORTAL_STRIKE.name, actionId: MORTAL_STRIKE, expandable: false, expanded: false, onToggle: fn() },
	argTypes: { actionId: { control: false } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof MetricsActionCell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Plain: Story = {};

export const Collapsed: Story = {
	args: { expandable: true },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('expand-toggle'));
		await expect(args.onToggle).toHaveBeenCalledOnce();
	},
};

export const Expanded: Story = { args: { expandable: true, expanded: true } };

export const Aura: Story = {
	args: { name: FLURRY.name, actionId: FLURRY, useBuffAura: true },
	parameters: { smoke: false, smokeReason: 'An aura tooltip asks the spell database whether the spell has a buff' },
};
