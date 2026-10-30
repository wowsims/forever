import type { Meta, StoryObj } from '@storybook/react-vite';
import { DESIGN_BASE_SECONDS } from '@testing/storybook/fixtures/encounter_design';
import { cssVars } from '@ui-kit/utils/css';
import { fn } from 'storybook/test';

import { DesignerEventBar } from './DesignerEventBar';

const PPS = 5;

const meta = {
	title: 'Features/EncounterDesigner/DesignerEventBar',
	component: DesignerEventBar,
	tags: ['autodocs'],
	args: {
		name: 'Wave 1',
		window: { start: 20 / 180, end: 60 / 180 },
		row: 0,
		baseSeconds: DESIGN_BASE_SECONDS,
		snapSeconds: 1,
		pixelsPerSecond: () => PPS,
		valueText: () => 'Wave 1, 20s to 60s',
		tabIndex: 0,
		onChange: fn(),
		onSelect: fn(),
		onActivate: fn(),
		onRemove: fn(),
		onDuplicate: fn(),
		onRenameStart: fn(),
		onRename: fn(),
		onSibling: fn(),
	},
	decorators: [
		Story => (
			<div
				className="relative h-(--designer-lane-height) w-timeline-total"
				style={cssVars({ '--pps': `${PPS}px`, '--duration': String(DESIGN_BASE_SECONDS), '--bar': 'var(--designer-lane-waves-bar)' })}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DesignerEventBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Span: Story = {};

export const Selected: Story = { args: { selected: true } };

export const Invalid: Story = { args: { invalid: true } };

export const SelectedInvalid: Story = { args: { selected: true, invalid: true } };

export const Jitter: Story = { args: { jitter: { start: 4 / 180, end: 6 / 180 } } };

export const Short: Story = { args: { window: { start: 90 / 180, end: 90.5 / 180 } } };

export const HealthSegment: Story = {
	args: {
		name: 'Boss health 1',
		label: '100 % → 65 %',
		startFixed: true,
		window: { start: 0, end: 0.5 },
		valueText: () => 'Boss health 1, 100 % to 65 % from 0s to 90s',
	},
};

export const Renaming: Story = { args: { renaming: true } };
