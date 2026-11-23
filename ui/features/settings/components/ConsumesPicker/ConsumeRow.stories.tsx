import type { Meta, StoryObj } from '@storybook/react-vite';
import { OVERWRITTEN_TOOLTIP, PARTLY_OVERWRITTEN_TOOLTIP } from '@testing/storybook/fixtures/overwritten';
import { storyIcon, StoryIconEnum } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { IconRow } from '@ui-kit/IconRow';

import { ConsumeRowView } from './ConsumeRow';

const FLASKS = [
	{ value: 0, tooltip: 'No Flask' },
	{ value: 1, icon: storyIcon('inv_potion_62', 'Flask of the Titans', 'item', 13510), tooltip: 'Flask of the Titans' },
	{ value: 2, icon: storyIcon('inv_potion_41', 'Flask of Supreme Power', 'item', 13512), tooltip: 'Flask of Supreme Power' },
];

const CRIT_ELIXIRS = [
	{ value: 0, tooltip: 'None' },
	{ value: 1, icon: storyIcon('inv_potion_32', 'Elixir of the Mongoose', 'item', 13452), tooltip: 'Elixir of the Mongoose' },
	{ value: 2, icon: storyIcon('inv_potion_160', 'Elixir of the Grizzly', 'item', 250351), tooltip: 'Elixir of the Grizzly' },
];

const STRENGTH_ELIXIRS = [
	{ value: 0, tooltip: 'None' },
	{ value: 1, icon: storyIcon('inv_potion_161', 'Elixir of Ferocity', 'item', 250350), tooltip: 'Elixir of Ferocity' },
	{ value: 2, icon: storyIcon('inv_misc_monsterscales_11', 'Juju Power', 'item', 12451), tooltip: 'Juju Power' },
];

const AGILITY_ELIXIRS = [
	{ value: 0, tooltip: 'None' },
	{ value: 1, icon: storyIcon('inv_potion_94', 'Elixir of Greater Agility', 'item', 9187), tooltip: 'Elixir of Greater Agility' },
];

const INTELLECT_SCROLLS = [
	{ value: 0, tooltip: 'None' },
	{ value: 1, icon: storyIcon('inv_scroll_01', 'Scroll of Intellect IV', 'item', 10308), tooltip: 'Scroll of Intellect IV' },
	{ value: 2, icon: storyIcon('inv_scroll_01', 'Scroll of Intellect III', 'item', 4419), tooltip: 'Scroll of Intellect III' },
];

const meta = {
	title: 'Features/Settings/ConsumeRow',
	component: ConsumeRowView,
	tags: ['autodocs'],
	args: {
		label: 'Flask & Elixirs',
		nowrap: true,
		hidden: false,
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={FLASKS} initial={1} />
				<StoryIconEnum options={STRENGTH_ELIXIRS} initial={1} overwrittenTooltip={PARTLY_OVERWRITTEN_TOOLTIP} overwrittenPartly />
				<StoryIconEnum options={AGILITY_ELIXIRS} />
				<StoryIconEnum options={CRIT_ELIXIRS} initial={1} />
			</IconRow>
		),
	},
	argTypes: { children: { control: false } },
} satisfies Meta<typeof ConsumeRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The flask and one pick per elixir group; Elixir of Ferocity loses its Agility to the Mongoose, so its badge shows on a lit swatch. */
export const Default: Story = {};

/** Swatches that wrap onto a second line leave the caption level with the first. */
export const Wrapped: Story = {
	decorators: [
		Story => (
			<div className="w-72">
				<Story />
			</div>
		),
	],
	args: {
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={FLASKS} initial={1} />
				<StoryIconEnum options={STRENGTH_ELIXIRS} initial={1} />
				<StoryIconEnum options={AGILITY_ELIXIRS} initial={1} />
				<StoryIconEnum options={CRIT_ELIXIRS} initial={1} />
				<StoryIconEnum options={INTELLECT_SCROLLS} initial={1} />
				<StoryIconEnum options={FLASKS} initial={2} />
			</IconRow>
		),
	},
};

/** A pick every stat of which other picks match is dimmed with the badge. */
export const Overwritten: Story = {
	args: {
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={STRENGTH_ELIXIRS} initial={1} overwrittenTooltip={OVERWRITTEN_TOOLTIP} />
				<StoryIconEnum options={AGILITY_ELIXIRS} initial={1} />
				<StoryIconEnum options={CRIT_ELIXIRS} initial={2} />
			</IconRow>
		),
	},
};

/** One pick per scroll line; a scroll a raid buff of its stat that is on matches is dimmed with a badge naming it. */
export const Scrolls: Story = {
	args: {
		label: 'Scrolls',
		children: (
			<IconRow className="justify-end">
				<StoryIconEnum options={INTELLECT_SCROLLS} initial={1} overwrittenTooltip="Intellect overwritten by Arcane Intellect" />
			</IconRow>
		),
	},
};

/** A row whose pickers all hide stays mounted with the `hidden` utility, so the canvas is empty. */
export const Hidden: Story = { args: { hidden: true } };
