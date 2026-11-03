import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import { APLRotation } from '@generated/proto/apl';
import { Debuffs } from '@generated/proto/buffs';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defaultHealerIndividualBuffs, defaultHealerPartyBuffs, defaultHealerRaidBuffs } from '@sim/proto/utils';
import { defineSpec } from '@sim/spec_config';

import * as Presets from './presets';

// Gear planner only: no healing spells are implemented, so there is no simulation for this spec.
export default defineSpec<Spec.SpecHolyPaladin>({
	spec: Spec.SpecHolyPaladin,
	// Nothing is simulated, so the incoming-healing model stays off.
	enableHealing: false,

	className: 'holy-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],

	// All stats for which EP should be calculated.
	epStats: [Stat.StatIntellect, Stat.StatSpirit, Stat.StatHealingPower, Stat.StatCritRating, Stat.StatHasteRating, Stat.StatMP5],
	// Reference stat against which to calculate EP.
	epReferenceStat: Stat.StatHealingPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[Stat.StatHealth, Stat.StatMana, Stat.StatStamina, Stat.StatIntellect, Stat.StatSpirit, Stat.StatHealingPower, Stat.StatSpellDamage, Stat.StatMP5],
		[PseudoStat.PseudoStatSpellCritPercent, PseudoStat.PseudoStatSpellHastePercent],
	),

	defaults: {
		// Default equipped gear.
		gear: Presets.GEAR_DEFAULT.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: new Stats(),
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: SavedTalents.create(),
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
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	// Stamina is not an EP stat for healers, but the buff still belongs in the stats panel.
	includeBuffDebuffInputs: [BuffDebuffInputs.PowerWordFortitude],
	// Nothing is simulated, so buffs that only matter inside an encounter (damage, cooldowns,
	// mana returns over a fight) would only mislead.
	excludeBuffDebuffInputs: [BuffDebuffInputs.Thorns, BuffDebuffInputs.Innervate, BuffDebuffInputs.PowerInfusion, BuffDebuffInputs.ManaTideTotem],
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [],
	},
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [],
		// Preset talents that the user can quickly select.
		talents: [],
		// Preset rotations that the user can quickly select.
		rotations: [],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.GEAR_DEFAULT],
	},

	autoRotation: (_: Player<Spec.SpecHolyPaladin>): APLRotation => {
		return APLRotation.create();
	},
});
