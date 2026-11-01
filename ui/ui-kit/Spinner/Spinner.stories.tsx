import type { Meta, StoryObj } from '@storybook/react-vite';

import { Spinner } from './Spinner';

const SIZES = ['xs', 'sm', 'md', 'lg'] as const;

const meta = {
	title: 'Kit/Spinner',
	component: Spinner,
	tags: ['autodocs'],
	argTypes: {
		size: { control: 'select', options: SIZES },
		tone: { control: 'inline-radio', options: ['default', 'strong'] },
		label: { control: 'text' },
		className: { control: 'text' },
	},
} satisfies Meta<typeof Spinner>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Sizes: Story = {
	render: () => (
		<div className="flex items-end gap-6 bg-surface p-4">
			{SIZES.map(size => (
				<div key={size} className="flex flex-col items-center gap-1">
					<Spinner size={size} />
					<code className="text-ui">{size}</code>
				</div>
			))}
		</div>
	),
};

export const Strong: Story = {
	args: { tone: 'strong', label: 'Running' },
	render: args => (
		<div className="flex items-end gap-6 bg-surface p-4">
			{SIZES.map(size => (
				<div key={size} className="flex flex-col items-center gap-1">
					<Spinner {...args} size={size} />
					<code className="text-ui">{size}</code>
				</div>
			))}
		</div>
	),
};
