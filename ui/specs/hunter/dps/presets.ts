import * as PresetUtils from '@app/preset_utils';
import { PseudoStat, Spec, Stat } from '@generated/proto/common';
import { Hunter_Rotation_Playstyle as Playstyle } from '@generated/proto/hunter';
import { Stats } from '@sim/proto/stats';

import CleanWeaveAPL from './apls/clean_weave.apl.siml';
import WeaveAPL from './apls/default.apl.siml';
import WeaveHawkAPL from './apls/default_hawks.apl.siml';
import WeaveMmAPL from './apls/default_mm.apl.siml';
import MeleeAPL from './apls/melee.apl.siml';
import NormalWeaveAPL from './apls/normal_weave.apl.siml';
import TurretAPL from './apls/turret.apl.siml';
import TurretHawkAPL from './apls/turret_hawk.apl.siml';
import TurretMultiAPL from './apls/turret_multi.apl.siml';
import BmBuild from './builds/bm.siml';
import BmWeaveBuild from './builds/bm_weave.siml';
import MmBuild from './builds/mm.siml';
import MmWeaveBuild from './builds/mm_weave.siml';
import SvBuild from './builds/sv.siml';
import SvMeleeBuild from './builds/sv_melee.siml';
import SvWeaveBuild from './builds/sv_weave.siml';
import DefaultsJson from './defaults.siml';
import PreBisMeleeGearJson from './gear_sets/prebis_melee.gear.siml';
import PreBisMeleeCheapGearJson from './gear_sets/prebis_melee_cheap.gear.siml';
import PreBisTurretGearJson from './gear_sets/prebis_turret.gear.siml';
import PreBisTurretCheapGearJson from './gear_sets/prebis_turret_cheap.gear.siml';
import PreBisWeaveGearJson from './gear_sets/prebis_weave.gear.siml';
import PreBisWeaveCheapGearJson from './gear_sets/prebis_weave_cheap.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecHunter, DefaultsJson);

// The best Phase 1 gear outside raids for each playstyle, from batch sims of the SV Weave, MM and SV Melee builds.
// Cheap leaves out exalted reputation, battleground quartermaster gear, epic crafts, the Beastmaster vendor pieces and
// world drops other than random-suffix ones. Neither takes Warsong Gulch or Alterac Valley gear that needs Exalted.
// The weave's two-hander and bow come as a pair: the melee swing resets Auto Shot, so each weave lasts as long as the
// slower of the two swings, and a slow two-hander only pays with a slow bow. The turret sets keep a two-hander for its
// stats and a fast bow, so under the Auto rotation, which weaves with any two-hander, they need a turret rotation.
export const PreBisWeaveGear = PresetUtils.makePresetGear('Pre-BiS', PreBisWeaveGearJson, { group: 'Weave' });
export const PreBisWeaveCheapGear = PresetUtils.makePresetGear('Pre-BiS (Cheap)', PreBisWeaveCheapGearJson, { group: 'Weave' });
export const PreBisTurretGear = PresetUtils.makePresetGear('Pre-BiS', PreBisTurretGearJson, { group: 'Turret' });
export const PreBisTurretCheapGear = PresetUtils.makePresetGear('Pre-BiS (Cheap)', PreBisTurretCheapGearJson, { group: 'Turret' });
export const PreBisMeleeGear = PresetUtils.makePresetGear('Pre-BiS', PreBisMeleeGearJson, { group: 'Melee' });
export const PreBisMeleeCheapGear = PresetUtils.makePresetGear('Pre-BiS (Cheap)', PreBisMeleeCheapGearJson, { group: 'Melee' });

