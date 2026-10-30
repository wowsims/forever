package dbcenums

// The misc value of A_ADD_FLAT_MODIFIER and A_ADD_PCT_MODIFIER: the spell property the modifier
// changes. The game files ship no name list, so every op is declared under TrinityCore's names - its
// 3.3.5 ones up to 30 and its current ones above - and checked against the talents that use them.
// The comment says what the op modifies where that is known, and names talents that use it.
type SpellModOp int32

const (
	SPELLMOD_DAMAGE                    SpellModOp = 0  // damage and healing done - Fire Power, Piercing Ice, Contagion
	SPELLMOD_DURATION                  SpellModOp = 1  // aura duration - Permafrost, Improved Gouge, Brutal Impact
	SPELLMOD_THREAT                    SpellModOp = 2  // threat generated - Subtlety, Improved Drain Soul
	SPELLMOD_EFFECT1                   SpellModOp = 3  // the first effect's value - Arcane Potency, Improved Concentration Aura
	SPELLMOD_CHARGES                   SpellModOp = 4  // aura charges - Improved Shield Block, Improved Holy Shield
	SPELLMOD_RANGE                     SpellModOp = 5  // range - Arctic Reach, Flame Throwing, Grim Reach
	SPELLMOD_RADIUS                    SpellModOp = 6  // area radius - Arctic Reach, Holy Reach, Booming Voice
	SPELLMOD_CRITICAL_CHANCE           SpellModOp = 7  // critical strike chance - Arcane Impact, Improved Flamestrike, Incineration
	SPELLMOD_ALL_EFFECTS               SpellModOp = 8  // every effect's value - Frost Warding, Magic Attunement, Demonic Aegis
	SPELLMOD_NOT_LOSE_CASTING_TIME     SpellModOp = 9  // pushback taken while casting - Burning Soul, Fel Concentration, Intensity
	SPELLMOD_CASTING_TIME              SpellModOp = 10 // cast time - Accelerant, Improved Frostbolt
	SPELLMOD_COOLDOWN                  SpellModOp = 11 // cooldown - Improved Fire Blast, Improved Frost Nova, Ice Floes
	SPELLMOD_EFFECT2                   SpellModOp = 12 // the second effect's value - Malediction, Mana Feed
	SPELLMOD_IGNORE_ARMOR              SpellModOp = 13
	SPELLMOD_COST                      SpellModOp = 14 // power cost - Frost Channeling, Cataclysm
	SPELLMOD_CRIT_DAMAGE_BONUS         SpellModOp = 15 // critical strike damage bonus - Ice Shards, Ruin, Vengeance
	SPELLMOD_RESIST_MISS_CHANCE        SpellModOp = 16 // chance to hit - Shadow Focus, Suppression
	SPELLMOD_JUMP_TARGETS              SpellModOp = 17 // chain targets
	SPELLMOD_CHANCE_OF_SUCCESS         SpellModOp = 18 // Improved Poisons, Improved Nature's Grasp
	SPELLMOD_ACTIVATION_TIME           SpellModOp = 19 // Improved Fire Totems
	SPELLMOD_DAMAGE_MULTIPLIER         SpellModOp = 20
	SPELLMOD_GLOBAL_COOLDOWN           SpellModOp = 21 // global cooldown - Improved Slam
	SPELLMOD_DOT                       SpellModOp = 22 // periodic damage and healing - Emberstorm, Contagion, Fire Power
	SPELLMOD_EFFECT3                   SpellModOp = 23 // the third effect's value - Improved Faerie Fire, Savage Fury
	SPELLMOD_BONUS_MULTIPLIER          SpellModOp = 24 // spell power coefficient - Empowered Arcane Missiles / Fireball / Frostbolt / Corruption
	SPELLMOD_TRIGGER_DAMAGE            SpellModOp = 25
	SPELLMOD_PROC_PER_MINUTE           SpellModOp = 26 // procs per minute
	SPELLMOD_VALUE_MULTIPLIER          SpellModOp = 27 // Improved Mana Shield
	SPELLMOD_RESIST_DISPEL_CHANCE      SpellModOp = 28 // chance to resist a dispel - Vile Poisons, Sanctified Seals
	SPELLMOD_CRIT_DAMAGE_BONUS_2       SpellModOp = 29
	SPELLMOD_SPELL_COST_REFUND_ON_FAIL SpellModOp = 30
	SPELLMOD_DOSES                     SpellModOp = 31
	SPELLMOD_EFFECT4                   SpellModOp = 32 // the fourth effect's value
	SPELLMOD_EFFECT5                   SpellModOp = 33 // the fifth effect's value
	SPELLMOD_COST2                     SpellModOp = 34
	SPELLMOD_JUMP_DISTANCE             SpellModOp = 35
	SPELLMOD_AREATRIGGER_MAX_SUMMONS   SpellModOp = 36
	SPELLMOD_MAX_AURA_STACKS           SpellModOp = 37 // maximum aura stacks
	SPELLMOD_PROC_COOLDOWN             SpellModOp = 38 // internal cooldown between procs
	SPELLMOD_COST3                     SpellModOp = 39
	SPELLMOD_MAX_TARGETS               SpellModOp = 40 // maximum targets
)
