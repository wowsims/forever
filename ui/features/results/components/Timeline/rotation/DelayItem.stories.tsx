import type { Meta, StoryObj } from '@storybook/react-vite';
import { castRotationRow, delayRowItem, HEROIC_STRIKE, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { DelayItem } from './DelayItem';

const meta = {
	title: 'Features/Results/Timeline/Components/DelayItem',
	component: DelayItem,
	tags: ['autodocs'],
	args: { item: delayRowItem(1, 1.35), index: 0 },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(HEROIC_STRIKE, []), 4)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof DelayItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Short: Story = {};

export const Long: Story = { args: { item: delayRowItem(0.5, 2.5) } };
