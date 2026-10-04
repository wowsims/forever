import type { Meta, StoryObj } from '@storybook/react-vite';
import { gcdRotationRow, gcdSegmentRowItem, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { GcdSegmentItem } from './GcdSegmentItem';

const meta = {
	title: 'Features/Results/Timeline/Components/GcdSegmentItem',
	component: GcdSegmentItem,
	tags: ['autodocs'],
	args: { item: gcdSegmentRowItem(1), index: 0 },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(gcdRotationRow([]), 4)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof GcdSegmentItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OneGcd: Story = {};
