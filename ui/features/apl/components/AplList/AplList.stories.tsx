import { GroupList } from '@features/apl/components/GroupList';
import { PriorityPane } from '@features/apl/components/PriorityPane';
import { VariablesList } from '@features/apl/components/VariablesList';
import {
	APLAction,
	APLActionMove,
	APLGroup,
	APLListItem,
	APLPrepullAction,
	APLRotation,
	APLValue,
	APLValueCompare_ComparisonOperator as Op,
} from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withAplHost } from '@testing/storybook/fixtures/apl';
import { expect, fireEvent, waitFor, within } from 'storybook/test';

const constant = (val: string) => APLValue.create({ value: { oneofKind: 'const', const: { val } } });
const afterTime = (val: string) =>
	APLValue.create({ value: { oneofKind: 'cmp', cmp: { op: Op.OpGt, lhs: { value: { oneofKind: 'currentTime', currentTime: {} } }, rhs: constant(val) } } });
const wait = (val: string) => APLAction.create({ action: { oneofKind: 'wait', wait: { duration: constant(val) } } });

/**
 * Three levels: the pane's sections, their rows, and the lists inside a row — an `and` condition's
 * operands and a sequence's actions. Nothing in it names a spell, so it renders without the database.
 */
const threeLevels = () =>
	APLRotation.create({
		prepullActions: [
			APLPrepullAction.create({
				action: { action: { oneofKind: 'move', move: APLActionMove.create({ rangeFromTarget: constant('5') }) } },
				doAtValue: constant('-2s'),
			}),
		],
		priorityList: [
			APLListItem.create({
				action: {
					condition: {
						value: { oneofKind: 'and', and: { vals: [afterTime('10s'), { value: { oneofKind: 'not', not: { val: afterTime('60s') } } }] } },
					},
					action: wait('1.5s').action,
				},
			}),
			APLListItem.create({ action: { action: { oneofKind: 'sequence', sequence: { name: 'opener', actions: [wait('1s'), wait('2s')] } } } }),
			APLListItem.create({ hide: true, action: wait('3s') }),
		],
		groups: [
			APLGroup.create({ name: 'burst', actions: [APLListItem.create({ action: wait('1s') }), APLListItem.create({ action: wait('2s') })] }),
			APLGroup.create({ name: 'filler', actions: [] }),
		],
		valueVariables: [{ name: 'late_fight', value: afterTime('120s') }],
	});

const meta = {
	title: 'Features/Apl/AplList',
	component: PriorityPane,
	tags: ['autodocs'],
	decorators: [withAplHost(threeLevels)],
} satisfies Meta<typeof PriorityPane>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The priority pane: the pre-pull and priority sections share one tree, so a row can move between them. */
export const PriorityPaneTree: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const priority = canvas.getByTestId('apl-priority-list-picker-root');
		const rows = within(priority).getAllByTestId('tree-item');
		await expect(rows.length).toBeGreaterThanOrEqual(3);
		const toggle = rows[0].querySelector<HTMLButtonElement>('button[aria-expanded]')!;
		fireEvent.click(toggle);
		await waitFor(() => expect(rows[0].hasAttribute('data-open')).toBe(false));
		fireEvent.click(toggle);
		await waitFor(() => expect(rows[0].hasAttribute('data-open')).toBe(true));
	},
};

/** The action groups pane: each group is an item with its actions as children, and an add button inside them. */
export const ActionGroups: Story = { render: () => <GroupList /> };

/** The variables pane: each variable's value is the item's body. */
export const Variables: Story = { render: () => <VariablesList /> };

/** At a phone's width the rows stack by the tree's own width. */
export const Narrow: Story = {
	decorators: [
		Story => (
			<div className="w-104">
				<Story />
			</div>
		),
	],
};
