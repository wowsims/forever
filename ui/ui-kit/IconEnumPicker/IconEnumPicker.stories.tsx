import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { IconEnumPicker, type IconEnumPickerProps } from './IconEnumPicker';
import { type IconEnumOption, IconEnumPickerDirection } from './types';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, itemId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/item=${itemId}`,
});

const FLASKS: Array<IconEnumOption<number>> = [
	{ value: 0, tooltip: 'No Flask' },
	{ value: 1, icon: icon('inv_potion_62', 'Flask of the Titans', 13510), tooltip: 'Flask of the Titans' },
	{ value: 2, icon: icon('inv_potion_41', 'Flask of Supreme Power', 13512), tooltip: 'Flask of Supreme Power' },
	{ value: 3, icon: icon('inv_potion_97', 'Flask of Distilled Wisdom', 13511), tooltip: 'Flask of Distilled Wisdom' },
	{ value: 4, icon: icon('inv_potion_48', 'Flask of Chromatic Resistance', 13513), tooltip: 'Flask of Chromatic Resistance' },
];

const CURSES: Array<IconEnumOption<number>> = [
	{ value: 0, color: '#1a1a1a', tooltip: 'None' },
	{ value: 1, color: '#6b2fa0', tooltip: 'Curse of Shadow' },
	{ value: 2, color: '#2f6ba0', tooltip: 'Curse of the Elements' },
	{ value: 3, iconUrl: `${ICONS}/spell_shadow_curseofachimonde.jpg`, tooltip: 'Curse of Doom' },
];

const IconEnumPickerStory = ({ value: initialValue, onChange, ...props }: IconEnumPickerProps<number>) => {
	const [value, setValue] = useState(initialValue);
	return (
		<IconEnumPicker
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
	title: 'Kit/IconEnumPicker',
	component: IconEnumPicker<number>,
	tags: ['autodocs'],
	args: {
		options: FLASKS,
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		value: 1,
		onChange: fn(),
		tooltip: 'Flask',
	},
	argTypes: {
		direction: { control: 'select', options: [IconEnumPickerDirection.Vertical, IconEnumPickerDirection.Horizontal] },
		numColumns: { control: 'number' },
		disabled: { control: 'boolean' },
		options: { control: false },
		value: { control: false },
		equals: { control: false },
		backupIcon: { control: false },
		className: { control: 'text' },
	},
	render: args => <IconEnumPickerStory {...args} />,
} satisfies Meta<typeof IconEnumPicker<number>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Vertical: Story = {};
export const Horizontal: Story = { args: { direction: IconEnumPickerDirection.Horizontal } };
export const Columns: Story = { args: { numColumns: 2 } };
export const WithColorOptions: Story = { args: { options: CURSES, value: 1, tooltip: 'Curse' } };
export const WithHiddenOption: Story = { args: { options: FLASKS.map(option => (option.value === 2 ? { ...option, hidden: true } : option)) } };
export const Labelled: Story = { args: { label: 'Flask' } };
