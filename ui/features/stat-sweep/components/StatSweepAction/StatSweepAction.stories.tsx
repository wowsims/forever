import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeStatSweepHost } from '@testing/storybook/fixtures/stat-sweep';

import { StatSweepAction } from './StatSweepAction';

const fixture = makeStatSweepHost({ showPowerTools: true });

const meta = {
	title: 'Features/StatSweep/StatSweepAction',
	component: StatSweepAction,
	tags: ['autodocs'],
	args: { settings: fixture.settings },
	argTypes: { settings: { control: false } },
	decorators: [
		Story => (
			<SimHostProvider host={fixture.host}>
				<Story />
			</SimHostProvider>
		),
	],
} satisfies Meta<typeof StatSweepAction>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The sidebar button while Show power tools is on; it opens the stat sweep on a sim that answers every step at once. */
export const Default: Story = {};
