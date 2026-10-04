import { rotationSource } from '@features/apl/utils';
import { APLAction, APLRotation, APLValue } from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeAplStoryPlayer } from '@testing/storybook/fixtures/apl';

import { ActionFieldList } from './ActionFieldList';

const wait = (val: string) =>
	APLAction.create({ action: { oneofKind: 'wait', wait: { duration: APLValue.create({ value: { oneofKind: 'const', const: { val } } }) } } });

/** A sequence's actions, held beside a story player whose rotation notifies on every write. */
const actionsOf = (initial: Array<APLAction>) => {
	const player = makeAplStoryPlayer(APLRotation.create());
	let actions = initial;
	const config = {
		id: 'story-actions',
		storeSubscribe: rotationSource,
		getValue: () => actions,
		setValue: (_subject: unknown, next: Array<APLAction>) => {
			actions = next;
			player.touchRotation();
		},
	};
	return { player: player as never, config: config as never };
};

const meta = {
	title: 'Features/Apl/FieldGroup/Components/ActionFieldList',
	component: ActionFieldList,
	tags: ['autodocs'],
	args: actionsOf([wait('1s'), wait('1.5s')]),
	argTypes: { player: { control: false }, config: { control: false } },
} satisfies Meta<typeof ActionFieldList>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Two waits in a sequence, each an inline row with a drag handle and the row menu. */
export const Default: Story = {};

/** An empty sequence shows only the add button. */
export const Empty: Story = { args: actionsOf([]) };
