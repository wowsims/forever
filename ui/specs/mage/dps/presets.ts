import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession } from '@generated/proto/common';
import { Mage_Options as MageOptions, MageArmor } from '@generated/proto/mage';
import { SavedTalents } from '@generated/proto/ui';
import type { PresetSettings } from '@sim/presets/types';

import arcaneFire from './apls/arcaneFireFill.apl.json';
import arcaneFrost from './apls/arcaneFrostFill.apl.json';
import deepFire from './apls/deepFire.apl.json';
import ffb from './apls/ffb.apl.json';
import frost from './apls/frost.apl.json';
import P1ArcaneGearJson from './gear_sets/p1Arcane.gear.json';

export const P1ArcaneGear = PresetUtils.makePresetGear('P1 Arcane', P1ArcaneGearJson, { devOnly: true });

export const ArcaneFrostAplPreset = PresetUtils.makePresetAPLRotation('Arcane Frost', arcaneFrost);
export const ArcaneFireAplPreset = PresetUtils.makePresetAPLRotation('Arcane Fire', arcaneFire);
export const FFBAplPreset = PresetUtils.makePresetAPLRotation('FFB', ffb);
export const FireAplPreset = PresetUtils.makePresetAPLRotation('Fire', deepFire);
export const FrostAplPreset = PresetUtils.makePresetAPLRotation('Frost', frost);

export const ArcaneFrostTalents = PresetUtils.makePresetTalents('Arcane Frost', SavedTalents.create({ talentsString: '0552205003100311531-03-0533' }));
export const ArcaneFireTalents = PresetUtils.makePresetTalents('Arcane Fire', SavedTalents.create({ talentsString: '0502205003100311531-230510301001-003' }));
export const FFBTalents = PresetUtils.makePresetTalents('FFB', SavedTalents.create({ talentsString: '-2355000013011503-2035010311013002' }));
export const FireTalents = PresetUtils.makePresetTalents('Fire', SavedTalents.create({ talentsString: '05022050001001-23550000130135031-003' }));
export const FrostTalents = PresetUtils.makePresetTalents('Frost', SavedTalents.create({ talentsString: '05022050011003--053500021100331251' }));

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
	greaterBlessingOfSalvation: true,
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

export const ArcaneFrostBuild = PresetUtils.makePresetBuild('Arcane Frost', {
	group: 'Arcane',
	gear: P1ArcaneGear,
	talents: ArcaneFrostTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: ArcaneFrostAplPreset,
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
		distanceFromTarget: 30,
	},
};

export const FrostBuild = PresetUtils.makePresetBuild('Frost', {
	group: 'Frost',
	gear: P1ArcaneGear,
	talents: FrostTalents,
	//rotationType: APLRotationType.TypeAuto,
	rotation: FrostAplPreset,
	settings: FrostSettings,
});
