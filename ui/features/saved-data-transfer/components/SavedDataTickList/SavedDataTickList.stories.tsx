import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { SavedDataTickList } from './SavedDataTickList';

const meta = {
	title: 'Features/SavedDataTransfer/SavedDataTickList',
	component: SavedDataTickList,
	tags: ['autodocs'],
	args: {
		onChange: fn(),
		rows: [
			{ name: 'Phase 1', checked: true },
			{ name: 'Fire Resistance', note: 'Saves as "Fire Resistance (2)"', checked: true },
			{ name: 'Tank Swap', note: 'Already saved', checked: false },
		],
	},
} satisfies Meta<typeof SavedDataTickList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ManyRows: Story = {
	args: {
		rows: [60, 90, 120, 180, 240, 300].flatMap(seconds =>
			['', ' - 2T', ' - 3T', ' - 5T'].map(targets => ({ name: `${seconds}/5${targets}`, checked: true })),
		),
	},
};
