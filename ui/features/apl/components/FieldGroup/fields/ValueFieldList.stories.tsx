import { rotationSource } from '@features/apl/utils';
import { APLRotation, APLValue } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeAplStoryPlayer } from '@testing/storybook/fixtures/apl';
import { expect, userEvent, within } from 'storybook/test';

import { ValueFieldList } from './ValueFieldList';

const constant = (uuid: string, val: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val } } });

/** The operands of an `and`, held beside a story player whose rotation notifies on every write. */
const operandsOf = (initial: Array<APLValue>) => {
	const player = makeAplStoryPlayer(APLRotation.create());
	let vals = initial;
	const config = {
		id: 'story-vals',
		storeSubscribe: rotationSource,
		getValue: () => vals,
		setValue: (_subject: unknown, next: Array<APLValue | undefined>) => {
			vals = next as Array<APLValue>;
			player.touchRotation();
		},
	};
	return { player: player as never, config: config as never };
};

const meta = {
	title: 'Features/Apl/FieldGroup/Components/ValueFieldList',
	component: ValueFieldList,
	tags: ['autodocs'],
	args: operandsOf([constant('a', '1'), constant('b', '10s')]),
	argTypes: { player: { control: false }, config: { control: false } },
	decorators: [
		(Story, { args }) => (
			<SimHostProvider host={{ player: args.player, rootElem: document.body } as never}>
				<Story />
			</SimHostProvider>
		),
	],
} satisfies Meta<typeof ValueFieldList>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Two operands, each an inline row with a drag handle, the row menu and the add button below. */
export const Default: Story = {};

/** An empty list shows only the add button. */
export const Empty: Story = { args: operandsOf([]) };

/** The add button appends an empty operand. */
export const AddOperand: Story = {
	args: operandsOf([constant('a', '1')]),
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const before = canvas.getAllByTestId('apl-value-list-item').length;
		await userEvent.click(canvas.getByTestId('tree-add'));
		await expect(canvas.getAllByTestId('apl-value-list-item')).toHaveLength(before + 1);
	},
};
