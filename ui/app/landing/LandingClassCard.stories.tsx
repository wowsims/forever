import { PlayerClasses } from '@sim/player/classes/index';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { LandingClassCard } from './LandingClassCard';

const meta = {
	title: 'App/Landing/Components/LandingClassCard',
	component: LandingClassCard,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
	args: { playerClass: PlayerClasses.Druid, index: 0 },
	argTypes: { playerClass: { control: false } },
	decorators: [
		Story => (
			<div className="grid grid-cols-landing-cards gap-6">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof LandingClassCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MultiSpec: Story = {};

export const SingleSpec: Story = { args: { playerClass: PlayerClasses.Mage } };
