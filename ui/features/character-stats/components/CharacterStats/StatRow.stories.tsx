import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { Stats, UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeRangedWeapon, makeStatAttribution, showStat } from '@testing/storybook/fixtures/character-stats';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { Table } from '@ui-kit/Table';

import { BonusStatsLink } from './BonusStatsLink';
import { StatRowView } from './StatRow';

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

const meta = {
	title: 'Features/CharacterStats/Components/StatRow',
	component: StatRowView,
	tags: ['autodocs'],
	args: {
		displayStat: { stat: STRENGTH },
		playerClass: Class.ClassWarrior,
		bonusStats: new Stats(),
		attribution: makeStatAttribution(),
		show: showStat,
		rangedWeapon: null,
		dualWielding: false,
	},
	argTypes: {
		displayStat: { control: false },
		bonusStats: { control: false },
		attribution: { control: false },
		rangedWeapon: { control: false },
		bonusStatsLink: { control: false },
	},
	decorators: [
		Story => (
			<Table variant="plain" className="max-w-64">
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof StatRowView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Pending: Story = { args: { pending: true } };

export const WithBonus: Story = { args: { bonusStats: new Stats().withStat(Stat.StatStrength, 20) } };

/** A combined melee row shows each hand on its own line, prefixed MH and OH, while dual wielding, each with its own breakdown. */
export const DualWielding: Story = {
	args: {
		displayStat: {
			stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
			offHandStat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent),
		},
		dualWielding: true,
	},
};

/** A hunter's ranged haste breakdown ends in the bow's effective weapon speed. */
export const HunterRangedHaste: Story = {
	args: {
		displayStat: { stat: UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent) },
		playerClass: Class.ClassHunter,
		rangedWeapon: makeRangedWeapon(),
	},
};

/** The bonus-stat picker writes through the player, so this one runs on a stub host. */
export const EditableBonus: Story = {
	args: { bonusStatsLink: <BonusStatsLink unitStat={STRENGTH} /> },
	decorators: [withSimHost({ player: { getBonusStats: () => new Stats(), setBonusStats: () => {} } })],
};
