import { StatCapType } from '@generated/proto/api';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { StatCap } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withTooltipSurface } from '@testing/storybook/decorators';

import { ReforgeSoftCapsTooltip } from './ReforgeSoftCapsTooltip';

const HIT = StatCap.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, {
	breakpoints: [5, 9],
	capType: StatCapType.TypeThreshold,
	postCapEPs: [0.4],
});

const HASTE = StatCap.fromStat(Stat.StatHasteRating, {
	breakpoints: [10, 15, 20],
	capType: StatCapType.TypeSoftCap,
	postCapEPs: [0.8, 0.6, 0.4],
});

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeSoftCapsTooltip',
	component: ReforgeSoftCapsTooltip,
	tags: ['autodocs'],
	args: { playerClass: Class.ClassWarrior, softCaps: [HIT, HASTE], additionalInformation: {} },
	argTypes: { softCaps: { control: false }, additionalInformation: { control: false } },
	decorators: [
		Story => (
			<div className="w-76">
				<Story />
			</div>
		),
		withTooltipSurface,
	],
} satisfies Meta<typeof ReforgeSoftCapsTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithSpecNote: Story = {
	args: { softCaps: [HASTE], additionalInformation: { [Stat.StatHasteRating]: () => 'Counts the haste from Flurry at full uptime.' } },
};
