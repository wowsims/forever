import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean, StoryNumber } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { fn } from 'storybook/test';

import { ReforgeStatCapRowView } from './ReforgeStatCapRow';
import { ReforgeStatCapsView } from './ReforgeStatCaps';

const CAPS: Array<[string, number, boolean]> = [
	['Hit', 9, true],
	['Crit', 0, false],
	['Haste', 0, false],
	['Expertise', 6.5, true],
];

const meta = {
	title: 'Features/Reforge/ReforgePanel/Components/ReforgeStatCaps',
	component: ReforgeStatCapsView,
	tags: ['autodocs'],
	args: {
		onReset: fn(),
		children: CAPS.map(([statName, percentage, undershoot]) => (
			<ReforgeStatCapRowView
				key={statName}
				statName={statName}
				tooltip={statName === 'Haste' ? 'Final percentage value including all buffs/gear.' : undefined}
				percentage={<StoryNumber float showZeroes={false} positive inputClassName="pr-2 pl-2 text-right" initial={percentage} />}
				undershoot={<StoryBoolean label="" initial={undershoot} />}
			/>
		)),
	},
	argTypes: { children: { control: false } },
} satisfies Meta<typeof ReforgeStatCapsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
