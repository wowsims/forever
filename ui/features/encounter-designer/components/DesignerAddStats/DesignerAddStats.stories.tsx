import { Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { designPresetTargets } from '@testing/storybook/fixtures/encounter_design';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DesignerAddStats, type DesignerAddStatsProps } from './DesignerAddStats';

// Holds the add the way the panel does, so an edit shows in the field it came from.
const EditableAddStats = ({ target: initial, onModify, ...props }: DesignerAddStatsProps) => {
	const [target, setTarget] = useState(initial);
	return (
		<div className="w-80">
			<DesignerAddStats
				{...props}
				target={target}
				onModify={edit => {
					onModify(edit);
					const next = Target.clone(target!);
					edit(next);
					setTarget(next);
				}}
			/>
		</div>
	);
};

const meta = {
	title: 'Features/EncounterDesigner/DesignerAddStats',
	component: DesignerAddStats,
	tags: ['autodocs'],
	args: { target: Target.clone(designPresetTargets[0].target!), idPrefix: 'story-add', onModify: fn() },
	argTypes: { target: { control: false } },
	render: args => <EditableAddStats {...args} />,
} satisfies Meta<typeof DesignerAddStats>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const EditsAStat: Story = {
	play: async ({ args, canvasElement }) => {
		// The first text field is the first stat, Health.
		const health = within(canvasElement).getAllByRole('textbox')[0];
		await userEvent.clear(health);
		await userEvent.type(health, '6000');
		await userEvent.tab();
		await expect(args.onModify).toHaveBeenCalled();
	},
};
