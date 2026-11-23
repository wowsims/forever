import type { Meta, StoryObj } from '@storybook/react-vite';
import { OVERWRITTEN_TOOLTIP } from '@testing/storybook/fixtures/overwritten';
import { storyIcon, StoryIconToggle } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { IconRow } from '@ui-kit/IconRow';
import { expect, userEvent, within } from 'storybook/test';

import { ElixirPickerView } from './ElixirPicker';

const MONGOOSE = storyIcon('inv_potion_32', 'Elixir of the Mongoose', 'item', 13452);
const GRIZZLY = storyIcon('inv_potion_32', 'Elixir of the Grizzly', 'item', 250351);
const JUJU_POWER = storyIcon('inv_misc_monsterscales_11', 'Juju Power', 'item', 12451);
const FEROCITY = storyIcon('inv_potion_161', 'Elixir of Ferocity', 'item', 250350);
const GREATER_STRENGTH = storyIcon('inv_potion_59', 'Elixir of Greater Strength', 'item', 9206);

const meta = {
	title: 'Features/Settings/ElixirPicker',
	component: ElixirPickerView,
	tags: ['autodocs'],
	args: {
		picks: (
			<>
				<StoryIconToggle icon={MONGOOSE} initial={1} />
				<StoryIconToggle icon={JUJU_POWER} initial={1} />
				<StoryIconToggle icon={FEROCITY} initial={1} overwrittenTooltip={OVERWRITTEN_TOOLTIP} />
			</>
		),
		groups: [
			{
				key: 'critical_strike',
				label: 'Critical Strike',
				children: (
					<>
						<StoryIconToggle icon={MONGOOSE} initial={1} />
						<StoryIconToggle icon={GRIZZLY} />
					</>
				),
			},
			{
				key: 'stat 0',
				label: 'Strength',
				children: (
					<>
						<StoryIconToggle icon={JUJU_POWER} initial={1} />
						<StoryIconToggle icon={FEROCITY} initial={1} overwrittenTooltip={OVERWRITTEN_TOOLTIP} />
						<StoryIconToggle icon={GREATER_STRENGTH} />
					</>
				),
			},
		],
	},
	argTypes: { picks: { control: false }, groups: { control: false } },
	decorators: [
		Story => (
			<IconRow className="justify-end">
				<Story />
			</IconRow>
		),
	],
} satisfies Meta<typeof ElixirPickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The picked elixirs, one the others match on every stat dimmed with a badge naming them, and the add button. */
export const Default: Story = {};

/** The add menu, one line per stat; a click on an elixir picks or unpicks it and leaves the menu open. */
export const Open: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('consumes-elixirs-add'));
		const menu = await within(canvasElement.ownerDocument.body).findByTestId('consumes-elixirs-menu');
		await expect(within(menu).getAllByTestId('consumes-elixir-group')).toHaveLength(2);
	},
};

/** No elixir picked: the row holds the add button alone. */
export const NonePicked: Story = { args: { picks: null } };
