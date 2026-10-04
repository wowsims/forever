import { Stats } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withTooltipSurface } from '@testing/storybook/decorators';
import { makePresetBuild } from '@testing/storybook/fixtures/app';

import { PresetBuildTooltip } from './PresetBuildTooltip';

const meta = {
	title: 'App/Presets/Components/PresetBuildTooltip',
	component: PresetBuildTooltip,
	tags: ['autodocs'],
	args: { build: makePresetBuild('Phase 1', { epWeights: { name: 'Phase 1', epWeights: new Stats() } }) },
	argTypes: { build: { control: false } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
		withTooltipSurface,
	],
} satisfies Meta<typeof PresetBuildTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
