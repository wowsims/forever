import type { Meta, StoryObj } from '@storybook/react-vite';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';

import { SimTabPane } from './SimTabPane';

const meta = {
	title: 'Kit/SimTabPane',
	component: SimTabPane,
	tags: ['autodocs'],
	args: {
		id: 'story-tab',
		children: (
			<>
				<TabPanelColumns.Left>
					<ContentBlock config={{ header: { title: 'Encounter' } }}>Patchwerk, 180 seconds</ContentBlock>
					<ContentBlock config={{ header: { title: 'Buffs' } }}>Arcane Intellect, Mark of the Wild</ContentBlock>
				</TabPanelColumns.Left>
				<TabPanelColumns.Right>
					<ContentBlock config={{ header: { title: 'Presets' } }}>Phase 1</ContentBlock>
				</TabPanelColumns.Right>
			</>
		),
	},
	argTypes: {
		children: { control: false },
	},
} satisfies Meta<typeof SimTabPane>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const StackedContent: Story = { args: { contentClassName: 'flex-col' } };
