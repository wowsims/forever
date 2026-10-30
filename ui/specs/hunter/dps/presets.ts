import * as PresetUtils from '@app/preset_utils';
import { Spec } from '@generated/proto/common';
import { Hunter_Rotation_Playstyle as Playstyle } from '@generated/proto/hunter';

import CleanWeaveAPL from './apls/clean_weave.apl.siml';
import WeaveAPL from './apls/default.apl.siml';
import WeaveHawkAPL from './apls/default_hawks.apl.siml';
import MeleeAPL from './apls/melee.apl.siml';
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
import P1GearJson from './gear_sets/p1.gear.siml';
import P1MeleeGearJson from './gear_sets/p1_melee.gear.siml';

// The spec's defaults, from defaults.siml.
export const Defaults = PresetUtils.specDefaultsFromSettings(Spec.SpecHunter, DefaultsJson);

// The same armor twice, with the weapons each playstyle swings: a two-hander to weave with, which
// is also the better pair of stats at range, and two one-handers to stand in melee with.
export const P1Gear = PresetUtils.makePresetGear('P1', P1GearJson, { devOnly: true });
export const P1MeleeGear = PresetUtils.makePresetGear('P1 Dual Wield', P1MeleeGearJson, { devOnly: true });

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
export const WeaveAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave', WeaveAPL);
export const WeaveHawkAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (Hawks)', WeaveHawkAPL);
// The weave with no toggles: every shot goes ahead of the weave, and a weave is a walk in for the next
// swing that is back before the Auto Shot it delays is more than Max Auto Shot delay late (Max long weave
// delay when one walk in and out outlasts an Auto Shot cycle). Walk times come from the run speed.
export const CleanWeaveAPLRotation = PresetUtils.makePresetAPLRotation('Clean Weave', CleanWeaveAPL);
export const MeleeAPLRotation = PresetUtils.makePresetAPLRotation('Melee', MeleeAPL);

export const DefaultRotation = WeaveAPLRotation;

// The simple rotation of the spec's defaults.
export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecHunter, Defaults.simpleRotation!);

// A build from `json`, which leaves "in front of target" and item swap as the page has them: a .siml file has no
// unset bool, so its settings: section turns both off.
const build = (name: string, json: any, gear: PresetUtils.PresetGear): PresetUtils.PresetBuild => {
	const preset = PresetUtils.makePresetBuildFromJSON(name, Spec.SpecHunter, json, { gear });
	const { inFrontOfTarget: _inFrontOfTarget, enableItemSwap: _enableItemSwap, ...playerOptions } = preset.settings!.playerOptions!;
	return { ...preset, settings: { ...preset.settings!, playerOptions } };
};

// Gear, talents, rotation, pet and distance together, one per tree and playstyle: the best build found for each on
// the P1 gear over a three minute fight.
export const Builds = [
	build('SV Weave', SvWeaveBuild, P1Gear),
	build('MM Weave', MmWeaveBuild, P1Gear),
	build('BM Weave', BmWeaveBuild, P1Gear),
	build('MM', MmBuild, P1Gear),
	build('BM', BmBuild, P1Gear),
	build('SV', SvBuild, P1Gear),
	build('SV Melee', SvMeleeBuild, P1MeleeGear),
];

// Each build's talents, under its name.
export const BuildTalents = Builds.map(build => build.talents!);
