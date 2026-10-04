import { Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withTooltipSurface } from '@testing/storybook/decorators';
import { makeStatAttribution, showStat } from '@testing/storybook/fixtures/character-stats';

import { StatTooltipBody } from './StatTooltipBody';

const meta = {
	title: 'Features/CharacterStats/Components/StatTooltipBody',
	component: StatTooltipBody,
	tags: ['autodocs'],
	args: {
		hand: UnitStat.fromStat(Stat.StatStrength),
		attribution: makeStatAttribution(),
		show: showStat,
		bonusStats: new Stats(),
		bonusStatValue: 0,
		total: '370',
	},
	argTypes: {
		hand: { control: false },
		attribution: { control: false },
		bonusStats: { control: false },
	},
	decorators: [
		Story => (
			<div className="w-56">
				<Story />
			</div>
		),
		withTooltipSurface,
	],
} satisfies Meta<typeof StatTooltipBody>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithBonus: Story = {
	args: { bonusStats: new Stats().withStat(Stat.StatStrength, 20), bonusStatValue: 20, total: '390' },
};

export const MainHand: Story = { args: { handLabel: 'Main Hand' } };

export const EffectiveWeaponSpeed: Story = { args: { rangedWeaponSpeed: 2.9 } };
