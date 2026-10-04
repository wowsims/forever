package spelldata

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
)

// The reason given for a proc whose rows state no rate, which the generator also gives where the
// rows' rate is a sentinel its tooltip contradicts.
const ReasonStatesNoRate = "states no rate"

const ReasonPPMHearsNoWeaponHits = "a procs-per-minute rate hears no weapon hits in this mask"

// What a proc the game files hang on an item, an enchant or a set bonus does not model, as one reason
// per shape. An empty answer means the rows state enough to build the listener the game files
// describe.
//
// The one decision, for the generator that writes the registrations and for the sim that resolves
// them: both read this row, so a proc the generator emits is one the sim can build, and a proc it
// comments out is one nothing would have heard correctly.
//
// A class's own spell states a class mask to narrow its listener to spells the caster really casts;
// an item is worn by every class, so a class mask on its trigger is the filter of the one class it
// was written for and names nothing the sim can reproduce. An item proc therefore has no character
// to weigh the mask against.
//
// itemPPM is the procs-per-minute rate a weapon's item effect states, which answers a trigger whose
// row states none; 0 where the item states no rate.
func ItemProcUnsupportedWithRate(trigger *Spell, isWeaponProc bool, itemPPM float64) []string {
	return itemProcUnsupported(trigger, isWeaponProc, false, itemPPM)
}

// The same for an enchant's equip aura, whose procs-per-minute rate rolls on weapon hits only: the
// sim never procs one from a spell or a heal.
func EnchantAuraUnsupported(trigger *Spell) []string {
	return itemProcUnsupported(trigger, false, true, 0)
}

func itemProcUnsupported(trigger *Spell, isWeaponProc bool, isEnchant bool, itemPPM float64) []string {
	var unsupported []string

	if trigger == Nil || trigger.ID == 0 {
		return []string{"no row in the store"}
	}

	decoded := core.DecodeProcTypeMask(trigger.ProcFlags, trigger.ProcHint)

	// No ProcTypeMask can state either: a trigger restricted to one named ability, and one
	// restricted to an outcome the mask has no bit for.
	if trigger.ProcHint.Matches(core.ProcHintNamedAbility) {
		unsupported = append(unsupported, "named ability")
	}
	if trigger.ProcHint.Matches(core.ProcHintOutcomeTaken) {
		unsupported = append(unsupported, "an outcome the proc mask has no bit for")
	}
	if !rowClassFlags(trigger).IsZero() {
		unsupported = append(unsupported, "the trigger carries a class mask")
	}

	for _, bits := range decoded.Unsupported {
		unsupported = append(unsupported, fmt.Sprintf("ProcTypeMask %s", bits))
	}

	// A weapon proc hears the hits of whatever carries it, which is the one shape the row states no
	// listener for: the game casts a chance-on-hit effect and a combat enchant off the hit itself.
	if isWeaponProc {
		if itemPPM == 0 && !weaponProcRateStated(trigger) {
			unsupported = append(unsupported, ReasonStatesNoRate)
		}
		return unsupported
	}

	if decoded.Callback == core.CallbackEmpty {
		unsupported = append(unsupported, "no callback in the proc mask")
	}

	switch {
	case !procRateStated(trigger):
		unsupported = append(unsupported, ReasonStatesNoRate)
	case trigger.RPPM > 0 && decoded.ProcMask == core.ProcMaskUnknown:
		// Procs per minute are measured against the hits the listener hears, and an empty mask
		// counts none of them. On a weapon proc the weapon answers that; nothing else can.
		unsupported = append(unsupported, "no proc mask to measure its rate on")
	case isEnchant && trigger.RPPM > 0 && !decoded.ProcMask.Matches(core.ProcMaskMeleeOrRanged):
		unsupported = append(unsupported, ReasonPPMHearsNoWeaponHits)
	}

	return unsupported
}

// The same for a weapon proc, where the game files' 100 and 101 are no answer. Those mean "fires
// whenever its own condition is met", and a chance-on-hit effect has no condition: the game casts it
// off the weapon's hit without consulting the row at all. Arcanite Champion, Annihilator and the
// combat enchants all read 101, and every one of them is a rate the spell data does not carry.
func weaponProcRateStated(s *Spell) bool {
	if s.ProcChanceSource == ProcChanceAlways {
		return s.RPPM > 0
	}

	return procRateStated(s)
}

// Whether the row states a rate the trigger can be built from at all: a roll of its own, or the
// procs-per-minute an override wrote onto it. A row stating neither would fire on every hit, which
// is never what the game files mean.
func procRateStated(s *Spell) bool {
	return s.RPPM > 0 || s.StatedChance() != 0
}
