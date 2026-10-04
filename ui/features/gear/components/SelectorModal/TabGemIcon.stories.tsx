import { GemColor } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIGem } from '@testing/storybook/fixtures/protos';

import { TabGemIcon } from './TabGemIcon';

const meta = {
	title: 'Features/Gear/SelectorModal/Components/TabGemIcon',
	component: TabGemIcon,
	tags: ['autodocs'],
	args: { socketColor: GemColor.GemColorRed, gem: null },
	argTypes: { gem: { control: false } },
	// `socketColor` is a `GemColor` number, not a CSS colour, so the preview's colour matcher is narrowed.
	parameters: { controls: { matchers: { color: /^(background|color)$/i } } },
} satisfies Meta<typeof TabGemIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

export const EmptyRedSocket: Story = {};

export const EmptyMetaSocket: Story = { args: { socketColor: GemColor.GemColorMeta } };

export const Socketed: Story = { args: { gem: makeUIGem() } };
