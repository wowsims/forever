import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean } from '@testing/storybook/fixtures/settings_reforge_bulk';

import { ReforgeFrozenSlotsView } from './ReforgeFrozenSlots';

const SLOTS = ['Head', 'Neck', 'Shoulders', 'Back', 'Chest', 'Wrist', 'Hands', 'Waist', 'Legs', 'Feet', 'Finger 1', 'Finger 2'];

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeFrozenSlots',
	component: ReforgeFrozenSlotsView,
	tags: ['autodocs'],
	args: {
		hidden: false,
		pickers: SLOTS.map((slot, index) => (
			<StoryBoolean key={slot} id={`story-freeze-${index}`} label={slot} layout="inline" initial={slot === 'Neck' || slot === 'Back'} />
		)),
	},
	argTypes: { pickers: { control: false } },
} satisfies Meta<typeof ReforgeFrozenSlotsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** Slot freezing is off: the toggles stay mounted and hidden, so the canvas is empty. */
export const Hidden: Story = { args: { hidden: true } };
