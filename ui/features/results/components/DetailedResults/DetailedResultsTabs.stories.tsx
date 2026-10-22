import { Tabs } from '@base-ui/react/tabs';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { DetailedResultsTabs } from './DetailedResultsTabs';
import { DETAILED_RESULTS_TABS } from './utils';

const meta = {
	title: 'Features/Results/DetailedResults/Components/DetailedResultsTabs',
	component: DetailedResultsTabs,
	tags: ['autodocs'],
	args: { tabs: DETAILED_RESULTS_TABS },
	argTypes: { tabs: { control: false } },
	decorators: [
		Story => (
			<Tabs.Root defaultValue="damageTab" className="flex">
				<Story />
			</Tabs.Root>
		),
	],
} satisfies Meta<typeof DetailedResultsTabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AllTabs: Story = {};

export const DamageOnly: Story = {
	args: { tabs: DETAILED_RESULTS_TABS.filter(tab => !['threatTab', 'healingTab', 'healingSpellsTab', 'damageTakenTab'].includes(tab.id)) },
};
