import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import {
	auraRowItem,
	castRowItem,
	delayRowItem,
	gcdSegmentRowItem,
	resourceRowItem,
	resultsHostParts,
	tickRowItem,
} from '@testing/storybook/fixtures/results-views';

import { RowItemTooltip } from './RowItemTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/RowItemTooltip',
	component: RowItemTooltip,
	tags: ['autodocs'],
	args: { item: castRowItem(12) },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
		withSimHost(resultsHostParts(null)),
	],
} satisfies Meta<typeof RowItemTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Cast: Story = {};

export const Delay: Story = { args: { item: delayRowItem(8.5, 8.85) } };

export const GcdSegment: Story = { args: { item: gcdSegmentRowItem(6) } };

export const Tick: Story = { args: { item: tickRowItem(9) } };

export const Aura: Story = { args: { item: auraRowItem(4.25, 16.25) } };

export const Resource: Story = { args: { item: resourceRowItem(12, 18, 60) } };
