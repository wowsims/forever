import type { Meta, StoryObj } from '@storybook/react-vite';
import { ItemCell } from '@ui-kit/ItemCell';
import { fn } from 'storybook/test';

import { BulkResultRowView } from './BulkResultRow';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';

const cell = (icon: string, name: string, ilvl: number) => (
	<ItemCell
		key={name}
		className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
		ilvl={ilvl}
		icon={<a className="ui-item-picker-icon" style={{ backgroundImage: `url('${ICONS}/${icon}.jpg')` }} />}
		name={name}
	/>
);

const meta = {
	title: 'Features/Bulk/BulkResults/BulkResultRow',
	component: BulkResultRowView,
	tags: ['autodocs'],
	args: {
		showDamage: true,
		avg: 1412.37,
		plusMinus: 11.8,
		isBase: false,
		delta: { text: '+42.10 (3.07%)', tone: 'positive' },
		significance: 'Difference is significantly different (Z = 5.214).',
		onEquip: fn(),
		children: [cell('inv_sword_39', 'Thunderfury, Blessed Blade of the Windseeker', 80), cell('inv_jewelry_ring_36', 'Band of Accuria', 78)],
	},
	argTypes: { children: { control: false }, delta: { control: 'object' } },
} satisfies Meta<typeof BulkResultRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Upgrade: Story = {};

export const Downgrade: Story = { args: { avg: 1351.02, delta: { text: '-19.25 (1.42%)', tone: 'negative' } } };

export const NotSignificant: Story = {
	args: { avg: 1372.9, delta: { text: '+2.63 (0.19%)', tone: null }, significance: 'Difference is not significantly different (Z = 0.311).' },
};

/** The current gear: labelled instead of compared, with no item cells and no equip button. */
export const CurrentGear: Story = { args: { avg: 1370.27, isBase: true, children: null } };

export const DamageHidden: Story = { args: { showDamage: false } };
