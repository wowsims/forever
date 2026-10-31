import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession } from '@generated/proto/common';
import { Mage_Options as MageOptions, MageArmor } from '@generated/proto/mage';
import { SavedTalents } from '@generated/proto/ui';
import type { PresetSettings } from '@sim/presets/types';

import arcane from './apls/arcane.apl.json';
import arcaneFire from './apls/arcaneFireFill.apl.json';
import deepFire from './apls/deepFire.apl.json';
import ffb from './apls/ffb.apl.json';
import frost from './apls/frost.apl.json';
import P1ArcaneGearJson from './gear_sets/p1Arcane.gear.json';

export const P1ArcaneGear = PresetUtils.makePresetGear('P1 Arcane', P1ArcaneGearJson, { devOnly: true });

export const ArcaneAplPreset = PresetUtils.makePresetAPLRotation('Arcane', arcane);
export const ArcaneFireAplPreset = PresetUtils.makePresetAPLRotation('Arcane Fire', arcaneFire);
export const FFBAplPreset = PresetUtils.makePresetAPLRotation('FFB', ffb);
export const FireAplPreset = PresetUtils.makePresetAPLRotation('Fire', deepFire);
export const FrostAplPreset = PresetUtils.makePresetAPLRotation('Frost', frost);

export const ArcaneTalents = PresetUtils.makePresetTalents('Arcane', SavedTalents.create({ talentsString: '050205223101311531-5500003-003' }));
export const ArcaneFireTalents = PresetUtils.makePresetTalents('Arcane Fire', SavedTalents.create({ talentsString: '050205203100311531-525000012001-003' }));
export const FFBTalents = PresetUtils.makePresetTalents('FFB', SavedTalents.create({ talentsString: '-5550000123011303-2035010311013002' }));
export const FireTalents = PresetUtils.makePresetTalents('Fire', SavedTalents.create({ talentsString: '0502052001001-55500001230133031-003' }));
export const FrostIncinTalents = PresetUtils.makePresetTalents('Frost Incin', SavedTalents.create({ talentsString: '203205-5-053500011121331251' }));
export const FrostMedTalents = PresetUtils.makePresetTalents('Frost Med', SavedTalents.create({ talentsString: '2032250100003--053500001121331251' }));

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
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
	reactionTimeMs: 100,
};

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: Consumables.FlaskOfSupremePower.id,
	elixirIds: [Consumables.GreaterArcaneElixir.id, Consumables.ElixirOfTheSages.id, Consumables.ElixirOfTheWhale.id, Consumables.MagebloodElixir.id],
	foodId: Consumables.NightfinSoup.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	mhWeaponImbueId: Consumables.ScrollOfImbuePrecision.id,
	resourcePotId: Consumables.MajorManaPotion.id,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
	moonkinAura: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 2,
	manaTideTotems: 1,
	darkInspiration: 1,
});

export const DefaultIndividualBuffs = IndividualBuffs.create({
	greaterBlessingOfKings: true,
	greaterBlessingOfWisdom: true,
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
	playerOptions: OtherDefaults,
};

export const ArcaneBuild = PresetUtils.makePresetBuild('Arcane', {
	group: 'Arcane',
	gear: P1ArcaneGear,
	talents: ArcaneTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: ArcaneAplPreset,
	settings: DefaultSettings,
});

export const ArcaneFireBuild = PresetUtils.makePresetBuild('Arcane Fire', {
	group: 'Arcane',
	gear: P1ArcaneGear,
	talents: ArcaneFireTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: ArcaneFireAplPreset,
	settings: DefaultSettings,
});

export const FFBBuild = PresetUtils.makePresetBuild('FFB', {
	group: 'FFB',
	gear: P1ArcaneGear,
	talents: FFBTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FFBAplPreset,
	settings: DefaultSettings,
});

export const FireBuild = PresetUtils.makePresetBuild('Fire', {
	group: 'Fire',
	gear: P1ArcaneGear,
	talents: FireTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FireAplPreset,
	settings: DefaultSettings,
});

export const FrostSettings: PresetSettings = {
	name: 'Frost',
	playerOptions: {
		distanceFromTarget: 28,
	},
};

export const FrostIncinBuild = PresetUtils.makePresetBuild('Frost Incin', {
	group: 'Frost',
	gear: P1ArcaneGear,
	talents: FrostIncinTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FrostAplPreset,
	settings: FrostSettings,
});

export const FrostMedBuild = PresetUtils.makePresetBuild('Frost Med', {
	group: 'Frost',
	gear: P1ArcaneGear,
	talents: FrostMedTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FrostAplPreset,
	settings: FrostSettings,
});
