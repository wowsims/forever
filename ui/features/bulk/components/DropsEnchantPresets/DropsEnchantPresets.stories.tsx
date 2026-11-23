import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { DropsEnchantPresetsView } from './DropsEnchantPresets';

const row = (name: string, summary: string, checked: boolean) => ({ ref: { name, preset: false }, name, summary, checked, duplicate: false });

const meta = {
	title: 'Features/Bulk/DropsEnchantPresets',
	component: DropsEnchantPresetsView,
	tags: ['autodocs'],
	args: {
		rows: [row('Agility', '2 slots.', true), row('Cloaks', '1 slot.', false)],
		onCheckedChange: fn(),
		onNew: fn(),
		dialog: null,
	},
	argTypes: { rows: { control: false }, dialog: { control: false } },
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DropsEnchantPresetsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const NoPresets: Story = { args: { rows: [] } };
