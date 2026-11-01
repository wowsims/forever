package core

import (
	"fmt"

	"github.com/wowsims/forever/sim/core/dbcenums"
)

// The taken bits that name a direct hit arriving on the character. The two helpful takes are out
// because the sim models neither - one comes back as unsupported, the other is ignored - and the
// any-damage take has a rule of its own below.
const procFlagAnyDirectTaken = dbcenums.PROC_FLAG_TAKE_MELEE_SWING |
	dbcenums.PROC_FLAG_TAKE_MELEE_ABILITY |
	dbcenums.PROC_FLAG_TAKE_RANGED_ATTACK |
	dbcenums.PROC_FLAG_TAKE_RANGED_ABILITY |
	dbcenums.PROC_FLAG_TAKE_HARMFUL_ABILITY |
	dbcenums.PROC_FLAG_TAKE_HARMFUL_SPELL

// Every direct hit dealt.
const procFlagAnyDirectDealt = dbcenums.PROC_FLAG_DEAL_MELEE_SWING |
	dbcenums.PROC_FLAG_DEAL_MELEE_ABILITY |
	dbcenums.PROC_FLAG_DEAL_RANGED_ATTACK |
	dbcenums.PROC_FLAG_DEAL_RANGED_ABILITY |
	dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL

// What the mask alone cannot say, read off the spell's tooltip by the caller. The mask states
// which hits reach the listener; the wording around it states the trigger condition and, for the
// helpful bits, whether they mean anything at all.
type ProcHint uint16

const (
	// The tooltip names the cast itself as the trigger: "each time you cast a spell", "chance on
	// spell cast".
	ProcHintCastTrigger ProcHint = 1 << iota
	// The tooltip names a critical strike as the trigger.
	ProcHintCrit
	// The tooltip's trigger clause names healing, or an unrestricted "a spell". Either is the
	// evidence that a helpful-spell bit without the harmful one carries a real trigger rather than
	// a leftover.
	ProcHintHeals
	// The tooltip restricts the trigger to healing spells.
	ProcHintPureHeal
	// The trigger clause names one ability: "Your Shock spells", "Your Moonfire ability".
	ProcHintNamedAbility
	// The trigger clause states an attack outcome the mask has no bit for: a resist, a block, a
	// dodge or a parry.
	ProcHintOutcomeTaken
	// The trigger clause names the wearer's own attack being dodged: "when you are Parried or
	// Dodged".
	ProcHintAttackDodged
	// The same for the wearer's own attack being parried.
	ProcHintAttackParried
	// The trigger clause restricts the proc to spells with a damage component: "your damaging
	// spells", "your next 5 damage or healing spells".
	ProcHintDamage
)

// The wearer's own attack landing on nothing, which the decoder turns into the outcome it names.
const ProcHintAttackAvoided = ProcHintAttackDodged | ProcHintAttackParried

// Returns whether there is any overlap between the given hints.
func (h ProcHint) Matches(other ProcHint) bool {
	return (h & other) != 0
}

// The listener shape a ProcTypeMask describes, in the sim's own terms.
type ProcTypeInfo struct {
	Callback           AuraCallback
	ProcMask           ProcMask
	Outcome            HitOutcome
	RequireDamageDealt bool
	// The hit must deal damage or come from a spell with a DoT; see ProcTrigger.
	RequireDamageComponent bool
	Unsupported            []string // named bits the decoder does not model, empty when supported
}

