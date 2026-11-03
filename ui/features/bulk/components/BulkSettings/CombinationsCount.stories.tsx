import type { Meta, StoryObj } from '@storybook/react-vite';

import { CombinationsCountView, type CombinationsCountViewProps } from './CombinationsCount';

const BUILDS: NonNullable<CombinationsCountViewProps['builds']> = {
	setups: 6,
	bundles: 3,
	encounters: 2,
	runs: 6,
	builds: 12,
	totalIterations: 240000,
	largestRequestIterations: 60000,
};

const meta = {
	title: 'Features/Bulk/BulkSettings/Components/CombinationsCount',
	component: CombinationsCountView,
	tags: ['autodocs'],
	args: { pending: false, ready: true, combinations: 48, iterations: 96000, iterationsLimit: 1000000, builds: null },
} satisfies Meta<typeof CombinationsCountView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Single: Story = { args: { combinations: 1, iterations: 2000 } };

export const Pending: Story = { args: { pending: true } };

export const OverTheLimit: Story = { args: { combinations: 1200, iterations: 2400000 } };

export const OverTheLimitWithTopResults: Story = { args: { combinations: 1200, iterations: 2400000, topResultsHint: true } };

export const Builds: Story = { args: { builds: BUILDS } };

export const BuildsBeforeTheSimIsReady: Story = {
	args: { ready: false, builds: { setups: 0, bundles: 0, encounters: 1, runs: 0, builds: 0, totalIterations: 0, largestRequestIterations: 0 } },
};

export const BuildsWithNoBundleTicked: Story = {
	args: { builds: { setups: 0, bundles: 0, encounters: 2, runs: 0, builds: 0, totalIterations: 0, largestRequestIterations: 0 } },
};

/** One bundle of one gear setup under one encounter is a single build, which leaves Simulate Batch nothing to compare. */
export const BuildsWithOneBuild: Story = {
	args: { builds: { setups: 1, bundles: 1, encounters: 1, runs: 1, builds: 1, totalIterations: 20000, largestRequestIterations: 20000 } },
};

export const BuildsOverTheLimit: Story = { args: { builds: { ...BUILDS, totalIterations: 2400000, largestRequestIterations: 1200000 } } };
