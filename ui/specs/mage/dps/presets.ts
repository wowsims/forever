import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Profession, Spec } from '@generated/proto/common';
import { Mage_Options as MageOptions, Mage_Rotation, MageArmor } from '@generated/proto/mage';
import { SavedTalents } from '@generated/proto/ui';

import DefaultApl from './apls/default.apl.json';
import P1ArcaneGearJson from './gear_sets/p1Arcane.gear.json';

export const P1ArcaneGear = PresetUtils.makePresetGear('P1 Arcane', P1ArcaneGearJson);

export const ArcaneTalents = {
	name: 'Arcane',
	data: SavedTalents.create({
		talentsString: '0502205003100301531--055500032',
	}),
};

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const DEFAULT_APL = PresetUtils.makePresetAPLRotation('Default', DefaultApl);

export const ArcaneMageSimpleRotation = Mage_Rotation.create({
	conserveStart: 20,
	conserveEnd: 30,
	delayMajorCDs: 10,
});

export const APL_ARCANE_SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecMage, ArcaneMageSimpleRotation);

export const DefaultOptions = MageOptions.create({
	classOptions: {
		defaultMageArmor: MageArmor.MageArmorMageArmor,
	},
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Engineering,
	profession2: Profession.Tailoring,
};

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.GreaterArcaneElixir.id, Consumables.ElixirOfTheSages.id],
	foodId: Consumables.NightfinSoup.id,
	mhImbueId: Consumables.BrilliantWizardOil.id,
	resourcePotId: Consumables.MajorManaPotion.id,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	prayerOfSpirit: true,
	arcaneBrilliance: true,
	giftOfTheWild: true,
	prayerOfFortitude: true,
});

export const DefaultPartyBuffs = PartyBuffs.create({
	manaSpringTotem: 2,
	manaTideTotems: 1,
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
