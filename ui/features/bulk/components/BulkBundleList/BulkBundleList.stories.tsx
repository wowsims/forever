import type { Meta, StoryObj } from '@storybook/react-vite';
import { bundleListProps, DUPLICATE_BUNDLE, NAMED_BUNDLE, NAMED_DUPLICATE_BUNDLE } from '@testing/storybook/fixtures/bulk-builds';
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

// The third and last bundles sim what the first one does, so they show unticked and cannot be ticked, and keep their pencil
// and ×. The second and last are named: their label's tooltip lists their parts, after the last one's reason.
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
		const description = (name: string) => {
			const id = within(canvasElement).getByRole('checkbox', { name }).getAttribute('aria-describedby');
			return id ? canvasElement.ownerDocument.getElementById(id)!.textContent : null;
		};
		await expect(description(meta.args.rows[0].label)).toBeNull();
		const named = description(NAMED_BUNDLE)!;
		for (const part of ['Pre-BiS (Arms)', 'Arms', 'Arms APL', 'Pre-BiS - Arms']) await expect(named).toContain(part);
		const namedBox = within(canvasElement).getByRole('checkbox', { name: NAMED_BUNDLE });
		const [edit] = within(namedBox.closest<HTMLElement>('[data-testid="bulk-bundle"]')!).getAllByRole('button');
		await expect(edit.getAttribute('aria-describedby')).toBe(namedBox.getAttribute('aria-describedby'));
		const namedDuplicate = description(NAMED_DUPLICATE_BUNDLE)!;
		await expect(namedDuplicate.startsWith(`${description(DUPLICATE_BUNDLE)}. `)).toBe(true);
		await expect(namedDuplicate).toContain('Fury APL');
	},
};

export const CurrentOnly: Story = { args: { rows: [] } };
