import type { Meta, StoryObj } from '@storybook/react-vite';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import { MetricTooltip } from './MetricTooltip';
import { metricForAnchor } from './tooltipAnchor';

interface Metric {
	name: string;
	casts: number;
}

const METRICS = new Map<string, Metric>([
	['0', { name: 'Mortal Strike', casts: 42 }],
	['1', { name: 'Heroic Strike', casts: 118 }],
]);

const meta = {
	title: 'Features/Results/MetricsTable/Components/MetricTooltip',
	component: MetricTooltip<Metric>,
	tags: ['autodocs'],
	args: {
		id: 'story-metric-tooltip',
		forAnchor: metricForAnchor(METRICS),
		body: metric => (
			<span>
				{metric.name}: {metric.casts} casts
			</span>
		),
	},
	argTypes: { forAnchor: { control: false }, body: { control: false } },
	render: args => (
		<div className="flex gap-4">
			{[...METRICS.entries()].map(([rowId, metric]) => (
				<span key={rowId} className="underline" data-row-id={rowId} {...tooltipAnchorProps(args.id)}>
					Hover {metric.name}
				</span>
			))}
			<MetricTooltip {...args} />
		</div>
	),
} satisfies Meta<typeof MetricTooltip<Metric>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OneColumnTwoRows: Story = {};
