import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Race, Stat } from '@generated/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import BearweaveApl from './apls/bearweave.apl.json';
import DefaultApl from './apls/default.apl.json';
import EnergizedApl from './apls/energized.apl.json';
import JvmesApl from './apls/jvmes.apl.json';
import PounceApl from './apls/pounce.apl.json';
import RavageApl from './apls/ravage.apl.json';
import ShadowmeldApl from './apls/shadowmeld.apl.json';
import Level60GearJson from './gear_sets/default.gear.json';
import PreRaidGearJson from './gear_sets/pre_raid.gear.json';

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = FeralDruidOptions.create({ assumeBleedActive: true });

export const DefaultConsumables = ConsumesSpec.create({
	resourcePotId: Consumables.MightyRagePotion.id,
	flaskId: Consumables.FlaskOfNaturalPrecision.id,
	elixirIds: [Consumables.ElixirOfTheMongoose.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	mhImbueId: Consumables.ElementalSharpeningStone.id,
	conjuredId: Consumables.ThistleTea.id,
	extraIds: [Consumables.JujuPower.id, Consumables.JujuMight.id, Consumables.GroundScorpokAssay.id],
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	reactionTime: 250,
};

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

// Pre-raid leather on a Night Elf with The Unstoppable Force and Wolfshead Helm. Shifting Cloak, Dunestalker's
// Boots, Ferocious Watcher's Signet and the two hit and attack power trinkets stand in for Cape of the Black
// Baron, Swiftwalker Boots, Tarnished Elven Ring, Blackhand's Breadth and Hand of Justice, which the item
// database does not carry.
export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

// EP estimates with agility as 1. Hit is weighed 2% below the 8% cap.
// Feral attack power uses attack power's weight through the form-only dependency.
export const Level60Ep = Stats.fromMap({
	[Stat.StatStrength]: 0.98,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.41,
	[Stat.StatFeralAttackPower]: 0.41,
	[Stat.StatPhysicalDamage]: 3.72,
	[Stat.StatHitRating]: 0.73,
	[Stat.StatExpertiseRating]: 0.69,
	[Stat.StatCritRating]: 0.72,
	[Stat.StatHasteRating]: 0.38,
	[Stat.StatArmorPenetration]: 0.12,
});

export const Level60EpWeights = PresetUtils.makePresetEpWeights('Level 60', Level60Ep);

// EP estimates with agility as 1. Hit is weighed 2% below the 8% cap.
export const PreRaidEp = Stats.fromMap({
	[Stat.StatStrength]: 0.97,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.39,
	[Stat.StatFeralAttackPower]: 0.39,
	[Stat.StatPhysicalDamage]: 3.56,
	[Stat.StatHitRating]: 0.75,
	[Stat.StatExpertiseRating]: 0.73,
	[Stat.StatCritRating]: 0.75,
	[Stat.StatHasteRating]: 0.43,
	[Stat.StatArmorPenetration]: 0.13,
});

export const PreRaidEpWeights = PresetUtils.makePresetEpWeights('Pre-raid', PreRaidEp);

// The cat stays in Cat Form: Shifting Power, Rip, Rake and Shred, Faerie Fire for Omen of Clarity,
// and a self-Innervate when mana would run short.
export const MONOCAT_APL = PresetUtils.makePresetAPLRotation('Monocat', DefaultApl);
// Monocat with Shadowmeld into Ravage on cooldown, for a Night Elf.
export const SHADOWMELD_APL = PresetUtils.makePresetAPLRotation('Monocat, Shadowmeld Ravage (Night Elf)', ShadowmeldApl);
// Monocat with Read Ley Line, for a High Order Skyborne: cast in an idle GCD under 1400 mana. 0 to 4 DPS
// from 60 to 300 seconds.
export const ENERGIZED_APL = PresetUtils.makePresetAPLRotation('Monocat, Energized (High Order Skyborne)', EnergizedApl);
// Uses Thistle Tea and optional Sapper, with racial actions gated by availability.
// The Prowl opener requires prepull Clearcasting; Shadowmeld enables Ravage during combat.
export const JVMES_APL = {
	...PresetUtils.makePresetAPLRotation('Monocat, adaptive racials and Sapper', JvmesApl),
	tooltip:
		'Maintains Rip and Rake, saves Bite for fight end, and uses available racial abilities and optional Sapper. Opens with Ravage when prepull Clearcasting is enabled.',
};
export const RAVAGE_APL = PresetUtils.makePresetAPLRotation('Ravage opener', RavageApl);
export const POUNCE_APL = PresetUtils.makePresetAPLRotation('Pounce opener', PounceApl);
// Monocat with Dire Bear trips for Primal Bite and Maul, only on mana the Shifting Powers to the end
// of the fight do not need. Needs the Primal Bite talent.
export const BEARWEAVE_APL = PresetUtils.makePresetAPLRotation('Bearweave', BearweaveApl);

export const CatTalents = PresetUtils.makePresetTalents('Cat', SavedTalents.create({ talentsString: '050022-35200032121131212051-052' }));
// Primal Bite takes one additional point from Naturalist.
export const BearweaveTalents = PresetUtils.makePresetTalents('Bearweave', SavedTalents.create({ talentsString: '050022-55000032021131212051-053' }));
// Natural Shapeshifter and Reflection in place of the Balance points' melee hit and crit. About 47 DPS
// behind Monocat at 180 seconds.
export const ManaEfficientTalents = PresetUtils.makePresetTalents(
	'Mana Efficient 0/35/16',
	SavedTalents.create({ talentsString: '0000000000000000-35200032121032212051-0550330000000000' }),
);
// Tanks and plays cat on one set of talents: Thick Hide and Primal Bite, no Furor or Naturalist. About 34 DPS
// behind Monocat at 180 seconds.
export const HybridTalents = PresetUtils.makePresetTalents(
	'Hybrid 9/42/0',
	SavedTalents.create({ talentsString: '0500220000000000-35230332121132212051-0000000000000000' }),
);

// A bear form forfeits Furor's energy, so a trip costs the energy the cat went in with. Ahead of
// Monocat only on a fight of 45 seconds exactly (+3.3); with the length varying by 15% it is 2.1
// behind at 45 seconds and 6 to 7 behind from 60 seconds on.
export const BearweaveBuild = PresetUtils.makePresetBuild('Bearweave (45s fights)', {
	talents: BearweaveTalents,
	rotation: BEARWEAVE_APL,
});
