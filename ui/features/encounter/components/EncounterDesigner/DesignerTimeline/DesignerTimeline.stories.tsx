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
import { expect, fireEvent, fn, userEvent, waitFor, within } from 'storybook/test';

import type { DesignEventRef } from '../../../model/design/lanes';
import { addEvent } from '../../../model/design/lanes';
import { validateDesign } from '../../../model/design/validate';
import { designerText, eventTooltip } from '../utils/configs';
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
	args: {
		design: emptyDesign(),
		baseSeconds: DESIGN_BASE_SECONDS,
		issues: [],
		targetNames: designTargetNames,
		onChange: fn(),
		onSelect: fn(),
		onActivate: fn(),
		onAdd: fn(),
	},
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

// Both ends of the axis are drawn whole: the last ruler label reads "180s" inside the scroller, and nothing scrolls at fit zoom.
export const AxisEndsVisible: Story = {
	args: { design: customBossDesign() },
	parameters: { smoke: false, smokeReason: 'Measures layout, which the DOM test environment does not have.' },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const scroller = canvas.getByRole('region', { name: designerText('timeline') });
		await waitFor(() => expect(canvas.getAllByTestId('rotation-ruler-label').some(label => label.textContent === '180s')).toBe(true));
		const bounds = scroller.getBoundingClientRect();
		const last = canvas.getAllByTestId('rotation-ruler-label').find(label => label.textContent === '180s')!;
		await expect(last.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right + 0.5);
		await expect(scroller.scrollWidth).toBeLessThanOrEqual(scroller.clientWidth);
		// scrollWidth and clientWidth round alike, so only the boxes show a fraction of a pixel of overflow.
		await expect(scroller.firstElementChild!.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right);

		const lane = canvas.getByRole('group', { name: designerText('lanes.boss') });
		const health = within(lane).getAllByRole('slider');
		const labelColumn = lane.firstElementChild!.getBoundingClientRect();
		await expect(health[0].getBoundingClientRect().left).toBeGreaterThanOrEqual(labelColumn.right - 0.5);
		await expect(health.at(-1)!.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right + 0.5);
	},
};

const abilityBar = (canvasElement: HTMLElement) =>
	within(within(canvasElement).getByRole('group', { name: designerText('lanes.abilities') })).getAllByRole('slider')[0];

// The strings the lane builds, so the story reads the same under the stubbed locale of the smoke test.
const SPELL_HIT_TIP = eventTooltip(customBossDesign(), { kind: 'ability', index: 0 }, DESIGN_BASE_SECONDS, designTargetNames);

const expectTip = async () => {
	await expect(await within(document.body).findByText(SPELL_HIT_TIP.title)).toBeVisible();
	for (const line of SPELL_HIT_TIP.lines) await expect(within(document.body).getByText(line)).toBeVisible();
};

export const HoverTooltip: Story = {
	args: { design: customBossDesign() },
	play: async ({ canvasElement }) => {
		await userEvent.hover(abilityBar(canvasElement));
		await expectTip();
	},
};

export const FocusTooltip: Story = {
	args: { design: customBossDesign() },
	play: async ({ canvasElement }) => {
		abilityBar(canvasElement).focus();
		await expectTip();
	},
};

export const ContextMenuDuplicate: Story = {
	args: { design: dynamicAddsDesign() },
	play: async ({ args, canvasElement }) => {
		const bar = within(canvasElement).getByRole('slider', { name: 'Sons of Flame' });
		fireEvent.contextMenu(bar, { clientX: 200, clientY: 200 });
		await userEvent.click(await within(document.body).findByRole('menuitem', { name: designerText('menu.duplicate') }));
		const waves = within(within(canvasElement).getByRole('group', { name: designerText('lanes.waves') }));
		await waitFor(() => expect(waves.getAllByRole('slider')).toHaveLength(2));
		await expect(args.onChange).toHaveBeenCalledTimes(1);
		await waitFor(() => expect(waves.getAllByRole('slider')[1]).toHaveFocus());
	},
};

