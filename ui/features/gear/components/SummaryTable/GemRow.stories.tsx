import type { Meta, StoryObj } from '@storybook/react-vite';
import { wowheadIconUrl } from '@testing/storybook/fixtures/gear';
import { makeUIGem } from '@testing/storybook/fixtures/protos';

import { GemRow } from './GemRow';

const meta = {
	title: 'Features/Gear/SummaryTable/Components/GemRow',
	component: GemRow,
	tags: ['autodocs'],
	args: { gem: makeUIGem(), count: 3, iconUrl: wowheadIconUrl('inv_misc_gem_ruby_01') },
	argTypes: { gem: { control: false } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof GemRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const IconResolving: Story = { args: { iconUrl: undefined } };
