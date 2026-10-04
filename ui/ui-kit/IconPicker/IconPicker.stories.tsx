import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { IconPicker, type IconPickerProps } from './IconPicker';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, spellId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/spell=${spellId}`,
});

const KINGS = icon('spell_magic_magearmor', 'Blessing of Kings', 20217);
const MARK = icon('spell_nature_regeneration', 'Mark of the Wild', 9885);
const IMPROVED_MARK = icon('spell_nature_regeneration', 'Improved Mark of the Wild', 17055);
const BLOODLUST = icon('spell_nature_bloodlust', 'Bloodlust', 2825);
const IMPROVED_BLOODLUST = icon('spell_nature_bloodlust', 'Improved Bloodlust', 2826);
const GREATER_BLOODLUST = icon('spell_nature_bloodlust', 'Greater Bloodlust', 2827);
const TOTEM = icon('spell_nature_windfury', 'Windfury Totem', 10614);

const IconPickerStory = ({ value: initialValue, onChange, ...props }: IconPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<IconPicker
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
	title: 'Kit/IconPicker',
	component: IconPicker,
	tags: ['autodocs'],
	args: {
		states: 2,
		icon: KINGS,
		value: 0,
		onChange: fn(),
	},
	argTypes: {
		states: { control: 'number' },
		value: { control: false },
		disabled: { control: 'boolean' },
		icon: { control: false },
		improvedIcon: { control: false },
		improvedIcon2: { control: false },
		className: { control: 'text' },
	},
	render: args => <IconPickerStory {...args} />,
} satisfies Meta<typeof IconPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Boolean: Story = {};
export const Tristate: Story = { args: { states: 3, icon: MARK, improvedIcon: IMPROVED_MARK, value: 2 } };
export const Quadstate: Story = {
	args: { states: 4, icon: BLOODLUST, improvedIcon: IMPROVED_BLOODLUST, improvedIcon2: GREATER_BLOODLUST, value: 3 },
};
export const Counter: Story = { args: { states: 0, icon: TOTEM, value: 4 } };
export const Disabled: Story = { args: { disabled: true, value: 1 } };
export const Labelled: Story = { args: { label: 'Blessing of Kings', value: 1 } };
