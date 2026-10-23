package dbc

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func MapResistanceToStat(index int) (proto.Stat, bool) {
	switch index {
	case 0:
		return proto.Stat_StatBonusArmor, true
	// 1 = Holy (no proto stat)
	case 2:
		return proto.Stat_StatFireResistance, true
	case 3:
		return proto.Stat_StatNatureResistance, true
	case 4:
		return proto.Stat_StatFrostResistance, true
	case 5:
		return proto.Stat_StatShadowResistance, true
	case 6:
		return proto.Stat_StatArcaneResistance, true
	}
	return 0, false
}

var MapArmorSubclassToArmorType = map[int]proto.ArmorType{
	ITEM_SUBCLASS_ARMOR_CLOTH:   proto.ArmorType_ArmorTypeCloth,
	ITEM_SUBCLASS_ARMOR_LEATHER: proto.ArmorType_ArmorTypeLeather,
	ITEM_SUBCLASS_ARMOR_MAIL:    proto.ArmorType_ArmorTypeMail,
	ITEM_SUBCLASS_ARMOR_PLATE:   proto.ArmorType_ArmorTypePlate,
	0:                           proto.ArmorType_ArmorTypeUnknown,
}

// Ref: https://wowdev.wiki/Stat_Types
func MapMainStatToStat(index int) (proto.Stat, bool) {
	switch index {
	case 0:
		return proto.Stat_StatStrength, true
	case 1:
		return proto.Stat_StatAgility, true
	case 2:
		return proto.Stat_StatStamina, true
	case 3:
		return proto.Stat_StatIntellect, true
	case 4:
		return proto.Stat_StatSpirit, true
	}
	return 0, false
}
func MapBonusStatIndexToStat(index int) (proto.Stat, bool) {
	switch index {
	case ITEM_MOD_MANA: // Mana
		return proto.Stat_StatMana, true
	case ITEM_MOD_HEALTH: // Health
		return proto.Stat_StatHealth, true
	case ITEM_MOD_STAMINA: // Stamina
		return proto.Stat_StatStamina, true
	case ITEM_MOD_AGILITY: // Agility
		return proto.Stat_StatAgility, true
	case ITEM_MOD_STRENGTH: // Strength
		return proto.Stat_StatStrength, true
	case ITEM_MOD_INTELLECT: // Intellect
		return proto.Stat_StatIntellect, true
	case ITEM_MOD_SPIRIT: // Spirit
		return proto.Stat_StatSpirit, true

	case ITEM_MOD_DEFENSE_SKILL_RATING:
		return proto.Stat_StatDefenseRating, true
	case ITEM_MOD_DODGE_RATING:
		return proto.Stat_StatDodgeRating, true
	case ITEM_MOD_PARRY_RATING:
		return proto.Stat_StatParryRating, true
	case ITEM_MOD_BLOCK_RATING:
		return proto.Stat_StatBlockRating, true

	// Secondary ratings
	case ITEM_MOD_HIT_MELEE_RATING, ITEM_MOD_HIT_RANGED_RATING, ITEM_MOD_HIT_SPELL_RATING, ITEM_MOD_HIT_RATING:
		return proto.Stat_StatHitRating, true
	case ITEM_MOD_CRIT_MELEE_RATING, ITEM_MOD_CRIT_RANGED_RATING, ITEM_MOD_CRIT_SPELL_RATING, ITEM_MOD_CRIT_RATING:
		return proto.Stat_StatCritRating, true
	case ITEM_MOD_HASTE_MELEE_RATING, ITEM_MOD_HASTE_RANGED_RATING, ITEM_MOD_HASTE_SPELL_RATING, ITEM_MOD_HASTE_RATING:
		return proto.Stat_StatHasteRating, true
	case ITEM_MOD_EXPERTISE_RATING:
		return proto.Stat_StatExpertiseRating, true
	case ITEM_MOD_ARMOR_PENETRATION_RATING:
		return proto.Stat_StatArmorPenetration, true

	case ITEM_MOD_ATTACK_POWER: // AttackPower
		return proto.Stat_StatAttackPower, true
	case ITEM_MOD_RANGED_ATTACK_POWER: // RangedAttackPower
		return proto.Stat_StatRangedAttackPower, true
	case ITEM_MOD_FERAL_ATTACK_POWER:
		return proto.Stat_StatFeralAttackPower, true
	case ITEM_MOD_SPELL_HEALING_DONE:
		return proto.Stat_StatHealingPower, true
	case ITEM_MOD_SPELL_DAMAGE_DONE:
		return proto.Stat_StatSpellDamage, true
	case ITEM_MOD_EXTRA_ARMOR: // ExtraArmor maps to BonusArmor (green armor)
		return proto.Stat_StatBonusArmor, true

	// Forever moved the five resistances out of Item.Resistances_* and onto the item
	// stat array, which is why the schema no longer carries those columns. Index 53
	// (Holy Resistance) has no proto.Stat counterpart and stays unmapped.
	case ITEM_MOD_FIRE_RESISTANCE:
		return proto.Stat_StatFireResistance, true
	case ITEM_MOD_FROST_RESISTANCE:
		return proto.Stat_StatFrostResistance, true
	case ITEM_MOD_SHADOW_RESISTANCE:
		return proto.Stat_StatShadowResistance, true
	case ITEM_MOD_NATURE_RESISTANCE:
		return proto.Stat_StatNatureResistance, true
	case ITEM_MOD_ARCANE_RESISTANCE:
		return proto.Stat_StatArcaneResistance, true

	// The school-damage block. 83 is vanilla's "+N Weapon Damage"; StatPhysicalDamage
	// is slightly broader (sim/core/spell_result.go adds it to every physical spell,
	// not swings only), but the magnitudes involved are single digits.
	case ITEM_MOD_WEAPON_DAMAGE:
		return proto.Stat_StatPhysicalDamage, true
	case ITEM_MOD_HOLY_DAMAGE:
		return proto.Stat_StatHolyDamage, true
	case ITEM_MOD_FIRE_DAMAGE:
		return proto.Stat_StatFireDamage, true
	case ITEM_MOD_NATURE_DAMAGE:
		return proto.Stat_StatNatureDamage, true
	case ITEM_MOD_FROST_DAMAGE:
		return proto.Stat_StatFrostDamage, true
	case ITEM_MOD_SHADOW_DAMAGE:
		return proto.Stat_StatShadowDamage, true
	case ITEM_MOD_ARCANE_DAMAGE:
		return proto.Stat_StatArcaneDamage, true
	case ITEM_MOD_MANA_REGENERATION: // ManaRegeneration
		return proto.Stat_StatMP5, true
	case ITEM_MOD_HEALTH_REGEN:
		return proto.Stat_StatHP5, true
	case ITEM_MOD_SPELL_PENETRATION:
		return proto.Stat_StatSpellPiercing, true
	case ITEM_MOD_BLOCK_VALUE:
		return proto.Stat_StatBlockValue, true
	default:
		return 0, false
	}
}

