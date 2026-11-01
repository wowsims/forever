import type { Meta, StoryObj } from '@storybook/react-vite';
import { ItemCell, ItemCellIcon } from '@ui-kit/ItemCell';
import { fn } from 'storybook/test';

import { BulkResultRowView } from './BulkResultRow';
import { BulkResultTieGroup } from './BulkResultTieGroup';

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
	title: 'Features/Bulk/BulkResults/Components/BulkResultTieGroup',
	component: BulkResultTieGroup,
	tags: ['autodocs'],
	args: {
		children: [
			<BulkResultRowView
				key="sword"
				showDamage
				avg={1412.37}
				plusMinus={11.8}
				isBase={false}
				delta={{ text: '+42.10 (3.07%)', tone: 'positive' }}
				significance="Difference is not significant (Z = 0.412)."
				onEquip={fn()}>
				{cell('inv_sword_39', 'Thunderfury, Blessed Blade of the Windseeker', 80)}
			</BulkResultRowView>,
			<BulkResultRowView
				key="ring"
				showDamage
				avg={1409.91}
				plusMinus={11.6}
				isBase={false}
				delta={{ text: '+39.64 (2.89%)', tone: 'positive' }}
				significance="Difference is not significant (Z = 0.412)."
				onEquip={fn()}>
				{cell('inv_jewelry_ring_36', 'Band of Accuria', 78)}
			</BulkResultRowView>,
		],
	},
	argTypes: { children: { control: false } },
} satisfies Meta<typeof BulkResultTieGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

const listRow = (key: string, avg: number, beforeGroup = false) => (
	<BulkResultRowView
		key={key}
		showDamage
		avg={avg}
		plusMinus={11.8}
		isBase={false}
		delta={{ text: '+42.10 (3.07%)', tone: 'positive' }}
		significance="Difference is not significant (Z = 0.412)."
		beforeGroup={beforeGroup}
		onEquip={fn()}>
		{cell('inv_sword_39', 'Thunderfury, Blessed Blade of the Windseeker', 80)}
	</BulkResultRowView>
);

/** A row before a group draws no separator of its own, and two groups in a row each keep their box. */
export const InList: Story = {
	render: () => (
		<div>
			{listRow('first', 1500)}
			{listRow('before', 1450, true)}
			<BulkResultTieGroup>
				{listRow('a1', 1412)}
				{listRow('a2', 1410)}
			</BulkResultTieGroup>
			<BulkResultTieGroup>
				{listRow('b1', 1380)}
				{listRow('b2', 1378)}
			</BulkResultTieGroup>
			{listRow('after', 1300)}
			{listRow('last', 1250)}
		</div>
	),
};
