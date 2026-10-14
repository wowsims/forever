import type { Meta, StoryObj } from '@storybook/react-vite';
import { Icon } from '@ui-kit/Icon';
import { Fragment } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { Button } from './Button';
import type { ButtonVariant } from './types';

const VARIANTS: Array<ButtonVariant> = [
	'primary',
	'secondary',
	'danger',
	'cancel',
	'link',
	'link-danger',
	'outline-primary',
	'outline-secondary',
	'outline-light',
	'outline-cancel',
	'unstyled',
];

const SIZES = ['default', 'sm', 'xs', 'none'] as const;

const meta = {
	title: 'Kit/Button',
	component: Button,
	tags: ['autodocs'],
	args: {
		children: (
			<>
				<Icon name="gear" /> Settings
			</>
		),
	},
	argTypes: {
		variant: { control: 'select', options: [...VARIANTS, null] },
		size: { control: 'select', options: [undefined, 'sm', 'xs', 'none'] },
		iconOnly: { control: 'boolean' },
		loading: { control: 'boolean' },
		disabled: { control: 'boolean' },
		className: { control: 'text' },
		href: { control: 'text' },
		children: { control: false },
		as: { control: false },
	},
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Primary: Story = { args: { variant: 'primary' } };
export const Secondary: Story = { args: { variant: 'secondary' } };
export const Danger: Story = { args: { variant: 'danger' } };
export const Cancel: Story = { args: { variant: 'cancel' } };
export const Link: Story = { args: { variant: 'link', size: 'sm' } };
export const LinkDanger: Story = { args: { variant: 'link-danger', size: 'sm' } };
export const OutlinePrimary: Story = { args: { variant: 'outline-primary' } };
export const OutlineSecondary: Story = { args: { variant: 'outline-secondary' } };
export const OutlineLight: Story = { args: { variant: 'outline-light' } };
export const OutlineCancel: Story = { args: { variant: 'outline-cancel' } };
export const Unstyled: Story = { args: { variant: 'unstyled' } };

export const Small: Story = { args: { size: 'sm' } };
export const Disabled: Story = { args: { disabled: true } };
/** Disabled but still focusable and hoverable, so a tooltip on it can say why. */
export const FocusableWhenDisabled: Story = { args: { variant: 'outline-primary', disabled: true, focusableWhenDisabled: true } };
export const Loading: Story = { args: { loading: true, children: 'Stat Weights' } };
export const IconOnly: Story = { args: { iconOnly: true, 'aria-label': 'Settings', children: <Icon name="gear" size="lg" /> } };
export const IconOnlyLoading: Story = {
	args: { iconOnly: true, loading: true, 'aria-label': 'Settings', children: <Icon name="gear" size="lg" /> },
	render: args => (
		<span className="inline-flex items-center gap-2 border border-surface-border p-1">
			<Button {...args} />
			<Button {...args} loading={false} />
		</span>
	),
};
export const IconOnlyLink: Story = {
	args: { iconOnly: true, variant: 'link', 'aria-label': 'Help', children: <Icon name="circle-question" style="regular" /> },
};
export const WithHref: Story = { args: { href: 'https://www.wowhead.com/forever', target: '_blank', children: 'Wowhead' } };
export const AsLabel: Story = {
	args: { as: 'label', htmlFor: 'button-story-file', children: 'Choose a file' },
	render: args => (
		<>
			<Button {...args} />
			<input id="button-story-file" type="file" hidden />
		</>
	),
};

export const Matrix: Story = {
	render: args => (
		<div className="grid grid-cols-5 items-center gap-2">
			<span />
			{SIZES.map(size => (
				<span key={size}>{size}</span>
			))}
			{VARIANTS.map(variant => (
				<Fragment key={variant}>
					<span>{variant}</span>
					{SIZES.map(size => (
						<Button key={size} variant={variant} size={size === 'default' ? undefined : size}>
							{args.children}
						</Button>
					))}
				</Fragment>
			))}
		</div>
	),
};

export const Clicks: Story = {
	args: { onClick: fn() },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Settings' }));
		await expect(args.onClick).toHaveBeenCalledOnce();
	},
};
