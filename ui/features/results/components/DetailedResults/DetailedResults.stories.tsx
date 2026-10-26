import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { TabPanel } from '@ui-kit/TabNav';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { DetailedResultsView, type DetailedResultsViewProps } from './DetailedResults';
import { DETAILED_RESULTS_TABS } from './utils';

const StatefulView = ({
	activeId: initial,
	onActiveIdChange,
	groupRanks: initialGroupRanks,
	onGroupRanksChange,
	showApmBreakdown: initialShowApmBreakdown,
	onShowApmBreakdownChange,
	...props
}: DetailedResultsViewProps) => {
	const [activeId, setActiveId] = useState(initial);
	const [groupRanks, setGroupRanks] = useState(initialGroupRanks);
	const [showApmBreakdown, setShowApmBreakdown] = useState(initialShowApmBreakdown);
	const change = (next: string) => {
		setActiveId(next);
		onActiveIdChange(next);
	};
	const changeGroupRanks = (next: boolean) => {
		setGroupRanks(next);
		onGroupRanksChange(next);
	};
	const changeShowApmBreakdown = (next: boolean) => {
		setShowApmBreakdown(next);
		onShowApmBreakdownChange(next);
	};
	return (
		<DetailedResultsView
			{...props}
			activeId={activeId}
			onActiveIdChange={change}
			groupRanks={groupRanks}
			onGroupRanksChange={changeGroupRanks}
			showApmBreakdown={showApmBreakdown}
			onShowApmBreakdownChange={changeShowApmBreakdown}
		/>
	);
};

const meta = {
	title: 'Features/Results/DetailedResults/DetailedResults',
	component: DetailedResultsView,
	tags: ['autodocs'],
	args: {
		tabs: DETAILED_RESULTS_TABS,
		activeId: 'damageTab',
		onActiveIdChange: fn(),
		hasResults: true,
		iterationDisabled: false,
		deathDisabled: true,
		onSimulateIteration: fn(),
		onSimulateDeath: fn(),
		targetPicker: <span className="px-2">All units</span>,
		groupRanks: true,
		onGroupRanksChange: fn(),
		showApmBreakdown: false,
		onShowApmBreakdownChange: fn(),
		children: DETAILED_RESULTS_TABS.map(tab => (
			<TabPanel key={tab.id} value={tab.id} className="pt-0 pb-0">
				<p className="p-4">The {tab.id} pane.</p>
			</TabPanel>
		)),
	},
	argTypes: { tabs: { control: false }, targetPicker: { control: false }, children: { control: false } },
	render: args => <StatefulView {...args} />,
} satisfies Meta<typeof DetailedResultsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithResults: Story = {
	play: async ({ args, canvasElement }) => {
		const logTab = within(canvasElement)
			.getAllByRole('tab')
			.find(tab => tab.id === 'logTab-nav')!;
		await userEvent.click(logTab);
		await expect(args.onActiveIdChange).toHaveBeenCalledWith('logTab');
	},
};

export const NoResults: Story = { args: { hasResults: false } };

export const DeathOnOffer: Story = { args: { deathDisabled: false } };

export const SimDisabled: Story = { args: { iterationDisabled: true } };

export const RanksUngrouped: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox'));
		await expect(args.onGroupRanksChange).toHaveBeenCalledWith(false);
	},
};

export const ApmBreakdownShown: Story = {
	args: { activeId: 'castsTab' },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('checkbox', { name: i18n.t('results_tab.details.show_apm_breakdown') }));
		await expect(args.onShowApmBreakdownChange).toHaveBeenCalledWith(true);
	},
};
