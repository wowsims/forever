import type { Meta, StoryObj } from '@storybook/react-vite';

import { BulkDpsCell } from './BulkDpsCell';

const meta = {
	title: 'Features/Bulk/BulkDpsCell',
	component: BulkDpsCell,
	tags: ['autodocs'],
	args: {
		className: 'w-52',
		avg: 767.5,
		plusMinus: 2.4,
		diff: { text: '+17.50 (2.33%)', tone: 'positive', significance: 'Difference is significantly different (Z = 17.500).' },
	},
	argTypes: { diff: { control: 'object' } },
} satisfies Meta<typeof BulkDpsCell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** The current gear's row: labelled, with no difference. */
export const CurrentGear: Story = { args: { label: 'Current gear', avg: 750, diff: null } };
