import { EncounterDesign, Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	customBossDesign,
	DESIGN_BASE_SECONDS,
	designPresetTargets,
	designTargetNames,
	dynamicAddsDesign,
	emptyDesign,
	overlappingDowntimesDesign,
} from '@testing/storybook/fixtures/encounter_design';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { NumberField, StringField } from '@ui-kit/TextField';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { DesignerSelection } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { validateDesign } from '../../model/validate';
import { EncounterDesignerView, type EncounterDesignerViewProps } from './EncounterDesigner';

// Holds the draft and the selection the way the container does, so a drag, an edit or an add lands in the story.
const EditableDesigner = ({ design: initial, selected: initialSelected, onChange, onSelect, onAdd, ...props }: EncounterDesignerViewProps) => {
	const [design, setDesign] = useState<EncounterDesign>(initial);
	const [selected, setSelected] = useState<DesignerSelection | undefined>(initialSelected);
	return (
		<EncounterDesignerView
			{...props}
			design={design}
			selected={selected}
			issues={validateDesign(design, designTargetNames.length)}
			onChange={next => {
				onChange(next);
				setDesign(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
			onAdd={kind => {
				onAdd(kind);
				const next = EncounterDesign.clone(design);
				setSelected(addEvent(next, kind));
				setDesign(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/EncounterDesigner/EncounterDesigner',
	component: EncounterDesignerView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		name: <StringField id="story-designer-name" label="Name" value="" onChange={fn()} />,
		duration: (
			<>
				<NumberField id="story-designer-duration" label="Duration" value={DESIGN_BASE_SECONDS} onChange={fn()} />
				<NumberField id="story-designer-duration-variation" label="Duration +/-" value={5} onChange={fn()} />
			</>
		),
		preset: (
			<ChoiceField
				id="story-designer-preset"
				label="Encounter"
				options={[
					{ value: 'custom', label: 'Custom' },
					{ label: 'Presets', options: [{ value: 'preset:0', label: 'Default/Dynamic Adds' }] },
				]}
				value="preset:0"
				onChange={fn()}
			/>
		),
		design: emptyDesign(),
		baseSeconds: DESIGN_BASE_SECONDS,
		issues: [],
		targetNames: designTargetNames,
		presetTargets: designPresetTargets,
		newAdd: () => Target.clone(designPresetTargets[0].target!),
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
		design: { control: false },
		issues: { control: false },
		presetTargets: { control: false },
	},
	render: args => <EditableDesigner {...args} />,
} satisfies Meta<typeof EncounterDesignerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const DynamicAdds: Story = { args: { design: dynamicAddsDesign(), selected: { kind: 'wave', index: 0 } } };

export const CustomBoss: Story = { args: { design: customBossDesign(), selected: { kind: 'ability', index: 0 } } };

export const WithIssues: Story = { args: { design: overlappingDowntimesDesign(), selected: { kind: 'downtime', index: 1 } } };

// A spec that reads the execute proportions gets them in the header.
export const WithExecuteBands: Story = {
	args: {
		execute: (
			<>
				<NumberField id="story-designer-execute-20" label="Execute Duration (20%)" value={20} onChange={fn()} />
				<NumberField id="story-designer-execute-35" label="Execute Duration (35%)" value={35} onChange={fn()} />
			</>
		),
	},
};

// The second target picked on the targets lane, with its fields in the panel.
export const TargetSelected: Story = {
	args: {
		design: customBossDesign(),
		selected: { kind: 'target', index: 1 },
		targetEditor: {
			target: Target.clone(designPresetTargets[0].target!),
			idPrefix: 'encounter-designer-target-1',
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
		name: <StringField id="story-designer-name" label="Name" value="Two adds" onChange={fn()} />,
		design: dynamicAddsDesign(),
		onSaveCopy: fn(),
		onDelete: fn(),
		deleteName: 'Two adds',
	},
	play: async ({ args }) => {
		// The label ends in "delete" with or without the English strings loaded.
		const deleteLabel = /delete$/i;
		const body = within(document.body);
		await userEvent.click(body.getByRole('button', { name: deleteLabel }));
		const confirm = await body.findByTestId('encounter-designer-delete-confirm');
		await userEvent.click(within(confirm).getByRole('button', { name: deleteLabel }));
		await expect(args.onDelete).toHaveBeenCalledTimes(1);
	},
};
