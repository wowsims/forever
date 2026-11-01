import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, screen, userEvent, within } from 'storybook/test';

import { ComboBox, type ComboBoxProps } from './ComboBox';

const ITEMS = [
	'Ashkandi, Greatsword of the Brotherhood',
	'Bonereaver’s Edge',
	'Drake Talon Cleaver',
	'Obsidian Edged Blade',
	'Sulfuras, Hand of Ragnaros',
	'Thunderfury, Blessed Blade of the Windseeker',
	'Typhoon',
	'Arcanite Reaper',
];

type StoryProps = ComboBoxProps<string>;

const ComboBoxStory = ({ value: initialValue, onChange, open: initiallyOpen, onOpenChange, items, ...props }: StoryProps) => {
	const [value, setValue] = useState(initialValue);
	const [open, setOpen] = useState(initiallyOpen);
	const matches = items.filter(item => item.toLowerCase().includes(value.toLowerCase()));
	return (
		<ComboBox
			{...props}
			items={matches}
			value={value}
			onChange={next => {
				setValue(next);
				setOpen(next.length > 0);
				onChange(next);
			}}
			open={open}
			onOpenChange={(next, reason) => {
				setOpen(next);
				onOpenChange(next, reason);
			}}
		/>
	);
};

const meta = {
	title: 'Kit/ComboBox',
	component: ComboBox<string>,
	tags: ['autodocs'],
	args: {
		id: 'combo-box-story',
		label: 'Name',
		placeholder: 'Search',
		value: '',
		onChange: fn(),
		items: ITEMS,
		itemKey: item => item,
		renderItem: item => <span className="p-2">{item}</span>,
		onItemSelect: fn(),
		open: false,
		onOpenChange: fn(),
	},
	argTypes: {
		value: { control: false },
		items: { control: false },
		itemKey: { control: false },
		renderItem: { control: false },
		open: { control: false },
		popupAnchor: { control: false },
		footer: { control: false },
		multiColumn: { control: 'boolean' },
		clearable: { control: 'boolean' },
		grow: { control: 'boolean' },
		className: { control: 'text' },
	},
	render: args => <ComboBoxStory {...args} />,
} satisfies Meta<StoryProps>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Sizes: Story = {
	render: args => (
		<div className="flex flex-col gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<ComboBoxStory key={size} {...args} id={`combo-box-${size}`} label={`Name (${size})`} size={size} />
			))}
		</div>
	),
};
export const MultiColumn: Story = { args: { value: 'a', open: true, multiColumn: true } };
export const Clearable: Story = {
	args: { value: 'Typhoon', clearable: true, clearLabel: 'Clear search' },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'Clear search' }));
		await expect(args.onChange).toHaveBeenLastCalledWith('');
		await expect(canvas.getByRole('combobox')).toHaveValue('');
	},
};
export const Opens: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.type(within(canvasElement).getByRole('combobox'), 'edge');
		const options = await screen.findAllByRole('option');
		await expect(options.map(option => option.textContent)).toEqual(['Bonereaver’s Edge', 'Obsidian Edged Blade']);
		await userEvent.click(options[1]);
		await expect(args.onItemSelect).toHaveBeenCalledWith('Obsidian Edged Blade');
	},
};
