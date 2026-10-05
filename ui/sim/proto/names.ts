import { StatCapType } from '@generated/proto/api';
import { ArmorType, Class, Profession, Race, Stat } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { DungeonDifficulty, RepLevel } from '@generated/proto/ui';

export const armorTypeNames: Map<ArmorType, string> = new Map([
	[ArmorType.ArmorTypeUnknown, 'Unknown'],
	[ArmorType.ArmorTypeCloth, 'Cloth'],
	[ArmorType.ArmorTypeLeather, 'Leather'],
	[ArmorType.ArmorTypeMail, 'Mail'],
	[ArmorType.ArmorTypePlate, 'Plate'],
]);

export const raceNames: Map<Race, string> = new Map([
	[Race.RaceUnknown, 'None'],
	[Race.RaceBloodElf, 'Blood Elf'],
	[Race.RaceDwarf, 'Dwarf'],
	[Race.RaceGnome, 'Gnome'],
	[Race.RaceHuman, 'Human'],
	[Race.RaceNightElf, 'Night Elf'],
	[Race.RaceOrc, 'Orc'],
	[Race.RaceTauren, 'Tauren'],
	[Race.RaceTroll, 'Troll'],
	[Race.RaceUndead, 'Undead'],
	[Race.RaceHighOrderSkyborne, 'High Order Skyborne'],
	[Race.RaceWindshaperSkyborne, 'Windshaper Skyborne'],
]);

export function nameToRace(name: string): Race {
	const normalized = name.toLowerCase().replaceAll(' ', '');
	for (const [key, value] of raceNames) {
		if (value.toLowerCase().replaceAll(' ', '') == normalized) {
			return key;
		}
	}
	return Race.RaceUnknown;
}

export const classNames: Map<Class, string> = new Map([
	[Class.ClassUnknown, 'None'],
	[Class.ClassDruid, 'Druid'],
	[Class.ClassHunter, 'Hunter'],
	[Class.ClassMage, 'Mage'],
	[Class.ClassPaladin, 'Paladin'],
	[Class.ClassPriest, 'Priest'],
	[Class.ClassRogue, 'Rogue'],
	[Class.ClassShaman, 'Shaman'],
	[Class.ClassWarlock, 'Warlock'],
	[Class.ClassWarrior, 'Warrior'],
]);

export function nameToClass(name: string): Class {
	const lower = name.toLowerCase();
	for (const [key, value] of classNames) {
		if (value.toLowerCase().replace(/\s+/g, '') == lower) {
			return key;
		}
	}
	return Class.ClassUnknown;
}

export const professionNames: Map<Profession, string> = new Map([
	[Profession.ProfessionUnknown, 'None'],
	[Profession.Alchemy, 'Alchemy'],
	[Profession.Blacksmithing, 'Blacksmithing'],
	[Profession.Enchanting, 'Enchanting'],
	[Profession.Engineering, 'Engineering'],
	[Profession.Herbalism, 'Herbalism'],
	[Profession.Inscription, 'Inscription'],
	[Profession.Leatherworking, 'Leatherworking'],
	[Profession.Mining, 'Mining'],
	[Profession.Skinning, 'Skinning'],
	[Profession.Tailoring, 'Tailoring'],
]);

export function nameToProfession(name: string): Profession {
	const lower = name.toLowerCase();
	for (const [key, value] of professionNames) {
		if (value.toLowerCase() == lower) {
			return key;
		}
	}
	return Profession.ProfessionUnknown;
}

export function getStatName(stat: Stat): string {
	if (stat == Stat.StatRangedAttackPower) {
		return 'Ranged AP';
	} else {
		return Stat[stat]
			.split(/(?<![A-Z])(?=[A-Z])/)
			.slice(1)
			.join(' ');
	}
}

enum SpellSchool {
	None = 0,
	Physical = 1,
	Holy = 2,
	Fire = 4,
	Nature = 8,
	Frost = 16,
	Shadow = 32,
	Arcane = 64,
}