export const CtrlDDuplicates: Story = {
	args: { design: overlappingDowntimesDesign() },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const bar = canvas.getAllByRole('slider')[0];
		await userEvent.click(bar);
		await userEvent.keyboard('{Control>}d{/Control}');
		await waitFor(() => expect(canvas.getAllByRole('slider')).toHaveLength(4));
		await expect(args.onChange).toHaveBeenCalledTimes(1);
		const copy = canvas
			.getAllByRole('slider')
			.find(slider => slider.getAttribute('aria-valuenow') === String(Number(bar.getAttribute('aria-valuenow')) + 1));
		await expect(copy).toBeDefined();
		await expect(copy).toHaveFocus();
	},
};

export const RenameInline: Story = {
	args: { design: dynamicAddsDesign() },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.dblClick(within(canvas.getByRole('slider', { name: 'Sons of Flame' })).getByText('Sons of Flame'));
		const input = canvas.getByRole('textbox', { name: designerText('menu.rename_label') });
		await userEvent.clear(input);
		await userEvent.type(input, 'Adds{Escape}');
		await expect(args.onChange).not.toHaveBeenCalled();
		await expect(canvas.getByRole('slider', { name: 'Sons of Flame' })).toBeVisible();

		await userEvent.dblClick(within(canvas.getByRole('slider', { name: 'Sons of Flame' })).getByText('Sons of Flame'));
		const second = canvas.getByRole('textbox', { name: designerText('menu.rename_label') });
		await userEvent.clear(second);
		await userEvent.type(second, 'Adds{Enter}');
		await expect(args.onChange).toHaveBeenCalledTimes(1);
		await waitFor(() => expect(canvas.getByRole('slider', { name: 'Adds' })).toBeVisible());
	},
};

// Zoomed in past the width, a drag on empty lane space scrolls the lanes, and a drag on a bar still moves the bar.
export const PanByDrag: Story = {
	args: { design: customBossDesign() },
	parameters: { smoke: false, smokeReason: 'Scrolls a laid-out scroller with pointer capture, which the DOM test environment does not have.' },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const zoomIn = canvas.getByRole('button', { name: designerText('toolbar.zoom_in') });
		for (let step = 0; step < 3; step++) await userEvent.click(zoomIn);
		const scroller = canvas.getByRole('region', { name: designerText('timeline') });
		await waitFor(() => expect(scroller.scrollWidth).toBeGreaterThan(scroller.clientWidth + 400));

		const drag = (target: HTMLElement, x: number, y: number, dx: number) => {
			const pointer = { pointerId: 1, pointerType: 'mouse', button: 0, clientY: y };
			fireEvent.pointerDown(target, { ...pointer, clientX: x });
			fireEvent.pointerMove(target, { ...pointer, clientX: x + dx / 2 });
			fireEvent.pointerMove(target, { ...pointer, clientX: x + dx });
			fireEvent.pointerUp(target, { ...pointer, clientX: x + dx });
		};
		const moves = canvas.getByRole('group', { name: designerText('lanes.moves') });
		const lane = moves.getBoundingClientRect();
		scroller.scrollLeft = 0;
		drag(moves, lane.right - 20, lane.top + 4, -200);
		await expect(scroller.scrollLeft).toBe(200);

		const bar = abilityBar(canvasElement);
		const start = bar.getAttribute('aria-valuenow');
		const box = bar.getBoundingClientRect();
		// The bar runs to the end of the fight, so only an earlier start moves it.
		drag(bar, box.left + 40, box.top + box.height / 2, -100);
		await expect(scroller.scrollLeft).toBe(200);
		await waitFor(() => expect(bar.getAttribute('aria-valuenow')).not.toBe(start));
		await expect(args.onChange).toHaveBeenCalledTimes(1);
	},
};
