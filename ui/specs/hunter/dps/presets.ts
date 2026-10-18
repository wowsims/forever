import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Race, Spec, TristateEffect } from '@generated/proto/common';
import {
	Hunter_Options as HunterOptions,
	Hunter_Rotation,
	Hunter_Rotation_Playstyle as Playstyle,
	HunterOptions_Ammo,
	HunterOptions_PetAttackSpeed,
	HunterOptions_PetType as PetType,
	HunterOptions_QuiverBonus,
} from '@generated/proto/hunter';
import { SavedTalents } from '@generated/proto/ui';

import WeaveAPL from './apls/default.apl.json';
import MeleeAPL from './apls/melee.apl.json';
import TurretAPL from './apls/turret.apl.json';
import TurretHawkAPL from './apls/turret_hawk.apl.json';
import P1GearJson from './gear_sets/p1.gear.json';
import P1MeleeGearJson from './gear_sets/p1_melee.gear.json';

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

// An APL with some of its variables set, so one priority list serves several presets.
const withVariables = (apl: any, values: Record<string, boolean>) => ({
	...apl,
	valueVariables: apl.valueVariables.map((variable: any) =>
		variable.name in values ? { ...variable, value: { const: { val: String(values[variable.name]) } } } : variable,
	),
});

export const TurretAPLRotation = PresetUtils.makePresetAPLRotation('Turret', TurretAPL);
export const TurretMultiAPLRotation = PresetUtils.makePresetAPLRotation(
	'Turret (Multi-Shot)',
	withVariables(TurretAPL, { 'Use Aimed Shot': false, 'Use Multi-Shot': true }),
);
export const TurretHawkAPLRotation = PresetUtils.makePresetAPLRotation('Turret (Hawks)', TurretHawkAPL);
export const WeaveAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave', WeaveAPL);
export const WeaveHawkAPLRotation = PresetUtils.makePresetAPLRotation('Melee Weave (Hawks)', withVariables(WeaveAPL, { 'Use Summon Hawk': true }));
export const MeleeAPLRotation = PresetUtils.makePresetAPLRotation('Melee', MeleeAPL);

export const DefaultRotation = WeaveAPLRotation;

export const DefaultSimpleRotation = Hunter_Rotation.create({
	playstyle: Playstyle.Weave,
	timeToWeave: 400,
	useMulti: true,
	useArcane: true,
});
export const SimpleRotation = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecHunter, DefaultSimpleRotation);

const talents = (name: string, talentsString: string) => ({ name, data: SavedTalents.create({ talentsString }) });

// The best build found for each tree and playstyle, on the P1 gear over a three minute fight.
export const SurvivalWeaveTalents = talents('SV Weave', '-00530501114-550200030050220151');
export const MarksmanshipWeaveTalents = talents('MM Weave', '-205305031150001-5502000300502201');
export const BeastMasteryWeaveTalents = talents('BM Weave', '5320001505101251-0053250001-4');
export const MarksmanshipTalents = talents('MM', '5-005355251152205-5');
export const BeastMasteryTalents = talents('BM', '5320001505101251-0053250001-4');
export const SurvivalTalents = talents('SV', '1-00531500114-52225303000021005');
export const SurvivalMeleeTalents = talents('SV Melee', '53200005021-005005-5502000300500201');

const classOptions = {
	ammo: HunterOptions_Ammo.ThoriumHeadedArrow,
	quiverBonus: HunterOptions_QuiverBonus.AncientSinewWrappedLamina,
	petUptime: 1,
	petAttackSpeed: HunterOptions_PetAttackSpeed.FasterAttackII,
	cobraReflexes: true,
	petAggression: 5,
};

// Lone Wolf wants no pet out; the Beast Mastery builds bring the raptor.
export const NoPetOptions = HunterOptions.create({ classOptions: { ...classOptions, petType: PetType.PetNone } });
export const PetOptions = HunterOptions.create({ classOptions: { ...classOptions, petType: PetType.Raptor } });

export const DefaultOptions = NoPetOptions;

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({});

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	prayerOfSpirit: true,
	giftOfTheWild: true,
	graceOfAirTotem: true,
	strengthOfEarthTotem: true,
	battleShout: TristateEffect.TristateEffectRegular,
	leaderOfThePack: true,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	curseOfRecklessness: true,
	exposeArmor: true,
	faerieFire: true,
	giftOfArthas: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	sunderArmor: true,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.GreaterMagebloodElixir.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	resourcePotId: Consumables.MajorManaPotion.id,
	conjuredId: Consumables.DemonicRune.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.ScrollOfStrengthIV.id],
});

export const OtherDefaults = {
	distanceFromTarget: 11,
	iterationCount: 25000,
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
};

const build = (
	name: string,
	gear: PresetUtils.PresetGear,
	talentPreset: { name: string; data: SavedTalents },
	rotation: PresetUtils.PresetRotation,
	options: HunterOptions,
	distanceFromTarget: number,
) =>
	PresetUtils.makePresetBuild(name, {
		gear,
		talents: talentPreset,
		rotation,
		settings: { name, specOptions: options, playerOptions: { distanceFromTarget } },
	});

// Gear, talents, rotation, pet and distance together, one per tree and playstyle.
export const Builds = [
	build('SV Weave', P1Gear, SurvivalWeaveTalents, WeaveAPLRotation, NoPetOptions, 11),
	build('MM Weave', P1Gear, MarksmanshipWeaveTalents, WeaveAPLRotation, NoPetOptions, 11),
	build('BM Weave', P1Gear, BeastMasteryWeaveTalents, WeaveHawkAPLRotation, PetOptions, 11),
	build('MM', P1Gear, MarksmanshipTalents, TurretAPLRotation, NoPetOptions, 11),
	build('BM', P1Gear, BeastMasteryTalents, TurretHawkAPLRotation, PetOptions, 11),
	build('SV', P1Gear, SurvivalTalents, TurretMultiAPLRotation, NoPetOptions, 11),
	build('SV Melee', P1MeleeGear, SurvivalMeleeTalents, MeleeAPLRotation, PetOptions, 5),
];
