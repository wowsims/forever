import { APLListItem, APLRotation, APLValue, APLValueCompare_ComparisonOperator as Op } from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import type { AplTestPlayer } from '@testing/apl/player';
import { makeAplStoryPlayer, withAplScope } from '@testing/storybook/fixtures/apl';
import { makeAplRotation } from '@testing/storybook/fixtures/protos';

import { AplListItem } from './AplListItem';

const constant = (val: string) => APLValue.create({ value: { oneofKind: 'const', const: { val } } });

/** Once the fight is 10 seconds in, wait 1.5 seconds. Nothing in it names a spell, so it renders without the database. */
const waitAfterTen = (hide = false) =>
	APLListItem.create({
		hide,
		action: {
			condition: {
				value: { oneofKind: 'cmp', cmp: { op: Op.OpGt, lhs: { value: { oneofKind: 'currentTime', currentTime: {} } }, rhs: constant('10s') } },
			},
			action: { oneofKind: 'wait', wait: { duration: constant('1.5s') } },
		},
	});

const firstRow = (player: AplTestPlayer) => ({
	getValue: () => player.aplRotation.priorityList[0],
	setValue: (_subject: unknown, next: APLListItem) => {
		player.aplRotation.priorityList[0] = next;
		player.touchRotation();
	},
});

const rowOf = (rotation: APLRotation) => {
	const player = makeAplStoryPlayer(rotation);
	return { player: player as never, config: firstRow(player) as never };
};

const meta = {
	title: 'Features/Apl/AplListItem',
	component: AplListItem,
	tags: ['autodocs'],
	args: rowOf(APLRotation.create({ priorityList: [waitAfterTen()] })),
} satisfies Meta<typeof AplListItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** A row the eye has disabled stays on screen, greyed out. */
export const Hidden: Story = { args: rowOf(APLRotation.create({ priorityList: [waitAfterTen(true)] })) };

/** A pre-pull row, which offers the pre-pull set of kinds and has no condition picker. */
export const PrePull: Story = {
	args: rowOf(APLRotation.create({ priorityList: [waitAfterTen()] })),
	decorators: [withAplScope({ isPrepull: true })],
};

const waitFor = (val: string) => ({ action: { oneofKind: 'wait' as const, wait: { duration: constant(val) } } });

/** A sequence keeps its own condition; the actions it runs have none. */
export const Sequence: Story = {
	args: rowOf(
		APLRotation.create({
			priorityList: [
				APLListItem.create({ action: { action: { oneofKind: 'sequence', sequence: { name: 'opener', actions: [waitFor('1s'), waitFor('2s')] } } } }),
			],
		}),
	),
};

/** Casting Bloodthirst: the spell picker names and paints the spell. */
export const CastSpell: Story = {
	args: rowOf(makeAplRotation()),
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};
