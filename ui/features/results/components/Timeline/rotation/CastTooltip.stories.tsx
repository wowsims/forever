import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { castLog, damageLog, resultsHostParts } from '@testing/storybook/fixtures/results-views';

import { CastTooltip } from './CastTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/CastTooltip',
	component: CastTooltip,
	tags: ['autodocs'],
	args: { log: castLog(12) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
		withSimHost(resultsHostParts(null)),
	],
} satisfies Meta<typeof CastTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const InstantWithDamage: Story = {};

export const CastWithTravel: Story = { args: { log: castLog(12, { castTime: 2.5, effectiveTime: 2.5, travelTime: 0.8 }) } };

export const Cancelled: Story = {
	args: {
		log: castLog(12, {
			castTime: 2.5,
			cancelTime: 1.1,
			castCancelledLog: { ...castLog(13.1).castBeganLog, kind: 'cast-cancelled', cancelTime: 1.1, timestamp: 13.1 },
			damageDealtLogs: [],
		}),
	},
};

const stackedHits = castLog(12, { damageDealtLogs: [damageLog(12.01, 244.59), damageLog(12.01, 489.19, { outcome: 'crit' })] });

/** Two casts at the same time, where the later one took both hits. */
export const Stacked: Story = { args: { log: stackedHits, stack: [castLog(12, { damageDealtLogs: [] }), stackedHits] } };
