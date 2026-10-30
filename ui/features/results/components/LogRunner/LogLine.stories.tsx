import { ResourceType } from '@generated/proto/spell';
import type { AuraLog, CastBeganLog, CastFailedLog, CombatLog, PlainLog, StatChangeLog } from '@sim/proto/combat_log';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { BATTLE_SHOUT, damageLog, DEEP_WOUNDS, MORTAL_STRIKE, PET, resourceLog, TARGET } from '@testing/storybook/fixtures/results-views';

import { LogLine } from './LogLine';

const base = damageLog(12.5, 0);

const meta = {
	title: 'Features/Results/LogRunner/Components/LogLine',
	component: LogLine,
	tags: ['autodocs'],
	args: { log: damageLog(12.5, 1420.5) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="whitespace-nowrap">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof LogLine>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Hit: Story = {};

export const Crit: Story = { args: { log: damageLog(14, 2980.75, { outcome: 'crit' }) } };

export const Tick: Story = { args: { log: damageLog(15, 310.25, { actionId: DEEP_WOUNDS, tick: true }) } };

export const Dodge: Story = { args: { log: damageLog(16, 0, { outcome: 'dodge', effect: null }) } };

export const DealtByTheTarget: Story = { args: { log: damageLog(17, 850, { actionId: MORTAL_STRIKE, source: TARGET, target: null }) } };

export const DealtByAPet: Story = { args: { log: damageLog(18, 120, { source: PET }) } };

export const RageSpent: Story = { args: { log: resourceLog(12.5, 60, 30, { actionId: MORTAL_STRIKE }) } };

export const HealthLost: Story = { args: { log: resourceLog(20, 4200, 3350, { resourceType: ResourceType.ResourceTypeHealth, total: 4200 }) } };

export const CastBegan: Story = {
	args: {
		log: { ...base, kind: 'cast-began', manaCost: 0, castTime: 0, gcd: 1.5, effectiveTime: 1.5, aplEntryPath: null } satisfies CastBeganLog as CombatLog,
	},
};

export const CastFailed: Story = {
	args: { log: { ...base, kind: 'cast-failed', reason: 'not enough rage' } satisfies CastFailedLog as CombatLog },
};

export const StatGained: Story = {
	args: { log: { ...base, actionId: BATTLE_SHOUT, kind: 'stat-change', isGain: true, stats: '{Attack Power: 232}' } satisfies StatChangeLog as CombatLog },
};

export const Plain: Story = {
	args: { log: { ...base, kind: 'plain', raw: '[12.50] [Player (#1)] Rotation step: queue Heroic Strike' } satisfies PlainLog as CombatLog },
};

export const AuraGained: Story = {
	args: { log: { ...base, actionId: BATTLE_SHOUT, kind: 'aura', isGained: true, isFaded: false, isRefreshed: false } satisfies AuraLog as CombatLog },
	parameters: { smoke: false, smokeReason: 'An aura link asks the spell database whether the spell has a buff' },
};
