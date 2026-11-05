import type { Meta, StoryObj } from '@storybook/react-vite';
import { DESIGN_BASE_SECONDS, designTargetNames } from '@testing/storybook/fixtures/encounter_design';
import { cssVars } from '@ui-kit/utils/css';
import { fn } from 'storybook/test';

import { DesignerTargetLane } from './DesignerTargetLane';

const PPS = 5;

const meta = {
	title: 'Features/EncounterDesigner/DesignerTargetLane',
	component: DesignerTargetLane,
	tags: ['autodocs'],
	args: {
		targetNames: designTargetNames,
		baseSeconds: DESIGN_BASE_SECONDS,
		onSelect: fn(),
		onActivate: fn(),
		onAdd: fn(),
	},
	decorators: [
		Story => (
			<div className="w-timeline-content-w" style={cssVars({ '--pps': `${PPS}px`, '--label-w': '12rem', '--duration': String(DESIGN_BASE_SECONDS) })}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DesignerTargetLane>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Selected: Story = { args: { selected: 1 } };

export const OneTarget: Story = { args: { targetNames: designTargetNames.slice(0, 1) } };
