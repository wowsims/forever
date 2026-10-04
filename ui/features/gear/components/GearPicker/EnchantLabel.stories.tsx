import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIEnchant } from '@testing/storybook/fixtures/protos';
import { fn } from 'storybook/test';

import { EnchantLabel } from './EnchantLabel';

const SMOKE_REASON = 'The label fetches /forever/assets/enchants/descriptions.json, and renders nothing without an enchant';

const meta = {
	title: 'Features/Gear/GearPicker/Components/EnchantLabel',
	component: EnchantLabel,
	tags: ['autodocs'],
	args: { className: 'ui-item-picker-label-muted', enchant: makeUIEnchant(), onActivate: fn() },
	argTypes: { enchant: { control: false } },
	parameters: { smoke: false, smokeReason: SMOKE_REASON },
} satisfies Meta<typeof EnchantLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
