import { SpellSchool } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { damageLog, DEEP_WOUNDS } from '@testing/storybook/fixtures/results-views';

import { DamageResult } from './DamageResult';

const meta = {
	title: 'Features/Results/DamageResult',
	component: DamageResult,
	tags: ['autodocs'],
	args: { log: damageLog(12, 1420.5) },
	argTypes: { log: { control: false } },
} satisfies Meta<typeof DamageResult>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Hit: Story = {};

export const Crit: Story = { args: { log: damageLog(12, 2841, { outcome: 'crit' }) } };

export const Tick: Story = { args: { log: damageLog(12, 310.25, { actionId: DEEP_WOUNDS, tick: true }) } };

export const PartialResist: Story = { args: { log: damageLog(12, 710, { spellSchool: SpellSchool.SpellSchoolFire, resist: 50 }) } };

export const Parry: Story = { args: { log: damageLog(12, 0, { outcome: 'parry', effect: null }) } };

export const Healed: Story = { args: { log: damageLog(12, 420, { effect: 'healing' }) } };

export const Shielded: Story = { args: { log: damageLog(12, 900, { effect: 'shielding' }) } };
