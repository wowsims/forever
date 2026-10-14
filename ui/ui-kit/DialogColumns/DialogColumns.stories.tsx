import type { Meta, StoryObj } from '@storybook/react-vite';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Dialog } from '@ui-kit/Dialog';

import { DialogColumns } from './DialogColumns';

const ColumnsStory = ({ rows }: { rows: number }) => (
	<Dialog open onOpenChange={() => {}} size="xl" scrollContents title="Dialog columns">
		<DialogColumns.Root>
			<DialogColumns.Main>
				{Array.from({ length: rows }, (_, row) => (
					<p key={row}>Main column row {row + 1}</p>
				))}
			</DialogColumns.Main>
			<DialogColumns.Sidebar>
				<ContentBlock config={{ header: { title: 'Settings' } }}>Sidebar settings</ContentBlock>
				<ContentBlock config={{ header: { title: 'History' } }}>Sidebar history</ContentBlock>
			</DialogColumns.Sidebar>
		</DialogColumns.Root>
	</Dialog>
);

const meta = {
	title: 'Kit/DialogColumns',
	component: ColumnsStory,
	tags: ['autodocs'],
	args: { rows: 5 },
} satisfies Meta<typeof ColumnsStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const MainScrolls: Story = { args: { rows: 80 } };
