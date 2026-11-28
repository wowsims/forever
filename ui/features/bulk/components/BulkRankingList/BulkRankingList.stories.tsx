import type { Meta, StoryObj } from '@storybook/react-vite';
import { ItemCell, ItemCellIcon } from '@ui-kit/ItemCell';
import { fn } from 'storybook/test';

import { BulkResultRowView } from '../BulkResults/BulkResultRow';
import { BulkRankingList } from './BulkRankingList';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';

const row = (avg: number) => (
	<BulkResultRowView
		showDamage
		avg={avg}
		plusMinus={11.8}
		isBase={false}
		delta={{ text: '+42.10 (3.07%)', tone: 'positive' }}
		significance="Difference is not significant (Z = 0.412)."
		onEquip={fn()}>
		<ItemCell
			className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
			icon={
				<ItemCellIcon ilvl={80}>
					<a className="ui-item-picker-icon" style={{ backgroundImage: `url('${ICONS}/inv_sword_39.jpg')` }} />
				</ItemCellIcon>
			}
			name="Thunderfury, Blessed Blade of the Windseeker"
		/>
	</BulkResultRowView>
);

const RankingList = BulkRankingList<number>;

const meta = {
	title: 'Features/Bulk/BulkRankingList',
	component: RankingList,
	tags: ['autodocs'],
	args: {
		chains: [[1500], [1450], [1412, 1410], [1380, 1378], [1300], [1250]],
		renderRow: row,
	},
	argTypes: { chains: { control: false }, renderRow: { control: false } },
} satisfies Meta<typeof RankingList>;

export default meta;
type Story = StoryObj<typeof meta>;

/** A row before a group draws no separator of its own, and two groups in a row each keep their box. */
export const WithTieGroups: Story = {};
