import { StatCapType } from '@generated/proto/api';
import { ArmorType, Class, Profession, Race, Stat } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import { DungeonDifficulty, RepLevel } from '@generated/proto/ui';
import i18n from '@i18n/config';

import { kebabCase } from '../utils/format';

// A name table whose names come from the locale at the time they are read.
export type NameTable<K> = Pick<ReadonlyMap<K, string>, 'get'>;
const translatedNames = <K>(entries: ReadonlyArray<readonly [K, () => string]>): NameTable<K> => {
	const names = new Map(entries);
	return { get: key => names.get(key)?.() };
};

export const armorTypeNames = translatedNames<ArmorType>([
	[ArmorType.ArmorTypeUnknown, () => i18n.t('common.armor_types.unknown')],
	[ArmorType.ArmorTypeCloth, () => i18n.t('common.armor_types.cloth')],
	[ArmorType.ArmorTypeLeather, () => i18n.t('common.armor_types.leather')],
	[ArmorType.ArmorTypeMail, () => i18n.t('common.armor_types.mail')],
	[ArmorType.ArmorTypePlate, () => i18n.t('common.armor_types.plate')],
]);

// English names, as import formats and the Wowhead gear planner spell them.
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

// English names, as import formats spell them.
const englishClassNames: Map<Class, string> = new Map([
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
	for (const [key, value] of englishClassNames) {
		if (value.toLowerCase().replace(/\s+/g, '') == lower) {
			return key;
		}
	}
	return Class.ClassUnknown;
}

export const classNames = translatedNames<Class>([
	[Class.ClassUnknown, () => i18n.t('common.none')],
	[Class.ClassDruid, () => i18n.t('classes.druid', { ns: 'character' })],
	[Class.ClassHunter, () => i18n.t('classes.hunter', { ns: 'character' })],
	[Class.ClassMage, () => i18n.t('classes.mage', { ns: 'character' })],
	[Class.ClassPaladin, () => i18n.t('classes.paladin', { ns: 'character' })],
	[Class.ClassPriest, () => i18n.t('classes.priest', { ns: 'character' })],
	[Class.ClassRogue, () => i18n.t('classes.rogue', { ns: 'character' })],
	[Class.ClassShaman, () => i18n.t('classes.shaman', { ns: 'character' })],
	[Class.ClassWarlock, () => i18n.t('classes.warlock', { ns: 'character' })],
	[Class.ClassWarrior, () => i18n.t('classes.warrior', { ns: 'character' })],
]);

// English names, as import formats spell them.
const englishProfessionNames: Map<Profession, string> = new Map([
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
	for (const [key, value] of englishProfessionNames) {
		if (value.toLowerCase() == lower) {
			return key;
		}
	}
	return Profession.ProfessionUnknown;
}

export const professionNames = translatedNames<Profession>([
	[Profession.ProfessionUnknown, () => i18n.t('professions.unknown', { ns: 'character' })],
	[Profession.Alchemy, () => i18n.t('professions.alchemy', { ns: 'character' })],
	[Profession.Blacksmithing, () => i18n.t('professions.blacksmithing', { ns: 'character' })],
	[Profession.Enchanting, () => i18n.t('professions.enchanting', { ns: 'character' })],
	[Profession.Engineering, () => i18n.t('professions.engineering', { ns: 'character' })],
	[Profession.Herbalism, () => i18n.t('professions.herbalism', { ns: 'character' })],
	[Profession.Inscription, () => i18n.t('professions.inscription', { ns: 'character' })],
	[Profession.Leatherworking, () => i18n.t('professions.leatherworking', { ns: 'character' })],
	[Profession.Mining, () => i18n.t('professions.mining', { ns: 'character' })],
	[Profession.Skinning, () => i18n.t('professions.skinning', { ns: 'character' })],
	[Profession.Tailoring, () => i18n.t('professions.tailoring', { ns: 'character' })],
]);

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

export const spellSchoolNames = translatedNames<number>([
	[SpellSchool.Physical, () => i18n.t('common.spell_schools.physical')],
	[SpellSchool.Arcane, () => i18n.t('common.spell_schools.arcane')],
	[SpellSchool.Fire, () => i18n.t('common.spell_schools.fire')],
	[SpellSchool.Frost, () => i18n.t('common.spell_schools.frost')],
	[SpellSchool.Holy, () => i18n.t('common.spell_schools.holy')],
	[SpellSchool.Nature, () => i18n.t('common.spell_schools.nature')],
	[SpellSchool.Shadow, () => i18n.t('common.spell_schools.shadow')],
	[SpellSchool.Nature + SpellSchool.Arcane, () => i18n.t('common.spell_schools.astral')],
	[SpellSchool.Shadow + SpellSchool.Fire, () => i18n.t('common.spell_schools.shadowflame')],
	[SpellSchool.Fire + SpellSchool.Arcane, () => i18n.t('common.spell_schools.spellfire')],
	[SpellSchool.Arcane + SpellSchool.Frost, () => i18n.t('common.spell_schools.spellfrost')],
	[SpellSchool.Frost + SpellSchool.Fire, () => i18n.t('common.spell_schools.frostfire')],
	[SpellSchool.Shadow + SpellSchool.Frost, () => i18n.t('common.spell_schools.shadowfrost')],
	[SpellSchool.Nature + SpellSchool.Shadow, () => i18n.t('common.spell_schools.plague')],
	[SpellSchool.Fire + SpellSchool.Nature, () => i18n.t('common.spell_schools.firestorm')],
	[SpellSchool.Fire + SpellSchool.Frost + SpellSchool.Nature, () => i18n.t('common.spell_schools.elemental')],
]);

// English names, as the combat log spells them; their kebab-case forms are the resource colour classes.
const englishResourceNames: Map<ResourceType, string> = new Map([
	[ResourceType.ResourceTypeNone, 'None'],
	[ResourceType.ResourceTypeHealth, 'Health'],
	[ResourceType.ResourceTypeMana, 'Mana'],
	[ResourceType.ResourceTypeEnergy, 'Energy'],
	[ResourceType.ResourceTypeRage, 'Rage'],
	[ResourceType.ResourceTypeComboPoints, 'Combo Points'],
	[ResourceType.ResourceTypeFocus, 'Focus'],
	[ResourceType.ResourceTypeGenericResource, 'Generic Resource'],
]);

export const resourceNames = translatedNames<ResourceType>([
	[ResourceType.ResourceTypeNone, () => i18n.t('common.none')],
	[ResourceType.ResourceTypeHealth, () => i18n.t('common.resource_types.health')],
	[ResourceType.ResourceTypeMana, () => i18n.t('common.resource_types.mana')],
	[ResourceType.ResourceTypeEnergy, () => i18n.t('common.resource_types.energy')],
	[ResourceType.ResourceTypeRage, () => i18n.t('common.resource_types.rage')],
	[ResourceType.ResourceTypeComboPoints, () => i18n.t('common.resource_types.combo_points')],
	[ResourceType.ResourceTypeFocus, () => i18n.t('common.resource_types.focus')],
	[ResourceType.ResourceTypeGenericResource, () => i18n.t('common.resource_types.generic_resource_name')],
]);

export const resourceCssName = (resourceType: ResourceType): string => kebabCase(englishResourceNames.get(resourceType) ?? '');

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
	for (const [key, val] of englishResourceNames) {
		byName.set(val.toLowerCase(), key);
	}
	return byName;
})();

