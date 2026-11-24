import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsRowsFixture } from '@testing/storybook/fixtures/bulk-drops';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DropsItemRowView } from './DropsItemRow';

const [upgrade, folded, current, stoppedEarly, loss] = dropsRowsFixture();

const meta = {
	title: 'Features/Bulk/DropsItemRow',
	component: DropsItemRowView,
	tags: ['autodocs'],
	args: upgrade,
	argTypes: { item: { control: false }, delta: { control: 'object' }, moreSources: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-full max-w-250">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsItemRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

// An item with a second source shows "+1" after its first.
export const Default: Story = {
	args: { onEquip: fn() },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'bulk_tab.results.equip_button' }));
		await expect(args.onEquip).toHaveBeenCalledOnce();
	},
};

/** The item's other variations are folded away; Show turns them on. */
export const VariationsHidden: Story = { args: folded };

/** The current gear: labelled, with no difference and no Equip. */
export const CurrentGear: Story = { args: current };

/** Dropped before the final stage: no margin, with a note. */
export const StoppedEarly: Story = { args: stoppedEarly };

export const Downgrade: Story = { args: loss };
