import * as PresetUtils from '@app/preset_utils';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import * as Mechanics from '@sim/constants/mechanics';
import { Stats } from '@sim/proto/stats';

import DefaultApl from './apls/default.apl.siml';
import HolyShockApl from './apls/holy_shock.apl.siml';
import UwSorSofApl from './apls/uw_sor_sof.apl.siml';
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
export const UwSorSofAplPreset = PresetUtils.makePresetAPLRotation('Unsealed Wrath SoR/SoF', UwSorSofApl);

export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecRetributionPaladin, Defaults.simpleRotation!);

// Pre-raid sets: dungeon drops, quest and vendor rewards, the Soulforge set, BoE world drops, at most two crafted
// pieces and at most one battleground reputation reward.
export const HolyShockGearPreset = PresetUtils.makePresetGear('Holy Shock', HolyShockGear, {
	tooltip: 'Pre-raid gear for Holy Shock.',
});
export const DefaultGearPreset = PresetUtils.makePresetGear('Default', DefaultGear, {
	tooltip: 'Pre-raid gear for Unsealed Wrath.',
});

// The Manual Crowd Pummeler, which the Holy Shock APL equips before the pull and swaps back in when its cooldown is nearly up.
export const HolyShockItemSwap = PresetUtils.makePresetItemSwapGear('Holy Shock', HolyShockItemSwapGear);

// Simmed on each build's gear set, talents and consumables, which reach the 8% melee hit cap or pass it. More hit adds
// only spell hit there, so the hit weights come from a run 2% under the cap and the other weights from the set itself.
export const HolyShockEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Shock',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.4,
			[Stat.StatAgility]: 1.91,
			[Stat.StatIntellect]: 1.77,
			[Stat.StatSpellDamage]: 2.47,
			[Stat.StatHolyDamage]: 2.06,
			[Stat.StatSpellPiercing]: 0.84,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 7.23,
			[Stat.StatCritRating]: 3,
			[Stat.StatHasteRating]: 2.73,
			[Stat.StatArmorPenetration]: 0.22,
			[Stat.StatExpertiseRating]: 4.81,
			[Stat.StatMP5]: 0.01,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 14.02,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 1.35 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
		},
	),
);

export const UnsealedWrathEpWeights = PresetUtils.makePresetEpWeights(
	'Unsealed Wrath',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 2.4,
			[Stat.StatAgility]: 1.91,
			[Stat.StatIntellect]: 1.48,
			[Stat.StatSpellDamage]: 1.95,
			[Stat.StatHolyDamage]: 1.59,
			[Stat.StatSpellPiercing]: 0.66,
			[Stat.StatAttackPower]: 1,
			[Stat.StatHitRating]: 7.76,
			[Stat.StatCritRating]: 2.74,
			[Stat.StatHasteRating]: 3.53,
			[Stat.StatArmorPenetration]: 0.23,
			[Stat.StatExpertiseRating]: 5.45,
			[Stat.StatMP5]: 0.01,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 14.14,
			[PseudoStat.PseudoStatSchoolHitPercentHoly]: 0.51 * Mechanics.HIT_RATING_PER_HIT_PERCENT,
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
