import { OtherAction } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { ActionIdIcon } from './ActionIdIcon';

const meta = {
	title: 'Features/Apl/FieldGroup/Components/ActionIdIcon',
	component: ActionIdIcon,
	tags: ['autodocs'],
	args: { useBuffAura: false },
} satisfies Meta<typeof ActionIdIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

/** An id that carries its own icon, so nothing is fetched. */
export const OtherActionIcon: Story = { args: { actionId: ActionId.fromOtherId(OtherAction.OtherActionAttack) } };

/** Bloodthirst: the icon and the wowhead link are filled from the spell. */
export const Spell: Story = {
	args: { actionId: ActionId.fromSpellId(23894) },
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};

/** The same spell as an aura, which asks wowhead for the buff's tooltip rather than the spell's. */
export const Aura: Story = {
	args: { actionId: ActionId.fromSpellId(23894), useBuffAura: true },
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database' },
};
