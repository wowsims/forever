import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { MultiComboBox } from './MultiComboBox';
import type { MultiComboBoxProps } from './types';

const AREAS = [
	{ name: 'Forest / Grassland', value: 1 },
	{ name: 'Mountainous', value: 2 },
	{ name: 'Snowy', value: 3 },
	{ name: 'Desert', value: 4 },
	{ name: 'Haunted', value: 7 },
];

const MORE_AREAS = [
	...AREAS,
	{ name: 'Swamp', value: 8 },
	{ name: 'Jungle', value: 9 },
	{ name: 'Volcanic', value: 10 },
	{ name: 'Underground', value: 11 },
	{ name: 'Coastal', value: 12 },
	{ name: 'Barrens', value: 13 },
];

const MultiComboBoxStory = ({ value: initialValue, onChange, ...props }: MultiComboBoxProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<MultiComboBox
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
	title: 'Kit/MultiComboBox',
	component: MultiComboBox,
	tags: ['autodocs'],
	args: {
		id: 'multi-combo-box-story',
		label: 'Area Types',
		placeholder: 'Pick area types',
		emptyText: 'No area type matches',
		removeLabel: 'Remove',
		openLabel: 'Show all area types',
		values: AREAS,
		value: [],
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		value: { control: false },
		values: { control: false },
		className: { control: 'text' },
	},
	render: args => <MultiComboBoxStory {...args} />,
} satisfies Meta<typeof MultiComboBox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};
export const Sizes: Story = {
	render: args => (
		<div className="flex flex-col gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<MultiComboBoxStory key={size} {...args} id={`multi-combo-box-${size}`} label={`Area Types (${size})`} size={size} />
			))}
		</div>
	),
};
export const WithPicks: Story = {
	args: { value: [2, 7] },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getAllByRole('button', { name: 'Remove' })[0]);
		await expect(args.onChange).toHaveBeenCalledWith([7]);
		await expect(canvas.getAllByTestId('multi-combo-box-chip').map(chip => chip.textContent)).toEqual(['Haunted']);
	},
};
export const ManyPicks: Story = { args: { values: MORE_AREAS, value: MORE_AREAS.map(area => area.value) } };
export const FitOptions: Story = { args: { fitOptions: true, values: MORE_AREAS, value: [1, 2, 7, 11] } };
export const Disabled: Story = {
	args: { value: [2, 3, 7], disabled: true },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const chip of canvas.getAllByTestId('multi-combo-box-chip')) await expect(chip).toBeVisible();
		await expect(canvas.queryAllByRole('button', { name: 'Remove' })).toEqual([]);
	},
};
