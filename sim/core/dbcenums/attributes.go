package dbcenums

// Spell attribute flags, named after the Attributes column they live in: ATTR_EX_3 is a flag
// in Attributes[3]. Read them through the Spell helpers rather than indexing Attributes.
const (
	// The spell is never cast: a stance's passive, a talent that only modifies other spells.
	ATTR_PASSIVE uint32 = 0x40

	// The spell cannot be cast while the caster is in any shapeshift form.
	ATTR_NOT_SHAPESHIFTED uint32 = 0x10000

	// The spell cannot be dodged, parried or blocked (No Active Defense).
	ATTR_NO_ACTIVE_DEFENSE uint32 = 0x200000

	// The two bits the game files mark a channel with. Arcane Missiles and Blizzard carry the first,
	// Evocation and Tranquility only the second, so a channel check has to read both.
	ATTR_EX_1_IS_CHANNELLED      uint32 = 0x4
	ATTR_EX_1_IS_SELF_CHANNELLED uint32 = 0x40

	// The server refunds 80% of the power cost when the spell misses: every rage special that
	// costs rage up front, Heroic Strike and Rend among them. Cleave and Whirlwind lack it.
	ATTR_EX_1_DISCOUNT_POWER_ON_MISS uint32 = 0x8000000

	ATTR_EX_2_CANT_CRIT uint32 = 0x20000000

	// The cast does not reset the auto attack timers.
	ATTR_EX_2_DO_NOT_RESET_COMBAT_TIMERS uint32 = 0x20000

	// The spell is castable while shapeshifted even though its form bars casting: the exception to
	// ATTR_NOT_SHAPESHIFTED.
	ATTR_EX_2_CASTABLE_IN_CASTER_FORM uint32 = 0x80000

	// Only a player can carry the aura: a target that is not a player, such as a pet, is invalid.
	// Strength of Earth and Grace of Air Totem carry it at every rank.
	ATTR_EX_3_ONLY_ON_PLAYER uint32 = 0x100

	// On a triggered spell: aura listeners treat its hits like a normal ability hit. Seal of
	// Command damage, every Judgement, Stormstrike's bonus hits and Sweeping Strikes carry it.
	ATTR_EX_3_NOT_A_PROC          uint32 = 0x200
	ATTR_EX_3_CAN_PROC_FROM_PROCS uint32 = 0x4000000

	// Weapon procs (Player::CastItemCombatSpell) ignore hits of this spell, as do auras marked
	// ATTR_EX_6_AURA_IS_WEAPON_PROC. In TBC that is the Seal of Blood, Righteousness and Martyr
	// damage spells plus Gouge, Sap, Scatter Shot and Maim.
	ATTR_EX_4_SUPPRESS_WEAPON_PROCS uint32 = 0x800000

	// An aura proc that honours ATTR_EX_4_SUPPRESS_WEAPON_PROCS anyway: Black Bow of the Betrayer
	// and the Sunwell melee neck.
	ATTR_EX_6_AURA_IS_WEAPON_PROC uint32 = 0x80

	// A periodic effect whose ticks roll a critical strike: Rend, Corruption, Rupture and the
	// other bleeds and DoTs the game files mark, 225 spells in this build.
	ATTR_EX_8_PERIODIC_CAN_CRIT uint32 = 0x200

	ATTR_EX_11_SCALES_WITH_ITEM_LEVEL uint32 = 0x4

	ATTR_EX_12_ONLY_PROC_FROM_CLASS_ABILITIES uint32 = 0x80000000
)

