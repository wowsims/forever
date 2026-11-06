import { EncounterPlan } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	areaDamagePlan,
	customBossPlan,
	dynamicAddsPlan,
	emptyPlan,
	movementPlan,
	overlappingDowntimesPlan,
	PLAN_BASE_SECONDS,
	planTargetNames,
} from '@testing/storybook/fixtures/encounter_plan';
import { useState } from 'react';
import { expect, fireEvent, fn, userEvent, waitFor, within } from 'storybook/test';

import type { PlannerSelection } from '../../model/lanes';
import { addEvent } from '../../model/lanes';
import { validatePlan } from '../../model/validate';
import { plannerText } from '../EncounterPlanner/utils/configs';
import { eventTooltip } from '../EncounterPlanner/utils/tips';
import { PlannerTimeline, type PlannerTimelineProps } from './PlannerTimeline';

// Holds the plan the way the dialog does, so a drag, a key press or an add lands in the story.
const EditableTimeline = ({ plan: initial, selected: initialSelected, onChange, onSelect, onAdd, ...props }: PlannerTimelineProps) => {
	const [plan, setPlan] = useState<EncounterPlan>(initial);
	const [selected, setSelected] = useState<PlannerSelection | undefined>(initialSelected);
	return (
		<PlannerTimeline
			{...props}
			plan={plan}
			selected={selected}
			issues={validatePlan(plan, planTargetNames.length)}
			onChange={next => {
				onChange(next);
				setPlan(next);
			}}
			onSelect={ref => {
				onSelect(ref);
				setSelected(ref);
			}}
			onAdd={(kind, targetIndex) => {
				onAdd(kind, targetIndex);
				const next = EncounterPlan.clone(plan);
				setSelected(addEvent(next, kind, targetIndex));
				setPlan(next);
			}}
		/>
	);
};

const meta = {
	title: 'Features/EncounterPlanner/PlannerTimeline',
	component: PlannerTimeline,
	tags: ['autodocs'],
	args: {
		plan: emptyPlan(),
		baseSeconds: PLAN_BASE_SECONDS,
		issues: [],
		targetNames: planTargetNames,
		onChange: fn(),
		onSelect: fn(),
		onActivate: fn(),
		onAdd: fn(),
		onAddTarget: fn(),
	},
	argTypes: { plan: { control: false }, issues: { control: false } },
	render: args => <EditableTimeline {...args} />,
} satisfies Meta<typeof PlannerTimeline>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const DynamicAdds: Story = { args: { plan: dynamicAddsPlan() } };

export const Movement: Story = { args: { plan: movementPlan() } };

export const CustomBoss: Story = { args: { plan: customBossPlan(), selected: { kind: 'ability', index: 2 } } };

export const OverlappingDowntimes: Story = { args: { plan: overlappingDowntimesPlan() } };

export const LongFight: Story = { args: { plan: dynamicAddsPlan(), baseSeconds: 600 } };

// A target's row label selects it on a click, and on Enter selects it and activates it; an arrow key moves nothing.
export const TargetSelected: Story = {
	args: { plan: customBossPlan(), selected: { kind: 'target', index: 0 } },
	play: async ({ args, canvasElement }) => {
		const lane = within(within(canvasElement).getByRole('group', { name: plannerText('targets') }));
		const second = lane.getByRole('button', { name: planTargetNames[1] });
		await userEvent.click(second);
		await expect(args.onSelect).toHaveBeenCalledWith({ kind: 'target', index: 1 });
		await expect(args.onActivate).not.toHaveBeenCalled();
		await expect(second).toHaveAttribute('aria-pressed', 'true');
		await userEvent.keyboard('{ArrowRight}{Enter}');
		await expect(args.onChange).not.toHaveBeenCalled();
		await expect(args.onActivate).toHaveBeenCalledWith({ kind: 'target', index: 1 });
	},
};

// Downtime 2 changes row, then passes Downtime 3 in time, and keeps focus through both; Enter then activates it.
const sameRowDowntimes = () => {
	const plan = overlappingDowntimesPlan();
	plan.downtimes[2].targetIndex = 0;
	return plan;
};

export const KeyboardMoveKeepsFocus: Story = {
	args: { plan: sameRowDowntimes() },
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
	args: { plan: overlappingDowntimesPlan() },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getAllByRole('slider')[0]);
		await userEvent.keyboard('{Delete}');
		await waitFor(() => expect(canvas.getAllByRole('slider')).toHaveLength(2));
		await waitFor(() => expect(canvas.getAllByRole('slider')[0]).toHaveFocus());
	},
};

export const ZoomedOut: Story = {
	args: { plan: customBossPlan() },
	play: async ({ canvasElement }) => {
		const zoomOut = within(canvasElement).getByRole('button', { name: plannerText('toolbar.zoom_out') });
		for (let step = 0; step < 4; step++) await userEvent.click(zoomOut);
	},
};

// Both ends of the axis are drawn whole: the last ruler label reads "180s" inside the scroller, and nothing scrolls at fit zoom.
export const AxisEndsVisible: Story = {
	args: { plan: customBossPlan() },
	parameters: { smoke: false, smokeReason: 'Measures layout, which the DOM test environment does not have.' },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const scroller = canvas.getByRole('region', { name: plannerText('timeline') });
		await waitFor(() => expect(canvas.getAllByTestId('rotation-ruler-label').some(label => label.textContent === '180s')).toBe(true));
		const bounds = scroller.getBoundingClientRect();
		const last = canvas.getAllByTestId('rotation-ruler-label').find(label => label.textContent === '180s')!;
		await expect(last.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right + 0.5);
		await expect(scroller.scrollWidth).toBeLessThanOrEqual(scroller.clientWidth);
		// scrollWidth and clientWidth round alike, so only the boxes show a fraction of a pixel of overflow.
		await expect(scroller.firstElementChild!.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right);

		const lane = canvas.getByRole('group', { name: planTargetNames[0] });
		const health = within(lane).getAllByRole('slider');
		const labelColumn = lane.firstElementChild!.getBoundingClientRect();
		await expect(health[0].getBoundingClientRect().left).toBeGreaterThanOrEqual(labelColumn.right - 0.5);
		await expect(health.at(-1)!.getBoundingClientRect().right).toBeLessThanOrEqual(bounds.right + 0.5);
	},
};

