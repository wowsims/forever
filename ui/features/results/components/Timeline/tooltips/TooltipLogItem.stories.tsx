import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, resourceLog } from '@testing/storybook/fixtures/results-views';
import { TIMELINE_SERIES_TEXT } from '@ui-kit/utils/colors';

import { TooltipLogItem } from './TooltipLogItem';

const meta = {
	title: 'Features/Results/Timeline/Components/TooltipLogItem',
	component: TooltipLogItem,
	tags: ['autodocs'],
	args: { log: damageLog(12, 1420.5), children: '1420.50 damage' },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<div className="ui-timeline-tooltip">
					<ul>
						<Story />
					</ul>
				</div>
			</div>
		),
	],
} satisfies Meta<typeof TooltipLogItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Damage: Story = {};

export const RageInItsColour: Story = { args: { log: resourceLog(12, 60, 30), children: '-30.0', seriesColorClass: TIMELINE_SERIES_TEXT.rage } };
