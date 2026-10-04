import { APLAction, APLListItem, APLRotation, APLValue, APLValueVariable } from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeAplStoryPlayer } from '@testing/storybook/fixtures/apl';

import { GroupVariablesField } from './GroupVariablesField';

const placeholder = (name: string) =>
	APLListItem.create({
		action: APLAction.create({ condition: APLValue.create({ value: { oneofKind: 'variablePlaceholder', variablePlaceholder: { name } } }) }),
	});

/** A reference to group `g1`, which declares the placeholders `names`; the rotation offers two variables. */
const referenceTo = (names: Array<string>) => {
	const player = makeAplStoryPlayer(
		APLRotation.create({
			groups: [{ name: 'g1', actions: names.map(placeholder), variables: [] }],
			valueVariables: [APLValueVariable.create({ name: 'boss' }), APLValueVariable.create({ name: 'add' })],
		}),
	);
	const reference = { groupName: 'g1', variables: [] as Array<unknown> };
	return {
		player: player as never,
		config: { id: 'story-group-variables', getValue: () => reference.variables, setValue: () => player.touchRotation() } as never,
		groupNameField: 'groupName',
		getParentValue: () => reference,
	};
};

const meta = {
	title: 'Features/Apl/FieldGroup/Components/GroupVariablesField',
	component: GroupVariablesField,
	tags: ['autodocs'],
	args: referenceTo(['target', 'focus']),
	argTypes: { player: { control: false }, config: { control: false }, getParentValue: { control: false } },
} satisfies Meta<typeof GroupVariablesField>;

export default meta;
type Story = StoryObj<typeof meta>;

/** One row per placeholder the group declares, each picking the rotation variable passed for it. */
export const Default: Story = {};
