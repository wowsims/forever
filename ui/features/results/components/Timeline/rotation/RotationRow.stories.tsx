import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	auraRowItem,
	castRotationRow,
	castRowItem,
	delayRowItem,
	gcdSegmentRowItem,
	MORTAL_STRIKE,
	resourceRowItem,
	rotationTrackStyle,
	tickRowItem,
	warriorRotationModel,
} from '@testing/storybook/fixtures/results-views';
import { Tooltip } from '@ui-kit/Tooltip';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { ContentRow } from '../../../model/timeline/rotation';
import { RotationRow, RowItemView } from './RotationRow';

const MODEL = warriorRotationModel();
const rowOf = (key: string) => MODEL.rows[MODEL.byKey.get(key)!] as ContentRow;
const allItems = (row: ContentRow) => row.items.map((_, index) => index);
const HIDE_TOOLTIP_ID = 'story-rotation-hide';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationRow',
	component: RotationRow,
	subcomponents: { RowItemView },
	tags: ['autodocs'],
	args: { row: rowOf('cast:Mortal Strike'), items: allItems(rowOf('cast:Mortal Strike')), onHide: fn(), hideTooltipId: HIDE_TOOLTIP_ID },
	argTypes: { row: { control: false }, items: { control: false } },
	decorators: [
		(Story, { args }) => (
			<div className="w-max text-white" style={rotationTrackStyle(args.row, MODEL.duration)}>
				<Story />
				<Tooltip id={HIDE_TOOLTIP_ID} />
			</div>
		),
	],
} satisfies Meta<typeof RotationRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Casts: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('rotation-row-hide'));
		await expect(args.onHide).toHaveBeenCalledWith('cast:Mortal Strike');
	},
};

export const Ticks: Story = { args: { row: rowOf('cast:Deep Wounds'), items: allItems(rowOf('cast:Deep Wounds')) } };

export const Rage: Story = { args: { row: rowOf('resource:rage'), items: allItems(rowOf('resource:rage')) } };

export const GcdStrip: Story = { args: { row: rowOf('gcd:player'), items: allItems(rowOf('gcd:player')) } };

export const WindowedToTheFirstTwo: Story = { args: { items: [0, 1] } };

type ItemStory = StoryObj<typeof RowItemView>;

const item: ItemStory = {
	args: { item: castRowItem(1), index: 0, iconUrl: MORTAL_STRIKE.iconUrl, cssName: '' },
	argTypes: { item: { control: false } },
	render: args => <RowItemView {...args} />,
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(MORTAL_STRIKE, []), 6)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
};

export const ItemCast: ItemStory = item;

export const ItemDelay: ItemStory = { ...item, args: { ...item.args, item: delayRowItem(1, 1.4) } };

export const ItemGcdExtension: ItemStory = { ...item, args: { ...item.args, item: { kind: 'gcdExtension', start: 1, end: 2.5 } } };

export const ItemGcdSegment: ItemStory = { ...item, args: { ...item.args, item: gcdSegmentRowItem(1) } };

export const ItemTick: ItemStory = { ...item, args: { ...item.args, item: tickRowItem(2) } };

export const ItemAura: ItemStory = { ...item, args: { ...item.args, item: auraRowItem(1, 5, [{ offset: 0, duration: 4, stacks: 2 }]) } };

export const ItemResource: ItemStory = { ...item, args: { ...item.args, item: resourceRowItem(1, 4, 60), cssName: 'rage' } };
