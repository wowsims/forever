import { Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { designPresetTargets } from '@testing/storybook/fixtures/encounter_design';
import { useState } from 'react';
import { expect, fn, screen, userEvent, waitFor, within } from 'storybook/test';

import { DesignerAdds, type DesignerAddsProps } from './DesignerAdds';

// Holds the adds the way the panel does, so an add, a copy or a removal shows in the list.
const EditableAdds = ({ adds: initial, onEdit, ...props }: DesignerAddsProps) => {
	const [adds, setAdds] = useState(initial);
	return (
		<div className="w-96">
			<DesignerAdds
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

const template = () => Target.clone(designPresetTargets[0].target!);

const meta = {
	title: 'Features/Encounter/EncounterDesigner/DesignerAdds',
	component: DesignerAdds,
	tags: ['autodocs'],
	args: { adds: [template()], presetTargets: designPresetTargets, newAdd: template, idPrefix: 'story-adds', onEdit: fn() },
	argTypes: { adds: { control: false }, presetTargets: { control: false } },
	render: args => <EditableAdds {...args} />,
} satisfies Meta<typeof DesignerAdds>;

export default meta;
type Story = StoryObj<typeof meta>;

export const OneAdd: Story = {};

export const MixedAdds: Story = {
	args: { adds: [template(), Target.clone(designPresetTargets[1].target!), template()] },
};

const editButtons = (canvasElement: HTMLElement) => within(canvasElement).getAllByRole('button', { name: /adds\.edit$|Edit add/ });

// A new add opens its dialog; once it closes, a copy lands beside its original and a removal takes one row.
export const AddCopyRemove: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: /adds\.add$|Add an add/ }));
		await userEvent.click(await screen.findByRole('button', { name: /adds\.done$|Done/ }));
		await waitFor(() => expect(screen.queryByTestId('encounter-designer-add-dialog')).toBeNull());
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
		const dialog = await screen.findByTestId('encounter-designer-add-dialog');
		await expect(within(dialog).getByTestId('sim-dialog-title').textContent).toBe(designPresetTargets[0].target!.name);
	},
};
