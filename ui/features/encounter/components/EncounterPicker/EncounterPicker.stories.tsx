import type { Meta, StoryObj } from '@storybook/react-vite';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { MultiComboBox } from '@ui-kit/MultiComboBox';
import { NumberField } from '@ui-kit/TextField';
import { fn } from 'storybook/test';

import { EncounterPickerView } from './EncounterPicker';

const number = (id: string, label: string, value: number) => <NumberField id={id} label={label} value={value} onChange={fn()} />;

const EXECUTE = (
	<>
		{number('story-execute-20', 'Execute Duration (20%)', 20)}
		{number('story-execute-25', 'Execute Duration (25%)', 25)}
		{number('story-execute-35', 'Execute Duration (35%)', 35)}
	</>
);

const meta = {
	title: 'Features/Encounter/EncounterPicker',
	component: EncounterPickerView,
	tags: ['autodocs'],
	args: {
		duration: (
			<>
				{number('story-duration', 'Duration', 180)}
				{number('story-duration-variation', 'Duration +/-', 5)}
			</>
		),
		zone: (
			<ChoiceField
				id="story-zone"
				label="Zone"
				options={[
					{ value: 0, label: 'Custom' },
					{ value: 2717, label: 'Molten Core' },
				]}
				value={2717}
				onChange={fn()}
			/>
		),
		areaTypes: (
			<MultiComboBox
				id="story-area-types"
				label="Area Types"
				className="w-full flex-col items-stretch"
				values={[
					{ name: 'Cavernous', value: 8 },
					{ name: 'Volcanic', value: 9 },
				]}
				value={[8, 9]}
				onChange={fn()}
			/>
		),
		preset: (
			<ChoiceField
				id="story-preset"
				label="Encounter Preset"
				options={[
					{ value: -1, label: 'Custom' },
					{ value: 0, label: 'Molten Core/Ragnaros' },
				]}
				value={0}
				onChange={fn()}
			/>
		),
		targetInputs: null,
		onAdvancedClick: fn(),
		onDesignClick: fn(),
	},
	argTypes: {
		duration: { control: false },
		execute: { control: false },
		zone: { control: false },
		areaTypes: { control: false },
		preset: { control: false },
		allies: { control: false },
		minBaseDamage: { control: false },
		targetInputs: { control: false },
	},
} satisfies Meta<typeof EncounterPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithExecuteBands: Story = { args: { execute: EXECUTE } };

export const Tank: Story = { args: { minBaseDamage: number('story-min-base-damage', 'Min Base Damage', 4192) } };

export const WithoutDamageMetrics: Story = { args: { preset: undefined } };