// allResistanceStats is what ItemModType 124 ("All Resistances") expands to. Confirmed
// from three independent tooltips: 22197 Heavy Obsidian Belt and 22198 Jagged Obsidian
// Shield both read "+5 All Resistances", and 12252 Staff of Protection spells out
// "+6 Arcane/Fire/Nature/Frost/Shadow".
var allResistanceStats = []proto.Stat{
	proto.Stat_StatArcaneResistance,
	proto.Stat_StatFireResistance,
	proto.Stat_StatFrostResistance,
	proto.Stat_StatNatureResistance,
	proto.Stat_StatShadowResistance,
}

// spellPowerStats is what ItemModType 45 expands to: 7655 Enchant Bracer - Spell Power reads
// "+12 Spell Power" for 45 = 12, damage and healing alike. Healing items state their extra
// healing as a separate 41 on top of it (Atiesh 22631: 45 and 41).
var spellPowerStats = []proto.Stat{
	proto.Stat_StatSpellDamage,
	proto.Stat_StatHealingPower,
}

// MapBonusStatIndexToStats is MapBonusStatIndexToStat for callers that have to cope with
// one index granting several stats. Every index maps to exactly one stat except 45 and 124.
// An enchant row whose argument is wrong is corrected before it gets here (enchantEffectArgFixes).
func MapBonusStatIndexToStats(index int) ([]proto.Stat, bool) {
	switch index {
	case ITEM_MOD_ALL_RESISTANCES:
		return allResistanceStats, true
	case ITEM_MOD_SPELL_POWER:
		return spellPowerStats, true
	}
	if stat, ok := MapBonusStatIndexToStat(index); ok {
		return []proto.Stat{stat}, true
	}
	return nil, false
}

