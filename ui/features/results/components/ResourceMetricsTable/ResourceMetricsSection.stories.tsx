import { ResourceType } from '@generated/proto/spell';
import type { ResourceMetrics } from '@sim/proto/sim_result';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { loadRageResult } from '@testing/storybook/fixtures/results-views';

import { createMetricsColumnHelper } from '../../hooks/useMetricsTable';
import type { SimResultData } from '../../model/result_data';
import { MetricsActionCell } from '../MetricsTable';
import { ResourceMetricsSection } from './ResourceMetricsSection';

const helper = createMetricsColumnHelper<ResourceMetrics>();
const COLUMNS = helper.columns([
	helper.accessor(row => row.metric.name, {
		id: 'name',
		header: 'Name',
		cell: info => (
			<MetricsActionCell
				name={info.row.original.metric.name}
				actionId={info.row.original.metric.actionId}
				expandable={false}
				expanded={false}
				onToggle={() => {}}
			/>
		),
	}),
	helper.accessor(row => row.metric.events, { id: 'casts', header: 'Casts', cell: info => info.getValue().toFixed(1) }),
	helper.accessor(row => row.metric.gain, { id: 'gain', header: 'Gain', cell: info => info.getValue().toFixed(1) }),
	helper.accessor(row => row.metric.wastedGain, { id: 'wasted-gain', header: 'Wasted gain', cell: info => info.getValue().toFixed(1) }),
]);

const meta = {
	title: 'Features/Results/ResourceMetricsSection',
	component: ResourceMetricsSection,
	tags: ['autodocs'],
	args: { resourceType: ResourceType.ResourceTypeRage, title: 'Rage', columns: COLUMNS, resultData: null },
	argTypes: { columns: { control: false }, resultData: { control: false } },
	loaders: [loadRageResult],
	render: (args, { loaded }) => <ResourceMetricsSection {...args} resultData={(loaded as { resultData: SimResultData }).resultData} />,
} satisfies Meta<typeof ResourceMetricsSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Rage: Story = {};

export const NoManaInThisRun: Story = { args: { resourceType: ResourceType.ResourceTypeMana, title: 'Mana' } };
