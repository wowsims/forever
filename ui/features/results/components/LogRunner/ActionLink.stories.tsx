import type { Meta, StoryObj } from '@storybook/react-vite';
import { HEROIC_STRIKE, MORTAL_STRIKE } from '@testing/storybook/fixtures/results-views';

import { ActionLink } from './ActionLink';

const meta = {
	title: 'Features/Results/LogRunner/Components/ActionLink',
	component: ActionLink,
	tags: ['autodocs'],
	args: { actionId: MORTAL_STRIKE },
	argTypes: { actionId: { control: false } },
} satisfies Meta<typeof ActionLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Spell: Story = {};

export const AnotherSpell: Story = { args: { actionId: HEROIC_STRIKE } };

export const Aura: Story = {
	args: { isAura: true },
	parameters: { smoke: false, smokeReason: 'An aura link asks the spell database whether the spell has a buff' },
};
