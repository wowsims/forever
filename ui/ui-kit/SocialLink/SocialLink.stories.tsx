import { SOCIALS } from '@sim/constants/other';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { SocialLink } from './SocialLink';

const [discord, github, patreon] = SOCIALS;

const meta = {
	title: 'Kit/SocialLink',
	component: SocialLink,
	tags: ['autodocs'],
	args: { social: discord },
	argTypes: { social: { control: false } },
} satisfies Meta<typeof SocialLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Discord: Story = {};
export const Github: Story = { args: { social: github } };
export const WithLabel: Story = { args: { social: patreon } };
export const Row: Story = {
	render: () => (
		<div className="flex items-center gap-4">
			{SOCIALS.map(social => (
				<SocialLink key={social.key} social={social} />
			))}
		</div>
	),
};
