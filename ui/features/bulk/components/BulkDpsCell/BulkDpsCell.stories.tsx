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
	argTypes: { diff: { control: 'object' }, lowPrecision: { control: 'object' } },
} satisfies Meta<typeof BulkDpsCell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** The current gear's row: labelled, with no difference. */
export const CurrentGear: Story = { args: { label: 'Current gear', avg: 750, diff: null } };

/** Stopped before the final stage: no margin, with a note. */
export const LowPrecision: Story = {
	args: {
		plusMinus: null,
		lowPrecision: {
			note: 'Stopped early',
			tooltip: 'The run dropped this candidate before the final stage, so its DPS carries a wider margin and shows no ±.',
		},
	},
};
