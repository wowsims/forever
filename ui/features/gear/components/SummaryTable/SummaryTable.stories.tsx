import type { Meta, StoryObj } from '@storybook/react-vite';
import { SummaryTableRow } from '@ui-kit/SummaryTableRow';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SummaryTable } from './SummaryTable';

const meta = {
	title: 'Features/Gear/SummaryTable/SummaryTable',
	component: SummaryTable,
	tags: ['autodocs'],
	args: {
		title: 'Item Sources',
		modifier: 'summary-table--sources',
		empty: false,
		reset: { label: 'Clear', onReset: fn() },
		children: (
			<>
				<SummaryTableRow>
					<span>Molten Core</span>
					<span>5</span>
				</SummaryTableRow>
				<SummaryTableRow>
					<span>Blackwing Lair</span>
					<span>3</span>
				</SummaryTableRow>
				<SummaryTableRow>
					<span>Crafted</span>
					<span>1</span>
				</SummaryTableRow>
			</>
		),
	},
	argTypes: { children: { control: false } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SummaryTable>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('summary-table-reset-button'));
		await expect(args.reset.onReset).toHaveBeenCalledOnce();
	},
};
