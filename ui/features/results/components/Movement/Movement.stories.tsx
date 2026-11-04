import type { Meta, StoryObj } from '@storybook/react-vite';

import type { Leg, MovementRun } from '../../model/movement/movement';
import { MovementView } from './Movement';

// A melee weave: out to 25 yd and back in the prepull, then 8 yd trips between range and melee every
// six seconds, at 7 yd per second. The cat runs in at the pull and stays in melee range.
const walk = (start: number, from: number, to: number): Leg => ({ start, end: start + Math.abs(to - from) / 7, from, to });

const weaveLegs: Array<Leg> = [walk(-19, 12, 25), walk(-2, 25, 12)];
for (let at = 2; at < 56; at += 6) weaveLegs.push(walk(at, 12, 4), walk(at + 2.5, 4, 12));

const weave: MovementRun = {
	start: -19,
	end: 60,
	reach: 25,
	deadZone: true,
	tracks: [
		{ name: 'Player', isPet: false, startYards: 12, legs: weaveLegs },
		{ name: 'Cat', isPet: true, startYards: 12, legs: [walk(0, 12, 4)] },
	],
};

const meta = {
	title: 'Features/Results/Movement/Movement',
	component: MovementView,
	tags: ['autodocs'],
	args: { run: weave, targetName: 'Target 1', active: true },
	argTypes: { run: { control: false } },
} satisfies Meta<typeof MovementView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Weave: Story = {};

export const NoWalks: Story = { args: { run: null } };
