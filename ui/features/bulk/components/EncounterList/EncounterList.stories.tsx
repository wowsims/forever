import type { Meta, StoryObj } from '@storybook/react-vite';
import { encounterListProps, sameAsCurrentEncounterRow } from '@testing/storybook/fixtures/bulk-builds';
import { expect, within } from 'storybook/test';

import { EncounterListView } from './EncounterList';

const meta = {
	title: 'Features/Bulk/EncounterList',
	component: EncounterListView,
	tags: ['autodocs'],
	args: encounterListProps(),
	argTypes: {
		current: { control: false },
		presets: { control: false },
		saved: { control: false },
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof EncounterListView>;

export default meta;
type Story = StoryObj<typeof meta>;

// "Patchwerk copy" is stored ticked but sims what "Patchwerk" does, so it shows unticked and cannot be ticked.
export const Default: Story = {
	play: async ({ canvasElement }) => {
		const box = within(canvasElement).getByRole('checkbox', { name: 'Patchwerk copy' });
		await expect(box.getAttribute('aria-checked')).toBe('false');
		await expect(box.getAttribute('aria-disabled')).toBe('true');
	},
};

const defaults = encounterListProps();

export const CurrentUnticked: Story = { args: { current: { ...defaults.current, checked: false } } };

// "Three minutes" sims what the current encounter does, so it shows unticked and cannot be ticked while that is ticked.
export const SameAsCurrentEncounter: Story = {
	args: { saved: [sameAsCurrentEncounterRow(true), ...defaults.saved] },
	play: async ({ canvasElement }) => {
		const box = within(canvasElement).getByRole('checkbox', { name: 'Three minutes' });
		await expect(box.getAttribute('aria-disabled')).toBe('true');
	},
};

// With the current encounter unticked, "Three minutes" runs under its own name.
export const SameAsCurrentEncounterUnticked: Story = {
	args: { current: { ...defaults.current, checked: false }, saved: [sameAsCurrentEncounterRow(false), ...defaults.saved] },
};

export const Empty: Story = { args: { presets: [], saved: [] } };

export const Narrow: Story = {
	decorators: [
		Story => (
			<div className="w-60">
				<Story />
			</div>
		),
	],
};
