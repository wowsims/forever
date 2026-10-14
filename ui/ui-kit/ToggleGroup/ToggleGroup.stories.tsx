import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ToggleGroup, type ToggleGroupProps } from './ToggleGroup';

type View = 'rotation' | 'dps' | 'threat';

const OPTIONS: ToggleGroupProps<View>['options'] = [
	{ value: 'rotation', label: 'Rotation' },
	{ value: 'dps', label: 'DPS' },
	{ value: 'threat', label: 'Threat' },
];

const StatefulToggleGroup = ({ value: initial, onChange, ...props }: ToggleGroupProps<View>) => {
	const [value, setValue] = useState(initial);
	return (
		<ToggleGroup
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
	title: 'Kit/ToggleGroup',
	component: ToggleGroup<View>,
	tags: ['autodocs'],
	args: { options: OPTIONS, value: 'rotation', onChange: fn(), 'aria-label': 'View' },
	argTypes: {
		options: { control: false },
		value: { control: 'select', options: OPTIONS.map(option => option.value) },
		size: { control: 'select', options: [undefined, 'sm', 'xs'] },
	},
	render: args => <StatefulToggleGroup {...args} />,
} satisfies Meta<typeof ToggleGroup<View>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		const dps = within(canvasElement).getByRole('button', { name: 'DPS' });
		await userEvent.click(dps);
		await expect(args.onChange).toHaveBeenCalledWith('dps');
		await expect(dps).toHaveAttribute('aria-pressed', 'true');
	},
};

export const Small: Story = { args: { size: 'sm' } };

export const Disabled: Story = { args: { disabled: true } };
