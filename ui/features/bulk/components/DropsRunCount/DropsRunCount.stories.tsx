import type { Meta, StoryObj } from '@storybook/react-vite';

import { DropsRunCountView } from './DropsRunCount';

const meta = {
	title: 'Features/Bulk/DropsRunCount',
	component: DropsRunCountView,
	tags: ['autodocs'],
	args: { pending: false, ready: true, picked: 12, candidates: 19, iterations: 380000, iterationsLimit: 1000000, error: null },
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

export const OverIterationsLimit: Story = { args: { picked: 60, candidates: 140, iterations: 2800000 } };

export const Error: Story = { args: { candidates: 0, iterations: 0, error: 'Too many combinations: 1,200,000 exceeds the limit of 1,000,000.' } };
