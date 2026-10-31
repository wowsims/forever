import type { Meta, StoryObj } from '@storybook/react-vite';
import { buildBundlesProps } from '@testing/storybook/fixtures/bulk-builds';

import { BuildBundlesView } from './BuildBundles';

const meta = {
	title: 'Features/Bulk/BuildBundles',
	component: BuildBundlesView,
	tags: ['autodocs'],
	args: buildBundlesProps(),
	argTypes: {
		list: { control: false },
		dialog: { control: false },
		enchantPicks: { control: false },
		enchantPresetDialog: { control: false },
	},
	decorators: [
		Story => (
			<div className="grid w-96 gap-4">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BuildBundlesView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoEnchantPresets: Story = { args: { enchantPicks: [] } };
