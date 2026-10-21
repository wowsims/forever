import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { BulkCompareButtonView } from './BulkCompareButton';

const meta = {
	title: 'Features/Bulk/BulkCompareButton',
	component: BulkCompareButtonView,
	tags: ['autodocs'],
	args: { onOpen: fn() },
} satisfies Meta<typeof BulkCompareButtonView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
