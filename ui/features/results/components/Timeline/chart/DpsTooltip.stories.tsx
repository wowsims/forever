import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, DEEP_WOUNDS, dpsLog, HEROIC_STRIKE } from '@testing/storybook/fixtures/results-views';

import { DpsTooltip } from './DpsTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/DpsTooltip',
	component: DpsTooltip,
	tags: ['autodocs'],
	args: {
		log: dpsLog(12, [
			damageLog(12, 1420.5),
			damageLog(12.2, 980.25, { actionId: HEROIC_STRIKE, outcome: 'crit' }),
			damageLog(12.8, 310.25, { actionId: DEEP_WOUNDS, tick: true }),
		]),
	},
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DpsTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeHits: Story = {};
