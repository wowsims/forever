import { CreatureFamily, MobType } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { UNDEAD_VERSUS } from '@testing/storybook/fixtures/results-views';

import { VersusTag } from './VersusTag';

const meta = {
	title: 'Features/Results/VersusTag',
	component: VersusTag,
	tags: ['autodocs'],
	args: { versus: UNDEAD_VERSUS },
	argTypes: { versus: { control: false } },
} satisfies Meta<typeof VersusTag>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MobTypeOnly: Story = {};

export const WithFamily: Story = {
	args: { versus: { ...UNDEAD_VERSUS, mobType: MobType.MobTypeHumanoid, creatureFamily: CreatureFamily.CreatureFamilyOrc } },
};
