import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { BulkResultsCsvButtonView } from './BulkResultsCsvButton';

const meta = {
	title: 'Features/Bulk/BulkResultsCsvButton',
	component: BulkResultsCsvButtonView,
	tags: ['autodocs'],
	args: { onExport: fn() },
} satisfies Meta<typeof BulkResultsCsvButtonView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