// The school a spell penetration stat pierces, which the sim keeps outside the stat array.
func MapBonusStatIndexToPiercingSchool(index int) (stats.SchoolIndex, bool) {
	switch index {
	case ITEM_MOD_FIRE_SPELL_PENETRATION:
		return stats.SchoolIndexFire, true
	case ITEM_MOD_FROST_SPELL_PENETRATION:
		return stats.SchoolIndexFrost, true
	}
	return stats.SchoolIndexNone, false
}

var MapProfessionIdToProfession = map[int]proto.Profession{
	0:   proto.Profession_ProfessionUnknown,
	164: proto.Profession_Blacksmithing,
	165: proto.Profession_Leatherworking,
	171: proto.Profession_Alchemy,
	182: proto.Profession_Herbalism,
	186: proto.Profession_Mining,
	197: proto.Profession_Tailoring,
	202: proto.Profession_Engineering,
	333: proto.Profession_Enchanting,
	393: proto.Profession_Skinning,
	773: proto.Profession_Inscription,
}

var MapInventoryTypeToItemType = map[int]proto.ItemType{
	0:                      proto.ItemType_ItemTypeUnknown,
	INVTYPE_HEAD:           proto.ItemType_ItemTypeHead,
	INVTYPE_NECK:           proto.ItemType_ItemTypeNeck,
	INVTYPE_SHOULDERS:      proto.ItemType_ItemTypeShoulder,
	INVTYPE_CHEST:          proto.ItemType_ItemTypeChest,
	INVTYPE_WAIST:          proto.ItemType_ItemTypeWaist,
	INVTYPE_LEGS:           proto.ItemType_ItemTypeLegs,
	INVTYPE_FEET:           proto.ItemType_ItemTypeFeet,
	INVTYPE_WRISTS:         proto.ItemType_ItemTypeWrist,
	INVTYPE_HANDS:          proto.ItemType_ItemTypeHands,
	INVTYPE_FINGER:         proto.ItemType_ItemTypeFinger,
	INVTYPE_TRINKET:        proto.ItemType_ItemTypeTrinket,
	INVTYPE_WEAPON:         proto.ItemType_ItemTypeWeapon,
	INVTYPE_SHIELD:         proto.ItemType_ItemTypeWeapon,
	INVTYPE_RANGED:         proto.ItemType_ItemTypeRanged,
	INVTYPE_CLOAK:          proto.ItemType_ItemTypeBack,
	INVTYPE_2HWEAPON:       proto.ItemType_ItemTypeWeapon,
	INVTYPE_BAG:            proto.ItemType_ItemTypeUnknown,
	INVTYPE_TABARD:         proto.ItemType_ItemTypeUnknown,
	INVTYPE_ROBE:           proto.ItemType_ItemTypeChest,
	INVTYPE_WEAPONMAINHAND: proto.ItemType_ItemTypeWeapon,
	INVTYPE_WEAPONOFFHAND:  proto.ItemType_ItemTypeWeapon,
	INVTYPE_HOLDABLE:       proto.ItemType_ItemTypeWeapon,
	INVTYPE_AMMO:           proto.ItemType_ItemTypeUnknown,
	INVTYPE_THROWN:         proto.ItemType_ItemTypeRanged,
	INVTYPE_RANGEDRIGHT:    proto.ItemType_ItemTypeRanged,
	INVTYPE_QUIVER:         proto.ItemType_ItemTypeRanged,
	INVTYPE_RELIC:          proto.ItemType_ItemTypeRanged,
}
var MapInventoryTypeFlagToItemType = map[InventoryTypeFlag]proto.ItemType{
	0:                proto.ItemType_ItemTypeUnknown,
	HEAD:             proto.ItemType_ItemTypeHead,
	NECK:             proto.ItemType_ItemTypeNeck,
	SHOULDER:         proto.ItemType_ItemTypeShoulder,
	CHEST:            proto.ItemType_ItemTypeChest,
	WAIST:            proto.ItemType_ItemTypeWaist,
	LEGS:             proto.ItemType_ItemTypeLegs,
	FEET:             proto.ItemType_ItemTypeFeet,
	WRIST:            proto.ItemType_ItemTypeWrist,
	HAND:             proto.ItemType_ItemTypeHands,
	FINGER:           proto.ItemType_ItemTypeFinger,
	TRINKET:          proto.ItemType_ItemTypeTrinket,
	MAIN_HAND:        proto.ItemType_ItemTypeWeapon,
	OFF_HAND:         proto.ItemType_ItemTypeWeapon,
	RANGED:           proto.ItemType_ItemTypeRanged,
	CLOAK:            proto.ItemType_ItemTypeBack,
	TWO_H_WEAPON:     proto.ItemType_ItemTypeWeapon,
	BAG:              proto.ItemType_ItemTypeUnknown,
	TABARD:           proto.ItemType_ItemTypeUnknown,
	ROBE:             proto.ItemType_ItemTypeChest,
	WEAPON_MAIN_HAND: proto.ItemType_ItemTypeWeapon,
	WEAPON_OFF_HAND:  proto.ItemType_ItemTypeWeapon,
	HOLDABLE:         proto.ItemType_ItemTypeWeapon,
	AMMO:             proto.ItemType_ItemTypeUnknown,
	THROWN:           proto.ItemType_ItemTypeRanged,
	RANGED_RIGHT:     proto.ItemType_ItemTypeRanged,
	QUIVER:           proto.ItemType_ItemTypeRanged,
	RELIC:            proto.ItemType_ItemTypeRanged,
}

