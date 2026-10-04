import type { Meta, StoryObj } from '@storybook/react-vite';
import { Icon } from '@ui-kit/Icon';
import { ToolbarButton } from '@ui-kit/Toolbar';
import { fn } from 'storybook/test';

import { FloatingActionBar } from './FloatingActionBar';

const meta = {
	title: 'Features/Results/FloatingActionBar',
	component: FloatingActionBar,
	tags: ['autodocs'],
	args: {
		testIdPrefix: 'story',
		rootClassName: 'ui-log-floating-action-bar',
		sheetClassName: 'ui-log-floating-action-bar-sheet',
		toolbarClassName: 'gap-2',
		toggleLabel: 'Filters',
		icon: <Icon name="filter" />,
		summary: 'No filters',
		preview: '',
		clear: { testId: 'story-floating-action-bar-clear', icon: <Icon name="xmark" />, label: 'Clear', hidden: true, onClick: fn() },
		controls: (
			<div className="flex items-center gap-2 md:ml-auto">
				<ToolbarButton className="whitespace-nowrap">Export</ToolbarButton>
			</div>
		),
		children: <div className="ui-floating-action-bar-drawer">The drawer's contents.</div>,
	},
	argTypes: { icon: { control: false }, clear: { control: false }, controls: { control: false }, children: { control: false } },
} satisfies Meta<typeof FloatingActionBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const Active: Story = {
	args: {
		summary: '2 filters',
		preview: 'Spell: Mortal Strike, Outcome: Crit',
		clear: { testId: 'story-floating-action-bar-clear', icon: <Icon name="xmark" />, label: 'Clear', hidden: false, onClick: fn() },
	},
};
