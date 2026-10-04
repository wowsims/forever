import { type EquipmentSpec, ItemSlot } from '@generated/proto/common';
import { SavedGearSet } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withStorage } from '@testing/storybook/decorators';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { makeEquipmentSpec } from '@testing/storybook/fixtures/protos';

import { SavedGear } from './SavedGear';

const STORAGE_KEY = 'storybook-savedGearSets';

const PRESET_GEAR = makeEquipmentSpec();
const SAVED_GEAR = makeEquipmentSpec({ [ItemSlot.ItemSlotHead]: { id: 12640 } });

const stubGear = (spec: EquipmentSpec) => ({ asSpec: () => spec });

const gearHost = (presets: Array<{ name: string; gear: EquipmentSpec; tooltip?: string }>) =>
	withSimHost({
		player: { getGear: () => stubGear(PRESET_GEAR), getBonusStats: () => new Stats(), setGear: () => {}, setBonusStats: () => {} },
		sim: { db: { lookupEquipmentSpec: stubGear } },
		individualConfig: { presets: { gear: presets } },
		getSavedGearStorageKey: () => STORAGE_KEY,
	});

const meta = {
	title: 'Features/Gear/SavedGear',
	component: SavedGear,
	tags: ['autodocs'],
} satisfies Meta<typeof SavedGear>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The equipped gear matches the preset, so its chip is the active one; the user's own set sits below. */
export const Default: Story = {
	decorators: [
		withStorage({ [STORAGE_KEY]: JSON.stringify({ Raiding: SavedGearSet.toJson(SavedGearSet.create({ gear: SAVED_GEAR })) }) }),
		gearHost([{ name: 'Phase 1', gear: PRESET_GEAR, tooltip: 'Best in slot for phase 1.' }]),
	],
};

export const Empty: Story = {
	decorators: [withStorage(), gearHost([])],
};
