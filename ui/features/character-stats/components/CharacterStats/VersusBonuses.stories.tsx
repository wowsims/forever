import { AreaType } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { VERSUS_BONUSES } from '@testing/storybook/fixtures/character-stats';

import { VersusBonusesView } from './VersusBonuses';

const meta = {
	title: 'Features/CharacterStats/Components/VersusBonuses',
	component: VersusBonusesView,
	tags: ['autodocs'],
	args: { bonuses: VERSUS_BONUSES, areaTypes: [] },
	argTypes: {
		bonuses: { control: false },
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

export const InTheArea: Story = { args: { areaTypes: [AreaType.AreaTypeForestGrassland] } };
