import type { Meta, StoryObj } from '@storybook/react-vite';
import { makePresetBuild } from '@testing/storybook/fixtures/app';
import { expect, fn, userEvent, within } from 'storybook/test';

import { PresetBuildChip, PresetBuildChipTooltip } from './PresetBuildChip';

const TOOLTIP_ID = 'preset-build-chip-story';
const build = makePresetBuild('Phase 1');

const meta = {
	title: 'App/Presets/Components/PresetBuildChip',
	component: PresetBuildChip,
	tags: ['autodocs'],
	args: { build, tooltipId: TOOLTIP_ID, onApply: fn() },
	argTypes: { build: { control: false } },
	decorators: [
		(Story, { args }) => (
			<div className="ui-saved-data-presets">
				<Story />
				<PresetBuildChipTooltip id={TOOLTIP_ID} builds={[args.build]} />
			</div>
		),
	],
} satisfies Meta<typeof PresetBuildChip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Inactive: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('saved-data-set-name'));
		await expect(args.onApply).toHaveBeenCalledWith(args.build);
	},
};

export const Active: Story = { args: { active: true } };
