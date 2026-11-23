import { BulkEncounterRef } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { batchEncounterProps } from '@testing/storybook/fixtures/bulk-builds';
import { expect, within } from 'storybook/test';

import { BatchEncounterView } from './BatchEncounter';

const defaults = batchEncounterProps();

const meta = {
	title: 'Features/Bulk/BatchEncounter',
	component: BatchEncounterView,
	tags: ['autodocs'],
	args: defaults,
	argTypes: {
		options: { control: false },
		picked: { control: false },
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BatchEncounterView>;

export default meta;
type Story = StoryObj<typeof meta>;

// Current encounter is picked: the batch uses the Encounter tab.
export const Off: Story = {};

export const On: Story = {
	args: { picked: BulkEncounterRef.create({ name: 'Patchwerk', preset: true }), summary: '5:00 ±30s · 1 target' },
	play: async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('combobox').textContent).toBe('Patchwerk');
	},
};

export const CurrentOnly: Story = { args: { options: defaults.options.slice(0, 1) } };
