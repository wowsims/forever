import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStatSweepHost } from '@testing/storybook/fixtures/stat-sweep';
import { fn } from 'storybook/test';

import { StatSweepDialog } from './StatSweepDialog';

const fixture = makeStatSweepHost();

const meta = {
	title: 'Features/StatSweep/StatSweepDialog',
	component: StatSweepDialog,
	tags: ['autodocs'],
	args: { open: true, onOpenChange: fn(), settings: fixture.settings },
	argTypes: { settings: { control: false } },
	decorators: [
		Story => (
			<SimHostProvider host={fixture.host}>
				<Story />
			</SimHostProvider>
		),
	],
} satisfies Meta<typeof StatSweepDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The dialog on a sim that answers every step at once, so Run draws a sweep straight away. */
export const Open: Story = {};
