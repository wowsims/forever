import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { DropsRunCountView } from './DropsRunCount';

const meta = {
	title: 'Features/Bulk/DropsRunCount',
	component: DropsRunCountView,
	tags: ['autodocs'],
	args: {
		pending: false,
		ready: true,
		picked: 12,
		twoHanderSkipped: 0,
		candidates: 19,
		iterations: 380000,
		iterationsLimit: 1000000,
		error: null,
		combinedGain: false,
		onCombinedGainChange: fn(),
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsRunCountView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const CombinedGain: Story = { args: { combinedGain: true } };

/** Under a worn two-hander: the picked one-handers and off-hand items are skipped. */
export const WornTwoHander: Story = { args: { twoHanderSkipped: 4 } };

export const OverIterationsLimit: Story = { args: { picked: 60, candidates: 140, iterations: 2800000 } };

export const Error: Story = { args: { candidates: 0, iterations: 0, error: 'Too many combinations: 1,200,000 exceeds the limit of 1,000,000.' } };

/** The run card's narrowest column: Minimum iterations wraps, and its help, number and warning sit on the label's first line. */
export const NarrowCard: Story = {
	args: { picked: 60, candidates: 140, iterations: 2800000 },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
};
