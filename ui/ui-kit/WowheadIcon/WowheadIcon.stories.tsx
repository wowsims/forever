import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect } from 'storybook/test';

import { WowheadIcon } from './WowheadIcon';

const meta = {
	title: 'Kit/WowheadIcon',
	component: WowheadIcon,
	tags: ['autodocs'],
	args: {
		iconUrl: 'https://wow.zamimg.com/images/wow/icons/large/ability_hunter_killcommand.jpg',
		href: 'https://www.wowhead.com/forever/spell=34026',
		label: 'Kill Command',
		className: 'h-6 w-6',
		testId: 'wowhead-icon-story',
	},
	argTypes: {
		as: { control: 'select', options: ['a', 'div'] },
		wowhead: { control: 'object' },
		children: { control: false },
	},
} satisfies Meta<typeof WowheadIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Anchor: Story = {};
export const AnchorWithTooltip: Story = {
	args: { wowhead: { 'data-wowhead': 'spell=34026&domain=forever&dataEnv=16' } },
	play: async ({ canvasElement }) => {
		const icon = canvasElement.querySelector('[data-testid="wowhead-icon-story"]')!;
		await expect(icon.tagName).toBe('A');
		await expect(icon.getAttribute('data-wowhead')).toBe('spell=34026&domain=forever&dataEnv=16');
	},
};
export const Div: Story = { args: { as: 'div', href: undefined } };
export const DivWithTooltip: Story = {
	args: { as: 'div', href: undefined, wowhead: { 'data-wowhead': 'item=19019&domain=forever&dataEnv=16' } },
	play: async ({ canvasElement }) => {
		const icon = canvasElement.querySelector<HTMLElement>('[data-testid="wowhead-icon-story"]')!;
		await expect(icon.tagName).toBe('DIV');
		await expect(icon.style.backgroundImage).toContain('ability_hunter_killcommand.jpg');
	},
};
