import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { resultsHostParts, warriorRotationModel } from '@testing/storybook/fixtures/results-views';
import { type ReactNode, useState } from 'react';

import { DrStickySlotContext } from '../../DetailedResults/context/DrStickySlotContext';
import { RotationTrack } from './RotationTrack';

// The ruler and the zoom buttons portal into the detailed results' sticky slot.
const StickySlot = ({ children }: { children: ReactNode }) => {
	const [slot, setSlot] = useState<HTMLDivElement | null>(null);
	return (
		<>
			<div ref={setSlot} />
			<DrStickySlotContext.Provider value={slot}>{children}</DrStickySlotContext.Provider>
		</>
	);
};

const meta = {
	title: 'Features/Results/Timeline/Components/RotationTrack',
	component: RotationTrack,
	tags: ['autodocs'],
	args: { model: warriorRotationModel(), active: true },
	argTypes: { model: { control: false } },
	decorators: [
		Story => (
			<StickySlot>
				<Story />
			</StickySlot>
		),
		// A cast's click reads the player's rotation, and its hover tooltip the sim's display flags.
		withSimHost(resultsHostParts(null)),
	],
} satisfies Meta<typeof RotationTrack>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Warrior: Story = {};

export const NoResult: Story = { args: { model: null } };