// The remaining named attribute bits.
const (
	ATTR_RANGED_ABILITY                     uint32 = 0x2
	ATTR_ABILITY                            uint32 = 0x10
	ATTR_TRADESKILL_ABILITY                 uint32 = 0x20
	ATTR_HIDDEN                             uint32 = 0x80
	ATTR_REQ_STEALTH                        uint32 = 0x20000
	ATTR_CANCEL_AUTO_ATTACK                 uint32 = 0x100000
	ATTR_NOT_IN_COMBAT                      uint32 = 0x10000000
	ATTR_NO_CANCEL                          uint32 = 0x80000000
	ATTR_EX_1_NO_STEALTH_BREAK              uint32 = 0x20
	ATTR_EX_1_MELEE_COMBAT_START            uint32 = 0x200
	ATTR_EX_1_NO_THREAT                     uint32 = 0x400
	ATTR_EX_1_DONT_DISPLAY_IN_AURA_BAR      uint32 = 0x10000000
	ATTR_EX_2_FOOD_AURA                     uint32 = 0x80000000
	ATTR_EX_3_REQ_MAIN_HAND                 uint32 = 0x400
	ATTR_EX_3_SUPPRESS_CASTER_PROCS         uint32 = 0x10000
	ATTR_EX_3_SUPPRESS_TARGET_PROCS         uint32 = 0x20000
	ATTR_EX_3_ALWAYS_HIT                    uint32 = 0x40000
	ATTR_EX_3_REQ_OFF_HAND                  uint32 = 0x1000000
	ATTR_EX_3_TREAT_AS_PERIODIC             uint32 = 0x2000000
	ATTR_EX_4_DISABLE_TARGET_MULT           uint32 = 0x100
	ATTR_EX_5_TICK_ON_APPLICATION           uint32 = 0x200
	ATTR_EX_5_DOT_HASTED                    uint32 = 0x2000
	ATTR_EX_5_TREAT_AS_AREA_EFFECT          uint32 = 0x8000
	ATTR_EX_5_REQ_LINE_OF_SIGHT             uint32 = 0x4000000
	ATTR_EX_6_IGNORE_FOR_MOD_TIME_RATE      uint32 = 0x10
	ATTR_EX_6_NO_PUSHBACK                   uint32 = 0x8000
	ATTR_EX_6_DISABLE_PLAYER_MULT           uint32 = 0x20000000
	ATTR_EX_7_NO_DODGE                      uint32 = 0x800000
	ATTR_EX_7_NO_PARRY                      uint32 = 0x1000000
	ATTR_EX_7_NO_MISS                       uint32 = 0x2000000
	ATTR_EX_7_CAN_PROC_FROM_SUPPRESSED_TGT  uint32 = 0x40000000
	ATTR_EX_8_NO_BLOCK                      uint32 = 0x1
	ATTR_EX_8_DURATION_HASTED               uint32 = 0x20000
	ATTR_EX_8_REQUIRES_EQUIPPED_ARMOR_TYPE  uint32 = 0x100000
	ATTR_EX_8_DOT_HASTED_MELEE              uint32 = 0x400000
	ATTR_EX_8_MASTERY_AFFECTS_POINTS        uint32 = 0x20000000
	ATTR_EX_9_FIXED_TRAVEL_TIME             uint32 = 0x10
	ATTR_EX_9_DISABLE_PLAYER_HEALING_MULT   uint32 = 0x1000000
	ATTR_EX_10_DISABLE_TARGET_POSITIVE_MULT uint32 = 0x2
	ATTR_EX_10_TARGET_SPECIFIC_COOLDOWN     uint32 = 0x400
	ATTR_EX_10_ROLLING_PERIODIC             uint32 = 0x4000
	ATTR_EX_12_ENABLE_PROCS_FROM_SUPPRESSED uint32 = 0x1 // requires CAN_PROC_FROM_SUPPRESSED on driver
	ATTR_EX_12_CAN_PROC_FROM_SUPPRESSED     uint32 = 0x2 // requires ENABLE_PROCS_FROM_SUPPRESSED on action
	ATTR_EX_13_ALLOW_CLASS_ABILITY_PROCS    uint32 = 0x1
	ATTR_EX_13_REFRESH_EXTENDS_DURATION     uint32 = 0x100000
	ATTR_EX_15_AURA_DOES_NOT_REFRESH        uint32 = 0x200
	ATTR_EX_15_ASYNCHRONOUS_STACKING_AURA   uint32 = 0x400
	ATTR_EX_15_IMPORTANT_SPELL              uint32 = 0x800
	ATTR_EX_15_IS_EXTERNAL_DEFENSIVE        uint32 = 0x80000
	ATTR_EX_16_IS_BIG_DEFENSIVE             uint32 = 0x1
)

// Attributes index each ATTR_EX_ flag above belongs to.
const (
	ATTR_INDEX_BASE  int = 0
	ATTR_INDEX_EX_1  int = 1
	ATTR_INDEX_EX_2  int = 2
	ATTR_INDEX_EX_3  int = 3
	ATTR_INDEX_EX_4  int = 4
	ATTR_INDEX_EX_6  int = 6
	ATTR_INDEX_EX_8  int = 8
	ATTR_INDEX_EX_9  int = 9
	ATTR_INDEX_EX_11 int = 11
	ATTR_INDEX_EX_12 int = 12
)
