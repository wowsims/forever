import { ResourceType } from '@generated/proto/spell';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { auraUptimeLog, BATTLE_SHOUT, resourceGroupLog, resourceLog } from '@testing/storybook/fixtures/results-views';

import { ResourceTooltip } from './ResourceTooltip';

const RAGE = resourceGroupLog(12, [resourceLog(12, 60, 30), resourceLog(12.5, 30, 45)]);

const meta = {
	title: 'Features/Results/Timeline/Components/ResourceTooltip',
	component: ResourceTooltip,
	tags: ['autodocs'],
	args: { log: RAGE, maxValue: 100, includeAuras: false },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ResourceTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rage: Story = {};

export const ManaAsAShare: Story = {
	args: {
		log: resourceGroupLog(12, [resourceLog(12, 2400, 2100, { resourceType: ResourceType.ResourceTypeMana, total: 4000 })]),
		maxValue: 4000,
	},
};

export const WithActiveAuras: Story = {
	args: { log: { ...RAGE, activeAuras: [auraUptimeLog(4, 16), auraUptimeLog(0, 180, BATTLE_SHOUT)] }, includeAuras: true },
};
