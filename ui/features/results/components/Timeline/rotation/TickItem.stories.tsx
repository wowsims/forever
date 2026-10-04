import type { Meta, StoryObj } from '@storybook/react-vite';
import { castRotationRow, DEEP_WOUNDS, rotationTrackStyle, tickRowItem } from '@testing/storybook/fixtures/results-views';

import { TickItem } from './TickItem';

const meta = {
	title: 'Features/Results/Timeline/Components/TickItem',
	component: TickItem,
	tags: ['autodocs'],
	args: { item: tickRowItem(2), index: 0 },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(DEEP_WOUNDS, []), 4)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof TickItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OneTick: Story = {};
