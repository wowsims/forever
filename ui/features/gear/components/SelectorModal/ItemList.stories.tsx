import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ItemListToolbarView } from './ItemList';

const toggle = (name: string, label: string, value: boolean) => ({
	id: `show-${name}-selector`,
	testId: `selector-modal-show-${name}`,
	label,
	value,
	onChange: fn(),
});

const meta = {
	title: 'Features/Gear/SelectorModal/Components/ItemList',
	component: ItemListToolbarView,
	tags: ['autodocs'],
	args: {
		search: '',
		onSearchChange: fn(),
		phase: 1,
		onPhaseChange: fn(),
		toggles: [toggle('ep-values', 'Show EP', true)],
		remove: { label: 'Unequip Item', onClick: fn() },
	},
	argTypes: { filtersMenu: { control: false } },
	decorators: [
		Story => (
			<div className="w-4xl">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ItemListToolbarView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EnchantsTab: Story = {
	args: { remove: { label: 'Remove Enchant', onClick: fn() } },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('selector-modal-remove-button'));
		await expect(args.remove?.onClick).toHaveBeenCalledOnce();
	},
};

export const MainHandItems: Story = {
	args: {
		search: 'sword',
		filtersMenu: <span />,
		onOpenFilters: fn(),
		toggles: [
			toggle('1h-weapons', 'Show 1-Handed Weapons', true),
			toggle('2h-weapons', 'Show 2-Handed Weapons', false),
			toggle('ep-values', 'Show EP', true),
		],
	},
};
