import type { Meta, StoryObj } from '@storybook/react-vite';
import { VERSUS_AREA_TYPES, VERSUS_BONUSES, VERSUS_TARGETS } from '@testing/storybook/fixtures/character-stats';

import { VersusBonusesView } from './VersusBonuses';

const meta = {
	title: 'Features/CharacterStats/Components/VersusBonuses',
	component: VersusBonusesView,
	tags: ['autodocs'],
	args: { bonuses: VERSUS_BONUSES, targets: VERSUS_TARGETS, areaTypes: VERSUS_AREA_TYPES },
	argTypes: {
		bonuses: { control: false },
		targets: { control: false },
		areaTypes: { control: false },
	},
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof VersusBonusesView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
