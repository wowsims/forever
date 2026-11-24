import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsItemListFixture } from '@testing/storybook/fixtures/bulk-drops';
import { fn } from 'storybook/test';

import { DropsItemsCardView } from './DropsItemsCard';

const meta = {
	title: 'Features/Bulk/DropsItemsCard',
	component: DropsItemsCardView,
	tags: ['autodocs'],
	args: { list: dropsItemListFixture(), hidden: 0, onChoose: fn() },
	argTypes: { list: { control: false } },
	decorators: [
		Story => (
			<div className="w-full max-w-170">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsItemsCardView>;

export default meta;
type Story = StoryObj<typeof meta>;

// The ticked count leaves out the worn item and the unticked one; each slot the list reaches gets a tile with its ticked count.
export const Default: Story = {};

// Two ticked items fail the dialog's filters; they still run.
export const HiddenByFilters: Story = { args: { hidden: 2 } };

// No source ticked: the button waits for a first item.
export const Empty: Story = { args: { list: { items: [] } } };
