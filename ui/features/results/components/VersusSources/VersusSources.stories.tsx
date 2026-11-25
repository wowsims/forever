import { CreatureFamily } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, UNDEAD_VERSUS } from '@testing/storybook/fixtures/results-views';

import { VersusSources } from './VersusSources';

const meta = {
	title: 'Features/Results/VersusSources',
	component: VersusSources,
	tags: ['autodocs'],
	args: { logs: [damageLog(12, 1420.5, { versus: UNDEAD_VERSUS })] },
	argTypes: { logs: { control: false } },
} satisfies Meta<typeof VersusSources>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Undead: Story = {};

export const Family: Story = {
	args: { logs: [damageLog(12, 1420.5, { versus: { ...UNDEAD_VERSUS, creatureFamily: CreatureFamily.CreatureFamilyWolf } })] },
};
