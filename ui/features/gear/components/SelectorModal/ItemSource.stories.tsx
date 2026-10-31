import { Profession } from '@generated/proto/common';
import { DungeonDifficulty, UIItemSource } from '@generated/proto/ui';
import type { Sim } from '@sim/sim';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIItem } from '@testing/storybook/fixtures/protos';

import { ItemSource } from './ItemSource';

const MOLTEN_CORE = { id: 2717, name: 'Molten Core' };
const RAGNAROS = { id: 11502, name: 'Ragnaros' };

const sim = { db: { getZone: () => MOLTEN_CORE, getNpc: () => RAGNAROS } } as unknown as Sim;

const withSource = (source: UIItemSource['source']) => makeUIItem({ sources: [UIItemSource.create({ source })] });

const meta = {
	title: 'Features/Gear/SelectorModal/Components/ItemSource',
	component: ItemSource,
	tags: ['autodocs'],
	args: {
		item: withSource({
			oneofKind: 'drop',
			drop: { difficulty: DungeonDifficulty.DifficultyRaid, npcId: RAGNAROS.id, zoneId: MOLTEN_CORE.id, otherName: '', category: '' },
		}),
		sim,
	},
	argTypes: { item: { control: false }, sim: { control: false } },
} satisfies Meta<typeof ItemSource>;

export default meta;
type Story = StoryObj<typeof meta>;

export const BossDrop: Story = {};

export const Crafted: Story = { args: { item: withSource({ oneofKind: 'crafted', crafted: { profession: Profession.Blacksmithing, spellId: 16991 } }) } };

export const SoldBy: Story = {
	args: { item: withSource({ oneofKind: 'soldBy', soldBy: { npcId: 12944, npcName: 'Lokhtos Darkbargainer', zoneId: 1584 } }) },
};
