import type { Meta, StoryObj } from '@storybook/react-vite';

import { SummaryTableRow } from './SummaryTableRow';

const meta = {
	title: 'Kit/SummaryTableRow',
	component: SummaryTableRow,
	tags: ['autodocs'],
	argTypes: { children: { control: false } },
	render: args => (
		<div className="flex w-72 flex-col gap-1">
			<SummaryTableRow {...args}>
				<span>Casts</span>
				<span>128</span>
			</SummaryTableRow>
			<SummaryTableRow {...args}>
				<span className="text-success">Hits</span>
				<span>97</span>
			</SummaryTableRow>
			<SummaryTableRow {...args}>
				<span className="text-danger">Misses</span>
				<span>31</span>
			</SummaryTableRow>
		</div>
	),
} satisfies Meta<typeof SummaryTableRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
