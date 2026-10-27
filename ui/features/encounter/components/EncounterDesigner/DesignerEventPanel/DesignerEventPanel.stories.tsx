import { DamageAmpType, DesignBossAbility_Kind, type EncounterDesign } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
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

import type { DesignEventRef } from '../../../model/design/lanes';
import { DesignerEventPanel, type DesignerEventPanelProps } from './DesignerEventPanel';

// Holds the design so an edit shows in the fields it came from.
const EditablePanel = ({ design: initial, selected: initialSelected, onChange, onSelect, ...props }: DesignerEventPanelProps) => {
	const [design, setDesign] = useState<EncounterDesign>(initial);
	const [selected, setSelected] = useState<DesignEventRef | undefined>(initialSelected);
	return (
		<DesignerEventPanel
			{...props}
			design={design}
			selected={selected}
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
	title: 'Features/Encounter/EncounterDesigner/DesignerEventPanel',
	component: DesignerEventPanel,
	tags: ['autodocs'],
	args: {
		design: emptyDesign(),
		baseSeconds: DESIGN_BASE_SECONDS,
		targetNames: designTargetNames,
		presetTargets: designPresetTargets,
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
