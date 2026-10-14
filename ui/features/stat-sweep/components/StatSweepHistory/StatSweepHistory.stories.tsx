import type { Meta, StoryObj } from '@storybook/react-vite';
import { SWEEP_HISTORY_BATCHES } from '@testing/storybook/fixtures/stat-sweep';
import { fn } from 'storybook/test';

import { StatSweepHistory } from './StatSweepHistory';

const meta = {
	title: 'Features/StatSweep/StatSweepHistory',
	component: StatSweepHistory,
	tags: ['autodocs'],
	args: {
		batches: SWEEP_HISTORY_BATCHES,
		onOverlayChange: fn(),
		onDelete: fn(),
		onExportCsv: fn(),
	},
	argTypes: { batches: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof StatSweepHistory>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The newest sweep open with every run drawn, then a closed one-stat sweep and a closed older sweep with one of its runs drawn. */
export const Default: Story = {};

export const Empty: Story = { args: { batches: [], onExportCsv: undefined } };
