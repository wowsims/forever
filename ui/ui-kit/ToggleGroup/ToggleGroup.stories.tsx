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
		size: { control: 'select', options: [undefined, 'xs', 'sm', 'md'] },
		tone: { control: 'select', options: ['accent', 'neutral'] },
		orientation: { control: 'select', options: ['horizontal', 'vertical'] },
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

export const Sizes: Story = {
	render: args => (
		<div className="flex flex-col items-start gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<StatefulToggleGroup key={size} {...args} aria-label={`View (${size})`} size={size} />
			))}
		</div>
	),
};

export const Disabled: Story = { args: { disabled: true } };

export const Neutral: Story = { args: { tone: 'neutral', size: 'sm' } };

export const Vertical: Story = {
	args: {
		tone: 'neutral',
		size: 'sm',
		orientation: 'vertical',
		options: [
			{
				value: 'rotation',
				label: (
					<span className="grid grid-cols-1">
						<span>Rotation</span>
						<span className="text-xs opacity-75">Spell order and cooldowns</span>
					</span>
				),
			},
			{
				value: 'dps',
				label: (
					<span className="grid grid-cols-1">
						<span>DPS</span>
						<span className="text-xs opacity-75">Damage per second over the fight</span>
					</span>
				),
			},
			{
				value: 'threat',
				label: (
					<span className="grid grid-cols-1">
						<span>Threat</span>
						<span className="text-xs opacity-75">Threat per second over the fight</span>
					</span>
				),
			},
		],
	},
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
};
