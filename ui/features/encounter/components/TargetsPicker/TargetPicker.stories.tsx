import type { Meta, StoryObj } from '@storybook/react-vite';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { NumberField } from '@ui-kit/TextField';
import { fn } from 'storybook/test';

import { TargetPickerView } from './TargetPicker';

const select = (id: string, label: string, names: string[]) => (
	<ChoiceField id={id} label={label} options={names.map((name, value) => ({ value, label: name }))} value={0} onChange={fn()} />
);

const number = (id: string, label: string, value: number, float = false) => <NumberField id={id} label={label} value={value} float={float} onChange={fn()} />;

const IDENTITY = (
	<>
		{select('story-target-npc', 'NPC', ['Custom', 'Ragnaros', 'Nefarian'])}
		{select('story-target-preset-target', 'Preset', ['None'])}
		{select('story-target-level', 'Level', ['63', '62', '61', '60'])}
		{select('story-target-mob-type', 'Mob Type', ['Demon', 'Dragonkin', 'Elemental', 'Humanoid'])}
	</>
);

const STATS = (
	<>
		{number('story-target-armor', 'Armor', 3731)}
		{number('story-target-arcane-resistance', 'Arcane Resistance', 0)}
		{number('story-target-fire-resistance', 'Fire Resistance', 0)}
		{number('story-target-frost-resistance', 'Frost Resistance', 0)}
	</>
);

const THREAT = (
	<>
		{number('story-target-swing-speed', 'Swing Speed', 2, true)}
		{number('story-target-min-base-damage', 'Min Base Damage', 4192)}
		<BooleanPicker id="story-target-dual-wield" label="Dual Wield" value={false} onChange={fn()} />
		{select('story-target-spell-school', 'Spell School', ['Physical', 'Fire', 'Shadow'])}
	</>
);

const meta = {
	title: 'Features/Encounter/TargetsPicker/Components/TargetPicker',
	component: TargetPickerView,
	tags: ['autodocs'],
	args: { identity: IDENTITY, stats: STATS, threat: THREAT },
	argTypes: {
		identity: { control: false },
		stats: { control: false },
		threat: { control: false },
	},
} satisfies Meta<typeof TargetPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Without threat metrics the third column stays empty so the first two keep their width. */
export const Default: Story = { args: { threat: undefined } };

export const ThreatMetrics: Story = {};
