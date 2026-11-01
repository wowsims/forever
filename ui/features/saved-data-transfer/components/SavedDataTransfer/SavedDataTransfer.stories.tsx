import { SavedTalents } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makePowerToolsHost } from '@testing/storybook/fixtures/saved-data-transfer';
import { fn } from 'storybook/test';

import { SavedDataTransfer } from './SavedDataTransfer';

const host = makePowerToolsHost();

const meta = {
	title: 'Features/SavedDataTransfer/SavedDataTransfer',
	component: SavedDataTransfer<SavedTalents>,
	tags: ['autodocs'],
	decorators: [
		Story => (
			<SimHostProvider host={host}>
				<div className="flex w-64">
					<Story />
				</div>
			</SimHostProvider>
		),
	],
	args: {
		type: 'talents',
		label: 'Talents',
		storage: { entries: [], storedNames: [], save: fn(), saveMany: () => true, remove: fn(), codec: SavedTalents },
	},
} satisfies Meta<typeof SavedDataTransfer<SavedTalents>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
