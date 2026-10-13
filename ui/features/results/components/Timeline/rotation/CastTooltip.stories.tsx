import { Entity } from '@sim/proto/combat_log';
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

const chainHit = (index: number, amount: number, overrides = {}) =>
	damageLog(12.01, amount, { target: new Entity('Raid Target', '', index, true, false), ...overrides });

/** One cast that hits three targets, each hit on its own line. */
export const MultipleTargets: Story = {
	args: { log: castLog(12, { damageDealtLogs: [chainHit(0, 612.4), chainHit(1, 428.68, { outcome: 'crit' }), chainHit(2, 300.08)] }) },
};
