import type { Meta, StoryObj } from '@storybook/react-vite';

import { DropsSummaryView } from './DropsSummary';

const meta = {
	title: 'Features/Bulk/DropsSummary',
	component: DropsSummaryView,
	tags: ['autodocs'],
	args: {
		current: { avg: 750, plusMinus: 2 },
		bestDrop: { name: 'Felstriker', dps: 780, pct: 4 },
		bestBoss: { label: 'Warchief Rend Blackhand', detail: 'Blackrock Spire', pct: 4 },
	},
	argTypes: { current: { control: 'object' }, bestDrop: { control: 'object' }, bestBoss: { control: 'object' } },
} satisfies Meta<typeof DropsSummaryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoUpgrade: Story = { args: { bestDrop: null, bestBoss: null } };

/** The best drop has no card, as when no ticked source reaches it. */
export const BestDropWithoutCard: Story = { args: { bestBoss: null } };
