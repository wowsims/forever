import type { Meta, StoryObj } from '@storybook/react-vite';
import { auraRowItem, castRotationRow, FLURRY, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { AuraItem } from './AuraItem';

const meta = {
	title: 'Features/Results/Timeline/Components/AuraItem',
	component: AuraItem,
	tags: ['autodocs'],
	args: { item: auraRowItem(1, 9), index: 0 },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(FLURRY, []), 12)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof AuraItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Uptime: Story = {};

export const Stacking: Story = {
	args: {
		item: auraRowItem(1, 9, [
			{ offset: 0, duration: 3, stacks: 1 },
			{ offset: 3, duration: 3, stacks: 2 },
			{ offset: 6, duration: 2, stacks: 3 },
		]),
	},
};

export const SharingACastRow: Story = { args: { item: { ...auraRowItem(1, 9, [{ offset: 0, duration: 8, stacks: 3 }]), sharesRowWithCast: true } } };
