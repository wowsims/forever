import type { Meta, StoryObj } from '@storybook/react-vite';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Input } from '@ui-kit/FormControl';
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
export const PartsExtraSmallWrappedControl: Story = {
	render: () => (
		<Field.Root id="field-parts-xs" size="xs" layout="inline">
			<Field.Label>Rotation</Field.Label>
			<Field.Control>
				<span className="flex gap-2">
					<Button>Simple</Button>
					<Button>APL</Button>
				</span>
			</Field.Control>
		</Field.Root>
	),
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByTestId('input-root').getAttribute('data-size')).toBe('xs');
		await expect(canvas.getByRole('button', { name: 'APL' }).getAttribute('data-size')).toBe('xs');
	},
};

/** A sized field hands its size to every control inside it, so each row's input, button and checkbox line up. */
export const Sizes: Story = {
	render: () => (
		<div className="flex flex-col gap-4">
			{(['xs', 'sm', 'md'] as const).map(size => (
				<Field.Root key={size} id={`field-size-${size}`} size={size} layout="inline">
					<Field.Label>{size}</Field.Label>
					<Field.Control>
						<span className="flex items-center gap-2">
							<Input defaultValue="0.1" />
							<Button>Apply</Button>
							<BooleanPicker value onChange={() => {}} />
						</span>
					</Field.Control>
				</Field.Root>
			))}
		</div>
	),
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
