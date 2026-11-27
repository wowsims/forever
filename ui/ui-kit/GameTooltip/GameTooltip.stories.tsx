import { ItemQuality } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { GameTooltip } from './GameTooltip';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';

const meta = {
	title: 'Kit/GameTooltip',
	component: GameTooltip,
	tags: ['autodocs'],
	args: {
		name: 'Summon Hawk',
		rank: 'Rank 1',
		iconUrl: `${ICONS}/spell_nature_ravenform.jpg`,
		lines: [
			{ left: '11% of base mana', right: '35 yd range' },
			{ left: 'Instant', right: '6 sec cooldown' },
			{
				left: 'Command a hawk to dive-bomb your targeted enemy, dealing 133 Physical damage and continuing its assault for 18 sec. Only 2 hawks can be active at once.',
				tone: 'gold',
			},
		],
		testId: 'game-tooltip-story',
	},
	decorators: [
		Story => (
			<div className="ui-tooltip inline-block max-w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof GameTooltip>;

export default meta;

type Story = StoryObj<typeof meta>;

export const Spell: Story = {};

export const TalentWithNextRank: Story = {
	args: {
		name: 'Improved Revenge',
		rank: 'Rank 1/3',
		iconUrl: `${ICONS}/ability_warrior_revenge.jpg`,
		lines: [
			{ left: 'Increases damage dealt by your Revenge ability by 20%.', tone: 'gold' },
			{ left: 'Next rank:', spaced: true },
			{ left: 'Increases damage dealt by your Revenge ability by 40%.', tone: 'gold' },
		],
	},
};

export const Item: Story = {
	args: {
		name: 'Devilsaur Gauntlets',
		quality: ItemQuality.ItemQualityRare,
		rank: undefined,
		iconUrl: `${ICONS}/inv_gauntlets_26.jpg`,
		lines: [
			{ left: 'Item Level 58', tone: 'gold' },
			{ left: 'Binds when equipped' },
			{ left: 'Hands', right: 'Leather' },
			{ left: '103 Armor' },
			{ left: '+9 Stamina' },
			{ left: 'Requires Level 53' },
			{ left: 'Equip: Improves your chance to get a critical strike by 1%.', tone: 'green' },
			{ left: 'Equip: +28 Attack Power.', tone: 'green' },
			{ left: 'Devilsaur Armor (1/2)', tone: 'gold', spaced: true },
			{ left: 'Devilsaur Leggings', tone: 'grey', indent: true },
			{ left: 'Devilsaur Gauntlets', indent: true },
			{ left: '(2) Set: Improves your chance to hit by 2%.', tone: 'grey', spaced: true },
		],
	},
};
