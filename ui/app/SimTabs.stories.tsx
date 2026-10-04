import type { Meta, StoryObj } from '@storybook/react-vite';
import { SimTabPane } from '@ui-kit/SimTabPane';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { SimTabDef } from './SimTabDef';
import { SimTabs, SimTabsPanels, SimTabsStrip } from './SimTabs';

interface SimTabsFrameProps {
	tabs: Array<{ id: string; title: string; badge?: string }>;
}

// The strip and the panels are ordinary children of one `SimTabs`, in the header and `main` the shell renders them into.
const SimTabsFrame = ({ tabs }: SimTabsFrameProps) => (
	<SimTabs
		defs={tabs.map(tab => (
			<SimTabDef key={tab.id} id={tab.id} title={tab.title} badge={tab.badge}>
				<SimTabPane id={tab.id}>
					<p>{tab.title} pane</p>
				</SimTabPane>
			</SimTabDef>
		))}>
		<div className="flex flex-col">
			<header className="flex h-sim-header items-stretch font-bold whitespace-nowrap">
				<SimTabsStrip />
			</header>
			<main className="flex grow">
				<SimTabsPanels />
			</main>
		</div>
	</SimTabs>
);

const meta = {
	title: 'App/Tabs/SimTabs',
	component: SimTabsFrame,
	subcomponents: { SimTabs, SimTabDef },
	tags: ['autodocs'],
	args: {
		tabs: [
			{ id: 'gear-tab', title: 'Gear' },
			{ id: 'settings-tab', title: 'Settings' },
			{ id: 'talents-tab', title: 'Talents' },
			{ id: 'bulk-tab', title: 'Batch Sim', badge: 'Beta' },
		],
	},
} satisfies Meta<typeof SimTabsFrame>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(await canvas.findByRole('tab', { name: 'Talents' }));
		const pane = await canvas.findByText('Talents pane');
		await waitFor(() => expect(pane).toBeVisible());
	},
};
