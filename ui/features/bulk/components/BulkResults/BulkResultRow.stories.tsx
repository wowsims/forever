import { warriorTalentsConfig } from '@sim/talents/warrior';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { ItemCell, ItemCellIcon } from '@ui-kit/ItemCell';
import { fn } from 'storybook/test';

import { BulkResultRowView } from './BulkResultRow';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';

const cell = (icon: string, name: string, ilvl: number) => (
	<ItemCell
		key={name}
		className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
		icon={
			<ItemCellIcon ilvl={ilvl}>
				<a className="ui-item-picker-icon" style={{ backgroundImage: `url('${ICONS}/${icon}.jpg')` }} />
			</ItemCellIcon>
		}
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
	argTypes: {
		children: { control: false },
		delta: { control: 'object' },
		talents: { control: 'object' },
		rotation: { control: 'object' },
		settings: { control: 'object' },
	},
} satisfies Meta<typeof BulkResultRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Upgrade: Story = {};

export const Downgrade: Story = { args: { avg: 1351.02, delta: { text: '-19.25 (1.40%)', tone: 'negative' } } };

export const NotSignificant: Story = {
	args: { avg: 1372.9, delta: { text: '+2.63 (0.19%)', tone: null }, significance: 'Difference is not significantly different (Z = 0.311).' },
};

/** The current gear: labelled instead of compared, with no item cells and no equip button. */
export const CurrentGear: Story = { args: { avg: 1370.27, isBase: true, children: null } };

export const DamageHidden: Story = { args: { showDamage: false } };

/** A Builds row: its gear setup with the enchants it changes, then its talents, rotation and settings above the gear. */
export const Build: Story = {
	args: {
		gearSetup: 'P1 Fury',
		enchants: ['Back: Enchant Cloak - Greater Agility'],
		talents: { label: 'Arms', talentsString: '30305001302-05050005025012251', config: warriorTalentsConfig },
		rotation: { label: 'Opener' },
		settings: { label: 'Pre-BiS - Fury' },
	},
};

/** A Builds row that wears the current gear with other talents: no gear changes, but a difference and an equip button for the build. */
export const CurrentGearOtherTalents: Story = {
	args: {
		avg: 1388.4,
		delta: { text: '+18.13 (1.32%)', tone: 'positive' },
		gearSetup: 'Current gear',
		talents: { label: 'Arms', talentsString: '30305001302-05050005025012251', config: warriorTalentsConfig },
		rotation: { label: 'Current rotation' },
		settings: { label: 'Current settings' },
		gearUnchanged: true,
		children: null,
	},
};

const FURY_TALENTS = { label: 'Fury', talentsString: '-0505000540501', config: warriorTalentsConfig };

/** The best row standing as the reference when the batch did not sim the current setup: no delta, but its items and equip button. */
export const BestAsReference: Story = {
	args: {
		isReference: true,
		gearSetup: 'Pre-BiS (Fury)',
		talents: FURY_TALENTS,
		rotation: { label: 'Fury APL' },
		settings: { label: 'Pre-BiS - Fury' },
	},
};

/** A ranking with the Current bundle unticked: the best row is the reference, and every other row shows its loss against it. */
export const RankingAgainstBest: Story = {
	render: args => (
		<div>
			<BulkResultRowView
				{...args}
				isReference
				gearSetup="Pre-BiS (Fury)"
				talents={FURY_TALENTS}
				rotation={{ label: 'Fury APL' }}
				settings={{ label: 'Pre-BiS - Fury' }}
			/>
			<BulkResultRowView
				{...args}
				avg={1405.61}
				delta={{ text: '-6.76 (0.48%)', tone: null }}
				significance="Difference is not significantly different (Z = 0.402)."
				gearSetup="Pre-BiS (Fury)"
				enchants={['Back: Enchant Cloak - Greater Agility']}
				talents={FURY_TALENTS}
				rotation={{ label: 'Fury APL' }}
				settings={{ label: 'Pre-BiS - Fury' }}
			/>
			<BulkResultRowView
				{...args}
				avg={1351.02}
				delta={{ text: '-61.35 (4.34%)', tone: 'negative' }}
				significance="Difference is significantly different (Z = 3.671)."
				gearSetup="Pre-BiS (Arms)"
				talents={{ label: 'Arms', talentsString: '30305001302-05050005025012251', config: warriorTalentsConfig }}
				rotation={{ label: 'Arms APL' }}
				settings={{ label: 'Pre-BiS - Arms' }}
			/>
		</div>
	),
};

/** The widest DPS line the column is sized for, next to a full set of seventeen slots wrapping across the remaining width. */
export const WidestValueAllSlots: Story = {
	args: {
		avg: 9999.99,
		plusMinus: 1,
		delta: { text: '+1,234.56 (14.08%)', tone: 'positive' },
		children: Array.from({ length: 17 }, (_, index) => cell(index % 2 ? 'inv_jewelry_ring_36' : 'inv_sword_39', `Slot ${index + 1}`, 78)),
	},
};
