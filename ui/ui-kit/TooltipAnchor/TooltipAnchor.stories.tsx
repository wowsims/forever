import type { Meta, StoryObj } from '@storybook/react-vite';
import { Tooltip } from '@ui-kit/Tooltip';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { TooltipAnchor } from './TooltipAnchor';

const meta = {
	title: 'Kit/TooltipAnchor',
	component: TooltipAnchor,
	tags: ['autodocs'],
	args: { tooltipId: 'tooltip-anchor-story', children: '13.00%', testId: 'tooltip-anchor' },
	argTypes: {
		className: { control: 'text' },
		children: { control: 'text' },
	},
	render: args => (
		<>
			<TooltipAnchor {...args} />
			<Tooltip id="tooltip-anchor-story" content="Glancing 10.00%, to hit cap 3.00%" />
		</>
	),
	play: async ({ canvasElement }) => {
		await userEvent.hover(within(canvasElement).getByTestId('tooltip-anchor'));
		const tooltip = await within(document.body).findByText('Glancing 10.00%, to hit cap 3.00%');
		await waitFor(() => expect(tooltip).toBeVisible());
	},
} satisfies Meta<typeof TooltipAnchor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
