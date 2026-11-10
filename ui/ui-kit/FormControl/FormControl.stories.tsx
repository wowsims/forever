import type { Meta, StoryObj } from '@storybook/react-vite';
import { Fragment } from 'react';

import * as FormControl from './index';

const meta = {
	title: 'Kit/FormControl',
	component: FormControl.Input,
	tags: ['autodocs'],
	args: { id: 'form-control-story' },
	argTypes: {
		disabled: { control: 'boolean' },
		placeholder: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof FormControl.Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Input: Story = {
	args: { placeholder: 'Search items' },
	render: args => (
		<>
			<FormControl.FieldLabel htmlFor="form-control-story">Name</FormControl.FieldLabel>
			<FormControl.Input {...args} />
			<FormControl.HelpText>Matches item names and ids.</FormControl.HelpText>
		</>
	),
};
export const TextArea: Story = {
	render: ({ id, disabled }) => (
		<>
			<FormControl.FieldLabel htmlFor={id}>Import</FormControl.FieldLabel>
			<FormControl.TextArea id={id} disabled={disabled} rows={4} defaultValue='{"gear":{"items":[]}}' />
		</>
	),
};
export const FieldLabel: Story = {
	render: ({ id }) => (
		<>
			<FormControl.FieldLabel htmlFor={id}>A label for an input</FormControl.FieldLabel>
			<FormControl.FieldLabel as="span">A caption for a group</FormControl.FieldLabel>
		</>
	),
};
export const HelpText: Story = {
	render: () => <FormControl.HelpText>Shown under a field to say what it takes.</FormControl.HelpText>,
};
export const Disabled: Story = { args: { disabled: true, defaultValue: 'Thunderfury' } };
/** The text area keeps its content height and takes the size's padding. */
export const Sizes: Story = {
	render: () => (
		<div className="grid grid-cols-2 gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<Fragment key={size}>
					<FormControl.Input size={size} aria-label={`Input ${size}`} defaultValue={size} />
					<FormControl.TextArea size={size} aria-label={`Text area ${size}`} rows={2} defaultValue={size} />
				</Fragment>
			))}
		</div>
	),
};
