import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyIcon, StoryIconEnum, StoryIconToggle } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { IconRow } from '@ui-kit/IconRow';

import { ConsumeRowView } from './ConsumeRow';

const FLASKS = [
	{ value: 0, tooltip: 'No Flask' },
	{ value: 1, icon: storyIcon('inv_potion_62', 'Flask of the Titans', 'item', 13510), tooltip: 'Flask of the Titans' },
	{ value: 2, icon: storyIcon('inv_potion_41', 'Flask of Supreme Power', 'item', 13512), tooltip: 'Flask of Supreme Power' },
];

const meta = {
	title: 'Features/Settings/ConsumeRow',
	component: ConsumeRowView,
	tags: ['autodocs'],
	args: {
		label: 'Elixirs',
		hidden: false,
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={FLASKS} initial={1} />
				<StoryIconToggle icon={storyIcon('inv_potion_32', 'Elixir of the Mongoose', 'item', 13452)} initial={1} />
				<StoryIconToggle icon={storyIcon('inv_potion_61', 'Elixir of Giants', 'item', 9206)} />
			</IconRow>
		),
	},
	argTypes: { children: { control: false } },
} satisfies Meta<typeof ConsumeRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

/** A row whose pickers all hide stays mounted with the `hidden` utility, so the canvas is empty. */
export const Hidden: Story = { args: { hidden: true } };
