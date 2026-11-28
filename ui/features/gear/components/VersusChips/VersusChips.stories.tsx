import { VersusEffect, VersusEffectKind } from '@generated/proto/api';
import { AreaStats, AreaType, CreatureFamily, MobType, MobTypeStats, ScalingItemProperties, Stat, Target } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIItem } from '@testing/storybook/fixtures/protos';

import { VersusChipsView } from './VersusChips';

const ITEM = makeUIItem({
	id: 1,
	name: 'Seal of the Dawn',
	mobTypes: [MobType.MobTypeUndead, MobType.MobTypeDemon],
	creatureFamilies: [CreatureFamily.CreatureFamilyWolf],
	areaTypes: [AreaType.AreaTypeSnowy],
	scalingOptions: {
		0: ScalingItemProperties.create({
			mobTypeStats: [MobTypeStats.create({ mobTypes: [MobType.MobTypeUndead], stats: { [Stat.StatAttackPower]: 81 } })],
			areaStats: [AreaStats.create({ areaType: AreaType.AreaTypeSnowy, stats: { [Stat.StatStamina]: 10 } })],
		}),
	},
});

const meta = {
	title: 'Features/Gear/VersusChips',
	component: VersusChipsView,
	tags: ['autodocs'],
	args: {
		item: ITEM,
		targets: [Target.create({ mobType: MobType.MobTypeUndead }), Target.create({ mobType: MobType.MobTypeHumanoid })],
		areaTypes: [],
		versusEffects: [],
	},
	argTypes: {
		item: { control: false },
		targets: { control: false },
		areaTypes: { control: false },
		versusEffects: { control: false },
	},
} satisfies Meta<typeof VersusChipsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SomeTargets: Story = {};

export const AllActive: Story = {
	args: {
		targets: [
			Target.create({ mobType: MobType.MobTypeUndead }),
			Target.create({ mobType: MobType.MobTypeDemon }),
			Target.create({ mobType: MobType.MobTypeBeast, creatureFamily: CreatureFamily.CreatureFamilyWolf }),
		],
		areaTypes: [AreaType.AreaTypeSnowy],
	},
};

export const WornWithFactor: Story = {
	args: {
		targets: [Target.create({ mobType: MobType.MobTypeBeast, creatureFamily: CreatureFamily.CreatureFamilyWolf })],
		versusEffects: [
			VersusEffect.create({
				itemId: 1,
				itemName: 'Seal of the Dawn',
				creatureFamilies: [CreatureFamily.CreatureFamilyWolf],
				multiplier: 1.5,
				kind: VersusEffectKind.VersusEffectKindDamage,
			}),
		],
	},
};

export const NoTargets: Story = { args: { targets: [] } };
