import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, HealingModel, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import { ProtectionWarrior_Options as ProtectionWarriorOptions, WarriorStance } from '@generated/proto/warrior';
import { Stats } from '@sim/proto/stats';
import { OtherDefaults as SimUIOtherDefaults } from '@sim/spec_config';

import * as WarriorPresets from '../shared/presets';
import GenericApl from './apls/default.apl.json';
import PreBiSGearJson from './gear_sets/prebis.gear.json';
import PreBiSThreatGearJson from './gear_sets/prebis_threat.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

// The rotation hides Autocast Other Cooldowns, so an on-use trinket fires only when the "CD: Offensive Trinkets" group
// names its item id.
export const ROTATION_DEFAULT = PresetUtils.makePresetAPLRotation('Generic', GenericApl);

// Dungeon and world drops, quest rewards, the warrior's dungeon tier set, crafted items whose reagents include no raid
// materials and vendor items; no reputation, PvP or Silithus rewards. The most threat at 140 defense, the least that
// makes a level 63 boss unable to crit. Cragfists of Defense drops in Uldaman and the other "of Defense" pieces are
// world-drop greens; Blackstone Ring drops in Maraudon; Howler's Furs, Ring of Awareness and Defender's Grip Stabilizer
// are sold by the Undermine Trader in Stranglethorn Vale; Rune of the Guard Captain is a Hinterlands quest reward. Horn
// of Ve'han, Furbolg Berserker's Boots and the Sentinel's Wristguards and Waistguard plans have no source in the item
// database.
export const PreBiSGear = PresetUtils.makePresetGear('Pre-BiS (Defense Cap)', PreBiSGearJson, { devOnly: true });

// The same sources with no defense floor: the most threat that survives the default target on the default healing.
// Vambraces of the Sadist and Crest of Retribution drop in Stratholme, Blackstone Ring in Maraudon and Riphook in
// Lower Blackrock Spire; Royal Seal of Eldre'Thalas is a Dire Maul quest reward and Mark of Fordring the reward for In
// Dreams. Howler's Furs, Valiant's Waistguard, Cutthroat's Signet and Adaptive Combat Assistant are sold by the
// Undermine Trader. Spaulders of Containment, Gauntlets of Heroism, Furbolg Berserker's Boots, the Dawn Armor pattern
// and its Pristine Scale and Fel Leather have no source in the item database.
export const PreBiSThreatGear = PresetUtils.makePresetGear('Pre-BiS (Threat)', PreBiSThreatGearJson, { devOnly: true });

export const ProtectionTalents = PresetUtils.makePresetTalents('Protection', SavedTalents.create({ talentsString: '05-05050005-150531120300001351' }));

// Measured on each Pre-BiS gear set with the preset talents and rotation, against the default target on the default
// healing: 0.6 x the threat weights (Strength = 1) plus 1.15 x the TMI weights (Stamina = 1). On the defense cap gear,
// defense is its value past the cap; both sets sit at the 8% hit cap, and hit is its value past it. ArmorNoMods weighs
// as Armor; the healing model's Inspiration scales only Armor, and the two measure less than 0.01 apart.
export const PreBiSEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS (Defense Cap)',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.62,
			[Stat.StatAgility]: 0.53,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 0.03,
			[Stat.StatCritRating]: 0.44,
			[Stat.StatHasteRating]: 0.2,
			[Stat.StatExpertiseRating]: 2.13,
			[Stat.StatArmorPenetration]: 0.11,
			[Stat.StatDefenseRating]: 0.22,
			[Stat.StatBlockRating]: 0.02,
			[Stat.StatBlockValue]: 0.62,
			[Stat.StatDodgeRating]: 0.16,
			[Stat.StatParryRating]: 0.16,
			[Stat.StatArmor]: 0.04,
			[Stat.StatBonusArmor]: 0.04,
			[Stat.StatArmorNoMods]: 0.04,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.28,
		},
	),
);

export const PreBiSThreatEpWeights = PresetUtils.makePresetEpWeights(
	'Pre-BiS (Threat)',
	Stats.fromMap(
		{
			[Stat.StatStamina]: 1.15,
			[Stat.StatStrength]: 0.62,
			[Stat.StatAgility]: 0.5,
			[Stat.StatAttackPower]: 0.27,
			[Stat.StatHitRating]: 0.04,
			[Stat.StatCritRating]: 0.43,
			[Stat.StatHasteRating]: 0.22,
			[Stat.StatExpertiseRating]: 2.19,
			[Stat.StatArmorPenetration]: 0.11,
			[Stat.StatDefenseRating]: 0.29,
			[Stat.StatBlockRating]: 0.03,
			[Stat.StatBlockValue]: 0.61,
			[Stat.StatDodgeRating]: 0.13,
			[Stat.StatParryRating]: 0.14,
			[Stat.StatArmor]: 0.04,
			[Stat.StatBonusArmor]: 0.04,
			[Stat.StatArmorNoMods]: 0.04,
		},
		{
			[PseudoStat.PseudoStatMainHandDps]: 1.3,
		},
	),
);

export const DefaultOptions = ProtectionWarriorOptions.create({
	classOptions: {
		queueDelay: 250,
		startingRage: 0,
		useBattleShout: true,
		defaultStance: WarriorStance.WarriorStanceDefensive,
		hasBsT2: false,
		stanceSnapshot: true,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	...WarriorPresets.DefaultConsumables,
	offensivePotId: Consumables.MajorFrenzyPotion.id,
	elixirIds: [
		Consumables.ElixirOfFerocity.id,
		Consumables.ElixirOfTheMongoose.id,
		Consumables.WinterfallFirewater.id,
		Consumables.ElixirOfGreaterDefense.id,
		Consumables.ElixirOfThePhalanx.id,
	],
	mhImbueId: Consumables.DenseSharpeningStone.id,
	conjuredId: Consumables.MajorHealthstone9421.id,
	extraIds: [Consumables.GoblinSapperCharge.id, Consumables.RumseyRum.id],
});

// Classic Era Molten Core logs: the healing that lands on the main tank, overhealing included, comes to 2.4 x
// its damage taken, in heals about 0.3 s apart, 1.5% of it as absorbs. The HPS is 2.4 x the 315 damage taken per
// second of the Pre-BiS (Defense Cap) gear against the default target, rounded down to 750; the Inspiration uptime is
// not from the logs.
export const OtherDefaults: Partial<SimUIOtherDefaults> = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	race: Race.RaceOrc,
	distanceFromTarget: 0,
	healingModel: HealingModel.create({
		hps: 750,
		cadenceSeconds: 0.3,
		cadenceVariation: 0.64,
		absorbFrac: 0.015,
		burstWindow: 6,
		inspirationUptime: 0.25,
	}),
};
