package dbc

import (
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

const (
	ITEM_ENCHANTMENT_NONE             int = 0
	ITEM_ENCHANTMENT_COMBAT_SPELL     int = 1
	ITEM_ENCHANTMENT_DAMAGE           int = 2
	ITEM_ENCHANTMENT_EQUIP_SPELL      int = 3
	ITEM_ENCHANTMENT_RESISTANCE       int = 4
	ITEM_ENCHANTMENT_STAT             int = 5
	ITEM_ENCHANTMENT_TOTEM            int = 6
	ITEM_ENCHANTMENT_USE_SPELL        int = 7
	ITEM_ENCHANTMENT_PRISMATIC_SOCKET int = 8
	ITEM_ENCHANTMENT_RELIC_RANK       int = 9
	ITEM_ENCHANTMENT_APPLY_BONUS      int = 11
	ITEM_ENCHANTMENT_RELIC_EVIL       int = 12 // Scaling relic +ilevel, see enchant::initialize_relic
)

// ItemBonus.Type values of the rows a random suffix's bonus list carries.
const (
	ITEM_BONUS_STAT               int = 2  // Value_0 the ITEM_MOD stat, Value_1 its allocation per 10000 of the item's random property points.
	ITEM_BONUS_SUFFIX             int = 5  // Value_0 the ItemNameDescription the item's name ends with.
	ITEM_BONUS_ITEM_EFFECT        int = 23 // Value_0 the ItemEffect whose spell the item gains.
	ITEM_BONUS_REDUCED_CASTER_DPS int = 54 // No value: a weapon with the list deals the REDUCED_CASTER_DPS share of its damage table.
)

const (
	ITEM_SPELLTRIGGER_ON_USE          int = 0 // use after equip cooldown
	ITEM_SPELLTRIGGER_ON_EQUIP        int = 1
	ITEM_SPELLTRIGGER_CHANCE_ON_HIT   int = 2
	ITEM_SPELLTRIGGER_SOULSTONE       int = 4
	ITEM_SPELLTRIGGER_ON_NO_DELAY_USE int = 5 // no equip cooldown
	ITEM_SPELLTRIGGER_LEARN_SPELL_ID  int = 6
)

// The game files' spell enums live in sim/core/dbcenums, which the sim reads as well. The aliases
// keep the field types of this package's rows spelled as its own.
type SpellEffectType = dbcenums.SpellEffectType
type EffectAuraType = dbcenums.EffectAuraType

type ItemSubClass int

// Define each item subclass as a bit flag (only those with a name).
const (
	OneHandedAxes    ItemSubClass = 1 << 0  // 1    from "One-Handed Axes" (SubClassID 0)
	TwoHandedAxes    ItemSubClass = 1 << 1  // 2    from "Two-Handed Axes" (SubClassID 1)
	Bows             ItemSubClass = 1 << 2  // 4    from "Bows" (SubClassID 2)
	Guns             ItemSubClass = 1 << 3  // 8    from "Guns" (SubClassID 3)
	OneHandedMaces   ItemSubClass = 1 << 4  // 16   from "One-Handed Maces" (SubClassID 4)
	TwoHandedMaces   ItemSubClass = 1 << 5  // 32   from "Two-Handed Maces" (SubClassID 5)
	Polearms         ItemSubClass = 1 << 6  // 64   from "Polearms" (SubClassID 6)
	OneHandedSwords  ItemSubClass = 1 << 7  // 128  from "One-Handed Swords" (SubClassID 7)
	TwoHandedSwords  ItemSubClass = 1 << 8  // 256  from "Two-Handed Swords" (SubClassID 8)
	Relic            ItemSubClass = 1 << 9  // 512  from "Relic" (SubClassID 9)
	Staves           ItemSubClass = 1 << 10 // 1024 from "Staves" (SubClassID 10)
	OneHandedExotics ItemSubClass = 1 << 11 // 2048 from "One-Handed Exotics" (SubClassID 11)
	TwoHandedExotics ItemSubClass = 1 << 12 // 4096 from "Two-Handed Exotics" (SubClassID 12)
	FistWeapons      ItemSubClass = 1 << 13 // 8192 from "Fist Weapons" (SubClassID 13)
	Daggers          ItemSubClass = 1 << 15 // 32768 from "Daggers" (SubClassID 15)
)

type ItemQuality int

const (
	JUNK      ItemQuality = 0
	COMMON    ItemQuality = 1
	UNCOMMON  ItemQuality = 2
	RARE      ItemQuality = 3
	EPIC      ItemQuality = 4
	LEGENDARY ItemQuality = 5
	ARTIFACT  ItemQuality = 6
	HEIRLOOM  ItemQuality = 7
)

func (raw ItemQuality) ToProto() proto.ItemQuality {
	switch raw {
	case JUNK:
		return proto.ItemQuality_ItemQualityJunk
	case COMMON:
		return proto.ItemQuality_ItemQualityCommon
	case UNCOMMON:
		return proto.ItemQuality_ItemQualityUncommon
	case RARE:
		return proto.ItemQuality_ItemQualityRare
	case EPIC:
		return proto.ItemQuality_ItemQualityEpic
	case LEGENDARY:
		return proto.ItemQuality_ItemQualityLegendary
	case ARTIFACT:
		return proto.ItemQuality_ItemQualityArtifact
	case HEIRLOOM:
		return proto.ItemQuality_ItemQualityHeirloom
	}
	return proto.ItemQuality_ItemQualityUncommon
}

const (
	ITEM_CLASS_CONSUMABLE = iota
	ITEM_CLASS_CONTAINER
	ITEM_CLASS_WEAPON = dbcenums.ITEM_CLASS_WEAPON
	ITEM_CLASS_GEM    = iota
	ITEM_CLASS_ARMOR
	ITEM_CLASS_REAGENT
	ITEM_CLASS_PROJECTILE
	ITEM_CLASS_TRADE_GOODS
	ITEM_CLASS_GENERIC
	ITEM_CLASS_RECIPE
	ITEM_CLASS_MONEY
	ITEM_CLASS_QUIVER
	ITEM_CLASS_QUEST
	ITEM_CLASS_KEY
	ITEM_CLASS_PERMANENT
	ITEM_CLASS_MISC
)

const (
	ITEM_SUBCLASS_WEAPON_AXE = iota
	ITEM_SUBCLASS_WEAPON_AXE2
	ITEM_SUBCLASS_WEAPON_BOW
	ITEM_SUBCLASS_WEAPON_GUN
	ITEM_SUBCLASS_WEAPON_MACE
	ITEM_SUBCLASS_WEAPON_MACE2
	ITEM_SUBCLASS_WEAPON_POLEARM
	ITEM_SUBCLASS_WEAPON_SWORD
	ITEM_SUBCLASS_WEAPON_SWORD2
	ITEM_SUBCLASS_WEAPON_WARGLAIVE
	ITEM_SUBCLASS_WEAPON_STAFF
	ITEM_SUBCLASS_WEAPON_EXOTIC
	ITEM_SUBCLASS_WEAPON_EXOTIC2
	ITEM_SUBCLASS_WEAPON_FIST
	ITEM_SUBCLASS_WEAPON_MISC
	ITEM_SUBCLASS_WEAPON_DAGGER
	ITEM_SUBCLASS_WEAPON_THROWN
	ITEM_SUBCLASS_WEAPON_SPEAR
	ITEM_SUBCLASS_WEAPON_CROSSBOW
	ITEM_SUBCLASS_WEAPON_WAND
	ITEM_SUBCLASS_WEAPON_FISHING_POLE
)

const ITEM_SUBCLASS_WEAPON_INVALID = 31

const (
	ITEM_SUBCLASS_ARMOR_MISC = iota
	ITEM_SUBCLASS_ARMOR_CLOTH
	ITEM_SUBCLASS_ARMOR_LEATHER
	ITEM_SUBCLASS_ARMOR_MAIL
	ITEM_SUBCLASS_ARMOR_PLATE
	ITEM_SUBCLASS_ARMOR_COSMETIC
	ITEM_SUBCLASS_ARMOR_SHIELD
	ITEM_SUBCLASS_ARMOR_LIBRAM
	ITEM_SUBCLASS_ARMOR_IDOL
	ITEM_SUBCLASS_ARMOR_TOTEM
	ITEM_SUBCLASS_ARMOR_SIGIL
	ITEM_SUBCLASS_ARMOR_RELIC
)

const (
	ITEM_SUBCLASS_CONSUMABLE = iota
	ITEM_SUBCLASS_POTION
	ITEM_SUBCLASS_ELIXIR
	ITEM_SUBCLASS_FLASK
	ITEM_SUBCLASS_SCROLL
	ITEM_SUBCLASS_FOOD
	ITEM_SUBCLASS_ITEM_ENHANCEMENT
	ITEM_SUBCLASS_BANDAGE
	ITEM_SUBCLASS_CONSUMABLE_OTHER
)

const (
	INVTYPE_NON_EQUIP = iota
	INVTYPE_HEAD
	INVTYPE_NECK
	INVTYPE_SHOULDERS
	INVTYPE_BODY
	INVTYPE_CHEST
	INVTYPE_WAIST
	INVTYPE_LEGS
	INVTYPE_FEET
	INVTYPE_WRISTS
	INVTYPE_HANDS
	INVTYPE_FINGER
	INVTYPE_TRINKET
	INVTYPE_WEAPON
	INVTYPE_SHIELD
	INVTYPE_RANGED
	INVTYPE_CLOAK
	INVTYPE_2HWEAPON
	INVTYPE_BAG
	INVTYPE_TABARD
	INVTYPE_ROBE
	INVTYPE_WEAPONMAINHAND
	INVTYPE_WEAPONOFFHAND
	INVTYPE_HOLDABLE
	INVTYPE_AMMO
	INVTYPE_THROWN
	INVTYPE_RANGEDRIGHT
	INVTYPE_QUIVER
	INVTYPE_RELIC
	INVTYPE_MAX
)

type ConsumableClass int

const (
	EXPLOSIVES_AND_DEVICES ConsumableClass = iota
	POTION
	ELIXIR
	FLASK
	SCROLL
	FOOD
	ITEM_ENHANCEMENT
	BANDAGE
	OTHER
	HERB
	DEVICE ConsumableClass = 14
)

const (
	TRADE_GOODS_EXPLOSIVES = 2
	// The game files' shared cooldown category for explosives, 60 s.
	EXPLOSIVES_SPELL_CATEGORY = 24
	// The categories potions, food, drink and conjured items (healthstones, mana gems, runes, Thistle
	// Tea) share their cooldown through.
	POTION_SPELL_CATEGORY   = 4
	FOOD_SPELL_CATEGORY     = 11
	DRINK_SPELL_CATEGORY    = 59
	CONJURED_SPELL_CATEGORY = 1153
)

// Power type in EffectMiscValue_0 of A_MOD_INCREASE_ENERGY. Only mana maps to a stat.
const POWER_TYPE_MANA int = 0

// https://wowdev.wiki/Stat_Types
type ItemModType uint

const (
	ITEM_MOD_MANA                     = 0
	ITEM_MOD_HEALTH                   = 1
	ITEM_MOD_AGILITY                  = 3
	ITEM_MOD_STRENGTH                 = 4
	ITEM_MOD_INTELLECT                = 5
	ITEM_MOD_SPIRIT                   = 6
	ITEM_MOD_STAMINA                  = 7
	ITEM_MOD_DEFENSE_SKILL_RATING     = 12
	ITEM_MOD_DODGE_RATING             = 13
	ITEM_MOD_PARRY_RATING             = 14
	ITEM_MOD_BLOCK_RATING             = 15
	ITEM_MOD_HIT_MELEE_RATING         = 16
	ITEM_MOD_HIT_RANGED_RATING        = 17
	ITEM_MOD_HIT_SPELL_RATING         = 18
	ITEM_MOD_CRIT_MELEE_RATING        = 19
	ITEM_MOD_CRIT_RANGED_RATING       = 20
	ITEM_MOD_CRIT_SPELL_RATING        = 21
	ITEM_MOD_HIT_TAKEN_MELEE_RATING   = 22
	ITEM_MOD_HIT_TAKEN_RANGED_RATING  = 23
	ITEM_MOD_HIT_TAKEN_SPELL_RATING   = 24
	ITEM_MOD_CRIT_TAKEN_MELEE_RATING  = 25
	ITEM_MOD_CRIT_TAKEN_RANGED_RATING = 26
	ITEM_MOD_CRIT_TAKEN_SPELL_RATING  = 27
	ITEM_MOD_HASTE_MELEE_RATING       = 28
	ITEM_MOD_HASTE_RANGED_RATING      = 29
	ITEM_MOD_HASTE_SPELL_RATING       = 30
	ITEM_MOD_HIT_RATING               = 31
	ITEM_MOD_CRIT_RATING              = 32
	ITEM_MOD_HIT_TAKEN_RATING         = 33
	ITEM_MOD_CRIT_TAKEN_RATING        = 34
	ITEM_MOD_RESILIENCE_RATING        = 35
	ITEM_MOD_HASTE_RATING             = 36
	ITEM_MOD_EXPERTISE_RATING         = 37
	ITEM_MOD_ATTACK_POWER             = 38
	ITEM_MOD_RANGED_ATTACK_POWER      = 39
	ITEM_MOD_FERAL_ATTACK_POWER       = 40
	ITEM_MOD_SPELL_HEALING_DONE       = 41
	ITEM_MOD_SPELL_DAMAGE_DONE        = 42
	ITEM_MOD_MANA_REGENERATION        = 43
	ITEM_MOD_ARMOR_PENETRATION_RATING = 44
	ITEM_MOD_SPELL_POWER              = 45
	ITEM_MOD_HEALTH_REGEN             = 46
	ITEM_MOD_SPELL_PENETRATION        = 47
	ITEM_MOD_BLOCK_VALUE              = 48
	ITEM_MOD_MAX                      = 49

	// Indices above ITEM_MOD_MAX that the Forever game files actually put on items.
	// 50-56 are standard (wowdev.wiki/Stat_Types); 50 is already handled below.
	// 83+ has no upstream definition -- the canonical enum stops at 74 -- so these
	// were identified from self-labelling test items and enchant rows in the game files.
	ITEM_MOD_EXTRA_ARMOR       = 50
	ITEM_MOD_FIRE_RESISTANCE   = 51
	ITEM_MOD_FROST_RESISTANCE  = 52
	ITEM_MOD_HOLY_RESISTANCE   = 53
	ITEM_MOD_SHADOW_RESISTANCE = 54
	ITEM_MOD_NATURE_RESISTANCE = 55
	ITEM_MOD_ARCANE_RESISTANCE = 56

	ITEM_MOD_WEAPON_DAMAGE = 83
	ITEM_MOD_HOLY_DAMAGE   = 84
	ITEM_MOD_FIRE_DAMAGE   = 85
	ITEM_MOD_NATURE_DAMAGE = 86
	ITEM_MOD_FROST_DAMAGE  = 87
	ITEM_MOD_SHADOW_DAMAGE = 88
	ITEM_MOD_ARCANE_DAMAGE = 89

	// Spell penetration on one school, in points of the target's resistance: 21516 Dagger of Spell
	// Penetration - Fire 150 Resist (TEST) budgets 150 of 119, and 21518's Frost twin 150 of 121.
	ITEM_MOD_FIRE_SPELL_PENETRATION  = 119
	ITEM_MOD_FROST_SPELL_PENETRATION = 121

	ITEM_MOD_ALL_RESISTANCES = 124
)

type Race int

// Define each race as a bit flag
// Table: ChrRaces
const (
	Human    Race = 1 << 0  // 1
	Orc      Race = 1 << 1  // 2
	Dwarf    Race = 1 << 2  // 4
	NightElf Race = 1 << 3  // 8
	Undead   Race = 1 << 4  // 16
	Tauren   Race = 1 << 5  // 32
	Gnome    Race = 1 << 6  // 64
	Troll    Race = 1 << 7  // 128
	Goblin   Race = 1 << 8  // 256
	BloodElf Race = 1 << 9  // 512
	Draenei  Race = 1 << 10 // 1024

	HighOrderSkyborne  Race = 1 << 32
	WindshaperSkyborne Race = 1 << 33

	AlliedRaces Race = Human | Dwarf | NightElf | Gnome | Draenei | HighOrderSkyborne
	HordeRaces  Race = Undead | Orc | Tauren | Troll | Goblin | BloodElf | WindshaperSkyborne
)

// Returns whether there is any overlap between the given masks.
func (race Race) Matches(other Race) bool {
	return (race & other) != 0
}

// SpellEffect.ImplicitTarget_0, restricted to the values a buff or debuff aura
// reaches its targets by. Table: SpellEffect, column ImplicitTarget.
type ImplicitTarget int

const (
	TARGET_UNIT_CASTER            ImplicitTarget = 1
	TARGET_UNIT_TARGET_ENEMY      ImplicitTarget = 6
	TARGET_UNIT_SRC_AREA_ENEMY    ImplicitTarget = 15
	TARGET_UNIT_CASTER_AREA_PARTY ImplicitTarget = 20
	TARGET_UNIT_TARGET_ALLY       ImplicitTarget = 21
	TARGET_SRC_CASTER             ImplicitTarget = 22
	TARGET_UNIT_TARGET_ANY        ImplicitTarget = 25
	TARGET_UNIT_CASTER_AREA_RAID  ImplicitTarget = 56
	TARGET_UNIT_TARGET_RAID       ImplicitTarget = 57
	TARGET_153                    ImplicitTarget = 153
)