// DecodeProcTypeMask reads SpellAuraOptions.ProcTypeMask (two 32-bit words).
//
// ProcHintNamedAbility is decoded past: a trigger restricted to one named ability is a shape no
// ProcTypeMask can state, so the caller refuses those spells rather than the decoder. Which
// outcome ProcHintOutcomeTaken names is the caller's too, for the same reason; all the decode
// takes from that hint is that the trigger is an outcome carrying no damage.
func DecodeProcTypeMask(mask [2]uint32, hint ProcHint) ProcTypeInfo {
	info := ProcTypeInfo{RequireDamageDealt: true}
	word := mask[0]

	if word&dbcenums.PROC_FLAG_DEAL_MELEE_SWING != 0 {
		info.ProcMask |= ProcMaskMeleeWhiteHit
	}

	if word&dbcenums.PROC_FLAG_DEAL_MELEE_ABILITY != 0 {
		info.ProcMask |= ProcMaskMeleeSpecial
	}

	if word&dbcenums.PROC_FLAG_DEAL_RANGED_ATTACK != 0 {
		info.ProcMask |= ProcMaskRangedAuto
	}

	if word&dbcenums.PROC_FLAG_DEAL_RANGED_ABILITY != 0 {
		info.ProcMask |= ProcMaskRangedSpecial
	}

	if word&(dbcenums.PROC_FLAG_DEAL_HARMFUL_PERIODIC|dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL) != 0 {
		info.ProcMask |= ProcMaskSpellDamage
	}

	if word&procFlagAnyDirectTaken != 0 {
		info.Callback |= CallbackOnSpellHitTaken

		if word&dbcenums.PROC_FLAG_TAKE_MELEE_SWING != 0 {
			info.ProcMask |= ProcMaskMeleeWhiteHit
		}

		if word&dbcenums.PROC_FLAG_TAKE_MELEE_ABILITY != 0 {
			info.ProcMask |= ProcMaskMeleeSpecial
		}

		if word&dbcenums.PROC_FLAG_TAKE_RANGED_ATTACK != 0 {
			info.ProcMask |= ProcMaskRangedAuto
		}

		if word&dbcenums.PROC_FLAG_TAKE_RANGED_ABILITY != 0 {
			info.ProcMask |= ProcMaskRangedSpecial
		}

		if word&dbcenums.PROC_FLAG_TAKE_HARMFUL_SPELL != 0 {
			info.ProcMask |= ProcMaskSpellDamage
		}
	}

	if word&dbcenums.PROC_FLAG_TAKE_HARMFUL_PERIODIC != 0 {
		info.Callback |= CallbackOnPeriodicDamageTaken
	}

	// Damage of any kind, however it arrived, which is both of the taken callbacks. The bit names
	// damage rather than a hit, so a landed hit dealing none does not count - the default this
	// decode starts from, and no game files mask pairs this bit with one that clears it.
	if word&dbcenums.PROC_FLAG_TAKE_ANY_DAMAGE != 0 {
		info.Callback |= CallbackOnSpellHitTaken | CallbackOnPeriodicDamageTaken
	}

	// The damage-class-none bit beside the magic one names the same casts a second way and says
	// nothing about the trigger, so the cast question is asked of the mask without it: 0x11000 is
	// the same shape as 0x10000, which is what unsupportedProcFlags says of that bit too.
	castWord := word &^ dbcenums.PROC_FLAG_DEAL_HARMFUL_ABILITY

	// A mask made of nothing but the spell-cast bits. The harmful one has to be present: a
	// helpful-only mask carries no evidence that casting is the trigger at all, and the helpful
	// branch below already demands tooltip evidence before it believes one.
	spellCastMask := castWord&dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL != 0 &&
		castWord&^(dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL|dbcenums.PROC_FLAG_DEAL_HELPFUL_SPELL) == 0

	// Whether the cast itself is the trigger. A mask of only the harmful-spell bit does not care
	// whether the spell landed. Adding the helpful bit settles nothing either way, and the two
	// items that pin it down disagree despite carrying the identical mask: Memento of Tyrande
	// procs off resists in logs, while Band of the Eternal Restorer does not proc on a miss or a
	// full resist. What separates them is that the first names the cast as the trigger and the
	// second does not, so for that pair the tooltip decides.
	castOnly := spellCastMask && (castWord == dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL || hint.Matches(ProcHintCastTrigger))

	// A tooltip naming an outcome is the exception to all of it: a crit is only known once the
	// hit resolves, so those stay on hit-dealt.
	if castOnly && !hint.Matches(ProcHintCrit) {
		info.Callback |= CallbackOnCastComplete
		info.RequireDamageDealt = false
	} else if word&procFlagAnyDirectDealt != 0 {
		info.Callback |= CallbackOnSpellHitDealt

		if word&dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL != 0 {
			info.RequireDamageDealt = false
		}
	}

	if word&dbcenums.PROC_FLAG_DEAL_HARMFUL_PERIODIC != 0 {
		info.Callback |= CallbackOnPeriodicDamageDealt
	}

	// The helpful-spell bit beside the harmful one is the game files' "any spell", heals included,
	// which the mask states on its own. A helpful bit without the harmful one needs the tooltip's
	// evidence that it is a trigger at all.
	helpfulSpells := word&dbcenums.PROC_FLAG_DEAL_HELPFUL_SPELL != 0 &&
		(word&dbcenums.PROC_FLAG_DEAL_HARMFUL_SPELL != 0 || hint.Matches(ProcHintHeals))

	if helpfulSpells {
		info.RequireDamageDealt = false
		info.ProcMask |= ProcMaskSpellHealing

		// Casting the heal is already the trigger above, so adding heal-dealt on top would
		// proc twice for one heal.
		if !info.Callback.Matches(CallbackOnCastComplete) {
			info.Callback |= CallbackOnHealDealt

			// handle HoTs only with direct heals for now, there are some odd cases with HoT / DoT overlaps
			if word&dbcenums.PROC_FLAG_DEAL_HELPFUL_PERIODIC != 0 {
				info.Callback |= CallbackOnPeriodicHealDealt
			}

			// Check if we have periodic damage flag but only heal paired with it
			// This usually indicates a pure heal proc mask
			if word&procFlagAnyDirectDealt == 0 {
				info.Callback &= ^CallbackOnPeriodicDamageDealt
				info.Callback &= ^CallbackOnSpellHitDealt
				info.ProcMask &= ^ProcMaskSpellDamage
			}
		}
	}

	// A mask naming one hand hears that hand's melee hits only. The bit restricts melee and
	// nothing else: a mask pairing it with a spell or ranged bit keeps those hits, so this drops
	// the other hand's melee bits rather than intersecting the whole proc mask. Naming both hands
	// is the same as naming neither, so only a mask with exactly one of the two restricts
	// anything.
	switch word & (dbcenums.PROC_FLAG_MAIN_HAND_WEAPON_SWING | dbcenums.PROC_FLAG_OFF_HAND_WEAPON_SWING) {
	case dbcenums.PROC_FLAG_MAIN_HAND_WEAPON_SWING:
		info.ProcMask &= ^ProcMaskMeleeOH
	case dbcenums.PROC_FLAG_OFF_HAND_WEAPON_SWING:
		info.ProcMask &= ^ProcMaskMeleeMH
	}
	if mask[1]&dbcenums.PROC_FLAG_2_MAIN_HAND_ONLY != 0 {
		info.ProcMask &= ^ProcMaskMeleeOH
	}

	// An avoidance outcome - a dodge, a parry, a miss, a full block or a resist - is a hit that
	// landed on nothing, so a listener whose trigger is one hears a hit that dealt no damage.
	// Which of them it is stays the caller's: no ProcTypeMask has a bit for any of them.
	if hint.Matches(ProcHintOutcomeTaken | ProcHintAttackAvoided) {
		info.RequireDamageDealt = false
	}

	// A trigger restricted to spells with a damage component hears a dealt hit that deals damage,
	// and the landed application of a DoT, which deals none. A cast has no hit to read that from,
	// and an avoided attack deals nothing by definition, so neither takes the rule.
	if hint.Matches(ProcHintDamage) && info.Callback.Matches(CallbackOnSpellHitDealt) &&
		!hint.Matches(ProcHintOutcomeTaken|ProcHintAttackAvoided) {
		info.RequireDamageDealt = false
		info.RequireDamageComponent = true
	}

	// An outcome the listener can be given. The crit hint names one, and so do the two naming the
	// wearer's own attack dodged or parried: the mask itself states which hits arrive, never how
	// they resolved, so everything else listens to a landed hit. A cast has not resolved into a
	// hit at all and takes no outcome.
	switch {
	case info.Callback.Matches(CallbackOnCastComplete):
		info.Outcome = OutcomeEmpty
	case hint.Matches(ProcHintAttackAvoided):
		if hint.Matches(ProcHintAttackDodged) {
			info.Outcome |= OutcomeDodge
		}
		if hint.Matches(ProcHintAttackParried) {
			info.Outcome |= OutcomeParry
		}
	case hint.Matches(ProcHintCrit):
		info.Outcome = OutcomeCrit
	default:
		info.Outcome = OutcomeLanded
	}

	if hint.Matches(ProcHintPureHeal) {
		info.Callback &= ^CallbackOnSpellHitDealt
		info.Callback &= ^CallbackOnPeriodicDamageDealt
	}

	info.Unsupported = unsupportedProcFlags(mask)

	return info
}

