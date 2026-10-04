import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, HEROIC_STRIKE, threatLogGroup } from '@testing/storybook/fixtures/results-views';

import { ThreatTooltip } from './ThreatTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/ThreatTooltip',
	component: ThreatTooltip,
	tags: ['autodocs'],
	args: { log: threatLogGroup(12, [damageLog(12, 1420.5), damageLog(12.2, 980.25, { actionId: HEROIC_STRIKE })]) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ThreatTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoHits: Story = {};
