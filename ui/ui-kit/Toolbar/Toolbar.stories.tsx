import type { Meta, StoryObj } from '@storybook/react-vite';
import { Icon } from '@ui-kit/Icon';
import { expect, userEvent, within } from 'storybook/test';

import { Toolbar } from './Toolbar';
import { ToolbarButton } from './ToolbarButton';

const meta = {
	title: 'Kit/Toolbar',
	component: Toolbar,
	tags: ['autodocs'],
	args: { testId: 'toolbar-story', 'aria-label': 'Rotation list' },
	argTypes: {
		className: { control: 'text' },
	},
	render: args => (
		<Toolbar {...args}>
			<ToolbarButton variant="primary" testId="toolbar-story-new">
				<Icon name="plus" />
				New Rotation
			</ToolbarButton>
			<ToolbarButton variant="outline-primary" testId="toolbar-story-copy">
				<Icon name="copy" />
				Copy
			</ToolbarButton>
			<ToolbarButton variant="link-danger" size="sm" testId="toolbar-story-reset">
				<Icon name="xmark" />
				Reset
			</ToolbarButton>
		</Toolbar>
	),
} satisfies Meta<typeof Toolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Horizontal: Story = { args: { className: 'items-center gap-2' } };

export const DisabledButton: Story = {
	render: args => (
		<Toolbar {...args} className="items-center gap-2">
			<ToolbarButton variant="primary">Enabled</ToolbarButton>
			<ToolbarButton variant="primary" disabled>
				Disabled but focusable
			</ToolbarButton>
		</Toolbar>
	),
};

export const RovingFocus: Story = {
	args: { className: 'items-center gap-2' },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const first = canvas.getByTestId('toolbar-story-new');
		await userEvent.click(first);
		await expect(first).toHaveFocus();
		await userEvent.keyboard('{ArrowRight}');
		await expect(canvas.getByTestId('toolbar-story-copy')).toHaveFocus();
		await userEvent.keyboard('{ArrowRight}');
		await expect(canvas.getByTestId('toolbar-story-reset')).toHaveFocus();
		await userEvent.keyboard('{ArrowLeft}');
		await expect(canvas.getByTestId('toolbar-story-copy')).toHaveFocus();
	},
};
