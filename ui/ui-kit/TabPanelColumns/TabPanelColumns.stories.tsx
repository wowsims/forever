import type { Meta, StoryObj } from '@storybook/react-vite';
import { ContentBlock } from '@ui-kit/ContentBlock';

import { TabPanelColumns, type TabPanelColumnsGap, type TabPanelColumnsLeftVariant } from './TabPanelColumns';

const GAPS: Array<TabPanelColumnsGap> = ['default', 'apl'];
const LEFT_VARIANTS: Array<TabPanelColumnsLeftVariant> = ['default', 'auto-columns', 'settings-columns', 'stacked'];

interface ColumnsStoryProps {
	gap?: TabPanelColumnsGap;
	leftVariant?: TabPanelColumnsLeftVariant;
}

const Block = ({ title }: { title: string }) => <ContentBlock config={{ header: { title } }}>Contents of {title}</ContentBlock>;

const ColumnsStory = ({ gap, leftVariant }: ColumnsStoryProps) => (
	<TabPanelColumns.Root gap={gap} testId="story-columns">
		<TabPanelColumns.Left variant={leftVariant}>
			<TabPanelColumns.Col>
				<Block title="Player" />
				<Block title="Race" />
			</TabPanelColumns.Col>
			<TabPanelColumns.Col>
				<Block title="Buffs" />
			</TabPanelColumns.Col>
			<TabPanelColumns.Col>
				<Block title="Debuffs" />
			</TabPanelColumns.Col>
		</TabPanelColumns.Left>
		<TabPanelColumns.Right>
			<Block title="Presets" />
		</TabPanelColumns.Right>
	</TabPanelColumns.Root>
);

const meta = {
	title: 'Kit/TabPanelColumns',
	component: ColumnsStory,
	tags: ['autodocs'],
	argTypes: {
		gap: { control: 'select', options: GAPS },
		leftVariant: { control: 'select', options: LEFT_VARIANTS },
	},
} satisfies Meta<typeof ColumnsStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const AplGap: Story = { args: { gap: 'apl' } };
export const AutoColumns: Story = { args: { leftVariant: 'auto-columns' } };
export const SettingsColumns: Story = { args: { leftVariant: 'settings-columns' } };
export const Stacked: Story = { args: { leftVariant: 'stacked' } };

export const Matrix: Story = {
	render: () => (
		<div className="flex flex-col gap-8">
			{GAPS.flatMap(gap =>
				LEFT_VARIANTS.map(leftVariant => (
					<section key={`${gap}-${leftVariant}`} className="flex flex-col gap-2">
						<h6 className="font-bold">
							gap {gap}, left {leftVariant}
						</h6>
						<ColumnsStory gap={gap} leftVariant={leftVariant} />
					</section>
				)),
			)}
		</div>
	),
};
