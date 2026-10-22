import { ActionId } from '@sim/proto/action_id';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { HealingSpellTableView } from './HealingSpellTable';

const spells = [
	{
		id: ActionId.fromSpellId(2061, 7),
		name: 'Flash Heal',
		rank: 7,
		periodic: false,
		absorb: false,
		normal: { min: 812, avg: 913.5, max: 1015 },
		critical: { min: 1218, avg: 1370.25, max: 1522.5 },
		critChance: 18.4,
		customMetrics: { example_absorb: { min: 243.6, avg: 274.05, max: 304.5 } },
	},
	{
		id: ActionId.fromSpellId(139, 10),
		name: 'Renew',
		rank: 10,
		periodic: true,
		absorb: false,
		normal: { min: 345.6, avg: 345.6, max: 345.6 },
		critical: { min: 518.4, avg: 518.4, max: 518.4 },
		critChance: 18.4,
	},
	{
		id: ActionId.fromSpellId(17, 14),
		name: 'Power Word: Shield',
		rank: 14,
		periodic: false,
		absorb: true,
		normal: { min: 2200, avg: 2200, max: 2200 },
		critChance: 0,
	},
];

const meta = {
	title: 'Features/Results/HealingSpellTable',
	component: HealingSpellTableView,
	tags: ['autodocs'],
	args: {
		spells,
	},
} satisfies Meta<typeof HealingSpellTableView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Empty: Story = { args: { spells: [] } };
