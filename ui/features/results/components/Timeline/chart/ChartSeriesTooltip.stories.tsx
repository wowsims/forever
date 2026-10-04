import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, dpsLog, HEROIC_STRIKE, resourceRowItem, threatLogGroup } from '@testing/storybook/fixtures/results-views';

import { ChartSeriesTooltip } from './ChartSeriesTooltip';

const HITS = [damageLog(12, 1420.5), damageLog(12.2, 980.25, { actionId: HEROIC_STRIKE })];

const meta = {
	title: 'Features/Results/Timeline/Components/ChartSeriesTooltip',
	component: ChartSeriesTooltip,
	tags: ['autodocs'],
	args: { spec: { kind: 'dps' }, log: dpsLog(12, HITS) },
	argTypes: { spec: { control: false }, log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ChartSeriesTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Dps: Story = {};

export const Threat: Story = { args: { spec: { kind: 'threat' }, log: threatLogGroup(12, HITS) } };

export const Rage: Story = { args: { spec: { kind: 'resource', maxValue: 100, includeAuras: true }, log: resourceRowItem(12, 18, 60).log } };
