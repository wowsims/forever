import { Tabs } from '@base-ui/react/tabs';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { Menu, MenuItem } from '@ui-kit/Menu';
import { expect, fireEvent, within } from 'storybook/test';

import { TabNav, type TabNavProps } from './TabNav';
import { TabPanel } from './TabPanel';
import { TabPanels } from './TabPanels';

const TABS = [
	{ id: 'story-gear', label: 'Gear' },
	{ id: 'story-settings', label: 'Settings' },
	{ id: 'story-talents', label: 'Talents' },
	{ id: 'story-rotation', label: 'Rotation' },
];

const TabsStory = (props: TabNavProps) => (
	<Tabs.Root defaultValue={props.tabs[0].id}>
		<TabNav {...props} />
		<TabPanels testId="story-panels">
			{props.tabs.map(tab => (
				<TabPanel key={tab.id} value={tab.id} testId={`${tab.id}-panel`}>
					<p>{tab.label} content</p>
				</TabPanel>
			))}
		</TabPanels>
	</Tabs.Root>
);

const meta = {
	title: 'Kit/TabNav',
	component: TabNav,
	tags: ['autodocs'],
	args: { tabs: TABS, testId: 'story-tabs', bordered: false, wrap: false },
	argTypes: {
		variant: { control: 'select', options: ['nav-tabs', 'sim'] },
		tabs: { control: 'object' },
	},
	render: args => <TabsStory {...args} />,
} satisfies Meta<typeof TabNav>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NavTabs: Story = {};

export const Sim: Story = { args: { variant: 'sim' } };

export const Bordered: Story = { args: { bordered: true } };

export const Wrap: Story = {
	args: {
		wrap: true,
		tabs: [...TABS, { id: 'story-results', label: 'Results' }, { id: 'story-bulk', label: 'Bulk' }, { id: 'story-log', label: 'Log' }],
	},
	render: args => (
		<div className="w-80">
			<TabsStory {...args} />
		</div>
	),
};

export const WithBadge: Story = {
	args: { tabs: [...TABS, { id: 'story-bulk', label: 'Bulk', badge: 'Beta' }] },
};

export const SimTabAsMenuTrigger: Story = {
	args: {
		variant: 'sim',
		tabs: [
			...TABS,
			{
				id: 'story-bulk',
				label: 'Batch',
				renderTab: tab => (
					<Menu surface="menu" triggerRender={tab} triggerProps={{ openOnHover: true, delay: 0, nativeButton: true }} align="start">
						<MenuItem layout="block">Gear</MenuItem>
						<MenuItem layout="block">Builds</MenuItem>
					</Menu>
				),
			},
		],
	},
	play: async ({ canvasElement }) => {
		const tab = within(canvasElement).getByRole('tab', { name: 'Batch' });
		await expect(tab.getAttribute('aria-haspopup')).toBe('menu');
		await fireEvent.click(tab);
		await expect(tab.getAttribute('aria-selected')).toBe('true');
	},
};

export const Switches: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByTestId('story-gear-panel').hasAttribute('hidden')).toBe(false);

		await fireEvent.click(canvas.getByText('Talents'));

		await expect(canvas.getByText('Talents').hasAttribute('data-active')).toBe(true);
		await expect(canvas.getByText('Gear').hasAttribute('data-active')).toBe(false);
		await expect(canvas.getByTestId('story-talents-panel').hasAttribute('hidden')).toBe(false);
	},
};
