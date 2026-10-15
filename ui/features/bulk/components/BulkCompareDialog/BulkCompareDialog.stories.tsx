import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeBulkCompareHost } from '@testing/storybook/fixtures/bulk-compare';
import { fn } from 'storybook/test';

import { BulkCompareDialog } from './BulkCompareDialog';

const host = makeBulkCompareHost();

const meta = {
	title: 'Features/Bulk/BulkCompareDialog',
	component: BulkCompareDialog,
	tags: ['autodocs'],
	args: { open: true, onOpenChange: fn() },
	parameters: { smoke: false, smokeReason: 'chart.js needs a canvas' },
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

/** Three finished encounters, one of them health-based, with sets that rank in one, two or all three of them. */
export const Open: Story = {};
