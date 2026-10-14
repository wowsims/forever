import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { PrecisionOptions } from './PrecisionOptions';

const ep = (key: string) => i18n.t(`sidebar.buttons.stat_weights.precision.${key}`);
const sweep = (key: string) => i18n.t(`sidebar.buttons.stat_sweep.precision.${key}`, { metric: i18n.t('sidebar.results.metrics.tps.label') });

const meta = {
	title: 'Features/StatWeights/PrecisionOptions',
	component: PrecisionOptions,
	tags: ['autodocs'],
	args: {
		value: { precisionMode: 'iterations', iterations: 0, targetCi: 0.01, maxIterationMultiplier: 0 },
		onChange: fn(),
		defaultIterations: 10000,
		idPrefix: 'story-precision',
		targetLabel: ep('target_ci'),
		targetTooltip: ep('target_ci_tooltip'),
		iterationsTooltip: ep('iterations_tooltip'),
		maxMultiplierTooltip: ep('max_multiplier_tooltip'),
		maxDecimalDigits: 4,
	},
	argTypes: { value: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PrecisionOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Iterations: Story = {};

export const TargetPrecision: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetCi: 0.01, maxIterationMultiplier: 0 } },
};

export const SweepTarget: Story = {
	args: {
		value: { precisionMode: 'target', iterations: 0, targetCi: 2, maxIterationMultiplier: 4 },
		targetLabel: sweep('target_ci'),
		targetTooltip: sweep('target_ci_tooltip'),
		iterationsTooltip: sweep('iterations_tooltip'),
		maxMultiplierTooltip: sweep('max_multiplier_tooltip'),
		maxDecimalDigits: 2,
	},
};

export const Disabled: Story = {
	args: { value: { precisionMode: 'target', iterations: 0, targetCi: 0.005, maxIterationMultiplier: 4 }, disabled: true },
};
