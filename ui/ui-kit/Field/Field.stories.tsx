import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Field } from './Field';

const meta = {
	title: 'Kit/Field',
	component: Field,
	tags: ['autodocs'],
	args: {
		id: 'field-story',
		label: 'Cast Delay',
		children: <input id="field-story" className="ui-input" defaultValue="0.1" />,
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split', 'icon'] },
		disabled: { control: 'boolean' },
		grow: { control: 'boolean' },
		className: { control: 'text' },
		children: { control: false },
		leading: { control: false },
		after: { control: false },
		clear: { control: false },
		ref: { control: false },
	},
} satisfies Meta<typeof Field>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Inline: Story = { args: { layout: 'inline' } };
export const Split: Story = { args: { layout: 'split' } };
export const WithTooltip: Story = { args: { labelTooltip: 'Seconds of <b>human</b> reaction time before each cast.' } };
export const WithDescription: Story = { args: { description: 'Delay before the first cast.' } };
export const IconField: Story = {
	args: { layout: 'icon', leading: <span className="ui-icon-picker-swatch" />, children: undefined },
};
export const WithClear: Story = {
	args: {
		label: 'Search',
		children: <input id="field-story" className="ui-input" defaultValue="Thunderfury" />,
		clear: { show: true, label: 'Clear', onClear: fn() },
	},
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Clear' }));
		await expect(args.clear?.onClear).toHaveBeenCalledOnce();
	},
};
export const Disabled: Story = { args: { disabled: true } };
export const Group: Story = {
	args: {
		id: undefined,
		label: 'Elixirs',
		role: 'group',
		layout: 'inline',
		children: (
			<span className="flex gap-2">
				<Button size="sm">Flask of Supreme Power</Button>
				<Button size="sm">Elixir of the Mongoose</Button>
			</span>
		),
	},
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('group', { name: 'Elixirs' })).toBeTruthy();
	},
};
export const Hidden: Story = { args: { hidden: true } };

export const Parts: Story = {
	render: () => (
		<Field.Root id="field-parts" layout="split">
			<Field.Label tooltip="Seconds of <b>human</b> reaction time before each cast.">Cast Delay</Field.Label>
			<Field.Description>Delay before the first cast.</Field.Description>
			<Field.Control>
				<input className="ui-input" defaultValue="0.1" />
			</Field.Control>
			<Field.After>s</Field.After>
		</Field.Root>
	),
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const input = canvas.getByRole('textbox', { name: 'Cast Delay' });
		await expect(input.id).toBe('field-parts');
		await expect(input.getAttribute('aria-describedby')).toBe(canvas.getByTestId('input-description').id);
	},
};
export const PartsCompactWrappedControl: Story = {
	render: () => (
		<Field.Root id="field-parts-compact" density="compact" layout="inline">
			<Field.Label>Rotation</Field.Label>
			<Field.Control>
				<span className="flex gap-2">
					<Button size="sm">Simple</Button>
					<Button size="sm">APL</Button>
				</span>
			</Field.Control>
		</Field.Root>
	),
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByTestId('input-root').getAttribute('data-density')).toBe('compact');
	},
};
export const PartsLeadingAndClear: Story = {
	render: () => (
		<Field.Root id="field-parts-search" grow={false}>
			<Field.Leading>
				<span className="ui-icon-picker-swatch" />
			</Field.Leading>
			<Field.Label>Search</Field.Label>
			<Field.Clear show label="Clear" onClear={fn()}>
				<Field.Control>
					<input className="ui-input" defaultValue="Thunderfury" />
				</Field.Control>
			</Field.Clear>
		</Field.Root>
	),
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('button', { name: 'Clear' })).toBeTruthy();
	},
};
