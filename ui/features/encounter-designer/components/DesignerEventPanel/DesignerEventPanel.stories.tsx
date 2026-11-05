import { DamageAmpType, DesignBossAbility_Kind, type EncounterDesign, Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStoryEncounter } from '@testing/storybook/fixtures/encounter';
import {
	areaDamageDesign,
	customBossDesign,
	DESIGN_BASE_SECONDS,
	designPresetTargets,
	designTargetNames,
	dynamicAddsDesign,
	emptyDesign,
	movementDesign,
	overlappingDowntimesDesign,
} from '@testing/storybook/fixtures/encounter_design';
import { useState } from 'react';
import { fn } from 'storybook/test';

import type { DesignerSelection } from '../../model/lanes';
import { encounterTargetContext, npcConfig, presetTargetConfig } from '../DesignerAddStats/utils/configs';
import { DesignerEventPanel, type DesignerEventPanelProps, type DesignerTargetEditor } from './DesignerEventPanel';

// Holds the design and the target so an edit shows in the fields it came from.
const EditablePanel = ({ design: initial, selected: initialSelected, targetEditor, onChange, onSelect, ...props }: DesignerEventPanelProps) => {
	const [design, setDesign] = useState<EncounterDesign>(initial);
	const [selected, setSelected] = useState<DesignerSelection | undefined>(initialSelected);
	const [target, setTarget] = useState(targetEditor?.target);
	return (
		<DesignerEventPanel
			{...props}
			design={design}
			selected={selected}
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
				setDesign(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
		/>
	);
};

const meta = {
	title: 'Features/EncounterDesigner/DesignerEventPanel',
	component: DesignerEventPanel,
	tags: ['autodocs'],
	args: {
		design: emptyDesign(),
		baseSeconds: DESIGN_BASE_SECONDS,
		targetNames: designTargetNames,
		presetTargets: designPresetTargets,
		newAdd: () => Target.clone(designPresetTargets[0].target!),
		onChange: fn(),
		onSelect: fn(),
	},
	argTypes: { design: { control: false }, presetTargets: { control: false } },
	render: args => <EditablePanel {...args} />,
} satisfies Meta<typeof DesignerEventPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

const withAbilityKind = (kind: DesignBossAbility_Kind) => {
	const design = customBossDesign();
	design.bossAbilities[0].kind = kind;
	return design;
};

const withAmpType = (type: DamageAmpType) => {
	const design = dynamicAddsDesign();
	design.damageAmps[0].type = type;
	return design;
};

export const NothingSelected: Story = {};

export const Downtime: Story = { args: { design: overlappingDowntimesDesign(), selected: { kind: 'downtime', index: 0 } } };

export const BossHealth: Story = { args: { design: customBossDesign(), selected: { kind: 'bossHealth', index: 1 } } };

export const Wave: Story = { args: { design: dynamicAddsDesign(), selected: { kind: 'wave', index: 0 } } };

export const Move: Story = { args: { design: movementDesign(), selected: { kind: 'move', index: 0 } } };

export const TargetDebuffAmp: Story = { args: { design: dynamicAddsDesign(), selected: { kind: 'amp', index: 0 } } };

export const CasterBuffAmp: Story = { args: { design: withAmpType(DamageAmpType.DamageAmpTypeCasterBuff), selected: { kind: 'amp', index: 0 } } };

export const SpellHit: Story = { args: { design: customBossDesign(), selected: { kind: 'ability', index: 0 } } };

export const Strike: Story = { args: { design: withAbilityKind(DesignBossAbility_Kind.Strike), selected: { kind: 'ability', index: 0 } } };

export const Thrash: Story = { args: { design: withAbilityKind(DesignBossAbility_Kind.Thrash), selected: { kind: 'ability', index: 0 } } };

export const SwingTiming: Story = { args: { design: withAbilityKind(DesignBossAbility_Kind.SwingTiming), selected: { kind: 'ability', index: 0 } } };

export const Pulse: Story = { args: { design: areaDamageDesign(), selected: { kind: 'ability', index: 0 } } };

export const AreaHit: Story = { args: { design: areaDamageDesign(), selected: { kind: 'ability', index: 1 } } };

const storyTarget = Target.clone(designPresetTargets[0].target!);
const storyTargetContext = encounterTargetContext(makeStoryEncounter([Target.create({ name: 'Ragnaros' }), storyTarget]), 1);
const targetEditor: DesignerTargetEditor = {
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
