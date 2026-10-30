import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Spec, TristateEffect } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import {
	Warlock_Options as WarlockOptions,
	Warlock_Rotation as WarlockSimpleRotation,
	WarlockOptions_Armor,
	WarlockOptions_CurseOptions,
	WarlockOptions_Summon,
	WarlockRotation,
} from '@generated/proto/warlock';

import AffWrackBuildJson from './builds/aff_wrack.build.json';
import DemoFireBrandBuildJson from './builds/demo_fire_brand.build.json';
import DemoShadowBrandBuildJson from './builds/demo_shadow_brand.build.json';
import DestroBuildJson from './builds/destro.build.json';
import SmRuinBuildJson from './builds/sm_ruin.build.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const DEMO_FIRE_BRAND_APL = PresetUtils.makePresetAPLRotation('Demo (Fire Brand)', DemoFireBrandBuildJson.player.rotation);
export const DEMO_SHADOW_BRAND_APL = PresetUtils.makePresetAPLRotation('Demo (Shadow Brand)', DemoShadowBrandBuildJson.player.rotation);
export const DESTRO_APL = PresetUtils.makePresetAPLRotation('Destro', DestroBuildJson.player.rotation);
export const SM_RUIN_APL = PresetUtils.makePresetAPLRotation('SM/Ruin', SmRuinBuildJson.player.rotation);
export const AFF_WRACK_APL = PresetUtils.makePresetAPLRotation('Aff (Wrack)', AffWrackBuildJson.player.rotation);

// The build rotation each simple rotation choice runs.
export const SimpleRotationApls: Record<WarlockRotation, PresetUtils.PresetRotation> = {
	[WarlockRotation.WarlockRotationNone]: DEMO_FIRE_BRAND_APL,
	[WarlockRotation.WarlockRotationDemoFireBrand]: DEMO_FIRE_BRAND_APL,
	[WarlockRotation.WarlockRotationDemoShadowBrand]: DEMO_SHADOW_BRAND_APL,
	[WarlockRotation.WarlockRotationDestro]: DESTRO_APL,
	[WarlockRotation.WarlockRotationSmRuin]: SM_RUIN_APL,
	[WarlockRotation.WarlockRotationAffWrack]: AFF_WRACK_APL,
};

export const DefaultSimpleRotation = WarlockSimpleRotation.create({ rotation: WarlockRotation.WarlockRotationDemoFireBrand });

const talentsOf = (name: string, json: { player: { talentsString: string } }) =>
	PresetUtils.makePresetTalents(name, SavedTalents.create({ talentsString: json.player.talentsString }));

export const DemoFireBrandTalents = talentsOf('Demo (Fire Brand)', DemoFireBrandBuildJson);
export const DemoShadowBrandTalents = talentsOf('Demo (Shadow Brand)', DemoShadowBrandBuildJson);
export const DestroTalents = talentsOf('Destro', DestroBuildJson);
export const SmRuinTalents = talentsOf('SM/Ruin', SmRuinBuildJson);
export const AffWrackTalents = talentsOf('Aff (Wrack)', AffWrackBuildJson);

// Defaults
export const DefaultOptions = WarlockOptions.create({
	classOptions: {
		armor: WarlockOptions_Armor.DemonArmor,
		curseOptions: WarlockOptions_CurseOptions.Elements,
		sacrificedDemon: WarlockOptions_Summon.Succubus,
		summon: WarlockOptions_Summon.Imp,
		// A full Felcloth Bag (21341), the largest Soul Bag a warlock can carry.
		soulShards: 32,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	foodId: Consumables.NightfinSoup.id,
	conjuredId: Consumables.DemonicRune.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	offensivePotId: Consumables.MajorSpellblastingPotion.id,
	explosiveId: Consumables.EzThroDynamiteII.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultRaidBuffs = RaidBuffs.create({
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	prayerOfSpirit: true,
	moonkinAura: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: TristateEffect.TristateEffectRegular,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
});

export const DefaultDebuffs = Debuffs.create({
	judgementOfWisdom: true,
});

// Each build file holds the talents, demons, consumables, buffs and rotation of one tree and playstyle.
const build = (name: string, json: unknown) => PresetUtils.makePresetBuildFromJSON(name, Spec.SpecWarlock, json);

export const Builds = [
	build('Demo (Fire Brand)', DemoFireBrandBuildJson),
	build('Demo (Shadow Brand)', DemoShadowBrandBuildJson),
	build('Destro', DestroBuildJson),
	build('SM/Ruin', SmRuinBuildJson),
	build('Aff (Wrack)', AffWrackBuildJson),
];
