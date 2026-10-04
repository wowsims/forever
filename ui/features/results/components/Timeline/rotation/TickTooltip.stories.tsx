import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { damageLog, DEEP_WOUNDS, resultsHostParts, TARGET } from '@testing/storybook/fixtures/results-views';

import { TickTooltip } from './TickTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/TickTooltip',
	component: TickTooltip,
	tags: ['autodocs'],
	args: { log: damageLog(9, 310.25, { actionId: DEEP_WOUNDS, tick: true }) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
		withSimHost(resultsHostParts(null)),
	],
} satisfies Meta<typeof TickTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithThreat: Story = {};

export const CritTick: Story = { args: { log: damageLog(9, 620.5, { actionId: DEEP_WOUNDS, tick: true, outcome: 'crit' }) } };

export const DealtByTheTarget: Story = { args: { log: damageLog(9, 120, { actionId: DEEP_WOUNDS, tick: true, source: TARGET, target: null }) } };