export function stringToResourceType(str: string): ResourceType {
	return resourceTypesByLowerName.get(str.toLowerCase()) ?? ResourceType.ResourceTypeNone;
}

export const difficultyNames = translatedNames<DungeonDifficulty>([
	[DungeonDifficulty.DifficultyUnknown, () => i18n.t('common.dungeon_difficulties.unknown')],
	[DungeonDifficulty.DifficultyNormal, () => i18n.t('common.dungeon_difficulties.normal')],
	[DungeonDifficulty.DifficultyHeroic, () => i18n.t('common.dungeon_difficulties.heroic')],
	[DungeonDifficulty.DifficultyCelestial, () => i18n.t('common.dungeon_difficulties.celestial')],
	[DungeonDifficulty.DifficultyRaid10, () => i18n.t('common.dungeon_difficulties.raid_10')],
	[DungeonDifficulty.DifficultyRaid10H, () => i18n.t('common.dungeon_difficulties.raid_10_heroic')],
	[DungeonDifficulty.DifficultyRaid25RF, () => i18n.t('common.dungeon_difficulties.raid_finder')],
	[DungeonDifficulty.DifficultyRaid25, () => i18n.t('common.dungeon_difficulties.raid_25')],
	[DungeonDifficulty.DifficultyRaid25H, () => i18n.t('common.dungeon_difficulties.raid_25_heroic')],
	[DungeonDifficulty.DifficultyRaidFlex, () => i18n.t('common.dungeon_difficulties.raid_flex')],
]);

export const REP_LEVEL_NAMES: Readonly<Record<RepLevel, string>> = {
	get [RepLevel.RepLevelUnknown]() {
		return i18n.t('common.rep_levels.unknown');
	},
	get [RepLevel.RepLevelHated]() {
		return i18n.t('common.rep_levels.hated');
	},
	get [RepLevel.RepLevelHostile]() {
		return i18n.t('common.rep_levels.hostile');
	},
	get [RepLevel.RepLevelUnfriendly]() {
		return i18n.t('common.rep_levels.unfriendly');
	},
	get [RepLevel.RepLevelNeutral]() {
		return i18n.t('common.rep_levels.neutral');
	},
	get [RepLevel.RepLevelFriendly]() {
		return i18n.t('common.rep_levels.friendly');
	},
	get [RepLevel.RepLevelHonored]() {
		return i18n.t('common.rep_levels.honored');
	},
	get [RepLevel.RepLevelRevered]() {
		return i18n.t('common.rep_levels.revered');
	},
	get [RepLevel.RepLevelExalted]() {
		return i18n.t('common.rep_levels.exalted');
	},
};

export const statCapTypeNames = translatedNames<StatCapType>([
	[StatCapType.TypeHardCap, () => i18n.t('common.stat_cap_types.hard_cap')],
	[StatCapType.TypeSoftCap, () => i18n.t('common.stat_cap_types.soft_cap')],
	[StatCapType.TypeThreshold, () => i18n.t('common.stat_cap_types.threshold')],
]);