var unsupportedProcFlagNames = map[uint32]string{
	dbcenums.PROC_FLAG_HEARTBEAT:          "HEARTBEAT",
	dbcenums.PROC_FLAG_KILL:               "KILL",
	dbcenums.PROC_FLAG_TAKE_HELPFUL_SPELL: "TAKE_HELPFUL_SPELL",
	dbcenums.PROC_FLAG_DEATH:              "DEATH",
	dbcenums.PROC_FLAG_JUMP:               "JUMP",
	dbcenums.PROC_FLAG_ENTER_COMBAT:       "ENTER_COMBAT",
	dbcenums.PROC_FLAG_ENCOUNTER_START:    "ENCOUNTER_START",
	dbcenums.PROC_FLAG_CAST_ENDED:         "CAST_ENDED",
	dbcenums.PROC_FLAG_LOOTED:             "LOOTED",
}

// The bits the shape above says nothing about. The three damage-class-none bits (0x400 and 0x1000
// dealt, 0x800 taken) are left out on purpose: next to real bits they change nothing the decode
// says, and on their own they leave the callback empty, which the caller refuses anyway.
func unsupportedProcFlags(mask [2]uint32) []string {
	// Everything from the death bit up is a state change rather than a hit, and of word 1 the sim
	// models the main-hand restriction alone.
	unsupported := mask[0] & (dbcenums.PROC_FLAG_HEARTBEAT | dbcenums.PROC_FLAG_KILL | dbcenums.PROC_FLAG_TAKE_HELPFUL_SPELL | ^(dbcenums.PROC_FLAG_DEATH - 1))

	var names []string
	for bit := range 32 {
		if unsupported&(1<<bit) == 0 {
			continue
		}

		if name, ok := unsupportedProcFlagNames[1<<bit]; ok {
			names = append(names, name)
		} else {
			names = append(names, fmt.Sprintf("bit %d", bit))
		}
	}

	word1 := mask[1] &^ dbcenums.PROC_FLAG_2_MAIN_HAND_ONLY
	for bit := range 32 {
		if word1&(1<<bit) != 0 {
			names = append(names, fmt.Sprintf("bit %d", 32+bit))
		}
	}

	return names
}
