import type { Meta, StoryObj } from '@storybook/react-vite';
import { FLURRY, MORTAL_STRIKE } from '@testing/storybook/fixtures/results-views';

import { RotationRowIcon } from './RotationRowIcon';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationRowIcon',
	component: RotationRowIcon,
	tags: ['autodocs'],
	args: { actionId: MORTAL_STRIKE, tooltip: 'spell' },
	argTypes: { actionId: { control: false } },
} satisfies Meta<typeof RotationRowIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Spell: Story = {};

export const SectionHeader: Story = { args: { tooltip: undefined } };

export const BuffAura: Story = {
	args: { actionId: FLURRY, tooltip: 'buffAura' },
	parameters: { smoke: false, smokeReason: 'An aura tooltip asks the spell database whether the spell has a buff' },
};
