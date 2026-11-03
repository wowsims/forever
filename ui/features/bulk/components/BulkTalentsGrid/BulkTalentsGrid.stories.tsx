import { warriorTalentsConfig } from '@sim/talents/warrior';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { BulkTalentsGrid } from './BulkTalentsGrid';

const meta = {
	title: 'Features/Bulk/BulkTalentsGrid',
	component: BulkTalentsGrid,
	tags: ['autodocs'],
	args: { config: warriorTalentsConfig, talentsString: '30305013002-25053005052010501' },
	argTypes: { config: { control: false } },
	parameters: { smoke: false, smokeReason: 'Each talent fills its icon from the database through useActionId' },
	render: args => (
		<div className="w-fit max-w-76 bg-black p-2 text-white">
			<BulkTalentsGrid {...args} />
		</div>
	),
} satisfies Meta<typeof BulkTalentsGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TwoTrees: Story = {};

export const ThreeTrees: Story = { args: { talentsString: '30325213132515201-0505-2' } };

export const OneTree: Story = { args: { talentsString: '--050531120300021351' } };

export const NoPoints: Story = { args: { talentsString: '' } };
