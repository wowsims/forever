import type { Meta, StoryObj } from '@storybook/react-vite';
import { castLog, castRotationRow, castRowItem, MORTAL_STRIKE, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { CastItem } from './CastItem';

const meta = {
	title: 'Features/Results/Timeline/Components/CastItem',
	component: CastItem,
	tags: ['autodocs'],
	args: { item: castRowItem(1), index: 0, iconUrl: MORTAL_STRIKE.iconUrl },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(MORTAL_STRIKE, []), 6)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof CastItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Hit: Story = {};

export const Crit: Story = { args: { item: castRowItem(1, { outcome: 'crit' }) } };

export const Miss: Story = { args: { item: castRowItem(1, { outcome: 'miss' }) } };

export const Cancelled: Story = { args: { item: castRowItem(1, { cancelled: true, outcome: 'none' }) } };

export const WithTravel: Story = { args: { item: castRowItem(1, { end: 3, travelStart: 3, travelDuration: 0.8 }) } };

const stackTop = castLog(1);

/** Three casts inside the stack window: the last one draws over the others and counts them. */
export const Stacked: Story = { args: { item: castRowItem(1, { log: stackTop, stack: [castLog(1), castLog(1), stackTop] }) } };
