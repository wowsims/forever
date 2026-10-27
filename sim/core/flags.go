package core

import (
	"strconv"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

//go:generate stringer -type=ProcMask
//go:generate stringer -type=AuraCallback
//go:generate stringer -type=SpellFlag
//go:generate stringer -type=SpellSchool
type ProcMask uint32

// Returns whether there is any overlap between the given masks.
func (pm ProcMask) Matches(other ProcMask) bool {
	return (pm & other) != 0
}

// Actual Blizzard flag values:
// 1  1        = Triggered by script
// 2  2        = Triggers on kill
// 3  4        = Melee auto attack
// 4  8        = On take melee auto attack
// 5  16       = Melee special attack / melee damage
// 6  32       = On take melee special attack
// 7  64       = Ranged auto attack
// 8  128      = On take ranged auto attack
// 9  256      = Ranged special attack / ranged damage
// 10 512      = On take ranged special attack
// 11 1024     = ???? On use combo points? Shapeshift? Change stance? Gain buff? Some rogue stuff
// 12 2048     = ???? Rogue related? Script?
// 13 4096     = ???? Stealth related? Script? On gain/lose stealth? Also possibly on stance change
// 14 8192     = On spell hit on you
// 15 16384    = Cast heal
// 16 32768    = On get healed
// 17 65536    = Deal spell damage
// 18 131072   = On take spell damage
// 19 262144   = Deal periodic damage
// 20 524288   = On take periodic damage
// 21 1048576  = On take any damage
// 22 2097152  = On Apply debuff
// 23 4194304  = ???? On have debuff applied to you? really bizarre mask
// 24 8388608  = On offhand attack
// 25 16777216 = What the fuck?

// Single-bit masks. These don't need to match Blizzard's values.
const (
	// Default value is invalid, to force authors to think about proc masks.
	ProcMaskUnknown ProcMask = 0

	ProcMaskEmpty ProcMask = 1 << iota
	ProcMaskMeleeMHAuto
	ProcMaskMeleeOHAuto
	ProcMaskMeleeMHSpecial
	ProcMaskMeleeOHSpecial
	ProcMaskRangedAuto
	ProcMaskRangedSpecial
	ProcMaskSpellDamage
	ProcMaskSpellHealing
	// Spell damage dealt by a weapon imbue or poison (Flametongue, Frostbrand, rogue poisons). A hit
	// kind, not proc-ness (that is SpellFlagProc). It is deliberately not a member of
	// ProcMaskSpellDamage, so a listener whose mask is ProcMaskSpellDamage never hears these hits
	// even when it can proc from procs; that is the whole reason the bit survives.
	ProcMaskSpellDamageProc

	ProcMaskLast
)

const (
	ProcMaskMeleeMH = ProcMaskMeleeMHAuto | ProcMaskMeleeMHSpecial
	ProcMaskMeleeOH = ProcMaskMeleeOHAuto | ProcMaskMeleeOHSpecial
	// Equivalent to in-game mask of 4.
	ProcMaskMeleeWhiteHit = ProcMaskMeleeMHAuto | ProcMaskMeleeOHAuto
	// Equivalent to in-game mask of 68.
	ProcMaskWhiteHit = ProcMaskMeleeMHAuto | ProcMaskMeleeOHAuto | ProcMaskRangedAuto
	// Equivalent to in-game mask of 16.
	ProcMaskMeleeSpecial = ProcMaskMeleeMHSpecial | ProcMaskMeleeOHSpecial
	// Equivalent to in-game mask of 272.
	ProcMaskMeleeOrRangedSpecial = ProcMaskMeleeSpecial | ProcMaskRangedSpecial
	// Equivalent to in-game mask of 20.
	ProcMaskMelee = ProcMaskMeleeWhiteHit | ProcMaskMeleeSpecial
	// Equivalent to in-game mask of 320.
	ProcMaskRanged = ProcMaskRangedAuto | ProcMaskRangedSpecial
	// Equivalent to in-game mask of 340.
	ProcMaskMeleeOrRanged = ProcMaskMelee | ProcMaskRanged

	ProcMaskDirect = ProcMaskMelee | ProcMaskRanged | ProcMaskSpellDamage

	ProcMaskSpecial = ProcMaskMeleeOrRangedSpecial | ProcMaskSpellDamage
)

// Possible outcomes of any hit/damage roll.
type HitOutcome uint16

// Returns whether there is any overlap between the given masks.
func (ho HitOutcome) Matches(other HitOutcome) bool {
	return (ho & other) != 0
}

// Single-bit outcomes.
const (
	OutcomeEmpty HitOutcome = 0

	// These bits are set by the hit roll
	OutcomeMiss HitOutcome = 1 << iota
	OutcomeHit
	OutcomeDodge
	OutcomeGlance
	OutcomeParry
	OutcomeBlock

	// These bits are set by the crit and damage rolls.
	OutcomeCrit
	OutcomeCrush

	// The tenths of the damage a partial resist took (1 to 10), as a 4-bit number.
	OutcomePartial1
	OutcomePartial2
	OutcomePartial4
	OutcomePartial8
)

const (
	OutcomePartial = OutcomePartial1 | OutcomePartial2 | OutcomePartial4 | OutcomePartial8
	OutcomeLanded  = OutcomeHit | OutcomeCrit | OutcomeCrush | OutcomeGlance | OutcomeBlock
)

func (ho HitOutcome) String() string {
	if ho.Matches(OutcomeMiss) {
		return "Miss"
	} else if ho.Matches(OutcomeDodge) {
		return "Dodge"
	} else if ho.Matches(OutcomeParry) {
		return "Parry"
	} else if ho.Matches(OutcomeBlock) && ho.Matches(OutcomeCrit) {
		return "BlockedCrit"
	} else if ho.Matches(OutcomeBlock) {
		return "Block"
	} else if ho.Matches(OutcomeGlance) {
		return "Glance" + ho.PartialResistString()
	} else if ho.Matches(OutcomeCrit) {
		return "Crit" + ho.PartialResistString()
	} else if ho.Matches(OutcomeHit) {
		return "Hit" + ho.PartialResistString()
	} else if ho.Matches(OutcomeCrush) {
		return "Crush"
	} else {
		return "Empty"
	}
}

func partialResistOutcome(tenths int) HitOutcome {
	return HitOutcome(tenths) * OutcomePartial1
}

func (ho HitOutcome) ResistedTenths() int {
	return int((ho & OutcomePartial) / OutcomePartial1)
}

func (ho HitOutcome) PartialResistString() string {
	if tenths := ho.ResistedTenths(); tenths > 0 {
		return " (" + strconv.Itoa(tenths*10) + "% Resist)"
	}
	return ""
}

// Other flags
type SpellFlag uint64

// Returns whether there is any overlap between the given masks.
func (se SpellFlag) Matches(other SpellFlag) bool {
	return (se & other) != 0
}

const (
	SpellFlagNone                     SpellFlag = 0
	SpellFlagIgnoreResists            SpellFlag = 1 << iota // skip spell resist/armor
	SpellFlagIgnoreTargetModifiers                          // skip target damage modifiers
	SpellFlagIgnoreAttackerModifiers                        // skip attacker damage modifiers
	SpellFlagApplyArmorReduction                            // Forces damage reduction from armor to apply, even if it otherwise wouldn't.
	SpellFlagCannotBeDodged                                 // Ignores dodge in physical hit rolls
	SpellFlagIncludeTargetBonusDamage                       // Spell benefits from Gift of Arthas and Hemorrhage.
	SpellFlagBinary                                         // The hit roll resists it in full at the target's average resist; a physical one is blocked whole.
	SpellFlagChanneled                                      // Spell is channeled
	SpellFlagDisease                                        // Spell is categorized as disease
	SpellFlagPoison                                         // Spell is categorized as poison
	SpellFlagHauntSE                                        // Spell benefits from haunt/SE effects
	SpellFlagHelpful                                        // For healing spells / buffs.
	SpellFlagMeleeMetrics                                   // Marks a spell as a melee ability for metrics.
	SpellFlagNoOnCastComplete                               // Disables the OnCastComplete callback.
	SpellFlagNoMetrics                                      // Disables metrics for a spell.
	SpellFlagNoLogs                                         // Disables logs for a spell.
	SpellFlagAPL                                            // Indicates this spell can be used from an APL rotation.
	SpellFlagMCD                                            // Indicates this spell is a MajorCooldown.
	SpellFlagReactive                                       // Allows a spell flagged as an MCD to be cast off-GCD. Used for instant cast defensive CDs.
	SpellFlagNoOnDamageDealt                                // Disables the damage callbacks for this spell, dealt and taken alike: OnCastResult, OnSpellHitDealt, OnPeriodicDamageDealt, OnSpellHitTaken and OnPeriodicDamageTaken.
	SpellFlagPrepullOnly                                    // Indicates this spell should only be used during prepull. Not enforced, just a signal for the APL UI.
	SpellFlagEncounterOnly                                  // Indicates this spell should only be used during the encounter (not prepull). Not enforced, just a signal for the APL UI.
	SpellFlagPotion                                         // Indicates this spell is a potion spell.
	SpellFlagConjured                                       // Indicates this spell is a conjured item spell.
	SpellFlagExplosive                                      // Indicates this spell is an engineering explosive.
	SpellFlagNoSpellMods                                    // Indicates that no spell mods should be applied to this spell
	SpellFlagCanCastWhileMoving                             // Allows the cast to be casted while moving
	SpellFlagPassiveSpell                                   // Indicates this spell is applied/cast as a result of another spell
	SpellFlagSuppressWeaponProcs                            // The Suppress Weapon Procs attribute (Attributes[4] 0x800000): weapon procs, and auras marked "aura is weapon proc", ignore this spell's hits. In TBC that is Seal of Blood, Righteousness and Martyr damage plus a few CC spells.
	SpellFlagProc                                           // Not a DBC attribute. The spell is cast by an aura or item effect and lacks Not a Proc (Attributes[3] 0x200), so aura procs only hear its hits when they can proc from procs. Weapon procs do not care.
	SpellFlagSupressDoTApply                                // If present this spell will not apply dots (Used for DTR dot supression)
	SpellFlagSwapped                                        // Indicates that this spell is not useable because it is from a currently swapped item
	SpellFlagCastWhileIncapacitated                         // Allows the cast while the unit is incapacitated, e.g. Berserker Rage breaking a Fear.
	SpellFlagFullResistanceTicks                            // Its ticks meet the target's full resistance, as direct hits and channel ticks do, where other ticks meet a tenth (Insect Swarm).
	SpellFlagPushback                                       // Damage taken while hardcasting pushes the cast back: SpellInterrupts.InterruptFlags 0x2.
	SpellFlagNoPartialResist                                // The spell's hits never lose damage to a partial resist, e.g. a wand's shots.
	SpellFlagKeepsSwingTimer                                // Completing a cast of the spell leaves the melee swing timer running: Do Not Reset Combat Timers (Attributes[2] 0x20000).
	SpellFlagMovesToTarget                                  // The cast moves the caster toward its target, e.g. Charge or Blink. Not enforced, just a signal for the APL validation.

	// Used to let agents categorize their spells.
	SpellFlagAgentReserved1
	SpellFlagAgentReserved2
	SpellFlagAgentReserved3
	SpellFlagAgentReserved4

	SpellFlagIgnoreModifiers = SpellFlagIgnoreAttackerModifiers | SpellFlagIgnoreTargetModifiers
)

type SpellSchool byte

// The bit values are the game files'. SpellMisc.SchoolMask uses exactly this order, so a generated
// spell row carries the number the DBC states rather than a translation of it. Nothing outside this
// file depends on the numbers: the conversions to stats.SchoolIndex, the resistance stat and the
// school damage stat are all switches over these names, stats.SchoolIndex is its own enum for array
// positions, and proto.SpellSchool is separate again.
const (
	SpellSchoolNone     SpellSchool = 0
	SpellSchoolPhysical SpellSchool = 1
	SpellSchoolHoly     SpellSchool = 2
	SpellSchoolFire     SpellSchool = 4
	SpellSchoolNature   SpellSchool = 8
	SpellSchoolFrost    SpellSchool = 16
	SpellSchoolShadow   SpellSchool = 32
	SpellSchoolArcane   SpellSchool = 64

	SpellSchoolChaos       SpellSchool = SpellSchoolArcane | SpellSchoolFire | SpellSchoolFrost | SpellSchoolHoly | SpellSchoolNature | SpellSchoolShadow
	SpellSchoolShadowFlame SpellSchool = SpellSchoolFire | SpellSchoolShadow
	SpellSchoolShadowFrost SpellSchool = SpellSchoolFrost | SpellSchoolShadow
	SpellSchoolPlague      SpellSchool = SpellSchoolNature | SpellSchoolShadow
	SpellSchoolFirestorm   SpellSchool = SpellSchoolFire | SpellSchoolNature
	SpellSchoolFrostfire   SpellSchool = SpellSchoolFire | SpellSchoolFrost
	SpellSchoolElemental   SpellSchool = SpellSchoolFire | SpellSchoolNature | SpellSchoolFrost
)

// Returns whether there is any overlap between the given masks.
func (ss SpellSchool) Matches(other SpellSchool) bool {
	return (ss & other) != 0
}

// Holy has no resistance stat, and neither do spell school combinations, which supposedly use the
// "path of the least resistance".
func (ss SpellSchool) ResistanceStat() (stats.Stat, bool) {
	switch ss {
	case SpellSchoolPhysical:
		return stats.ArmorPenetration, true
	case SpellSchoolArcane:
		return stats.ArcaneResistance, true
	case SpellSchoolFire:
		return stats.FireResistance, true
	case SpellSchoolFrost:
		return stats.FrostResistance, true
	case SpellSchoolNature:
		return stats.NatureResistance, true
	case SpellSchoolShadow:
		return stats.ShadowResistance, true
	}
	return 0, false
}

// The stat a target resistance mask pierces: ArmorPenetration for physical alone and SpellPiercing for
// a mask of every school with a resistance. Any other mask has no stat: a mask of physical and a spell
// school pierces nothing, and a mask of some spell schools pierces those schools alone.
func (ss SpellSchool) PiercingStat() (stats.Stat, bool) {
	const resistances = SpellSchoolFire | SpellSchoolNature | SpellSchoolFrost | SpellSchoolShadow | SpellSchoolArcane
	switch {
	case ss == SpellSchoolPhysical:
		return stats.ArmorPenetration, true
	case ss.Matches(SpellSchoolPhysical):
		return 0, false
	case ss&resistances == resistances:
		return stats.SpellPiercing, true
	}
	return 0, false
}

func (ss SpellSchool) SchoolDamage() stats.Stat {
	switch ss {
	case SpellSchoolArcane:
		return stats.ArcaneDamage
	case SpellSchoolFire:
		return stats.FireDamage
	case SpellSchoolFrost:
		return stats.FrostDamage
	case SpellSchoolHoly:
		return stats.HolyDamage
	case SpellSchoolNature:
		return stats.NatureDamage
	case SpellSchoolShadow:
		return stats.ShadowDamage
	default:
		return stats.SpellDamage
	}
}

var singleSchools = [...]SpellSchool{SpellSchoolPhysical, SpellSchoolArcane, SpellSchoolFire, SpellSchoolFrost,
	SpellSchoolHoly, SpellSchoolNature, SpellSchoolShadow}

// The school Spell.SchoolIndex reads: the first of physical, arcane, fire, frost, holy, nature, shadow.
func (ss SpellSchool) MultiplierSchool() SpellSchool {
	for _, school := range singleSchools {
		if ss.Matches(school) {
			return school
		}
	}
	return SpellSchoolNone
}

func (ss SpellSchool) SchoolIndex() stats.SchoolIndex {
	switch ss {
	case SpellSchoolNone:
		return stats.SchoolIndexNone
	case SpellSchoolPhysical:
		return stats.SchoolIndexPhysical
	case SpellSchoolArcane:
		return stats.SchoolIndexArcane
	case SpellSchoolFire:
		return stats.SchoolIndexFire
	case SpellSchoolFrost:
		return stats.SchoolIndexFrost
	case SpellSchoolHoly:
		return stats.SchoolIndexHoly
	case SpellSchoolNature:
		return stats.SchoolIndexNature
	case SpellSchoolShadow:
		return stats.SchoolIndexShadow
	default:
		return stats.SchoolIndexPhysical
	}
}

func SpellSchoolFromProto(p proto.SpellSchool) SpellSchool {
	switch p {
	case proto.SpellSchool_SpellSchoolPhysical:
		return SpellSchoolPhysical
	case proto.SpellSchool_SpellSchoolArcane:
		return SpellSchoolArcane
	case proto.SpellSchool_SpellSchoolFire:
		return SpellSchoolFire
	case proto.SpellSchool_SpellSchoolFrost:
		return SpellSchoolFrost
	case proto.SpellSchool_SpellSchoolHoly:
		return SpellSchoolHoly
	case proto.SpellSchool_SpellSchoolNature:
		return SpellSchoolNature
	case proto.SpellSchool_SpellSchoolShadow:
		return SpellSchoolShadow
	default:
		return SpellSchoolPhysical
	}
}

/*
outcome roll hit/miss/crit/glance (assigns Outcome mask) -> If Hit, Crit Roll -> damage (applies metrics) -> trigger proc

So in TBC it looks like they just gave it the cannot miss flag even though they also switched its defense type to physical (??)
the damage type is holy, which ignores armor and as it is magic so it can be partially resisted (due to level resistance).
however it also gains the physical bit mask as I explain in a post above

so there is no hit roll, there is a melee crit roll, a spell damage roll, and melee "on hit"

ok so I did some more testing on this.
Judgement of Blood correctly gets the "always hit" (aka cannot miss flag applied to it) --
its only mitigation events are partial resists at the correct rates
however Judgement of Command is broken. even though it has the "always hit" flag it seems to
be ignored because it is procced by an intermediary dummy spell which does not have the "cannot miss" flag applied to it lmao.
for some god forsaken reason Judgement of Command is ALSO a dummy which then casts the correct Judgement of Command
which deals damage, and this dummy can miss, lmao
I got ~16.4% resists in about almost 96 casts which suggests it uses the spell hit check,
which makes sense because its defensetype is set to 1, Magic

arcane shot - ranged hit, spell dmg, procs special ranged
	OutcomeRollRanged, School Arcane, ProcMask - RangedSpecial

judgement of blood - physical hit/crit, spell damage, "cannot miss", procs special melee and ranged
	Damage is (weapon damage + spell power)*0.7*(bonus holy damage against target)+flat bonus damage
	OtherFlagCannotMiss, OutcomeRollSpecial, School Holy (base damage = weapon damage range), Multiplier 70%

judgement of command - spell hit, melee crit, spell damage, procs special melee and ranged
	OutcomeRollSpell, School Holy


moonfire - spell hit, spell dmg, dot dmg, procs spell hit
stormstrike - melee hit, melee dmg, procs special melee
rupture -

wotlk
shadowflame - requires each 'effect' to have its own school.
*/
