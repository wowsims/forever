import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

import { ItemCell, ItemCellIcon } from './ItemCell';

const ICON = 'https://wow.zamimg.com/images/wow/icons/large/inv_sword_39.jpg';

const meta = {
	title: 'Kit/ItemCell',
	component: ItemCell,
	tags: ['autodocs'],
	args: {
		icon: (
			<ItemCellIcon ilvl="66">
				<a className="ui-item-picker-icon" style={{ backgroundImage: `url('${ICON}')` }} />
			</ItemCellIcon>
		),
		name: (
			<span className="ui-item-picker-name-container text-quality-epic">
				<span className="ui-item-picker-name">Thunderfury, Blessed Blade of the Windseeker</span>
			</span>
		),
		testId: 'item-cell-story',
	},
	argTypes: {
		icon: { control: false },
		name: { control: false },
		labels: { control: false },
		action: { control: false },
	},
} satisfies Meta<typeof ItemCell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const MediumIcon: Story = {
	args: {
		icon: (
			<ItemCellIcon ilvl="66" size="md">
				<a className="ui-item-picker-icon size-full" style={{ backgroundImage: `url('${ICON}')` }} />
			</ItemCellIcon>
		),
	},
};

export const SmallIcon: Story = {
	args: {
		icon: (
			<ItemCellIcon ilvl="66" size="sm">
				<a className="ui-item-picker-icon size-full" style={{ backgroundImage: `url('${ICON}')` }} />
			</ItemCellIcon>
		),
	},
};

export const WithLabels: Story = {
	args: {
		labels: (
			<>
				<span className="ui-item-picker-label-muted">Crusader</span>
				<span className="ui-item-picker-label-muted">Reforged: Crit → Haste</span>
			</>
		),
	},
};

export const WithAction: Story = {
	args: {
		action: (
			<Button variant="link" iconOnly aria-label="Favourite">
				<Icon name="star" />
			</Button>
		),
	},
};

export const AlignEnd: Story = {
	args: {
		align: 'end',
		className: 'w-96',
		labels: <span className="ui-item-picker-label-muted">Crusader</span>,
	},
};
