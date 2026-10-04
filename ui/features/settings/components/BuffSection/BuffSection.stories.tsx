import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyIcon, StoryIconToggle } from '@testing/storybook/fixtures/settings_reforge_bulk';

import { BuffSection } from './BuffSection';

const BUFFS = [
	<StoryIconToggle key="kings" icon={storyIcon('spell_magic_magearmor', 'Blessing of Kings', 'spell', 20217)} initial={1} />,
	<StoryIconToggle key="motw" icon={storyIcon('spell_nature_regeneration', 'Mark of the Wild', 'spell', 9885)} states={3} initial={2} />,
	<StoryIconToggle key="fort" icon={storyIcon('spell_holy_wordfortitude', 'Power Word: Fortitude', 'spell', 10938)} />,
];

const DEBUFFS = [
	<StoryIconToggle key="sunder" icon={storyIcon('ability_warrior_sunder', 'Sunder Armor', 'spell', 11597)} initial={1} />,
	<StoryIconToggle key="faerie" icon={storyIcon('spell_nature_faeriefire', 'Faerie Fire', 'spell', 9907)} />,
];

const meta = {
	title: 'Features/Settings/BuffSection',
	component: BuffSection,
	tags: ['autodocs'],
	args: { section: 'raid_buffs', empty: false, children: BUFFS },
	argTypes: { section: { control: 'select', options: ['raid_buffs', 'party_buffs', 'debuffs'] }, children: { control: false } },
} satisfies Meta<typeof BuffSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const RaidBuffs: Story = {};

export const PartyBuffs: Story = { args: { section: 'party_buffs', children: BUFFS.slice(0, 1) } };

export const Debuffs: Story = { args: { section: 'debuffs', children: DEBUFFS } };

export const Empty: Story = { args: { empty: true } };