var MapWeaponSubClassToWeaponType = map[int]proto.WeaponType{
	ITEM_SUBCLASS_WEAPON_AXE:          proto.WeaponType_WeaponTypeAxe,
	ITEM_SUBCLASS_WEAPON_AXE2:         proto.WeaponType_WeaponTypeAxe,
	ITEM_SUBCLASS_WEAPON_BOW:          proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_GUN:          proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_MACE:         proto.WeaponType_WeaponTypeMace,
	ITEM_SUBCLASS_WEAPON_MACE2:        proto.WeaponType_WeaponTypeMace,
	ITEM_SUBCLASS_WEAPON_POLEARM:      proto.WeaponType_WeaponTypePolearm,
	ITEM_SUBCLASS_WEAPON_SWORD:        proto.WeaponType_WeaponTypeSword,
	ITEM_SUBCLASS_WEAPON_SWORD2:       proto.WeaponType_WeaponTypeSword,
	ITEM_SUBCLASS_WEAPON_WARGLAIVE:    proto.WeaponType_WeaponTypePolearm, // assuming polearm idk
	ITEM_SUBCLASS_WEAPON_STAFF:        proto.WeaponType_WeaponTypeStaff,
	ITEM_SUBCLASS_WEAPON_EXOTIC:       proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_EXOTIC2:      proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_FIST:         proto.WeaponType_WeaponTypeFist,
	ITEM_SUBCLASS_WEAPON_MISC:         proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_DAGGER:       proto.WeaponType_WeaponTypeDagger,
	ITEM_SUBCLASS_WEAPON_THROWN:       proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_SPEAR:        proto.WeaponType_WeaponTypePolearm,
	ITEM_SUBCLASS_WEAPON_CROSSBOW:     proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_WAND:         proto.WeaponType_WeaponTypeUnknown,
	ITEM_SUBCLASS_WEAPON_FISHING_POLE: proto.WeaponType_WeaponTypeUnknown,
}

var MapMinReputationToRepLevel = map[int]proto.RepLevel{
	0: proto.RepLevel_RepLevelUnknown,
	1: proto.RepLevel_RepLevelHostile,
	2: proto.RepLevel_RepLevelUnfriendly,
	3: proto.RepLevel_RepLevelNeutral,
	4: proto.RepLevel_RepLevelFriendly,
	5: proto.RepLevel_RepLevelHonored,
	6: proto.RepLevel_RepLevelRevered,
	7: proto.RepLevel_RepLevelExalted,
}

type EnchantMetaType struct {
	ItemType   proto.ItemType
	WeaponType proto.WeaponType
}

