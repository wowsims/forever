import * as PresetUtils from '@app/preset_utils';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import * as Mechanics from '@sim/constants/mechanics';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.siml';
import HolyShockApl from './apls/holy_shock.apl.siml';
import HolyShockBuildJson from './builds/holy_shock.siml';
import UnsealedWrathBuildJson from './builds/unsealed_wrath.siml';
import DefaultsJson from './defaults.siml';
import DefaultGear from './gear_sets/default.gear.siml';
import HolyShockGear from './gear_sets/holy_shock.gear.siml';
import HolyShockItemSwapGear from './gear_sets/holy_shock_item_swap.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecRetributionPaladin, DefaultsJson);

export const HolyShockAplPreset = PresetUtils.makePresetAPLRotation('Holy Shock', HolyShockApl);
export const DefaultAplPreset = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecRetributionPaladin, Defaults.simpleRotation!);

export const HolyShockGearPreset = PresetUtils.makePresetGear('Holy Shock', HolyShockGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});
export const DefaultGearPreset = PresetUtils.makePresetGear('Default', DefaultGear, {
	devOnly: true,
	tooltip: 'Level 60 blues from dungeons, quests, reputations and crafting. A placeholder until the Phase 1 gear is known.',
});

// The Manual Crowd Pummeler, which the Holy Shock APL equips before the pull and swaps back in when its cooldown is nearly up.
export const HolyShockItemSwap = PresetUtils.makePresetItemSwapGear('Holy Shock', HolyShockItemSwapGear);

export const HolyShockEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Shock',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.5,
			[Stat.StatAgility]: 3.11,
			[Stat.StatIntellect]: 4.53,
			[Stat.StatSpellDamage]: 5.78,
			[Stat.StatSpellPiercing]: 0.01,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 10.25,
			[Stat.StatCritRating]: 4.66,
			[Stat.StatHasteRating]: 4.06,
			[Stat.StatArmorPenetration]: 0.26,
			[Stat.StatExpertiseRating]: 6.2,
			[Stat.StatMP5]: 0.01,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.04,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 2.32 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const UnsealedWrathEpWeights = PresetUtils.makePresetEpWeights(
	'Unsealed Wrath',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.4,
			[Stat.StatAgility]: 2.23,
			[Stat.StatIntellect]: 2.06,
			[Stat.StatSpellDamage]: 2.36,
			[Stat.StatSpellPiercing]: 0.01,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 6.83,
			[Stat.StatCritRating]: 3.08,
			[Stat.StatHasteRating]: 22.87,
			[Stat.StatArmorPenetration]: 0.24,
			[Stat.StatExpertiseRating]: 5.1,
			[Stat.StatMP5]: 0.23,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 13.85,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 0.6 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

// A build that keeps the page's simple rotation, which picks the Default or the Holy Shock APL from the talents. Item
// swap is the build's itemSwap, on with it and off without.
const build = (name: string, json: any, options: PresetUtils.PresetBuildOptions): PresetUtils.PresetBuild =>
	PresetUtils.makePresetBuildFromSiml(name, Spec.SpecRetributionPaladin, json, { rotationType: APLRotationType.TypeSimple, ...options });

export const DefaultBuild = build('Unsealed Wrath', UnsealedWrathBuildJson, { gear: DefaultGearPreset });
export const HolyShockBuild = build('Holy Shock', HolyShockBuildJson, { gear: HolyShockGearPreset, itemSwap: HolyShockItemSwap });

export const UnsealedWrathTalents = DefaultBuild.talents!;
export const HolyShockTalents = HolyShockBuild.talents!;

// Each build's consumables, buffs and player settings, the Holy Shock ones with the Manual Crowd Pummeler swap on.
export const DefaultSettings = { ...DefaultBuild.settings!, name: 'Default' };
export const HolyShockSettings = {
	...HolyShockBuild.settings!,
	playerOptions: { ...HolyShockBuild.settings!.playerOptions, enableItemSwap: true, itemSwap: HolyShockItemSwap.itemSwap },
};
