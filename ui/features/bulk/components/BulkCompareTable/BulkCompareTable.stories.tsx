import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { BulkCompareTable } from './BulkCompareTable';

const meta = {
	title: 'Features/Bulk/BulkCompareTable',
	component: BulkCompareTable,
	tags: ['autodocs'],
	args: {
		encounters: [
			{ label: 'Council', summary: '1:30 ±10s · 1 target' },
			{ label: 'Current', summary: '3:00 ±5s · 1 target' },
			{ label: 'Patchwerk', summary: 'Health · 1 target' },
		],
		rows: [
			{
				key: 'current',
				label: 'Current gear',
				changes: [],
				cells: [
					{ dps: '812.40', margin: '3', gain: null },
					{ dps: '776.30', margin: '2', gain: null },
					{ dps: '760.00', margin: '2', gain: null },
				],
			},
			{
				key: 'a',
				label: 'Set 1',
				changes: ['Band of Accuria', 'Blackhand’s Breadth'],
				cells: [{ dps: '838.10', margin: '3', gain: '+25.70 (3.16%)' }, { dps: '797.41', margin: '2', gain: '+21.11 (2.72%)' }, null],
			},
			{
				key: 'b',
				label: 'Set 2',
				changes: ['Crul’shorukh, Edge of Chaos (Crusader)'],
				cells: [null, { dps: '801.02', margin: '2', gain: '+24.72 (3.18%)' }, { dps: '771.80', margin: '2', gain: '+11.80 (1.55%)' }],
			},
		],
		highlighted: null,
		onHighlight: fn(),
	},
	argTypes: { encounters: { control: false }, rows: { control: false } },
} satisfies Meta<typeof BulkCompareTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThreeEncounters: Story = {};

export const SetPickedOut: Story = { args: { highlighted: 'a' } };
