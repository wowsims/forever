import type { Meta, StoryObj } from '@storybook/react-vite';

import { Pill, type PillSize, type PillTone } from './Pill';

const TONES: Array<PillTone> = ['neutral', 'positive', 'negative', 'warning', 'primary'];
const SIZES: Array<PillSize> = ['xs', 'sm', 'md'];

const meta = {
	title: 'Kit/Pill',
	component: Pill,
	tags: ['autodocs'],
	args: { children: '+4.21 vs ref' },
	argTypes: {
		tone: { control: 'select', options: TONES },
		size: { control: 'select', options: SIZES },
		rounded: { control: 'select', options: ['full', 'xs'] },
	},
} satisfies Meta<typeof Pill>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Tones: Story = {
	render: args => (
		<div className="flex flex-wrap gap-2">
			{TONES.map(tone => (
				<Pill key={tone} {...args} tone={tone}>
					{tone}
				</Pill>
			))}
		</div>
	),
};

export const Squared: Story = {
	args: { rounded: 'xs' },
};

export const Matrix: Story = {
	render: args => (
		<div className="grid grid-cols-5 items-center gap-2">
			{SIZES.flatMap(size =>
				TONES.map(tone => (
					<Pill key={`${size}-${tone}`} {...args} tone={tone} size={size}>
						{size}
					</Pill>
				)),
			)}
		</div>
	),
};
