import { Target as TargetProto } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStoryEncounter } from '@testing/storybook/fixtures/encounter';
import { expect, userEvent, within } from 'storybook/test';

import { TargetsPicker } from './TargetsPicker';

const meta = {
	title: 'Features/Encounter/TargetsPicker/TargetsPicker',
	component: TargetsPicker,
	tags: ['autodocs'],
	args: { encounter: makeStoryEncounter([TargetProto.create({ level: 63, swingSpeed: 2 })]) },
	argTypes: { encounter: { control: false } },
} satisfies Meta<typeof TargetsPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

/** One target: its menu offers copy only, because the list keeps at least one target. */
export const Default: Story = {};

/** Without threat metrics the third column stays empty. */
export const DamageOnly: Story = {
	args: { encounter: makeStoryEncounter([TargetProto.create({ level: 63 })], { threat: false }) },
};

/** The add button appends a default target. */
export const AddTarget: Story = {
	args: { encounter: makeStoryEncounter([TargetProto.create({ level: 63, swingSpeed: 2 })]) },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const before = canvas.getAllByTestId('targets-picker-item').length;
		await userEvent.click(canvas.getByTestId('tree-add'));
		await expect(canvas.getAllByTestId('targets-picker-item')).toHaveLength(before + 1);
	},
};
