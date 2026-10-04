import type { Meta, StoryObj } from '@storybook/react-vite';
import { Box } from '@ui-kit/Box';
import { useRef } from 'react';

import { VirtualList } from './VirtualList';

interface VirtualListStoryProps {
	count: number;
	rowHeight?: number;
}

const VirtualListStory = ({ count, rowHeight }: VirtualListStoryProps) => {
	const scroller = useRef<HTMLDivElement>(null);
	return (
		<Box ref={scroller} padding="none" className="h-80 w-96 overflow-y-auto" testId="story-scroller">
			<VirtualList
				count={count}
				rowHeight={rowHeight}
				getScrollElement={() => scroller.current}
				rowClassName={() => 'flex items-center px-2 data-[stripe=odd]:bg-white/5'}
				renderRow={index => (
					<>
						<span>Event {index + 1}</span>
						<span className="ml-auto text-secondary">{(index * 0.37).toFixed(2)}s</span>
					</>
				)}
			/>
		</Box>
	);
};

const meta = {
	title: 'Kit/VirtualList',
	component: VirtualListStory,
	tags: ['autodocs'],
	args: { count: 1000 },
} satisfies Meta<typeof VirtualListStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThousandRows: Story = {};

export const TallRows: Story = { args: { rowHeight: 48 } };