// Measured in Agility on each playstyle's Pre-BiS set and build (SV Weave, MM, SV Melee) over a three minute fight. The
// sets sit at the 8% hit cap, so hit is measured with 20 hit rating taken off, for its value below the cap. Intellect
// is mana and, through Careful Aim, attack power and ranged attack power.
export const WeaveEpWeights = PresetUtils.makePresetEpWeights(
	'Weave',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.21,
			[Stat.StatIntellect]: 0.43,
			[Stat.StatAttackPower]: 0.19,
			[Stat.StatRangedAttackPower]: 0.2,
			[Stat.StatHitRating]: 1.35,
			[Stat.StatCritRating]: 0.8,
			[Stat.StatHasteRating]: 0.51,
			[Stat.StatMP5]: 0.07,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 2.75,
			[PseudoStat.PseudoStatRangedDps]: 1.72,
		},
	),
);
export const TurretEpWeights = PresetUtils.makePresetEpWeights(
	'Turret',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatIntellect]: 0.42,
			[Stat.StatRangedAttackPower]: 0.34,
			[Stat.StatHitRating]: 1.14,
			[Stat.StatCritRating]: 0.84,
			[Stat.StatHasteRating]: 0.8,
			[Stat.StatMP5]: 0.14,
		},
		{
			[PseudoStat.PseudoStatRangedDps]: 2.87,
		},
	),
);
export const MeleeEpWeights = PresetUtils.makePresetEpWeights(
	'Melee',
	Stats.fromMap(
		{
			[Stat.StatAgility]: 1,
			[Stat.StatStrength]: 0.34,
			[Stat.StatIntellect]: 0.11,
			[Stat.StatAttackPower]: 0.31,
			[Stat.StatRangedAttackPower]: 0.13,
			[Stat.StatHitRating]: 1.93,
			[Stat.StatCritRating]: 0.92,
			[Stat.StatHasteRating]: 0.66,
			[Stat.StatMP5]: 0.22,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 3.28,
			[PseudoStat.PseudoStatOffHandDps]: 0.97,
			[PseudoStat.PseudoStatRangedDps]: 0.77,
		},
	),
);

// The priority list of each playstyle. Their "Use ..." variables are the simple rotation's toggles.
export const PlaystyleAPLs: Record<Playstyle, any> = {
	[Playstyle.Turret]: TurretAPL,
	[Playstyle.Weave]: WeaveAPL,
	[Playstyle.Melee]: MeleeAPL,
};
// At range with hawks out, the hawks go ahead of Serpent Sting and Aimed Shot.
export const TurretHawkPlaystyleAPL: any = TurretHawkAPL;

export const TurretAPLRotation = PresetUtils.makePresetAPLRotation('Turret', TurretAPL);
export const TurretMultiAPLRotation = PresetUtils.makePresetAPLRotation('Turret (Multi-Shot)', TurretMultiAPL);
export const TurretHawkAPLRotation = PresetUtils.makePresetAPLRotation('Turret (Hawks)', TurretHawkAPL);
// The max weave walks in after every Auto Shot; the normal weave walks in only once the shots are cooling and stays
// for the melee cooldowns, which suits a dual wielder.
export const WeaveAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (Max)', WeaveAPL);
export const NormalWeaveAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (Normal)', NormalWeaveAPL);
export const WeaveHawkAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (Hawks)', WeaveHawkAPL);
export const WeaveMmAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (MM)', WeaveMmAPL);
// The weave with no toggles: every shot goes ahead of the weave, which walks in once Auto Shot has gone off and back
// out once the hunter has swung.
export const CleanWeaveAPLRotation = PresetUtils.makePresetAPLRotation('Clean Weave', CleanWeaveAPL);
export const MeleeAPLRotation = PresetUtils.makePresetAPLRotation('Melee', MeleeAPL);

export const DefaultRotation = WeaveAPLRotation;

// The simple rotation of the spec's defaults.
export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecHunter, Defaults.simpleRotation!);

const build = (name: string, json: any, gear: PresetUtils.PresetGear): PresetUtils.PresetBuild =>
	PresetUtils.makePresetBuildFromSiml(name, Spec.SpecHunter, json, { gear });

// Gear, talents, rotation, pet and distance together, one per tree and playstyle, each with its playstyle's Pre-BiS set.
export const Builds = [
	build('SV Weave', SvWeaveBuild, PreBisWeaveGear),
	build('MM Weave', MmWeaveBuild, PreBisWeaveGear),
	build('BM Weave', BmWeaveBuild, PreBisWeaveGear),
	build('MM', MmBuild, PreBisTurretGear),
	build('BM', BmBuild, PreBisTurretGear),
	build('SV', SvBuild, PreBisTurretGear),
	build('SV Melee', SvMeleeBuild, PreBisMeleeGear),
];

// Each build's talents, under its name.
export const BuildTalents = Builds.map(build => build.talents!);
