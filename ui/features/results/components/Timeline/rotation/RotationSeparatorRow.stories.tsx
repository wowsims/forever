import type { Meta, StoryObj } from '@storybook/react-vite';
import { rotationTrackStyle, TARGET_SEPARATOR_ROW } from '@testing/storybook/fixtures/results-views';

import { RotationSeparatorRow } from './RotationSeparatorRow';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationSeparatorRow',
	component: RotationSeparatorRow,
	tags: ['autodocs'],
	args: { row: TARGET_SEPARATOR_ROW },
	argTypes: { row: { control: false } },
	decorators: [
		(Story, { args }) => (
			<div className="w-max" style={rotationTrackStyle(args.row)}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof RotationSeparatorRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const BeforeTheTargets: Story = {};
