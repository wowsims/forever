import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';

import { BonusStatsLink } from './BonusStatsLink';

const meta = {
	title: 'Features/CharacterStats/Components/BonusStatsLink',
	component: BonusStatsLink,
	tags: ['autodocs'],
	args: { unitStat: UnitStat.fromStat(Stat.StatStrength) },
	argTypes: { unitStat: { control: false } },
	decorators: [
		withSimHost({
			player: {
				getBonusStats: () => new Stats().withStat(Stat.StatStrength, 20).withStat(Stat.StatHitRating, 20),
				setBonusStats: () => {},
			},
		}),
	],
} satisfies Meta<typeof BonusStatsLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const HitRating: Story = {
	args: { unitStat: UnitStat.fromStat(Stat.StatHitRating) },
};
