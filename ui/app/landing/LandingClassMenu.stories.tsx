import { PlayerClasses } from '@sim/player/classes/index';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import { LandingClassMenu } from './LandingClassMenu';

const meta = {
	title: 'App/Landing/Components/LandingClassMenu',
	component: LandingClassMenu,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
	args: { playerClass: PlayerClasses.Warrior },
	argTypes: { playerClass: { control: false } },
	decorators: [
		Story => (
			<div className="flex flex-wrap">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof LandingClassMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MultiSpec: Story = {
	play: async ({ canvasElement }) => {
		const dropdown = within(canvasElement).getByTestId('sim-link-dropdown');
		await userEvent.click(dropdown.querySelector<HTMLElement>(':scope > [data-testid="sim-link"]')!);
		await expect(dropdown.querySelectorAll('.ui-landing-sim-link-popup a[role="menuitem"]').length).toBe(Object.keys(PlayerClasses.Warrior.specs).length);
	},
};

export const SingleSpec: Story = { args: { playerClass: PlayerClasses.Mage } };
