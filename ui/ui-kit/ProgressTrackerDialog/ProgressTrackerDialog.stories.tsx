import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ProgressStepList } from './ProgressStepList';
import { ProgressTrackerDialog } from './ProgressTrackerDialog';
import type { ProgressTrackerHandle, ProgressTrackerProgress } from './types';

const SAMPLE: ProgressTrackerProgress = { title: '3 / 12 simulations complete', current: 30, total: 120 };

const pushSample = (handle: ProgressTrackerHandle | null) => handle?.setProgress(SAMPLE);

const STEPS = ['Current · 3:00 ±5s · 1 target', 'Patchwerk · 5:00 ±30s · 1 target', 'Council · 4:00 ±10s · 3 targets'];

const meta = {
	title: 'Kit/ProgressTrackerDialog',
	component: ProgressTrackerDialog,
	tags: ['autodocs'],
	args: {
		open: true,
		title: 'Calculate Stat Weights',
		state: { stage: 'running' },
	},
	argTypes: {
		state: { control: false },
		warning: { control: false },
		belowBar: { control: false },
		ref: { control: false },
	},
} satisfies Meta<typeof ProgressTrackerDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Running: Story = {
	args: { state: { stage: 'running', message: 'Preparing the batch.' } },
};

export const WithProgress: Story = {
	args: { hasProgressBar: true, ref: pushSample },
};

export const WithTimeRemaining: Story = {
	args: { hasProgressBar: true, ref: handle => handle?.setProgress({ ...SAMPLE, secondsRemaining: 95 }) },
};

export const WithStepsBelowBar: Story = {
	args: {
		hasProgressBar: true,
		ref: pushSample,
		belowBar: <ProgressStepList labels={STEPS} running={1} />,
		state: { stage: 'running', message: '1,200 / 4,000 iterations complete' },
	},
};

export const WithWarning: Story = {
	args: {
		title: 'Optimizing Gear',
		warning: <p className="m-0">Optimizing can take a while. You may cancel at any time.</p>,
	},
};

export const Cancelable: Story = {
	args: { onCancel: fn() },
	render: ({ onCancel, ...args }) => <ProgressTrackerDialog {...args} onCancel={() => onCancel?.()} />,
	play: async ({ args, canvasElement }) => {
		const dialog = await within(canvasElement).findByRole('dialog', { name: 'Calculate Stat Weights' });
		await expect(canvasElement.querySelector('[data-testid="storybook-portal"]')?.contains(dialog)).toBe(true);
		await userEvent.click(within(dialog).getByTestId('progress-tracker-modal-cancel-btn'));
		await expect(args.onCancel).toHaveBeenCalledTimes(1);
	},
};
