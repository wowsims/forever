import type { Meta, StoryObj } from '@storybook/react-vite';

import { DEFINITIONS } from './definitions';
import { Icon } from './Icon';
import type { IconSize, IconStyle } from './types';

const STYLES: Array<{ style: IconStyle; name: 'circle-question' | 'github' }> = [
	{ style: 'solid', name: 'circle-question' },
	{ style: 'regular', name: 'circle-question' },
	{ style: 'brands', name: 'github' },
	{ style: 'base', name: 'circle-question' },
];

const SIZES: IconSize[] = ['2xs', 'xs', 'sm', '1x', 'lg', 'xl', '2xl', '2x', '3x', '4x', '5x'];

const meta = {
	title: 'Kit/Icon',
	component: Icon,
	tags: ['autodocs'],
	args: { name: 'star' },
	argTypes: {
		name: { control: 'select', options: Object.keys(DEFINITIONS) },
		style: { control: 'select', options: ['solid', 'regular', 'brands', 'base'] },
		size: { control: 'select', options: [undefined, ...SIZES] },
		title: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof Icon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Styles: Story = {
	render: () => (
		<div className="flex items-end gap-6">
			{STYLES.map(({ style, name }) => (
				<div key={style} className="flex flex-col items-center gap-1">
					<Icon name={name} style={style} size="2x" />
					<code className="text-ui">{style}</code>
				</div>
			))}
		</div>
	),
};

export const Sizes: Story = {
	render: args => (
		<div className="flex flex-wrap items-end gap-6">
			{SIZES.map(size => (
				<div key={size} className="flex flex-col items-center gap-1">
					<Icon {...args} size={size} />
					<code className="text-ui">{size}</code>
				</div>
			))}
		</div>
	),
};

export const FixedWidth: Story = {
	render: () => (
		<div className="flex flex-col gap-1">
			{(['circle-info', 'triangle-exclamation', 'star', 'user'] as const).map(name => (
				<div key={name} className="flex items-center gap-2">
					<Icon name={name} size="lg" fixedWidth className="bg-surface" />
					<code className="text-ui">{name}</code>
				</div>
			))}
		</div>
	),
};
