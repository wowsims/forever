import type { Meta, StoryObj } from '@storybook/react-vite';

import { GameTooltip } from './GameTooltip';

const meta = {
	title: 'Kit/GameTooltip',
	component: GameTooltip,
	tags: ['autodocs'],
	args: {
		name: 'Summon Hawk',
		rank: 'Rank 1',
		lines: [
			['11% of base mana', '35 yd range'],
			['Instant', '6 sec cooldown'],
		],
		text: 'Command a hawk to dive-bomb your targeted enemy, dealing 133 Physical damage and continuing its assault for 18 sec. Only 2 hawks can be active at once.',
		testId: 'game-tooltip-story',
	},
	decorators: [
		Story => (
			<div className="ui-tooltip inline-block">
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
		lines: undefined,
		text: 'Increases damage dealt by your Revenge ability by 20%.',
		next: { label: 'Next rank:', text: 'Increases damage dealt by your Revenge ability by 40%.' },
	},
};

export const Buff: Story = {
	args: { name: 'Strider Kick', rank: undefined, lines: undefined, text: 'Movement speed increased by 30%.' },
};
