import type { StatWeightsResult, StatWeightValues } from '@generated/proto/api';
import i18n from '@i18n/config';
import type { DisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyStatFilter } from '@testing/storybook/fixtures/stat-filters';
import { makeEpColumns, makeStatWeightsResult, STAT_WEIGHT_STATS } from '@testing/storybook/fixtures/stat-weights';
import { Tooltip } from '@ui-kit/Tooltip';
import { expect, fn, within } from 'storybook/test';

import { EpWeightsTableView } from './EpWeightsTable';
import { StatWeightRowView } from './StatWeightRow';
import { StatsType } from './types';
import { EP_TOOLTIP_ID } from './utils';

const NAMES = ['Strength', 'Agility', 'Attack Power', 'Crit Rating'];
const EP_RATIOS = [1, 0, 0, 0, 0, 0];
const DAMAGE: DisplayMetrics = { damage: true, healing: false, threat: false };
const ALL: DisplayMetrics = { damage: true, healing: true, threat: true };

const metricsOf = (result: StatWeightsResult | null, displayMetrics: DisplayMetrics) => {
	const all: Array<[StatWeightValues | undefined, keyof DisplayMetrics]> = [
		[result?.dps, 'damage'],
		[result?.hps, 'healing'],
		[result?.tps, 'threat'],
		[result?.dtps, 'threat'],
		[result?.tmi, 'threat'],
		[result?.pDeath, 'threat'],
	];
	return all
		.map(([statWeights, metric], ratioIndex) => ({ statWeights, metric, ratioIndex, epRatio: EP_RATIOS[ratioIndex] }))
		.filter(entry => displayMetrics[entry.metric]);
};

const rows = (result: StatWeightsResult | null, displayMetrics: DisplayMetrics, showThreatMetrics: boolean, showUpdateColumn = true) =>
	STAT_WEIGHT_STATS.map((stat, index) => (
		<StatWeightRowView
			key={NAMES[index]}
			stat={stat}
			fullName={NAMES[index]}
			shortName={NAMES[index]}
			metrics={showUpdateColumn ? metricsOf(result, displayMetrics) : []}
			iterations={1000}
			epDelta={0}
			currentEp={result ? stat.getProtoValue(result.dps!.epValues!) : 0}
			onCurrentEpChange={fn()}
			include={showUpdateColumn ? { value: true, disabled: index === 0, onChange: fn() } : undefined}
			showThreatMetrics={showThreatMetrics}
			showUpdateColumn={showUpdateColumn}
		/>
	));

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpWeightsTable',
	component: EpWeightsTableView,
	tags: ['autodocs'],
	args: {
		columns: makeEpColumns(DAMAGE),
		statsType: StatsType.Ep,
		epRatios: EP_RATIOS,
		onEpRatioChange: fn(),
		onComputeEp: fn(),
		showThreatMetrics: false,
		showEpRatios: true,
		statAction: storyStatFilter('ep-stat-filter', i18n.t('stat_filters.triggers.ep')),
		children: rows(null, DAMAGE, false),
	},
	argTypes: {
		columns: { control: false },
		statAction: { control: false },
		children: { control: false },
		statsType: { control: 'select', options: [StatsType.Ep, StatsType.Weight] },
	},
	render: args => (
		<>
			<EpWeightsTableView {...args} />
			<Tooltip id={EP_TOOLTIP_ID} />
		</>
	),
} satisfies Meta<typeof EpWeightsTableView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Calculated: Story = { args: { children: rows(makeStatWeightsResult(), DAMAGE, false) } };

export const Weights: Story = { args: { statsType: StatsType.Weight, children: rows(makeStatWeightsResult(), DAMAGE, false) } };

/** With threat metrics on every metric shows, and the inputs go compact to fit the columns. */
export const ThreatMetrics: Story = {
	args: { columns: makeEpColumns(ALL), showThreatMetrics: true, children: rows(makeStatWeightsResult(), ALL, true) },
};

/** Scrolled inside a short dialog: the header, and the rule under it, stay put while the rows pass beneath. */
export const ScrolledUnderHeader: Story = {
	args: { columns: makeEpColumns(ALL), showThreatMetrics: true, children: rows(makeStatWeightsResult(), ALL, true) },
	decorators: [
		Story => (
			<div className="flex h-44 flex-col">
				<Story />
			</div>
		),
	],
	play: async ({ canvasElement }) => {
		const container = within(canvasElement).getByTestId('results-ep-table-container');
		container.scrollTop = 60;
		await expect(container.scrollTop).toBeGreaterThan(0);
	},
};

/** A gear planner computes nothing: only the editable Current EP column is left. */
export const GearPlanner: Story = {
	args: {
		columns: makeEpColumns({ damage: false, healing: false, threat: false }),
		showEpRatios: false,
		showUpdateColumn: false,
		children: rows(null, DAMAGE, false, false),
	},
};
