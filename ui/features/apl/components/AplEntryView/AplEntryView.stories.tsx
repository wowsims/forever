import { resolveAplEntry } from '@features/apl/model/apl_entry';
import { APLAction, APLRotation, APLValue } from '@generated/proto/apl';
import { ActionID } from '@generated/proto/common';
import type { AplEntryPath } from '@sim/proto/combat_log';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { after, constant, sequence, waitFor } from '@testing/apl/builders';
import { withAplHost } from '@testing/storybook/fixtures/apl';

import { AplEntryView } from './AplEntryView';

const and = (...vals: Array<APLValue>) => APLValue.create({ value: { oneofKind: 'and', and: { vals } } });
const or = (...vals: Array<APLValue>) => APLValue.create({ value: { oneofKind: 'or', or: { vals } } });
const bloodthirst = (condition?: APLValue) =>
	APLAction.create({
		condition,
		action: { oneofKind: 'castSpell', castSpell: { spellId: ActionID.create({ rawId: { oneofKind: 'spellId', spellId: 23894 } }) } },
	});

const rotation = APLRotation.create({
	prepullActions: [{ action: waitFor('1s'), doAtValue: constant('-2s') }],
	priorityList: [
		{ action: bloodthirst(and(after('10s'), after('20s'))) },
		{
			action: waitFor(
				'1s',
				and(or(after('1s'), after('2s'), after('3s')), or(after('4s'), after('5s'), after('6s')), or(after('7s'), after('8s'), after('9s'))),
			),
		},
		{ action: sequence([waitFor('1s'), waitFor('2s'), waitFor('3s')]) },
	],
});

const entryAt = (...path: AplEntryPath) => ({ entry: resolveAplEntry(rotation, path)! });

const meta = {
	title: 'Features/Apl/AplEntryView',
	component: AplEntryView,
	tags: ['autodocs'],
	decorators: [withAplHost(() => rotation)],
	args: entryAt({ kind: 'priority', index: 1 }),
} satisfies Meta<typeof AplEntryView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Nine comparisons under an And of Ors run past the height cap, so the entry fades out at the bottom. */
export const LongCondition: Story = {};

/** Casting Bloodthirst once the fight is past 10 and 20 seconds: a Cast Spell row with an And condition. */
export const CastSpell: Story = {
	args: entryAt({ kind: 'priority', index: 0 }),
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};

/** The second step of the opener fired, so it is marked. */
export const SequenceStep: Story = {
	args: entryAt({ kind: 'priority', index: 2 }, { kind: 'step', index: 1 }),
};

/** A pre-pull row, with the time it happens at. */
export const PrePull: Story = { args: entryAt({ kind: 'prepull', index: 0 }) };
