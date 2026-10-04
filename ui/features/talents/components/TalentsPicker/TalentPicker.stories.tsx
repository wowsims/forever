import { mageTalentsConfig } from '@sim/talents/mage';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { TalentPicker } from './TalentPicker';

const TALENT = { ...mageTalentsConfig[0].talents[1], location: { rowIdx: 0, colIdx: 0 } };

const meta = {
	title: 'Features/Talents/TalentsPicker/Components/TalentPicker',
	component: TalentPicker,
	tags: ['autodocs'],
	args: { config: TALENT, points: 0, canAdd: true, allPointsSpent: false, onSetPoints: fn() },
	argTypes: { config: { control: false } },
	parameters: { smoke: false, smokeReason: 'The talent fills its icon from the database through useActionId' },
	decorators: [
		Story => (
			<div className="grid w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TalentPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Available: Story = {};

export const Disabled: Story = { args: { canAdd: false, allPointsSpent: true } };

export const Partial: Story = { args: { points: 2 } };

export const Full: Story = { args: { points: TALENT.maxPoints, canAdd: false } };