const abilityBar = (canvasElement: HTMLElement) =>
	within(within(canvasElement).getByRole('group', { name: plannerText('lanes.abilities') })).getAllByRole('slider')[0];

// The strings the lane builds, so the story reads the same under the stubbed locale of the smoke test.
const SPELL_HIT_TIP = eventTooltip(customBossPlan(), { kind: 'ability', index: 0 }, PLAN_BASE_SECONDS, planTargetNames);

const expectTip = async () => {
	await expect(await within(document.body).findByText(SPELL_HIT_TIP.title)).toBeVisible();
	for (const line of SPELL_HIT_TIP.lines) await expect(within(document.body).getByText(line)).toBeVisible();
};

export const HoverTooltip: Story = {
	args: { plan: customBossPlan() },
	play: async ({ canvasElement }) => {
		await userEvent.hover(abilityBar(canvasElement));
		await expectTip();
	},
};

export const FocusTooltip: Story = {
	args: { plan: customBossPlan() },
	play: async ({ canvasElement }) => {
		abilityBar(canvasElement).focus();
		await expectTip();
	},
};

const PULSE_TIP = eventTooltip(areaDamagePlan(), { kind: 'ability', index: 0 }, PLAN_BASE_SECONDS, planTargetNames);

// A pulse's window with its ticks, a move that steps out of it, and an area hit every 40 s; hovering the pulse shows its rules.
export const AreaDamage: Story = {
	args: { plan: areaDamagePlan() },
	play: async ({ canvasElement }) => {
		await userEvent.hover(within(canvasElement).getByRole('slider', { name: 'Lava Burst' }));
		await expect(await within(document.body).findByText(PULSE_TIP.title)).toBeVisible();
		for (const line of PULSE_TIP.lines) await expect(within(document.body).getByText(line)).toBeVisible();
	},
};

export const ContextMenuDuplicate: Story = {
	args: { plan: dynamicAddsPlan() },
	play: async ({ args, canvasElement }) => {
		const bar = within(canvasElement).getByRole('slider', { name: 'Sons of Flame' });
		fireEvent.contextMenu(bar, { clientX: 200, clientY: 200 });
		await userEvent.click(await within(document.body).findByRole('menuitem', { name: plannerText('menu.duplicate') }));
		const waves = within(within(canvasElement).getByRole('group', { name: plannerText('lanes.waves') }));
		await waitFor(() => expect(waves.getAllByRole('slider')).toHaveLength(2));
		await expect(args.onChange).toHaveBeenCalledTimes(1);
		await waitFor(() => expect(waves.getAllByRole('slider')[1]).toHaveFocus());
	},
};

export const CtrlDDuplicates: Story = {
	args: { plan: overlappingDowntimesPlan() },
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
	args: { plan: dynamicAddsPlan() },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.dblClick(within(canvas.getByRole('slider', { name: 'Sons of Flame' })).getByText('Sons of Flame'));
		const input = canvas.getByRole('textbox', { name: plannerText('menu.rename_label') });
		await userEvent.clear(input);
		await userEvent.type(input, 'Adds{Escape}');
		await expect(args.onChange).not.toHaveBeenCalled();
		await expect(canvas.getByRole('slider', { name: 'Sons of Flame' })).toBeVisible();

		await userEvent.dblClick(within(canvas.getByRole('slider', { name: 'Sons of Flame' })).getByText('Sons of Flame'));
		const second = canvas.getByRole('textbox', { name: plannerText('menu.rename_label') });
		await userEvent.clear(second);
		await userEvent.type(second, 'Adds{Enter}');
		await expect(args.onChange).toHaveBeenCalledTimes(1);
		await waitFor(() => expect(canvas.getByRole('slider', { name: 'Adds' })).toBeVisible());
	},
};

// Zoomed in past the width, a drag on empty lane space scrolls the lanes, and a drag on a bar still moves the bar.
export const PanByDrag: Story = {
	args: { plan: customBossPlan() },
	parameters: { smoke: false, smokeReason: 'Scrolls a laid-out scroller with pointer capture, which the DOM test environment does not have.' },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const zoomIn = canvas.getByRole('button', { name: plannerText('toolbar.zoom_in') });
		for (let step = 0; step < 3; step++) await userEvent.click(zoomIn);
		const scroller = canvas.getByRole('region', { name: plannerText('timeline') });
		await waitFor(() => expect(scroller.scrollWidth).toBeGreaterThan(scroller.clientWidth + 400));

		const drag = (target: HTMLElement, x: number, y: number, dx: number) => {
			const pointer = { pointerId: 1, pointerType: 'mouse', button: 0, clientY: y };
			fireEvent.pointerDown(target, { ...pointer, clientX: x });
			fireEvent.pointerMove(target, { ...pointer, clientX: x + dx / 2 });
			fireEvent.pointerMove(target, { ...pointer, clientX: x + dx });
			fireEvent.pointerUp(target, { ...pointer, clientX: x + dx });
		};
		const moves = canvas.getByRole('group', { name: plannerText('lanes.moves') });
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
