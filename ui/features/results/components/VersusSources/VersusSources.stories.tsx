import { CreatureFamily, MobType } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, SLAYER_RING, UNDEAD_VERSUS } from '@testing/storybook/fixtures/results-views';

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

export const CritChanceAndProcs: Story = {
	args: {
		logs: [
			damageLog(12, 1420.5, {
				versus: {
					mobType: MobType.MobTypeHumanoid,
					creatureFamily: CreatureFamily.CreatureFamilyOrc,
					sources: [
						{ actionId: SLAYER_RING, effect: { kind: 'crit_chance', amount: 10 } },
						{ actionId: SLAYER_RING, effect: { kind: 'proc_chance', multiplier: 2 } },
						{ actionId: SLAYER_RING, effect: { kind: 'proc_only' } },
					],
				},
			}),
		],
	},
};
