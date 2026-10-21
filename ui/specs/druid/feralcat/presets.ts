import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, PseudoStat, Race, Stat } from '@generated/proto/common';
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
	devOnly: true,
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

// Pre-raid leather on a Night Elf with The Unstoppable Force and Wolfshead Helm. Shifting Cloak, Dunestalker's
// Boots, Ferocious Watcher's Signet and the two hit and attack power trinkets stand in for Cape of the Black
// Baron, Swiftwalker Boots, Tarnished Elven Ring, Blackhand's Breadth and Hand of Justice, which the item
// database does not carry.
export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	devOnly: true,
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

// DPS weights for Dream, Cat 9/35/7 and 180s ±120s, averaged across the four races.
// Rating weights measure hit at 7% and expertise at 5.5%.
export const DefaultEp = Stats.fromMap(
	{
		[Stat.StatStrength]: 0.5699,
		[Stat.StatAgility]: 0.5815,
		[Stat.StatAttackPower]: 0.2351,
		[Stat.StatFeralAttackPower]: 0.2351,
		[Stat.StatPhysicalDamage]: 2.1248,
		[Stat.StatHitRating]: 4.8332 / 10,
		[Stat.StatExpertiseRating]: 4.57 / 10,
		[Stat.StatCritRating]: 5.9088 / 14,
		[Stat.StatHasteRating]: 4.1595 / 10,
	},
	{
		[PseudoStat.PseudoStatMainHandDps]: 2.2704,
	},
);

export const DefaultEpWeights = PresetUtils.makePresetEpWeights('Feral Cat', DefaultEp);

// The cat stays in Cat Form: Shifting Power, Rip, Rake and Shred, Faerie Fire for Omen of Clarity,
// and a self-Innervate when mana would run short.
export const MONOCAT_APL = PresetUtils.makePresetAPLRotation('Monocat', DefaultApl);
// Monocat with Shadowmeld into Ravage on cooldown, for a Night Elf.
export const SHADOWMELD_APL = PresetUtils.makePresetAPLRotation('Monocat, Shadowmeld Ravage (Night Elf)', ShadowmeldApl);
// Monocat with Read Ley Line, for a High Order Skyborne: cast in an idle GCD under 1400 mana. 0 to 4 DPS
// from 60 to 300 seconds.
export const ENERGIZED_APL = PresetUtils.makePresetAPLRotation('Monocat, Energized (High Order Skyborne)', EnergizedApl);
// Uses Thistle Tea and optional Sapper, with racial actions gated by availability.
// Opens with Berserk and Ravage from Prowl; during combat, Shadowmeld enables Ravage on Clearcasting.
export const JVMES_APL = {
	...PresetUtils.makePresetAPLRotation('Monocat, adaptive racials and Sapper', JvmesApl),
	tooltip:
		'Opens with Berserk and Ravage, maintains bleeds, and saves Bite for fight end. Later Berserks pool to 60 energy or Clearcasting when time permits. Named conditions explain racial abilities, optional Sapper and resource use.',
};
export const RAVAGE_APL = PresetUtils.makePresetAPLRotation('Ravage opener', RavageApl);
export const POUNCE_APL = PresetUtils.makePresetAPLRotation('Pounce opener', PounceApl);
// Monocat with Dire Bear trips for Primal Bite and Maul, only on mana the Shifting Powers to the end
// of the fight do not need. Needs the Primal Bite talent.
export const BEARWEAVE_APL = PresetUtils.makePresetAPLRotation('Bearweave', BearweaveApl);

// Naturalist in place of Natural Shapeshifter.
export const CatTalents = PresetUtils.makePresetTalents('Cat', SavedTalents.create({ talentsString: '050022-35200032121131212051-052' }));
export const HybridTalents = PresetUtils.makePresetTalents('Hybrid', SavedTalents.create({ talentsString: '050022-05230332121131212051-04' }));

export const BearweaveBuild = PresetUtils.makePresetBuild('Bearweave (45s fights)', {
	talents: CatTalents,
	rotation: BEARWEAVE_APL,
});
