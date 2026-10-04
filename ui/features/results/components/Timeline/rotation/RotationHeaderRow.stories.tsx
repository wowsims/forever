import type { Meta, StoryObj } from '@storybook/react-vite';
import { MORTAL_STRIKE, PLAYER_HEADER_ROW, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { RotationHeaderRow } from './RotationHeaderRow';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationHeaderRow',
	component: RotationHeaderRow,
	tags: ['autodocs'],
	args: { row: PLAYER_HEADER_ROW },
	argTypes: { row: { control: false } },
	decorators: [
		(Story, { args }) => (
			<div className="w-max text-white" style={rotationTrackStyle(args.row)}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof RotationHeaderRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Player: Story = {};

export const WithIcon: Story = { args: { row: { ...PLAYER_HEADER_ROW, label: 'Mortal Strike', actionId: MORTAL_STRIKE } } };
