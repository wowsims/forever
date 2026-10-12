import type { Meta, StoryObj } from '@storybook/react-vite';

import { CombinationsCountView } from './CombinationsCount';

const meta = {
	title: 'Features/Bulk/BulkSettings/Components/CombinationsCount',
	component: CombinationsCountView,
	tags: ['autodocs'],
	args: { pending: false, combinations: 48, iterations: 96000, iterationsLimit: 1000000, durations: 1 },
} satisfies Meta<typeof CombinationsCountView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Single: Story = { args: { combinations: 1, iterations: 2000 } };

export const MultipleDurations: Story = { args: { durations: 3 } };

export const Pending: Story = { args: { pending: true } };

export const OverTheLimit: Story = { args: { combinations: 1200, iterations: 2400000 } };
