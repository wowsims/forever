import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { BooleanPicker, type BooleanPickerProps } from './BooleanPicker';

const BooleanPickerStory = ({ value: initialValue, onChange, ...props }: BooleanPickerProps) => {
	const [value, setValue] = useState(initialValue);
	return (
		<BooleanPicker
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
	title: 'Kit/BooleanPicker',
	component: BooleanPicker,
	tags: ['autodocs'],
	args: {
		id: 'boolean-picker-story',
		label: 'Enable Item Swap',
		value: false,
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		reverse: { control: 'boolean' },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <BooleanPickerStory {...args} />,
} satisfies Meta<typeof BooleanPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox'));
		await expect(args.onChange).toHaveBeenLastCalledWith(true);
	},
};
export const Reverse: Story = { args: { label: 'Dual Wield', reverse: true, layout: 'inline', value: true } };
export const Inline: Story = { args: { label: 'Show All Stats', layout: 'inline' } };
export const Disabled: Story = { args: { disabled: true, value: true } };
// Tab reaches the disabled box, whose focus shows the reason on the label tooltip, right of the label text; Space leaves
// it unticked.
export const DisabledFocusable: Story = {
	args: {
		label: 'Patchwerk copy',
		labelTooltip: 'Same as Patchwerk',
		labelTooltipPlace: 'right',
		layout: 'inline',
		disabled: true,
		focusableWhenDisabled: true,
	},
	play: async ({ args, canvasElement }) => {
		await userEvent.tab();
		await expect(document.activeElement).toBe(within(canvasElement).getByRole('checkbox'));
		const tooltip = await within(document.body).findByRole('tooltip');
		await expect(tooltip.textContent).toBe('Same as Patchwerk');
		const label = within(canvasElement).getByText('Patchwerk copy');
		await expect(tooltip.getBoundingClientRect().left).toBeGreaterThanOrEqual(label.getBoundingClientRect().right);
		await userEvent.keyboard(' ');
		await expect(args.onChange).not.toHaveBeenCalled();
	},
};
export const Unlabelled: Story = { args: { label: undefined } };
export const Invalid: Story = { args: { label: 'Patchwerk again', layout: 'inline', invalid: true, value: true } };
export const Sizes: Story = {
	render: args => (
		<div className="flex flex-col gap-2">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<BooleanPickerStory key={size} {...args} id={`boolean-picker-${size}`} label={`Show All Stats (${size})`} layout="inline" size={size} value />
			))}
		</div>
	),
};
