import * as InputHelpers from '@features/inputs/input_helpers';
import { HealingSettings } from '@features/settings';
import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import { APLRotation } from '@generated/proto/apl';
import { Debuffs } from '@generated/proto/buffs';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PrayerOfHealingMode } from '@generated/proto/priest';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defaultHealerIndividualBuffs, defaultHealerPartyBuffs, defaultHealerRaidBuffs } from '@sim/proto/utils';
import { defineSpec } from '@sim/spec_config';

import * as PriestInputs from '../shared/inputs';
import * as Presets from './presets';

const oneToFiveValues = [1, 2, 3, 4, 5].map(value => ({ name: String(value), value }));

export default defineSpec<Spec.SpecHealerPriest>({
	spec: Spec.SpecHealerPriest,
	enableHealing: true,
	className: 'healer-priest-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Priest),
	knownIssues: ['Healing sim is Experimental, all results should be taken with a grain of salt', 'Downranking Penalty is not implemented yet'],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatIntellect, Stat.StatSpirit, Stat.StatHealingPower, Stat.StatCritRating, Stat.StatHasteRating, Stat.StatMP5, Stat.StatMana],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatHealingPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatMana, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpirit, Stat.StatHealingPower, Stat.StatSpellDamage, Stat.StatMP5],
		[PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),
	modifyDisplayStats: player => ({
		talents: new Stats().addPseudoStat(PseudoStat.PseudoStatSpellCritPercent, player.getTalents().holySpecialization),
	}),

	defaults: {
		// Default equipped gear.
		gear: Presets.PreRaidGear.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.HolyTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: defaultHealerRaidBuffs(),
		partyBuffs: defaultHealerPartyBuffs(),
		individualBuffs: defaultHealerIndividualBuffs(),
		debuffs: Debuffs.create({}),
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [PriestInputs.ArmorInput()],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	// Stamina is not an EP stat for healers, but the buff still belongs in the stats panel.
	includeBuffDebuffInputs: [BuffDebuffInputs.PrayerOfFortitude],
	// These encounter buffs do not affect the healer's current spell model.
	excludeBuffDebuffInputs: [
		BuffDebuffInputs.Thorns,
		BuffDebuffInputs.PowerInfusion,
		BuffDebuffInputs.AtieshDruid,
		BuffDebuffInputs.AtieshMage,
		BuffDebuffInputs.AtieshPriest,
		BuffDebuffInputs.AtieshWarlock,
	],
	// Spell-specific controls shown in the Healing Inputs section.
	healingInputs: {
		inputs: [
			InputHelpers.makeSpecOptionsEnumInput<Spec.SpecHealerPriest, number>({
				fieldName: 'holyNovaHealTargets',
				label: 'Holy Nova heal targets',
				labelTooltip: "Number of party members, including the priest, randomly selected to receive Holy Nova's heal.",
				getValue: player => player.getSpecOptions().holyNovaHealTargets || 5,
				values: oneToFiveValues,
				showWhen: player => player.getTalents().holyNova,
				storeField: ['specOptions', 'talentsString'] as const,
			}),
			InputHelpers.makeSpecOptionsEnumInput<Spec.SpecHealerPriest, PrayerOfHealingMode>({
				fieldName: 'prayerOfHealingMode',
				label: 'Prayer of Healing targets',
				labelTooltip: 'Randomly models movement by choosing between one and five party members on each cast.',
				values: [
					{ name: 'Full party', value: PrayerOfHealingMode.FullParty },
					{ name: 'Dynamic', value: PrayerOfHealingMode.Dynamic },
				],
			}),
			InputHelpers.makeSpecOptionsEnumInput<Spec.SpecHealerPriest, number>({
				fieldName: 'prayerOfMendingJumps',
				label: 'Prayer of Mending heals',
				labelTooltip: 'Number of heals per cast, the first one included. Each heal spends one of the 5 charges.',
				getValue: player => player.getSpecOptions().prayerOfMendingJumps || 5,
				values: oneToFiveValues,
				showWhen: player => player.getTalents().prayerOfMending,
				storeField: ['specOptions', 'talentsString'] as const,
			}),
			InputHelpers.makeSpecOptionsNumberInput<Spec.SpecHealerPriest>({
				fieldName: 'prayerOfMendingDelay',
				label: 'Prayer of Mending delay (seconds)',
				labelTooltip: 'Time between each heal and jump. The first heal occurs after this delay.',
				getValue: player => player.getSpecOptions().prayerOfMendingDelay ?? 2.5,
				float: true,
				positive: true,
				max: 5,
				maxDecimalDigits: 2,
				showWhen: player => player.getTalents().prayerOfMending,
				storeField: ['specOptions', 'talentsString'] as const,
			}),
		],
	},
	otherInputs: { inputs: [] },
	settingsComponentUnderPlayer: HealingSettings,
	hideDebuffs: true,
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.DiscRaidHealingEpWeights, Presets.DiscSingleTargetEpWeights, Presets.HolySingleTargetEpWeights, Presets.HolyRaidHealingEpWeights],
		// Preset talents that the user can quickly select.
		talents: [Presets.DiscTalents, Presets.HolyTalents],
		// Preset rotations that the user can quickly select.
		rotations: [Presets.ROTATION_DISC_DEFAULT, Presets.ROTATION_HOLY_DEFAULT],
		// Preset gear configurations that the user can quickly select.
		gear: [],
	},

	autoRotation: (priest: Player<Spec.SpecHealerPriest>): APLRotation => {
		const [disc, holy] = priest.getTalentTreePoints();

		if (disc > holy) {
			return Presets.ROTATION_DISC_DEFAULT.rotation.rotation!;
		}
		return Presets.ROTATION_HOLY_DEFAULT.rotation.rotation!;
	},

	// The gear optimizer.
	reforge: {},
});
