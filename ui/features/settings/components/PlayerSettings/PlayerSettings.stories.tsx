import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean, StoryEnum, storyIcon, StoryIconToggle } from '@testing/storybook/fixtures/settings_reforge_bulk';

import { PlayerSettingsView } from './PlayerSettings';

const RACES = [
	{ name: 'Human', value: 1 },
	{ name: 'Dwarf', value: 3 },
	{ name: 'Orc', value: 2 },
];

const PROFESSIONS = [
	{ name: 'None', value: 0 },
	{ name: 'Blacksmithing', value: 2 },
	{ name: 'Engineering', value: 4 },
];

const meta = {
	title: 'Features/Settings/PlayerSettings',
	component: PlayerSettingsView,
	tags: ['autodocs'],
	args: {
		iconPickers: [
			<StoryIconToggle key="shout" icon={storyIcon('ability_warrior_battleshout', 'Battle Shout', 'spell', 25289)} initial={1} />,
			<StoryIconToggle key="stance" icon={storyIcon('ability_racial_avatar', 'Berserker Stance', 'spell', 2458)} />,
		],
		race: <StoryEnum id="story-race" label="Race" values={RACES} />,
		inputs: [<StoryBoolean key="start" id="story-start-in-combat" label="Start in combat" layout="inline" />],
		professions: [
			<StoryEnum key="p1" id="story-profession-1" label="Profession 1" values={PROFESSIONS} initial={2} />,
			<StoryEnum key="p2" id="story-profession-2" label="Profession 2" values={PROFESSIONS} initial={4} />,
		],
	},
	argTypes: { iconPickers: { control: false }, race: { control: false }, inputs: { control: false }, professions: { control: false } },
} satisfies Meta<typeof PlayerSettingsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoSpecInputs: Story = { args: { iconPickers: [], inputs: [] } };
