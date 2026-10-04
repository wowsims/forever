import type { Meta, StoryObj } from '@storybook/react-vite';
import { rageRotationRow, resourceRowItem, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { ResourceItem } from './ResourceItem';

const meta = {
	title: 'Features/Results/Timeline/Components/ResourceItem',
	component: ResourceItem,
	tags: ['autodocs'],
	args: { item: resourceRowItem(1, 4, 60), index: 0, cssName: 'rage' },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(rageRotationRow([]), 6)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof ResourceItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rage: Story = {};

export const EnergyFill: Story = {
	args: { item: resourceRowItem(1, 4, 80, { display: 'fill', fillPercent: 80, text: '80' }), cssName: 'energy' },
};

export const ManaPercent: Story = { args: { item: resourceRowItem(1, 4, 2400, { display: 'percent', text: '62%' }), cssName: 'mana' } };
