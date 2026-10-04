import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, within } from 'storybook/test';

import { WowheadAnchor } from './WowheadAnchor';

const HREF = 'https://www.wowhead.com/classic/item=19019';

const meta = {
	title: 'Kit/WowheadAnchor',
	component: WowheadAnchor,
	tags: ['autodocs'],
	args: {
		href: HREF,
		icon: true,
		children: 'Thunderfury, Blessed Blade of the Windseeker',
		className: 'text-quality-legendary',
		testId: 'story-anchor',
	},
	argTypes: {
		href: { control: 'text' },
		icon: { control: 'boolean' },
		target: { control: 'select', options: [undefined, '_blank'] },
		onActivate: { control: false },
		children: { control: 'text' },
	},
} satisfies Meta<typeof WowheadAnchor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
	play: async ({ canvasElement, args }) => {
		const anchor = within(canvasElement).getByTestId('story-anchor');
		await expect(anchor.tagName).toBe('A');
		if (args.href) await expect(anchor.getAttribute('rel')).toBe('noopener noreferrer');
	},
};
