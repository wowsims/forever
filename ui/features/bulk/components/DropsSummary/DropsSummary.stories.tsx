import type { Meta, StoryObj } from '@storybook/react-vite';

import { DropsSummaryView } from './DropsSummary';

const meta = {
	title: 'Features/Bulk/DropsSummary',
	component: DropsSummaryView,
	tags: ['autodocs'],
	args: {
		current: { avg: 750, plusMinus: 2 },
		bestDrop: { name: 'Felstriker', dps: 780, pct: 4 },
		farmFirst: { label: 'Warchief Rend Blackhand', detail: 'Blackrock Spire', upgrades: 2, total: 5 },
	},
	argTypes: { current: { control: 'object' }, bestDrop: { control: 'object' }, farmFirst: { control: 'object' } },
} satisfies Meta<typeof DropsSummaryView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoUpgrade: Story = { args: { bestDrop: null, farmFirst: null } };

/** No card holds an upgrade, as when no ticked source reaches the best drop. */
export const BestDropWithoutCard: Story = { args: { farmFirst: null } };
