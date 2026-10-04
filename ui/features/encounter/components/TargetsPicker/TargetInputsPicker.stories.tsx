import { InputType, Target as TargetProto, TargetInput } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStoryEncounter } from '@testing/storybook/fixtures/encounter';

import { TargetInputsPicker } from './TargetInputsPicker';

const target = TargetProto.create({
	targetInputs: [
		TargetInput.create({ label: 'Stacks', tooltip: 'Stacks at pull', inputType: InputType.Number, numberValue: 3 }),
		TargetInput.create({ label: 'Enraged', tooltip: 'Starts enraged', inputType: InputType.Bool, boolValue: true }),
		TargetInput.create({ label: 'Phase', tooltip: 'Starting phase', inputType: InputType.Enum, enumValue: 1, enumOptions: ['One', 'Two'] }),
	],
});

const meta = {
	title: 'Features/Encounter/TargetsPicker/Components/TargetInputsPicker',
	component: TargetInputsPicker,
	tags: ['autodocs'],
	args: { encounter: makeStoryEncounter([target]), targetIndex: 0 },
	argTypes: { encounter: { control: false }, targetIndex: { control: false } },
} satisfies Meta<typeof TargetInputsPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

/** One picker per declared input, dispatched on its type, with no list wrapper of its own. */
export const Default: Story = {};
