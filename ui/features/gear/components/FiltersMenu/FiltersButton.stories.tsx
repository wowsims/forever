import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { FiltersButton } from './FiltersButton';

const meta = {
	title: 'Features/Gear/FiltersMenu/Components/FiltersButton',
	component: FiltersButton,
	tags: ['autodocs'],
	args: { active: 0, onClick: fn() },
} satisfies Meta<typeof FiltersButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

// Two filters narrow the list, so the label counts them.
export const Active: Story = { args: { active: 2 } };
