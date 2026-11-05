import { SavedTalents } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { SavedDataExporter } from './SavedDataExporter';

const entryOf = (name: string) => {
	const data = SavedTalents.create({ talentsString: '2305' });
	return { name, data, json: SavedTalents.toJsonString(data) };
};

const meta = {
	title: 'Features/SavedDataTransfer/SavedDataExporter',
	component: SavedDataExporter<SavedTalents>,
	tags: ['autodocs'],
	decorators: [withSimHost({ player: { getSpec: () => 1 } })],
	args: {
		type: 'talents',
		open: true,
		onOpenChange: fn(),
		storage: { entries: [entryOf('Raid'), entryOf('Dungeon')], codec: SavedTalents },
	},
} satisfies Meta<typeof SavedDataExporter<SavedTalents>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
