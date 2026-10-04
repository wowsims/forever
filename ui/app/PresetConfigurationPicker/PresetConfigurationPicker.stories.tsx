import type { Meta, StoryObj } from '@storybook/react-vite';
import { makePhasedPresetBuilds, makePresetBuild } from '@testing/storybook/fixtures/app';
import { expect, fn, userEvent, within } from 'storybook/test';

import { PresetConfigurationPickerView } from './PresetConfigurationPicker';

const meta = {
	title: 'App/Presets/PresetConfigurationPicker',
	component: PresetConfigurationPickerView,
	tags: ['autodocs'],
	args: {
		builds: [makePresetBuild('Pre-Raid'), makePresetBuild('Phase 1'), makePresetBuild('Phase 2')],
		active: [false, true, false],
		ready: true,
		phases: [],
		phase: 1,
		onSelectPhase: fn(),
		onApply: fn(),
	},
	argTypes: { builds: { control: false } },
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PresetConfigurationPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Flat: Story = {
	play: async ({ args, canvasElement }) => {
		const names = within(canvasElement).getAllByTestId('saved-data-set-name');
		await userEvent.click(names[0]);
		await expect(args.onApply).toHaveBeenCalledWith(args.builds[0]);
	},
};

export const Loading: Story = { args: { ready: false, active: null } };

export const GroupedByPhase: Story = {
	args: { builds: makePhasedPresetBuilds(), active: [true, false, false], phases: [1, 2], phase: 1 },
};
