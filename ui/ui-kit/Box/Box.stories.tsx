import type { Meta, StoryObj } from '@storybook/react-vite';

import { Box, type BoxBorder, type BoxPadding } from './Box';

const BORDERS: Array<BoxBorder> = ['none', 'border', 'border-muted', 'frozen', 'brand', 'warning'];
const PADDINGS: Array<BoxPadding> = ['none', 'xs', 'sm', 'md'];

const meta = {
	title: 'Kit/Box',
	component: Box,
	tags: ['autodocs'],
	args: { children: 'Health 7354' },
	argTypes: {
		border: { control: 'select', options: BORDERS },
		padding: { control: 'inline-radio', options: PADDINGS },
		children: { control: false },
	},
} satisfies Meta<typeof Box>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Matrix: Story = {
	render: args => (
		<div className="grid grid-cols-4 items-start gap-2">
			{BORDERS.flatMap(border =>
				PADDINGS.map(padding => (
					<Box key={`${border}-${padding}`} {...args} border={border} padding={padding}>
						{border} / {padding}
					</Box>
				)),
			)}
		</div>
	),
};
