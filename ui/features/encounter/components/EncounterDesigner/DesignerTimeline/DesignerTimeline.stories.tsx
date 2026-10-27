import { EncounterDesign } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	customBossDesign,
	DESIGN_BASE_SECONDS,
	designTargetNames,
	dynamicAddsDesign,
	emptyDesign,
	movementDesign,
	overlappingDowntimesDesign,
} from '@testing/storybook/fixtures/encounter_design';
import { useState } from 'react';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';

import type { DesignEventRef } from '../../../model/design/lanes';
import { addEvent } from '../../../model/design/lanes';
import { validateDesign } from '../../../model/design/validate';
import { designerText } from '../utils/configs';
import { DesignerTimeline, type DesignerTimelineProps } from './DesignerTimeline';

// Holds the design the way the dialog does, so a drag, a key press or an add lands in the story.
const EditableTimeline = ({ design: initial, selected: initialSelected, onChange, onSelect, onAdd, ...props }: DesignerTimelineProps) => {
	const [design, setDesign] = useState<EncounterDesign>(initial);
	const [selected, setSelected] = useState<DesignEventRef | undefined>(initialSelected);
	return (
		<DesignerTimeline
			{...props}
			design={design}
			selected={selected}
			issues={validateDesign(design, designTargetNames.length)}
			onChange={next => {
				onChange(next);
				setDesign(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
			onAdd={kind => {
				onAdd(kind);
				const next = EncounterDesign.clone(design);
				setSelected(addEvent(next, kind));
				setDesign(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Encounter/EncounterDesigner/DesignerTimeline',
	component: DesignerTimeline,
	tags: ['autodocs'],
	args: { design: emptyDesign(), baseSeconds: DESIGN_BASE_SECONDS, issues: [], onChange: fn(), onSelect: fn(), onActivate: fn(), onAdd: fn() },
	argTypes: { design: { control: false }, issues: { control: false } },
	render: args => <EditableTimeline {...args} />,
} satisfies Meta<typeof DesignerTimeline>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const DynamicAdds: Story = { args: { design: dynamicAddsDesign() } };

export const Movement: Story = { args: { design: movementDesign() } };

export const CustomBoss: Story = { args: { design: customBossDesign(), selected: { kind: 'ability', index: 2 } } };

export const OverlappingDowntimes: Story = { args: { design: overlappingDowntimesDesign() } };

export const LongFight: Story = { args: { design: dynamicAddsDesign(), baseSeconds: 600 } };

// Downtime 2 changes row, then passes Downtime 3 in time, and keeps focus through both; Enter then activates it.
export const KeyboardMoveKeepsFocus: Story = {
	args: { design: overlappingDowntimesDesign() },
	play: async ({ args, canvasElement }) => {
		const bar = within(canvasElement).getAllByRole('slider')[1];
		await expect(bar.style.getPropertyValue('--row')).toBe('2 / span 1');
		await userEvent.click(bar);
		await expect(args.onActivate).not.toHaveBeenCalled();
		for (let step = 0; step < 3; step++) await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}');
		await waitFor(() => expect(bar.style.getPropertyValue('--row')).toBe('1 / span 1'));
		await expect(bar).toHaveFocus();
		for (let step = 0; step < 6; step++) await userEvent.keyboard('{Shift>}{ArrowRight}{/Shift}');
		await waitFor(() => expect(bar).toHaveAttribute('aria-valuenow', '105'));
		await expect(bar).toHaveFocus();
		await userEvent.keyboard('{Enter}');
		await expect(args.onActivate).toHaveBeenCalledWith({ kind: 'downtime', index: 1 });
	},
};

export const DeleteMovesFocus: Story = {
	args: { design: overlappingDowntimesDesign() },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getAllByRole('slider')[0]);
		await userEvent.keyboard('{Delete}');
		await waitFor(() => expect(canvas.getAllByRole('slider')).toHaveLength(2));
		await waitFor(() => expect(canvas.getAllByRole('slider')[0]).toHaveFocus());
	},
};

export const ZoomedOut: Story = {
	args: { design: customBossDesign() },
	play: async ({ canvasElement }) => {
		const zoomOut = within(canvasElement).getByRole('button', { name: designerText('toolbar.zoom_out') });
		for (let step = 0; step < 4; step++) await userEvent.click(zoomOut);
	},
};
