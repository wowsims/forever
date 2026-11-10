import { APLAction, APLRotation } from '@generated/proto/apl';
import { type AplEntryPath, type CastLog, Entity } from '@sim/proto/combat_log';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { after, constant, sequence, waitFor } from '@testing/apl/builders';
import { makeAplStoryPlayer } from '@testing/storybook/fixtures/apl';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { castLog, damageLog, resultsHostParts } from '@testing/storybook/fixtures/results-views';

import { CastTooltip } from './CastTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/CastTooltip',
	component: CastTooltip,
	tags: ['autodocs'],
	args: { log: castLog(12) },
	argTypes: { log: { control: false }, aplRun: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
		// The APL player answers the reads the drawn entry's pickers make.
		withSimHost({ ...resultsHostParts(null), player: makeAplStoryPlayer() }),
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

const rotation = APLRotation.create({
	prepullActions: [{ action: waitFor('1s'), doAtValue: constant('-2s') }],
	priorityList: [
		{ action: APLAction.create({ action: { oneofKind: 'groupReference', groupReference: { groupName: 'execute', variables: [] } } }) },
		{ action: sequence([waitFor('1s'), waitFor('2s'), waitFor('3s')]) },
	],
	groups: [{ name: 'execute', actions: [{ action: waitFor('1s') }, { action: waitFor('2s', after('20s')) }], variables: [] }],
});

const IN_GROUP: AplEntryPath = [
	{ kind: 'priority', index: 0 },
	{ kind: 'group', group: 0, index: 1 },
];
const SECOND_STEP: AplEntryPath = [
	{ kind: 'priority', index: 1 },
	{ kind: 'step', index: 1 },
];

const ranEntry = (aplEntryPath: AplEntryPath, { changed = false, ...overrides }: Partial<CastLog> & { changed?: boolean } = {}) => ({
	log: castLog(12, { aplEntryPath, ...overrides }),
	aplRun: { rotation, changedSinceRun: () => changed },
});

/** A cast from the second entry of a group, with the path that reached it and the click hint. */
export const AplEntry: Story = { args: ranEntry(IN_GROUP) };

/** The rotation was edited after the run: the entry is drawn as it ran, and the hint says it changed. */
export const AplEntryChanged: Story = { args: ranEntry(IN_GROUP, { changed: true }) };

/** The second step of a sequence fired, so the whole sequence is drawn with that step marked. */
export const AplSequenceStep: Story = { args: ranEntry(SECOND_STEP) };

/** A pre-pull entry, with the time it happens at. */
export const AplPrepull: Story = { args: ranEntry([{ kind: 'prepull', index: 0 }]) };

/** Sixty hits under a sequence entry: the tooltip stops at 80% of the window's height and cuts off the last hits. */
export const AplEntryManyHits: Story = {
	args: ranEntry(SECOND_STEP, { damageDealtLogs: Array.from({ length: 60 }, (_, i) => chainHit(i % 3, 300 + i)) }),
};
