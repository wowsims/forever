import type { Meta, StoryObj } from '@storybook/react-vite';

import { MenuSection } from './MenuSection';

const meta = {
	title: 'Features/Gear/FiltersMenu/Components/MenuSection',
	component: MenuSection,
	tags: ['autodocs'],
	args: {
		name: 'Source',
		className: 'ui-filters-menu-section-bool-list',
		children: ['Crafting', 'Dungeon', 'Quest', 'Reputation', 'PvP', 'World Drop'].map(label => <span key={label}>{label}</span>),
	},
	argTypes: { children: { control: false } },
	decorators: [
		Story => (
			<div className="w-150">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof MenuSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoColumns: Story = {};

export const Stacked: Story = {
	args: { name: 'General', className: undefined, children: ['Min ilvl', 'Max ilvl'].map(label => <span key={label}>{label}</span>) },
};
