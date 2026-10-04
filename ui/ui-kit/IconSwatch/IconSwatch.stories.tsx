import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { expect, within } from 'storybook/test';

import { IconSwatch } from './IconSwatch';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, spellId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/spell=${spellId}`,
});

const BLOODLUST = icon('spell_nature_bloodlust', 'Bloodlust', 2825);
const IMPROVED_BLOODLUST = icon('spell_nature_bloodlust', 'Improved Bloodlust', 2826);
const GREATER_BLOODLUST = icon('spell_nature_bloodlust', 'Greater Bloodlust', 2827);

const meta = {
	title: 'Kit/IconSwatch',
	component: IconSwatch,
	tags: ['autodocs'],
	args: {
		icon: BLOODLUST,
		href: BLOODLUST.href,
		testId: 'story-swatch',
	},
	argTypes: {
		icon: { control: false },
		badges: { control: false },
		active: { control: 'boolean' },
		disabled: { control: 'boolean' },
		text: { control: 'text' },
	},
} satisfies Meta<typeof IconSwatch>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Plain: Story = {
	play: async ({ canvasElement }) => {
		const swatch = within(canvasElement).getByTestId('story-swatch');
		await expect(swatch.tagName).toBe('A');
		await expect(swatch.getAttribute('href')).toBe(BLOODLUST.href);
		await expect(swatch.dataset.whtticon).toBe('false');
	},
};
export const Active: Story = { args: { active: true } };
export const WithText: Story = { args: { active: true, text: '4' } };
export const WithBadges: Story = {
	args: {
		active: true,
		badges: (
			<>
				<IconSwatch icon={IMPROVED_BLOODLUST} href={IMPROVED_BLOODLUST.href} active className="absolute right-0 bottom-0 size-4 min-w-4" />
				<IconSwatch
					icon={GREATER_BLOODLUST}
					href={GREATER_BLOODLUST.href}
					className="absolute bottom-0 left-px size-4 min-w-4 border-gray-600 grayscale"
				/>
			</>
		),
	},
	play: async ({ canvasElement }) => {
		const swatch = within(canvasElement).getByTestId('story-swatch');
		await expect(swatch.parentElement?.querySelectorAll(':scope > a')).toHaveLength(3);
		await expect(canvasElement.querySelectorAll('a a')).toHaveLength(0);
	},
};
