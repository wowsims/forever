import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { TourStep } from '../../model/tour_steps';
import { OnboardingTour } from './OnboardingTour';

const STEPS: ReadonlyArray<TourStep> = [
	{ id: 'welcome' },
	{ id: 'gear', target: '[data-testid="story-gear"]' },
	{ id: 'simulate', target: '[data-testid="story-simulate"]' },
	{ id: 'finish' },
];

const WORDS: Record<string, { title: string; body: string }> = {
	welcome: { title: 'Welcome to WoWSims', body: 'A short look around the page.' },
	gear: { title: 'Your equipment', body: 'Click a slot to pick its item, enchant and gems.' },
	simulate: { title: 'Run the sim', body: 'Press Simulate to run it.' },
	finish: { title: 'You’re set', body: 'That is the whole page.' },
};

const FakePage = () => (
	<div className="flex h-120 gap-8 p-8">
		<div className="flex w-60 flex-col gap-4 border border-surface-border bg-surface p-4" data-testid="story-gear">
			Gear slots
		</div>
		<div className="mt-auto ml-auto border border-surface-border bg-surface p-4" data-testid="story-simulate">
			Simulate
		</div>
	</div>
);

const meta = {
	title: 'Features/Onboarding/OnboardingTour',
	component: OnboardingTour<TourStep>,
	args: {
		steps: STEPS,
		text: (step: TourStep) => WORDS[step.id],
		onEnd: fn(),
	},
	argTypes: { steps: { control: false }, text: { control: false }, prepare: { control: false } },
	parameters: { layout: 'fullscreen' },
	render: args => (
		<>
			<FakePage />
			<OnboardingTour {...args} />
		</>
	),
} satisfies Meta<typeof OnboardingTour<TourStep>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** Into the tour: the page dims except for a ring around the gear, and the card sits beside it. */
export const RingedStep: Story = {
	parameters: { smoke: false, smokeReason: 'The tour rings only an element with a laid-out box, and happy-dom lays nothing out.' },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement.ownerDocument.body);
		await userEvent.click(canvas.getByTestId('tour-card-next'));
		await expect(await canvas.findByTestId('tour-ring')).toBeTruthy();
	},
};
