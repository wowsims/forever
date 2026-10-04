import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean, StoryEnum, StoryNumber } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { Table } from '@ui-kit/Table';

import { ReforgeStatCapRowView } from './ReforgeStatCapRow';

const percentage = (initial: number) => <StoryNumber float showZeroes={false} positive inputClassName="pr-2 pl-2 text-right" initial={initial} />;

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeStatCapRow',
	component: ReforgeStatCapRowView,
	tags: ['autodocs'],
	args: {
		statName: 'Hit',
		percentage: percentage(9),
		undershoot: <StoryBoolean label="" />,
	},
	argTypes: { percentage: { control: false }, undershoot: { control: false }, presets: { control: false }, tooltip: { control: 'text' } },
	decorators: [
		Story => (
			<Table dense>
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof ReforgeStatCapRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithTooltip: Story = {
	args: { statName: 'Haste', tooltip: 'Final percentage value including all buffs/gear.', percentage: percentage(0) },
};

export const WithPresets: Story = {
	args: {
		statName: 'Expertise',
		presets: (
			<StoryEnum
				label=""
				values={[
					{ name: 'Select preset', value: 0 },
					{ name: 'Dodge cap - 6.50%', value: 6.5 },
				]}
			/>
		),
	},
};
