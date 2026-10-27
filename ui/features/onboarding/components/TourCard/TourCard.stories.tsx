import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { TourCard } from './TourCard';

const meta = {
	title: 'Features/Onboarding/TourCard',
	component: TourCard,
	tags: ['autodocs'],
	args: {
		kind: 'step',
		title: 'Your equipment',
		body: 'Click a slot to pick its item, enchant and gems. Hovering an item shows its tooltip.',
		position: 3,
		total: 13,
		anchor: null,
		onNext: fn(),
		onPrevious: fn(),
		onSkip: fn(),
	},
	argTypes: {
		kind: { control: 'inline-radio', options: ['welcome', 'step', 'finish'] },
		anchor: { control: false },
		media: { control: false },
	},
} satisfies Meta<typeof TourCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Step: Story = {
	play: async ({ args, canvasElement }) => {
		const page = within(canvasElement.ownerDocument.body);
		await userEvent.click(await page.findByTestId('tour-card-next'));
		await expect(args.onNext).toHaveBeenCalledOnce();
		await userEvent.keyboard('{Escape}');
		await expect(args.onSkip).toHaveBeenCalledOnce();
	},
};

export const Welcome: Story = {
	args: {
		kind: 'welcome',
		media: <img className="size-16 self-center" src="/forever/assets/img/WoW-Simulator-Icon.png" alt="" />,
		title: 'Welcome to WoWSims',
		body: 'A short look around the page, from your gear to a finished sim. Move through it with the left and right arrow keys, or skip it any time; Sim Options can start it again.',
		onPrevious: undefined,
	},
};

/** A step that waits for the player to run the sim: the primary button passes it over instead. */
export const Waiting: Story = { args: { title: 'Run the sim', body: 'Press Simulate. The tour moves on by itself when the run finishes.', waiting: true } };

export const Finish: Story = {
	args: {
		kind: 'finish',
		title: 'You’re set',
		body: 'There is more to find on the page as you go. If you have a question or something looks wrong, ask us on Discord.',
	},
};
