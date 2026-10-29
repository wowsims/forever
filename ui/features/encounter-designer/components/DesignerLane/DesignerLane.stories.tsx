import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	customBossDesign,
	DESIGN_BASE_SECONDS,
	designTargetNames,
	dynamicAddsDesign,
	movementDesign,
	overlappingDowntimesDesign,
} from '@testing/storybook/fixtures/encounter_design';
import { cssVars } from '@ui-kit/utils/css';
import { fn } from 'storybook/test';

import { DesignerLane } from './DesignerLane';

const PPS = 5;

const meta = {
	title: 'Features/EncounterDesigner/DesignerLane',
	component: DesignerLane,
	tags: ['autodocs'],
	args: {
		lane: 'boss',
		design: customBossDesign(),
		baseSeconds: DESIGN_BASE_SECONDS,
		snapSeconds: 1,
		targetNames: designTargetNames,
		tooltipId: 'designer-lane-tip',
		pixelsPerSecond: () => PPS,
		invalid: new Set<string>(),
		onSelect: fn(),
		onActivate: fn(),
		onChange: fn(),
		onAdd: fn(),
	},
	argTypes: { design: { control: false }, invalid: { control: false } },
	decorators: [
		Story => (
			<div className="w-timeline-content-w" style={cssVars({ '--pps': `${PPS}px`, '--label-w': '12rem', '--duration': String(DESIGN_BASE_SECONDS) })}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DesignerLane>;

export default meta;
type Story = StoryObj<typeof meta>;

export const BossHealthKeyframes: Story = {};

export const OverlappingDowntimes: Story = {
	args: { design: overlappingDowntimesDesign(), invalid: new Set(['downtime:1']), selected: { kind: 'downtime', index: 0 } },
};

export const Abilities: Story = { args: { lane: 'abilities' } };

export const RepeatingWave: Story = { args: { lane: 'waves', design: dynamicAddsDesign() } };

export const Amps: Story = { args: { lane: 'amps', design: dynamicAddsDesign() } };

export const Moves: Story = { args: { lane: 'moves', design: movementDesign() } };

export const EmptyLane: Story = { args: { lane: 'moves' } };
