import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { BULK_COMPARE_BEST_ENTRIES, makeBulkCompareHost } from '@testing/storybook/fixtures/bulk-compare';
import { fn } from 'storybook/test';

import { BulkCompareDialog } from './BulkCompareDialog';

const host = makeBulkCompareHost();

const meta = {
	title: 'Features/Bulk/BulkCompareDialog',
	component: BulkCompareDialog,
	tags: ['autodocs'],
	args: { open: true, onOpenChange: fn() },
	decorators: [
		Story => (
			<SimHostProvider host={host}>
				<Story />
			</SimHostProvider>
		),
	],
} satisfies Meta<typeof BulkCompareDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Three finished encounters with sets that rank in one, two or all three of them. */
export const Open: Story = {};

const bestHost = makeBulkCompareHost(BULK_COMPARE_BEST_ENTRIES);

/** A Builds batch with the Current bundle unticked: no current row, each encounter's best set at 0.00% and the rest losses. */
export const CurrentBundleUnticked: Story = {
	decorators: [
		Story => (
			<SimHostProvider host={bestHost}>
				<Story />
			</SimHostProvider>
		),
	],
};
