import type { ActionMetrics } from '@sim/proto/sim_result';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { type AttackRowsFixture, loadAttackRows } from '@testing/storybook/fixtures/results-views';

import { createMetricsColumnHelper } from '../../hooks/useMetricsTable';
import { attackMetricsColumns } from '../AttackMetricsColumns';
import { MetricsTable } from './MetricsTable';

const helper = createMetricsColumnHelper<ActionMetrics>();

const columnsFor = (maxDamage: number) =>
	helper.columns([
		attackMetricsColumns.name(),
		attackMetricsColumns.primary({
			id: 'damage-done',
			header: 'Damage done',
			total: metric => metric.avgDamage,
			value: metric => metric.damage,
			percentage: metric => metric.totalDamagePercent,
			max: { current: maxDamage },
			tooltipId: 'story-damage',
		}),
		attackMetricsColumns.casts(),
		attackMetricsColumns.rate({ id: 'dps', header: 'DPS', value: metric => metric.dps }),
	]);

const meta = {
	title: 'Features/Results/MetricsTable/MetricsTable',
	component: MetricsTable<ActionMetrics>,
	tags: ['autodocs'],
	args: { rootTestId: 'story-metrics-root', columns: columnsFor(0), rows: [], sortColumnId: 'damage-done', hasResult: true },
	argTypes: { columns: { control: false }, rows: { control: false } },
	loaders: [loadAttackRows],
	render: (args, { loaded }) => {
		const { rows, maxDamage } = loaded as AttackRowsFixture;
		return <MetricsTable {...args} columns={columnsFor(maxDamage)} rows={args.hasResult ? rows : []} />;
	},
} satisfies Meta<typeof MetricsTable<ActionMetrics>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ExpandedParent: Story = {};

export const BeforeTheFirstRun: Story = { args: { hasResult: false } };