export const spellSchoolNames: Map<number, string> = new Map([
	[SpellSchool.Physical, 'Physical'],
	[SpellSchool.Arcane, 'Arcane'],
	[SpellSchool.Fire, 'Fire'],
	[SpellSchool.Frost, 'Frost'],
	[SpellSchool.Holy, 'Holy'],
	[SpellSchool.Nature, 'Nature'],
	[SpellSchool.Shadow, 'Shadow'],
	[SpellSchool.Nature + SpellSchool.Arcane, 'Astral'],
	[SpellSchool.Shadow + SpellSchool.Fire, 'Shadowflame'],
	[SpellSchool.Fire + SpellSchool.Arcane, 'Spellfire'],
	[SpellSchool.Arcane + SpellSchool.Frost, 'Spellfrost'],
	[SpellSchool.Frost + SpellSchool.Fire, 'Frostfire'],
	[SpellSchool.Shadow + SpellSchool.Frost, 'Shadowfrost'],
	[SpellSchool.Nature + SpellSchool.Shadow, 'Plague'],
	[SpellSchool.Fire + SpellSchool.Nature, 'Firestorm'],
	[SpellSchool.Fire + SpellSchool.Frost + SpellSchool.Nature, 'Elemental'],
]);

export const resourceNames: Map<ResourceType, string> = new Map([
	[ResourceType.ResourceTypeNone, 'None'],
	[ResourceType.ResourceTypeHealth, 'Health'],
	[ResourceType.ResourceTypeMana, 'Mana'],
	[ResourceType.ResourceTypeEnergy, 'Energy'],
	[ResourceType.ResourceTypeRage, 'Rage'],
	[ResourceType.ResourceTypeComboPoints, 'Combo Points'],
	[ResourceType.ResourceTypeFocus, 'Focus'],
	[ResourceType.ResourceTypeGenericResource, 'Generic Resource'],
]);

export const resourceColors: Map<ResourceType, string> = new Map([
	[ResourceType.ResourceTypeNone, '#ffffff'],
	[ResourceType.ResourceTypeHealth, '#22ba00'],
	[ResourceType.ResourceTypeMana, '#2e93fa'],
	[ResourceType.ResourceTypeEnergy, '#ffd700'],
	[ResourceType.ResourceTypeRage, '#ff0000'],
	[ResourceType.ResourceTypeComboPoints, '#ffa07a'],
	[ResourceType.ResourceTypeFocus, '#cd853f'],
	[ResourceType.ResourceTypeGenericResource, '#ffffff'],
]);

// Built once. This is called for every resource line the combat log parser sees.
const resourceTypesByLowerName = (() => {
	const byName = new Map<string, ResourceType>();
	for (const [key, val] of resourceNames) {
		byName.set(val.toLowerCase(), key);
	}
	return byName;
})();

export function stringToResourceType(str: string): ResourceType {
	return resourceTypesByLowerName.get(str.toLowerCase()) ?? ResourceType.ResourceTypeNone;
}

export const difficultyNames: Map<DungeonDifficulty, string> = new Map([
	[DungeonDifficulty.DifficultyUnknown, 'Unknown'],
	[DungeonDifficulty.DifficultyNormal, 'N'],
	[DungeonDifficulty.DifficultyHeroic, 'H'],
	[DungeonDifficulty.DifficultyCelestial, 'CEL'],
	[DungeonDifficulty.DifficultyRaid10, '10N'],
	[DungeonDifficulty.DifficultyRaid10H, '10H'],
	[DungeonDifficulty.DifficultyRaid25RF, 'RF'],
	[DungeonDifficulty.DifficultyRaid25, 'RN'],
	[DungeonDifficulty.DifficultyRaid25H, 'RH'],
	[DungeonDifficulty.DifficultyRaidFlex, 'RFL'],
]);

export const REP_LEVEL_NAMES: Record<RepLevel, string> = {
	[RepLevel.RepLevelUnknown]: 'Unknown',
	[RepLevel.RepLevelHated]: 'Hated',
	[RepLevel.RepLevelHostile]: 'Hostile',
	[RepLevel.RepLevelUnfriendly]: 'Unfriendly',
	[RepLevel.RepLevelNeutral]: 'Neutral',
	[RepLevel.RepLevelFriendly]: 'Friendly',
	[RepLevel.RepLevelHonored]: 'Honored',
	[RepLevel.RepLevelRevered]: 'Revered',
	[RepLevel.RepLevelExalted]: 'Exalted',
};

export const statCapTypeNames = new Map<StatCapType, string>([
	[StatCapType.TypeHardCap, 'Hard cap'],
	[StatCapType.TypeSoftCap, 'Soft cap'],
	[StatCapType.TypeThreshold, 'Threshold'],
]);
