import type { Meta, StoryObj } from '@storybook/react-vite';
import { OVERWRITTEN_TOOLTIP } from '@testing/storybook/fixtures/overwritten';
import { storyIcon, StoryIconEnum, StoryIconToggle } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { IconRow } from '@ui-kit/IconRow';

import { ConsumeRowView } from './ConsumeRow';
import { ElixirPickerView } from './ElixirPicker';

const FLASKS = [
	{ value: 0, tooltip: 'No Flask' },
	{ value: 1, icon: storyIcon('inv_potion_62', 'Flask of the Titans', 'item', 13510), tooltip: 'Flask of the Titans' },
	{ value: 2, icon: storyIcon('inv_potion_41', 'Flask of Supreme Power', 'item', 13512), tooltip: 'Flask of Supreme Power' },
];

const STRENGTH_SCROLLS = [
	{ value: 0, tooltip: 'None' },
	{ value: 1, icon: storyIcon('inv_scroll_02', 'Scroll of Strength IV', 'item', 10310), tooltip: 'Scroll of Strength IV' },
	{ value: 2, icon: storyIcon('inv_scroll_02', 'Scroll of Strength III', 'item', 4426), tooltip: 'Scroll of Strength III' },
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
				<ElixirPickerView
					picks={
						<>
							<StoryIconToggle icon={storyIcon('inv_potion_32', 'Elixir of the Mongoose', 'item', 13452)} initial={1} />
							<StoryIconToggle icon={storyIcon('inv_misc_monsterscales_11', 'Juju Power', 'item', 12451)} initial={1} />
							<StoryIconToggle
								icon={storyIcon('inv_potion_161', 'Elixir of Ferocity', 'item', 250350)}
								initial={1}
								overwrittenTooltip={OVERWRITTEN_TOOLTIP}
							/>
						</>
					}
					groups={[]}
				/>
			</IconRow>
		),
	},
	argTypes: { children: { control: false } },
} satisfies Meta<typeof ConsumeRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The flask, the picked elixirs and the button that adds more. */
export const Default: Story = {};

/** One pick per scroll line; a scroll an elixir picked beside it matches is dimmed with a badge naming it. */
export const Scrolls: Story = {
	args: {
		label: 'Scrolls',
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={STRENGTH_SCROLLS} initial={1} overwrittenTooltip="Adds nothing: every stat it grants is matched by Juju Power." />
			</IconRow>
		),
	},
};

/** A row whose pickers all hide stays mounted with the `hidden` utility, so the canvas is empty. */
export const Hidden: Story = { args: { hidden: true } };
