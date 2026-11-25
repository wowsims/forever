import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ChoiceField } from './ChoiceField';
import type { ChoiceEntry, ChoiceFieldProps, ChoiceOption } from './types';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/medium';

const MOB_TYPES: Array<ChoiceOption<number>> = [
	{ value: 0, label: 'Unknown' },
	{ value: 1, label: 'Beast' },
	{ value: 2, label: 'Demon', tooltip: 'Takes extra damage from demon slaying effects.' },
	{ value: 8, label: 'Undead' },
];

const UNITS: Array<ChoiceOption<number>> = [
	{ value: 0, label: 'Self', icon: 'user' },
	{ value: 1, label: 'Current Target', icon: 'bullseye' },
	{ value: 2, label: 'Target 1', icon: { iconUrl: `${ICONS}/spell_shadow_metamorphosis.jpg`, name: 'Target 1' }, className: 'text-class-warrior' },
	{ value: 3, label: 'Wolf', icon: { iconUrl: `${ICONS}/ability_hunter_pet_wolf.jpg`, name: 'Wolf' } },
];

const VALUE_KINDS: Array<ChoiceEntry<number>> = [
	{ value: 0, label: 'None' },
	{
		label: 'Logic »',
		submenu: [
			{ value: 1, label: 'All of', tooltip: '<p>True when every sub-value is true.</p>' },
			{ value: 2, label: 'Any of', tooltip: '<p>True when any sub-value is true.</p>' },
		],
	},
	{ label: 'Resources »', submenu: [{ label: 'Mana »', submenu: [{ value: 3, label: 'Current Mana' }] }] },
];

const ENCOUNTERS: Array<ChoiceEntry<number>> = [
	{ value: -1, label: 'Custom' },
	{
		label: 'Presets',
		options: [
			{ value: 0, label: 'Raid Target' },
			{ value: 1, label: 'Dynamic Adds' },
		],
	},
	{
		label: 'My encounters',
		options: [
			{ value: 10, label: 'Ragnaros practice' },
			{ value: 11, label: 'Two adds' },
		],
	},
];

const ChoiceFieldStory = ({ value: initialValue, onChange, ...props }: ChoiceFieldProps<number>) => {
	const [value, setValue] = useState(initialValue);
	return (
		<ChoiceField<number>
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
	title: 'Kit/ChoiceField',
	component: ChoiceField<number>,
	tags: ['autodocs'],
	args: {
		id: 'choice-field-story',
		label: 'Mob Type',
		options: MOB_TYPES,
		value: 2,
		onChange: fn(),
	},
	argTypes: {
		variant: { control: 'inline-radio', options: ['select', 'link'] },
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		side: { control: 'select', options: [undefined, 'top', 'bottom', 'left', 'right'] },
		disabled: { control: 'boolean' },
		placeholder: { control: 'text' },
		className: { control: 'text' },
		options: { control: false },
		optionTree: { control: false },
		value: { control: false },
		equals: { control: false },
		renderValue: { control: false },
	},
	render: args => <ChoiceFieldStory {...args} />,
} satisfies Meta<typeof ChoiceField<number>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('combobox'));
		await userEvent.click(await within(canvasElement.ownerDocument.body).findByRole('option', { name: 'Undead' }));
		await expect(args.onChange).toHaveBeenLastCalledWith(8);
		await expect(within(canvasElement).getByRole('combobox').textContent).toBe('Undead');
	},
};
export const Sizes: Story = {
	render: args => (
		<div className="flex flex-col items-start gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<ChoiceFieldStory key={size} {...args} id={`choice-field-${size}`} label={`Mob Type (${size})`} size={size} />
			))}
		</div>
	),
};
export const Invalid: Story = { args: { invalid: true } };
export const DescribedByAMessage: Story = {
	args: { invalid: true, 'aria-describedby': 'choice-field-story-message' },
	render: args => (
		<div className="grid gap-2">
			<ChoiceFieldStory {...args} />
			<p id="choice-field-story-message" className="mb-0 text-xs text-danger">
				Already in the list
			</p>
		</div>
	),
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('combobox').getAttribute('aria-describedby')).toBe('choice-field-story-message');
	},
};
export const Link: Story = { args: { variant: 'link', label: undefined, placeholder: 'Mob type' } };
export const WithIcons: Story = { args: { variant: 'link', label: undefined, options: UNITS, value: 2, placeholder: 'Unit' } };
export const WithSubmenu: Story = { args: { variant: 'link', label: undefined, options: VALUE_KINDS, value: 3, placeholder: 'No condition' } };
export const WithSections: Story = {
	args: { label: 'Encounter', options: ENCOUNTERS, value: 0 },
	play: async ({ args, canvasElement }) => {
		const body = within(canvasElement.ownerDocument.body);
		await userEvent.click(within(canvasElement).getByRole('combobox'));
		const [, mine] = await body.findAllByRole('group');
		await expect(within(mine).getByText('My encounters')).toBeTruthy();
		await userEvent.click(within(mine).getByRole('option', { name: 'Two adds' }));
		await expect(args.onChange).toHaveBeenLastCalledWith(11);
	},
};
export const InAStartAlignedColumn: Story = {
	decorators: [
		Story => (
			<div className="flex flex-col items-start">
				<Story />
			</div>
		),
	],
};
