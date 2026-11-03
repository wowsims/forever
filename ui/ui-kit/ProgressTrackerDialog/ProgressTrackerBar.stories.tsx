import type { Meta, StoryObj } from '@storybook/react-vite';

import { ProgressTrackerBar } from './ProgressTrackerBar';

const meta = {
	title: 'Kit/ProgressTrackerBar',
	component: ProgressTrackerBar,
	tags: ['autodocs'],
	args: { running: true, progress: 0.4, title: 'Running high iteration rounds', readout: '10/26' },
	argTypes: {
		variant: { control: 'inline-radio', options: ['bar', 'line'] },
		progress: { control: { type: 'range', min: 0, max: 1, step: 0.05 } },
	},
} satisfies Meta<typeof ProgressTrackerBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Bar: Story = {};

/** Done, at 40% and not started, as the lines along the bottom of three table rows. */
export const Line: Story = {
	args: { variant: 'line', title: undefined, readout: undefined, label: 'Run progress' },
	render: args => (
		<div className="flex w-80 flex-col gap-4 bg-surface p-4">
			{[1, args.progress, 0].map((progress, index) => (
				<div key={index} className="relative h-8 border-b border-b-white/20">
					<ProgressTrackerBar {...args} progress={progress} className="absolute inset-x-0 -bottom-px" />
				</div>
			))}
		</div>
	),
};
