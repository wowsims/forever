import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ChartViewPicker, type ChartViewPickerProps } from './ChartViewPicker';

const StatefulPicker = ({ value: initial, onChange, ...props }: ChartViewPickerProps) => {
	const [value, setValue] = useState(initial);
	return (
		<ChartViewPicker
			{...props}
			value={value}
			onChange={next => {
				setValue(next);
				onChange(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Results/Timeline/Components/ChartViewPicker',
	component: ChartViewPicker,
	tags: ['autodocs'],
	args: { value: 'rotation', onChange: fn() },
	render: args => <StatefulPicker {...args} />,
} satisfies Meta<typeof ChartViewPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rotation: Story = {
	play: async ({ args, canvasElement }) => {
		const [, dps] = within(canvasElement).getAllByRole('button');
		await userEvent.click(dps);
		await expect(args.onChange).toHaveBeenCalledWith('dps');
		await expect(dps).toHaveAttribute('aria-pressed', 'true');
	},
};

export const Dps: Story = { args: { value: 'dps' } };
