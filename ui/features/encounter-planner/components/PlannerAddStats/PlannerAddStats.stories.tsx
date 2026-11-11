import { Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { planPresetTargets } from '@testing/storybook/fixtures/encounter_plan';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { usePlannerSectionState } from '../PlannerSections';
import { PlannerAddStats, type PlannerAddStatsProps } from './PlannerAddStats';

// Holds the add the way the panel does, so an edit shows in the field it came from.
const EditableAddStats = ({ target: initial, onModify, ...props }: PlannerAddStatsProps) => {
	const [target, setTarget] = useState(initial);
	return (
		<div className="w-80">
			<PlannerAddStats
				{...props}
				sections={usePlannerSectionState()}
				target={target}
				onModify={edit => {
					onModify(edit);
					const next = Target.clone(target);
					edit(next);
					setTarget(next);
				}}
			/>
		</div>
	);
};

const meta = {
	title: 'Features/EncounterPlanner/PlannerAddStats',
	component: PlannerAddStats,
	tags: ['autodocs'],
	args: { target: Target.clone(planPresetTargets[0].target!), idPrefix: 'story-add', sections: { toggles: {}, onToggle: fn() }, onModify: fn() },
	argTypes: { target: { control: false }, sections: { control: false } },
	render: args => <EditableAddStats {...args} />,
} satisfies Meta<typeof PlannerAddStats>;

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
