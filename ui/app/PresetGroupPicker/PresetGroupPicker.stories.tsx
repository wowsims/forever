import type { Meta, StoryObj } from '@storybook/react-vite';
import { Chip } from '@ui-kit/Chip';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { PresetGroupPicker } from './PresetGroupPicker';
import type { PresetGroupItem } from './types';

const chip = (key: string, phase?: 1 | 2, group?: string): PresetGroupItem => ({ key, phase, group, node: <Chip label={key} onSelect={() => {}} /> });

const meta = {
	title: 'App/Presets/Components/PresetGroupPicker',
	component: PresetGroupPicker,
	tags: ['autodocs'],
	args: {
		sections: [
			{
				title: 'Preset Configurations',
				tooltip: 'Loads a whole build at once.',
				items: [chip('Pre-Raid'), chip('Fury P1', 1, 'Fury'), chip('Arms P1', 1, 'Arms'), chip('Fury P2', 2, 'Fury')],
			},
			{ title: 'Gear Sets', items: [chip('Blue Gear'), chip('Molten Core', 1), chip('Blackwing Lair', 2)] },
		],
		phases: [1, 2],
		phase: 1,
		onSelectPhase: fn(),
	},
	argTypes: { sections: { control: false } },
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PresetGroupPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

type PickerProps = Story['args'] & Parameters<typeof PresetGroupPicker>[0];

const PhasedPicker = (args: PickerProps) => {
	const [phase, setPhase] = useState(args.phase);
	return (
		<PresetGroupPicker
			{...args}
			phase={phase}
			onSelectPhase={next => {
				setPhase(next);
				args.onSelectPhase(next);
			}}
		/>
	);
};

export const MultiPhase: Story = {
	render: args => <PhasedPicker {...args} />,
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const tabs = within(canvas.getByTestId('preset-group-phase-tabs')).getAllByRole('tab');
		await userEvent.click(tabs[1]);
		await expect(args.onSelectPhase).toHaveBeenCalledWith(2);
		await expect(canvas.getByText('Fury P2')).toBeVisible();
		await expect(canvas.queryByText('Fury P1')).not.toBeInTheDocument();
	},
};

export const SinglePhase: Story = { args: { phases: [1] } };
