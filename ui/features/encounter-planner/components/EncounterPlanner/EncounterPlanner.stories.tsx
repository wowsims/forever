import { EncounterPlan, Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	customBossPlan,
	dynamicAddsPlan,
	emptyPlan,
	overlappingDowntimesPlan,
	PLAN_BASE_SECONDS,
	planPresetTargets,
	planTargetNames,
} from '@testing/storybook/fixtures/encounter_plan';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { NumberField, StringField } from '@ui-kit/TextField';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { PlannerSelection } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { validatePlan } from '../../model/validate';
import { EncounterPlannerView, type EncounterPlannerViewProps } from './EncounterPlanner';

// Holds the draft and the selection the way the container does, so a drag, an edit or an add lands in the story.
const EditablePlanner = ({ plan: initial, selected: initialSelected, onChange, onSelect, onAdd, ...props }: EncounterPlannerViewProps) => {
	const [plan, setPlan] = useState<EncounterPlan>(initial);
	const [selected, setSelected] = useState<PlannerSelection | undefined>(initialSelected);
	return (
		<EncounterPlannerView
			{...props}
			plan={plan}
			selected={selected}
			issues={validatePlan(plan, planTargetNames.length)}
			onChange={next => {
				onChange(next);
				setPlan(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
			onAdd={(kind, targetIndex) => {
				onAdd(kind, targetIndex);
				const next = EncounterPlan.clone(plan);
				setSelected(addEvent(next, kind, targetIndex));
				setPlan(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/EncounterPlanner/EncounterPlanner',
	component: EncounterPlannerView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		name: <StringField id="story-planner-name" label="Name" value="" onChange={fn()} />,
		duration: (
			<>
				<NumberField id="story-planner-duration" label="Duration" value={PLAN_BASE_SECONDS} onChange={fn()} />
				<NumberField id="story-planner-duration-variation" label="Duration +/-" value={5} onChange={fn()} />
			</>
		),
		preset: (
			<ChoiceField
				id="story-planner-preset"
				label="Encounter"
				options={[
					{ value: 'custom', label: 'Custom' },
					{ label: 'Presets', options: [{ value: 'preset:0', label: 'Default/Dynamic Adds' }] },
				]}
				value="preset:0"
				onChange={fn()}
			/>
		),
		plan: emptyPlan(),
		baseSeconds: PLAN_BASE_SECONDS,
		issues: [],
		targetNames: planTargetNames,
		presetTargets: planPresetTargets,
		newAdd: () => Target.clone(planPresetTargets[0].target!),
		onSelect: fn(),
		onActivate: fn(),
		onChange: fn(),
		onAdd: fn(),
		onAddTarget: fn(),
		onSave: fn(),
	},
	argTypes: {
		name: { control: false },
		duration: { control: false },
		preset: { control: false },
		execute: { control: false },
		targetEditor: { control: false },
		plan: { control: false },
		issues: { control: false },
		presetTargets: { control: false },
	},
	render: args => <EditablePlanner {...args} />,
} satisfies Meta<typeof EncounterPlannerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const DynamicAdds: Story = { args: { plan: dynamicAddsPlan(), selected: { kind: 'wave', index: 0 } } };

export const CustomBoss: Story = { args: { plan: customBossPlan(), selected: { kind: 'ability', index: 0 } } };

export const WithIssues: Story = { args: { plan: overlappingDowntimesPlan(), selected: { kind: 'downtime', index: 1 } } };

// A spec that reads the execute proportions gets them in the header.
export const WithExecuteBands: Story = {
	args: {
		execute: (
			<>
				<NumberField id="story-planner-execute-20" label="Execute Duration (20%)" value={20} onChange={fn()} />
				<NumberField id="story-planner-execute-35" label="Execute Duration (35%)" value={35} onChange={fn()} />
			</>
		),
	},
};

// The second target picked on the targets lane, with its fields in the panel.
export const TargetSelected: Story = {
	args: {
		plan: customBossPlan(),
		selected: { kind: 'target', index: 1 },
		targetEditor: {
			target: Target.clone(planPresetTargets[0].target!),
			idPrefix: 'encounter-planner-target-1',
			identity: [],
			threatMetrics: true,
			canRemove: true,
			onModify: fn(),
			onMove: fn(),
			onCopy: fn(),
			onRemove: fn(),
		},
	},
};

// Opened on an entry of My encounters: Save as copy and Delete join Save, and Delete asks first.
export const OwnEncounter: Story = {
	args: {
		name: <StringField id="story-planner-name" label="Name" value="Two adds" onChange={fn()} />,
		plan: dynamicAddsPlan(),
		onSaveCopy: fn(),
		onDelete: fn(),
		deleteName: 'Two adds',
	},
	play: async ({ args }) => {
		// The label ends in "delete" with or without the English strings loaded.
		const deleteLabel = /delete$/i;
		const body = within(document.body);
		await userEvent.click(body.getByRole('button', { name: deleteLabel }));
		const confirm = await body.findByTestId('encounter-planner-delete-confirm');
		await userEvent.click(within(confirm).getByRole('button', { name: deleteLabel }));
		await expect(args.onDelete).toHaveBeenCalledTimes(1);
	},
};
