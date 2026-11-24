import type { Meta, StoryObj } from '@storybook/react-vite';
import { dropsItemListFixture, dropsNamesFixture, dropsTickedFixture, dropsTreeFixture } from '@testing/storybook/fixtures/bulk-drops';
import { fn } from 'storybook/test';

import { DropsItemsView } from '../DropsItems';
import { DropsSourcesView } from '../DropsSources';
import { DropsSetupView } from './DropsSetup';

const meta = {
	title: 'Features/Bulk/DropsSetup',
	component: DropsSetupView,
	tags: ['autodocs'],
	args: {
		sources: <DropsSourcesView tree={dropsTreeFixture()} ticked={dropsTickedFixture()} onTick={fn()} onQuickPick={fn()} />,
		items: (
			<DropsItemsView
				list={dropsItemListFixture()}
				grouping="slot"
				tree={dropsTreeFixture()}
				names={dropsNamesFixture()}
				onGroupingChange={fn()}
				onIncludedChange={fn()}
			/>
		),
	},
	argTypes: { sources: { control: false }, items: { control: false } },
} satisfies Meta<typeof DropsSetupView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Phone: Story = {
	decorators: [
		Story => (
			<div className="w-full max-w-96">
				<Story />
			</div>
		),
	],
};
