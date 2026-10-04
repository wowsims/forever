import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, dpsLog } from '@testing/storybook/fixtures/results-views';

import { TooltipAuras } from './TooltipAuras';

const meta = {
	title: 'Features/Results/Timeline/Components/TooltipAuras',
	component: TooltipAuras,
	tags: ['autodocs'],
	args: { log: dpsLog(12, [damageLog(12, 1420.5)]) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<div className="ui-timeline-tooltip">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof TooltipAuras>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoAuras: Story = {};

export const NoAuras: Story = { args: { log: damageLog(12, 1420.5) } };
