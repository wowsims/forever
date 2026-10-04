import type { Meta, StoryObj } from '@storybook/react-vite';
import { auraUptimeLog, BATTLE_SHOUT } from '@testing/storybook/fixtures/results-views';

import { AuraTooltip } from './AuraTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/AuraTooltip',
	component: AuraTooltip,
	tags: ['autodocs'],
	args: { log: auraUptimeLog(4.25, 16.25) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof AuraTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Flurry: Story = {};

export const WholeFight: Story = { args: { log: auraUptimeLog(0, 180, BATTLE_SHOUT) } };
