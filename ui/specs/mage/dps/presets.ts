import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession } from '@generated/proto/common';
import { Mage_Options as MageOptions, MageArmor } from '@generated/proto/mage';
import { SavedTalents } from '@generated/proto/ui';
import type { PresetSettings } from '@sim/presets/types';

import { Phase } from '../../../sim/constants/other';
import arcane from './apls/arcane.apl.json';
import arcaneFire from './apls/arcaneFireFill.apl.json';
import deepFire from './apls/deepFire.apl.json';
import ffb from './apls/ffb.apl.json';
import frost from './apls/frost.apl.json';
//GEAR
import PreRaidArcaneGearJson from './gear_sets/preraid.gear.json';
import PreRaidFireGearJson from './gear_sets/preraidFire.gear.json';
import PreRaidFrostGearJson from './gear_sets/preraidFrost.gear.json';

export const PreRaidArcaneGear = PresetUtils.makePresetGear('Preraid', PreRaidArcaneGearJson, { group: 'Arcane', phase: Phase.Phase1 });
export const PreRaidFireGear = PresetUtils.makePresetGear('Preraid', PreRaidFireGearJson, { group: 'Fire', phase: Phase.Phase1 });
export const PreRaidFrostGear = PresetUtils.makePresetGear('Preraid', PreRaidFrostGearJson, { group: 'Frost', phase: Phase.Phase1 });

//APL
export const ArcaneAplPreset = PresetUtils.makePresetAPLRotation('Arcane', arcane);
export const ArcaneFireAplPreset = PresetUtils.makePresetAPLRotation('Arcane Fire', arcaneFire);
export const FFBAplPreset = PresetUtils.makePresetAPLRotation('FFB', ffb);
export const FireAplPreset = PresetUtils.makePresetAPLRotation('Fire', deepFire);
export const FrostAplPreset = PresetUtils.makePresetAPLRotation('Frost', frost);

//TALENTS
export const ArcaneTalents = PresetUtils.makePresetTalents('Arcane', SavedTalents.create({ talentsString: '050205223101311531-5500003-003' }));
export const ArcaneFireTalents = PresetUtils.makePresetTalents('Arcane Fire', SavedTalents.create({ talentsString: '050205203100311531-525000012001-003' }));
export const FFBTalents = PresetUtils.makePresetTalents('FFB', SavedTalents.create({ talentsString: '-5550000123011303-2035010311013002' }));
export const FireTalents = PresetUtils.makePresetTalents('Fire', SavedTalents.create({ talentsString: '0502052001001-55500001230133031-003' }));
export const FrostIncinTalents = PresetUtils.makePresetTalents('Frost Incin', SavedTalents.create({ talentsString: '203205-5-053500011121331251' }));
export const FrostMedTalents = PresetUtils.makePresetTalents('Frost Med', SavedTalents.create({ talentsString: '2032250100003--053500001121331251' }));

export const SkyborneFireLT180Talents = PresetUtils.makePresetTalents(
	'Fire Skyborne LT 180',
	SavedTalents.create({ talentsString: '0502052011001-54500001230133031-003' }),
);
export const SkyborneFireGT180Talents = PresetUtils.makePresetTalents(
	'Fire Skyborne GT 180',
	SavedTalents.create({ talentsString: '0502052001002-54500001230133031-003' }),
);

/* export const ArcaneMageSimpleRotation = Mage_Rotation.create({
	conserveStart: 20,
	conserveEnd: 30,
	delayMajorCDs: 10,
});
export const APL_ARCANE_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecMage, ArcaneMageSimpleRotation);
*/

export const DefaultOptions = MageOptions.create({
	classOptions: {
		defaultMageArmor: MageArmor.MageArmorMageArmor,
	},
});

export const OtherDefaults = {
	channelClipDelay: 100,
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	reactionTimeMs: 100,
};

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	elixirIds: [Consumables.GreaterArcaneElixir.id, Consumables.ElixirOfTheWhale.id, Consumables.MagebloodElixir.id],
	foodId: Consumables.NightfinSoup.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	mhWeaponImbueId: Consumables.ScrollOfImbuePrecision.id,
	resourcePotId: Consumables.MajorManaPotion.id,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneIntellect: true,
	markOfTheWild: true,
	powerWordFortitude: true,
	moonkinAura: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 2,
	manaTideTotems: 1,
	darkInspiration: 5,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: 1,
	innervates: 1,
});

export const DefaultDebuffs = Debuffs.create({
	curseOfElements: true,
	judgementOfWisdom: true,
});

export const DefaultSettings: PresetSettings = {
	name: 'Default',
	specOptions: DefaultOptions,
	consumables: DefaultConsumables,
	buffs: DefaultIndividualBuffs,
	partyBuffs: DefaultPartyBuffs,
	raidBuffs: DefaultRaidBuffs,
	debuffs: DefaultDebuffs,
	playerOptions: { ...OtherDefaults, channelClipDelayMs: OtherDefaults.channelClipDelay },
};

export const ArcaneBuild = PresetUtils.makePresetBuild('Arcane', {
	group: 'Arcane',
	gear: PreRaidArcaneGear,
	talents: ArcaneTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: ArcaneAplPreset,
	settings: DefaultSettings,
});

export const ArcaneFireBuild = PresetUtils.makePresetBuild('Arcane Fire', {
	group: 'Arcane',
	gear: PreRaidArcaneGear,
	talents: ArcaneFireTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: ArcaneFireAplPreset,
	settings: DefaultSettings,
});

export const FFBBuild = PresetUtils.makePresetBuild('FFB', {
	group: 'FFB',
	gear: PreRaidFireGear,
	talents: FFBTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FFBAplPreset,
	settings: DefaultSettings,
});

export const FireBuild = PresetUtils.makePresetBuild('Fire', {
	group: 'Fire',
	gear: PreRaidFireGear,
	talents: FireTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FireAplPreset,
	settings: { ...DefaultSettings, consumables: { ...DefaultConsumables, mhWeaponImbueId: Consumables.ScrollOfImbueGreaterFlame.id } },
});

export const FrostSettings: PresetSettings = {
	name: 'Frost',
	playerOptions: {
		distanceFromTarget: 28,
	},
};

export const FrostIncinBuild = PresetUtils.makePresetBuild('Frost Incin', {
	group: 'Frost',
	gear: PreRaidFrostGear,
	talents: FrostIncinTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FrostAplPreset,
	settings: FrostSettings,
});

export const FrostMedBuild = PresetUtils.makePresetBuild('Frost Med', {
	group: 'Frost',
	gear: PreRaidFrostGear,
	talents: FrostMedTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FrostAplPreset,
	settings: FrostSettings,
});
