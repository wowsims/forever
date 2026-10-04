import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyIcon, StoryIconEnum } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { NumberListField } from '@ui-kit/TextField';
import { Tooltip } from '@ui-kit/Tooltip';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { CooldownRowView } from './CooldownRow';

const COOLDOWNS = [
	{ value: 0, color: '#808080' },
	{ value: 1, icon: storyIcon('ability_criticalstrike', 'Recklessness', 'spell', 1719) },
	{ value: 2, icon: storyIcon('spell_nature_bloodlust', 'Bloodlust', 'spell', 2825) },
];

const Timings = ({ initial }: { initial: Array<number> }) => {
	const [value, setValue] = useState(initial);
	return <NumberListField id="story-cooldown-timings" className="ml-auto w-2/5 min-w-2/5" placeholder="20, 40, ..." value={value} onChange={setValue} />;
};

const meta = {
	title: 'Features/Settings/CooldownRow',
	component: CooldownRowView,
	tags: ['autodocs'],
	args: {
		name: 'Recklessness',
		isAdd: false,
		deleteTooltipId: 'story-delete-cooldown',
		onDelete: fn(),
		actionPicker: <StoryIconEnum options={COOLDOWNS} initial={1} />,
		timingsPicker: <Timings initial={[20, 200]} />,
	},
	argTypes: { actionPicker: { control: false }, timingsPicker: { control: false } },
	decorators: [
		Story => (
			<>
				<Story />
				<Tooltip id="story-delete-cooldown" content="Delete Cooldown" />
			</>
		),
	],
} satisfies Meta<typeof CooldownRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('delete-cooldown'));
		await expect(args.onDelete).toHaveBeenCalledOnce();
	},
};

export const AddRow: Story = {
	args: { name: undefined, isAdd: true, actionPicker: <StoryIconEnum options={COOLDOWNS} />, timingsPicker: <Timings initial={[]} /> },
};
