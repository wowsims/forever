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
				talents={null}
				gearUnchanged={false}
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
				talents={null}
				gearUnchanged={false}
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
