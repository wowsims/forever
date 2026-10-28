import { Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { designPresetTargets } from '@testing/storybook/fixtures/encounter_design';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

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

// A copy opens beside its original; removing a row closes nothing else.
export const AddCopyRemove: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const rows = () => canvasElement.querySelectorAll('[data-testid$="-stats"]').length;
		await userEvent.click(canvas.getByRole('button', { name: /^(.*)adds\.add$|Add an add/ }));
		await expect(rows()).toBe(2);
		await userEvent.click(canvas.getAllByRole('button', { name: /duplicate/i })[0]);
		await expect(rows()).toBe(3);
		await userEvent.click(canvas.getAllByRole('button', { name: /remove/i })[0]);
		await expect(rows()).toBe(2);
	},
};
