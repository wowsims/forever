import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { AplListToolbarView } from './AplListToolbar';

const meta = {
	title: 'Features/Apl/AplListToolbar',
	component: AplListToolbarView,
	tags: ['autodocs'],
	args: { itemName: 'Action', onCreate: fn(), onReset: fn() },
} satisfies Meta<typeof AplListToolbarView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The priority list's bar: the new button appends an unnamed row at once, and the paste button appends the row the clipboard holds. */
export const PriorityListBar: Story = { args: { onPaste: fn() } };

/** The groups' and variables' bar: the new button asks for a name, and a taken one blocks the confirm button. */
export const NamedItemBar: Story = {
	args: {
		itemName: 'Group',
		nameDialog: { inputLabel: 'Name', existingNames: () => ['Execute', 'Opener'] },
	},
};
