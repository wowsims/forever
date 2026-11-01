import type { Meta, StoryObj } from '@storybook/react-vite';
import { bundleListProps, DUPLICATE_BUNDLE } from '@testing/storybook/fixtures/bulk-builds';
import { expect, within } from 'storybook/test';

import { BulkBundleList } from './BulkBundleList';

const meta = {
	title: 'Features/Bulk/BulkBundleList',
	component: BulkBundleList,
	tags: ['autodocs'],
	args: bundleListProps(),
	argTypes: {
		current: { control: false },
		rows: { control: false },
		testIds: { control: false },
	},
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkBundleList>;

export default meta;
type Story = StoryObj<typeof meta>;

// The last bundle sims what the first one does, so it shows unticked and cannot be ticked, and keeps its pencil and ×.
export const Default: Story = {
	play: async ({ canvasElement }) => {
		const box = within(canvasElement).getByRole('checkbox', { name: DUPLICATE_BUNDLE });
		await expect(box.getAttribute('aria-checked')).toBe('false');
		await expect(box.getAttribute('aria-disabled')).toBe('true');
		const row = box.closest<HTMLElement>('[data-testid="bulk-bundle"]')!;
		await expect(
			within(row)
				.getAllByRole('button')
				.map(button => (button as HTMLButtonElement).disabled),
		).toEqual([false, false]);
	},
};

export const CurrentOnly: Story = { args: { rows: [] } };
