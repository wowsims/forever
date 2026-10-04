import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryEnum } from '@testing/storybook/fixtures/settings_reforge_bulk';

import { ReforgeBreakpointLimitsView } from './ReforgeBreakpointLimits';

const limit = (statName: string, breakpoints: Array<number>) => ({
	key: statName,
	statName,
	picker: (
		<StoryEnum label="" values={[{ name: 'No limit set', value: 0 }, ...breakpoints.map(breakpoint => ({ name: `${breakpoint}%`, value: breakpoint }))]} />
	),
});

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeBreakpointLimits',
	component: ReforgeBreakpointLimitsView,
	tags: ['autodocs'],
	args: { rows: [limit('Hit', [5, 9]), limit('Spell Hit', [16])] },
	argTypes: { rows: { control: false } },
} satisfies Meta<typeof ReforgeBreakpointLimitsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const OneStat: Story = { args: { rows: [limit('Hit', [5, 9])] } };