var SpellSchoolToSpellDamageStat = map[SpellSchool]proto.Stat{
	FIRE:     proto.Stat_StatFireDamage,
	ARCANE:   proto.Stat_StatArcaneDamage,
	NATURE:   proto.Stat_StatNatureDamage,
	FROST:    proto.Stat_StatFrostDamage,
	SHADOW:   proto.Stat_StatShadowDamage,
	HOLY:     proto.Stat_StatHolyDamage,
	PHYSICAL: proto.Stat_StatPhysicalDamage,
}
var SpellSchoolToResistanceStat = map[SpellSchool]proto.Stat{
	FIRE:     proto.Stat_StatFireResistance,
	ARCANE:   proto.Stat_StatArcaneResistance,
	NATURE:   proto.Stat_StatNatureResistance,
	FROST:    proto.Stat_StatFrostResistance,
	SHADOW:   proto.Stat_StatShadowResistance,
	PHYSICAL: proto.Stat_StatArmor,
}

var MapInventoryTypeToEnchantMetaType = map[InventoryTypeFlag]EnchantMetaType{
	HEAD:     {ItemType: proto.ItemType_ItemTypeHead, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	NECK:     {ItemType: proto.ItemType_ItemTypeNeck, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	SHOULDER: {ItemType: proto.ItemType_ItemTypeShoulder, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	CHEST:    {ItemType: proto.ItemType_ItemTypeChest, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	WAIST:    {ItemType: proto.ItemType_ItemTypeWaist, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	LEGS:     {ItemType: proto.ItemType_ItemTypeLegs, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	FEET:     {ItemType: proto.ItemType_ItemTypeFeet, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	WRIST:    {ItemType: proto.ItemType_ItemTypeWrist, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	HAND:     {ItemType: proto.ItemType_ItemTypeHands, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	FINGER:   {ItemType: proto.ItemType_ItemTypeFinger, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	TRINKET:  {ItemType: proto.ItemType_ItemTypeTrinket, WeaponType: proto.WeaponType_WeaponTypeUnknown},

	WEAPON_MAIN_HAND: {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeUnknown}, // One-Hand
	WEAPON_OFF_HAND:  {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeShield},  // Off Hand
	RANGED:           {ItemType: proto.ItemType_ItemTypeRanged, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	CLOAK:            {ItemType: proto.ItemType_ItemTypeBack, WeaponType: proto.WeaponType_WeaponTypeUnknown},
	TWO_H_WEAPON:     {ItemType: proto.ItemType_ItemTypeWeapon, WeaponType: proto.WeaponType_WeaponTypeUnknown},
}
var consumableClassToProto = map[ConsumableClass]proto.ConsumableType{
	EXPLOSIVES_AND_DEVICES: proto.ConsumableType_ConsumableTypeExplosive,
	POTION:                 proto.ConsumableType_ConsumableTypePotion,
	FLASK:                  proto.ConsumableType_ConsumableTypeFlask,
	SCROLL:                 proto.ConsumableType_ConsumableTypeScroll,
	FOOD:                   proto.ConsumableType_ConsumableTypeFood,
	BANDAGE:                proto.ConsumableType_ConsumableTypeBandage,
	OTHER:                  proto.ConsumableType_ConsumableTypeOther,
	HERB:                   proto.ConsumableType_ConsumableTypeHerb,
	DEVICE:                 proto.ConsumableType_ConsumableTypeDevice,
}

var MapPowerTypeEnumToResourceType = map[int32]proto.ResourceType{
	0: proto.ResourceType_ResourceTypeMana,
	1: proto.ResourceType_ResourceTypeRage,
	2: proto.ResourceType_ResourceTypeFocus,
	3: proto.ResourceType_ResourceTypeEnergy,
	4: proto.ResourceType_ResourceTypeComboPoints,
}

func ClassNameFromDBC(dbc DbcClass) string {
	switch dbc.ID {
	case 1:
		return "Warrior"
	case 2:
		return "Paladin"
	case 3:
		return "Hunter"
	case 4:
		return "Rogue"
	case 5:
		return "Priest"
	case 7:
		return "Shaman"
	case 8:
		return "Mage"
	case 9:
		return "Warlock"
	case 11:
		return "Druid"
	default:
		return "Unknown"
	}
}

type DbcClass struct {
	ProtoClass proto.Class
	ID         int
}

var Classes = []DbcClass{
	{proto.Class_ClassWarrior, 1},
	{proto.Class_ClassPaladin, 2},
	{proto.Class_ClassHunter, 3},
	{proto.Class_ClassRogue, 4},
	{proto.Class_ClassPriest, 5},
	{proto.Class_ClassShaman, 7},
	{proto.Class_ClassMage, 8},
	{proto.Class_ClassWarlock, 9},
	{proto.Class_ClassDruid, 11},
}
