import { SavedTalents } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { SavedDataImporter } from './SavedDataImporter';

const meta = {
	title: 'Features/SavedDataTransfer/SavedDataImporter',
	component: SavedDataImporter<SavedTalents>,
	tags: ['autodocs'],
	decorators: [withSimHost()],
	args: {
		type: 'talents',
		label: 'Talents',
		open: true,
		onOpenChange: fn(),
		storage: { entries: [], storedNames: [], saveMany: () => true, codec: SavedTalents },
	},
} satisfies Meta<typeof SavedDataImporter<SavedTalents>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
