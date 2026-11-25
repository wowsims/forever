import { PlanBossAbility_Source } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	areaDamagePlan,
	customBossPlan,
	dynamicAddsPlan,
	movementPlan,
	overlappingDowntimesPlan,
	PLAN_BASE_SECONDS,
	planTargetNames,
	tankSwapPlan,
} from '@testing/storybook/fixtures/encounter_plan';
import { cssVars } from '@ui-kit/utils/css';
import { fn } from 'storybook/test';

import { PlannerLane } from './PlannerLane';

const PPS = 5;

const meta = {
	title: 'Features/EncounterPlanner/PlannerLane',
	component: PlannerLane,
	tags: ['autodocs'],
	args: {
		lane: 'boss',
		plan: customBossPlan(),
		baseSeconds: PLAN_BASE_SECONDS,
		snapSeconds: 1,
		targetNames: planTargetNames,
		tooltipId: 'planner-lane-tip',
		pixelsPerSecond: () => PPS,
		invalid: new Set<string>(),
		onSelect: fn(),
		onActivate: fn(),
		onChange: fn(),
		onAdd: fn(),
	},
	argTypes: { plan: { control: false }, invalid: { control: false } },
	decorators: [
		Story => (
			<div className="w-timeline-content-w" style={cssVars({ '--pps': `${PPS}px`, '--label-w': '12rem', '--duration': String(PLAN_BASE_SECONDS) })}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PlannerLane>;

export default meta;
type Story = StoryObj<typeof meta>;

const target = (index: number, selected = false) => ({ index, selected, onSelect: fn(), onActivate: fn() });

export const BossHealthKeyframes: Story = { args: { target: target(0) } };

export const OverlappingDowntimes: Story = {
	args: { plan: overlappingDowntimesPlan(), target: target(0), invalid: new Set(['downtime:1']), selected: { kind: 'downtime', index: 0 } },
};

// The second target's row: its own downtime, and no boss health to add.
export const SecondTarget: Story = { args: { plan: overlappingDowntimesPlan(), target: target(1) } };

// A tank swap on the boss's row, its repeats drawn as numbered ghosts to the fight's end.
export const TankSwaps: Story = { args: { plan: tankSwapPlan(), target: target(0) } };

export const TargetSelected: Story = { args: { target: target(0, true) } };

// The boss's row with its spell hit, strike, thrash and swing pause, and a pulse it casts, which ticks from its cast end.
const bossAbilitiesPlan = () => {
	const plan = customBossPlan();
	plan.bossAbilities.push({ ...areaDamagePlan().bossAbilities[0], source: PlanBossAbility_Source.Boss, castTime: 2 });
	return plan;
};

export const TargetAbilities: Story = { args: { plan: bossAbilitiesPlan(), target: target(0) } };

// The Environment pulse and area hit, which come from no target.
export const AreaDamage: Story = { args: { lane: 'abilities', plan: areaDamagePlan() } };

export const RepeatingWave: Story = { args: { lane: 'waves', plan: dynamicAddsPlan() } };

export const Amps: Story = { args: { lane: 'amps', plan: dynamicAddsPlan() } };

export const Moves: Story = { args: { lane: 'moves', plan: movementPlan() } };

export const EmptyLane: Story = { args: { lane: 'moves' } };
