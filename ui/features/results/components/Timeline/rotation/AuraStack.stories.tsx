import type { Meta, StoryObj } from '@storybook/react-vite';
import { castRotationRow, FLURRY, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { AuraStack } from './AuraStack';

const meta = {
	title: 'Features/Results/Timeline/Components/AuraStack',
	component: AuraStack,
	tags: ['autodocs'],
	args: { segment: { offset: 1, duration: 4, stacks: 3 } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(FLURRY, []), 8)}>
				<div className="ui-timeline-row-track">
					<div className="ui-timeline-item top-(--rotation-item-top) h-(--rotation-item-h) w-full bg-timeline-aura opacity-50">
						<Story />
					</div>
				</div>
			</div>
		),
	],
} satisfies Meta<typeof AuraStack>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeStacks: Story = {};

export const Indented: Story = { args: { indent: true } };
