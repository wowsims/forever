import { Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { planPresetTargets } from '@testing/storybook/fixtures/encounter_plan';
import { useState } from 'react';
import { expect, fn, screen, userEvent, waitFor, within } from 'storybook/test';

import { PlannerAdds, type PlannerAddsProps } from './PlannerAdds';

// Holds the adds the way the panel does, so an add, a copy or a removal shows in the list.
const EditableAdds = ({ adds: initial, onEdit, ...props }: PlannerAddsProps) => {
	const [adds, setAdds] = useState(initial);
	return (
		<div className="w-96">
			<PlannerAdds
				{...props}
				adds={adds}
				onEdit={edit => {
					onEdit(edit);
					const next = adds.map(add => Target.clone(add));
					edit(next);
					setAdds(next);
				}}
			/>
		</div>
	);
};

const template = () => Target.clone(planPresetTargets[0].target!);

const meta = {
	title: 'Features/EncounterPlanner/PlannerAdds',
	component: PlannerAdds,
	tags: ['autodocs'],
	args: { adds: [template()], presetTargets: planPresetTargets, newAdd: template, idPrefix: 'story-adds', onEdit: fn() },
	argTypes: { adds: { control: false }, presetTargets: { control: false } },
	render: args => <EditableAdds {...args} />,
} satisfies Meta<typeof PlannerAdds>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OneAdd: Story = {};

export const MixedAdds: Story = {
	args: { adds: [template(), Target.clone(planPresetTargets[1].target!), template()] },
};

const editButtons = (canvasElement: HTMLElement) => within(canvasElement).getAllByRole('button', { name: /adds\.edit$|Edit add/ });

// A new add opens its dialog; once it closes, a copy lands beside its original and a removal takes one row.
export const AddCopyRemove: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: /adds\.add$|Add an add/ }));
		await userEvent.click(await screen.findByRole('button', { name: /adds\.done$|Done/ }));
		await waitFor(() => expect(screen.queryByTestId('encounter-planner-add-dialog')).toBeNull());
		await expect(editButtons(canvasElement)).toHaveLength(2);
		await userEvent.click(canvas.getAllByRole('button', { name: /duplicate/i })[0]);
		await expect(editButtons(canvasElement)).toHaveLength(3);
		await userEvent.click(canvas.getAllByRole('button', { name: /remove/i })[0]);
		await expect(editButtons(canvasElement)).toHaveLength(2);
	},
};

// The add's template and stats open in a dialog over the list, titled with the add's name.
export const EditDialog: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(editButtons(canvasElement)[0]);
		const dialog = await screen.findByTestId('encounter-planner-add-dialog');
		await expect(within(dialog).getByTestId('sim-dialog-title').textContent).toBe(planPresetTargets[0].target!.name);
	},
};
