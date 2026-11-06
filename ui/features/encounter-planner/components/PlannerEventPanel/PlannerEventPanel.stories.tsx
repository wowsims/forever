import { DamageAmpType, type EncounterPlan, PlanBossAbility_Kind, Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStoryEncounter } from '@testing/storybook/fixtures/encounter';
import {
	areaDamagePlan,
	customBossPlan,
	dynamicAddsPlan,
	emptyPlan,
	movementPlan,
	overlappingDowntimesPlan,
	PLAN_BASE_SECONDS,
	planPresetTargets,
	planTargetNames,
} from '@testing/storybook/fixtures/encounter_plan';
import { useState } from 'react';
import { fn } from 'storybook/test';

import type { PlannerSelection } from '../../model/lanes';
import { validatePlan } from '../../model/validate';
import { encounterTargetContext, npcConfig, presetTargetConfig } from '../PlannerAddStats/utils/configs';
import { PlannerEventPanel, type PlannerEventPanelProps, type PlannerTargetEditor } from './PlannerEventPanel';

// Holds the plan and the target so an edit shows in the fields it came from.
const EditablePanel = ({ plan: initial, selected: initialSelected, targetEditor, onChange, onSelect, ...props }: PlannerEventPanelProps) => {
	const [plan, setPlan] = useState<EncounterPlan>(initial);
	const [selected, setSelected] = useState<PlannerSelection | undefined>(initialSelected);
	const [target, setTarget] = useState(targetEditor?.target);
	return (
		<PlannerEventPanel
			{...props}
			plan={plan}
			selected={selected}
			issues={validatePlan(plan, props.targetNames.length)}
			targetEditor={
				targetEditor &&
				target && {
					...targetEditor,
					target,
					onModify: edit => {
						targetEditor.onModify(edit);
						const next = Target.clone(target);
						edit(next);
						setTarget(next);
					},
				}
			}
			onChange={next => {
				onChange(next);
				setPlan(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
		/>
	);
};

const meta = {
	title: 'Features/EncounterPlanner/PlannerEventPanel',
	component: PlannerEventPanel,
	tags: ['autodocs'],
	args: {
		plan: emptyPlan(),
		issues: [],
		baseSeconds: PLAN_BASE_SECONDS,
		targetNames: planTargetNames,
		presetTargets: planPresetTargets,
		newAdd: () => Target.clone(planPresetTargets[0].target!),
		onChange: fn(),
		onSelect: fn(),
	},
	argTypes: { plan: { control: false }, issues: { control: false }, presetTargets: { control: false } },
	render: args => <EditablePanel {...args} />,
} satisfies Meta<typeof PlannerEventPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const withAbilityKind = (kind: PlanBossAbility_Kind) => {
	const plan = customBossPlan();
	plan.bossAbilities[0].kind = kind;
	return plan;
};

const withAmpType = (type: DamageAmpType) => {
	const plan = dynamicAddsPlan();
	plan.damageAmps[0].type = type;
	return plan;
};

export const NothingSelected: Story = {};

export const Downtime: Story = { args: { plan: overlappingDowntimesPlan(), selected: { kind: 'downtime', index: 0 } } };

export const BossHealth: Story = { args: { plan: customBossPlan(), selected: { kind: 'bossHealth', index: 1 } } };

export const Wave: Story = { args: { plan: dynamicAddsPlan(), selected: { kind: 'wave', index: 0 } } };

export const Move: Story = { args: { plan: movementPlan(), selected: { kind: 'move', index: 0 } } };

export const TargetDebuffAmp: Story = { args: { plan: dynamicAddsPlan(), selected: { kind: 'amp', index: 0 } } };

export const CasterBuffAmp: Story = { args: { plan: withAmpType(DamageAmpType.DamageAmpTypeCasterBuff), selected: { kind: 'amp', index: 0 } } };

export const SpellHit: Story = { args: { plan: customBossPlan(), selected: { kind: 'ability', index: 0 } } };

export const Strike: Story = { args: { plan: withAbilityKind(PlanBossAbility_Kind.Strike), selected: { kind: 'ability', index: 0 } } };

export const Thrash: Story = { args: { plan: withAbilityKind(PlanBossAbility_Kind.Thrash), selected: { kind: 'ability', index: 0 } } };

export const SwingTiming: Story = { args: { plan: withAbilityKind(PlanBossAbility_Kind.SwingTiming), selected: { kind: 'ability', index: 0 } } };

export const Pulse: Story = { args: { plan: areaDamagePlan(), selected: { kind: 'ability', index: 0 } } };

export const AreaHit: Story = { args: { plan: areaDamagePlan(), selected: { kind: 'ability', index: 1 } } };

const storyTarget = Target.clone(planPresetTargets[0].target!);
const storyTargetContext = encounterTargetContext(makeStoryEncounter([Target.create({ name: 'Ragnaros' }), storyTarget]), 1, 'target-1');
const targetEditor: PlannerTargetEditor = {
	target: storyTarget,
	idPrefix: storyTargetContext.idPrefix,
	identity: [npcConfig(storyTargetContext), presetTargetConfig(storyTargetContext)],
	threatMetrics: true,
	canRemove: true,
	onModify: fn(),
	onMove: fn(),
	onCopy: fn(),
	onRemove: fn(),
};

// The second encounter target: its NPC, defences and swing, Move up, Move down, Duplicate and Remove.
export const EncounterTarget: Story = {
	args: { selected: { kind: 'target', index: 1 }, targetEditor },
	argTypes: { targetEditor: { control: false } },
};

// With threat metrics off, the panel leaves out attack power, block value, the tank, the swing and the spell school.
export const EncounterTargetWithoutThreatMetrics: Story = {
	args: { selected: { kind: 'target', index: 1 }, targetEditor: { ...targetEditor, threatMetrics: false } },
	argTypes: { targetEditor: { control: false } },
};

// The first target has no Remove and no Move up.
export const FirstEncounterTarget: Story = {
	args: { selected: { kind: 'target', index: 0 }, targetEditor: { ...targetEditor, canRemove: false } },
	argTypes: { targetEditor: { control: false } },
};
