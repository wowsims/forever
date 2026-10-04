import type { Meta, StoryObj } from '@storybook/react-vite';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconPicker } from '@ui-kit/IconPicker';
import { fn } from 'storybook/test';

import { IconRow } from './IconRow';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, itemId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/item=${itemId}`,
});

const CONSUMES = [
	icon('inv_potion_62', 'Flask of the Titans', 13510),
	icon('inv_potion_97', 'Elixir of the Mongoose', 13452),
	icon('inv_misc_food_64', 'Smoked Desert Dumplings', 20452),
	icon('inv_stone_02', 'Dense Sharpening Stone', 12404),
];

const meta = {
	title: 'Kit/IconRow',
	component: IconRow,
	tags: ['autodocs'],
	argTypes: {
		className: { control: 'text' },
		children: { control: false },
	},
} satisfies Meta<typeof IconRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	args: {
		children: CONSUMES.map((consume, index) => <IconPicker key={consume.name} states={2} icon={consume} value={index % 2} onChange={fn()} />),
	},
};
