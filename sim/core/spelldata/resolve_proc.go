package spelldata

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// An addition to the resolved trigger for what the game files do not state, which on a proc is the
// rate: a row that states none needs one of PPM, ChanceFrom or Chance.
type ProcOpt func(*core.Character, *core.ProcTrigger)

// What the game files state about a proc, as the fields core registers a listener through: what it
// hears (ProcTypeMask through the decoder, with the row's tooltip hint), how often it fires, its
// internal cooldown and the attribute-driven gates. The caller adds what the row does not carry -
// the name and action the sim keys the effect by, the buff the handler applies - to the returned
// value before handing it to MakeProcTriggerAura. The row's own class flags go on the trigger too,
// so a SPELLMOD_CHANCE_OF_SUCCESS that names the row reaches its roll, and the character's mods are
// the ones counted wherever the trigger sits.
//
// ProcCharges is not a property of the trigger: it is how many times the buff acts before it drops,
// which AuraConfig reads as the aura's MaxStacks.
//
// The character is needed for the procs-per-minute managers, which are bound to the weapon carrying
// them, so a trigger is built where the sim has a character rather than at package init.
func ProcTrigger(character *core.Character, s *Spell, handler core.ProcHandler, opts ...ProcOpt) core.ProcTrigger {
	decoded := core.DecodeProcTypeMask(s.ProcFlags, s.ProcHint)

	trigger := core.ProcTrigger{
		Name:                s.Name,
		ActionID:            core.ActionID{SpellID: s.ID},
		Callback:            decoded.Callback,
		ProcMask:            decoded.ProcMask,
		Outcome:             decoded.Outcome,
		RequireDamageDealt:  decoded.RequireDamageDealt,
		ICD:                 s.ICD(),
		CanProcFromProcs:    s.CanProcFromProcs(),
		ClassSpellsOnly:     s.ClassSpellsOnly(),
		ClassFlags:          listenerClassFlags(character, s),
		ProcSpellClassFlags: s.ClassFlags,
		Caster:              &character.Unit,
		Handler:             handler,

		RequireDamageComponent: decoded.RequireDamageComponent,
	}

	if s.IsWeaponProcAura() {
		trigger.SpellFlagsExclude |= core.SpellFlagSuppressWeaponProcs
	}

	fillProcChance(s, &trigger)

	// The row first and the caller's options on top, so an option sees the mask the row decoded to
	// and can override a rate the row states wrongly or not at all.
	for _, opt := range opts {
		opt(character, &trigger)
	}

	settleProcRate(character, s, &trigger)

	return trigger
}

// A hand-supplied rate, for a proc whose real rate lives outside the spell data. It reads the
// trigger's own proc mask, so an option that narrows the mask has to be given before this one.
func PPM(ppm float64) ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		// The callback rolls the chance first and the manager only where the chance let it
		// through, so a column chance left next to a manager would gate the rate twice.
		trigger.ProcChance = 0
		trigger.DPM = character.NewLegacyPPMManager(ppm, trigger.ProcMask)
	}
}

// A hand-supplied rate on the main hand's speed as Seal of the Crusader leaves it, through
// core.NewWeaponSpeedPPMManager. Like PPM, it reads the trigger's proc mask.
func WeaponSpeedPPM(ppm float64) ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewWeaponSpeedPPMManager(ppm, trigger.ProcMask)
	}
}

// The chance an effect states, for a $mN the generator could not resolve into ProcChanceEffect.
func ChanceFrom(e *Effect) ProcOpt {
	return Chance(e.Percent())
}

// A weapon proc's listener, which no row states: the game casts a "Chance on hit" effect and a
// combat enchant off every eligible weapon hit, so the trigger hears them all and the weapon it
// sits on decides which ones count. A hit that lands is enough, so a special that deals no damage,
// such as Sunder Armor, rolls it too. The outcome stays what the row's tooltip hint read, and the
// aura-side proc-ness attribute goes: a weapon proc ignores proc-ness by its own rule. A rate option
// that reads the mask has to come after this one.
func WeaponProc() ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.Callback = core.CallbackOnSpellHitDealt
		trigger.ProcMask = core.ProcMaskUnknown
		trigger.RequireDamageDealt = false
		trigger.RequireDamageComponent = false
		trigger.CanProcFromProcs = false
		trigger.SpellFlagsExclude &^= core.SpellFlagSuppressWeaponProcs
		trigger.IsWeaponProc = true
	}
}

// A rate the caller states outright, which replaces whatever the row said, manager included.
func Chance(chance float64) ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = chance
		trigger.DPM = nil
	}
}

// The spells the listener fires on where the game files name them, which is the
// EffectSpellClassMask of the effect that carries the proc, read against the class wearing it.
func listenerClassFlags(character *core.Character, s *Spell) core.ClassFlags {
	flags := rowClassFlags(s)
	if othersFamily(character, flags) {
		return core.ClassFlags{}
	}

	return flags
}

// Only the proc effect's flags: another effect of the same spell states the spells it modifies,
// not the ones that feed the proc.
func rowClassFlags(s *Spell) core.ClassFlags {
	for i := range s.Effects {
		e := &s.Effects[i]

		if e.Aura.IsProcTrigger() || e.Aura == dbcenums.A_PROC_TRIGGER_SPELL_COPY ||
			e.Aura == dbcenums.A_PROC_TRIGGER_DAMAGE || e.Aura == dbcenums.A_DUMMY {
			if !e.ClassFlags.IsZero() {
				return e.ClassFlags
			}
		}
	}

	return core.ClassFlags{}
}

// Whether the mask names spells of a family this character's class never casts, which is how an
// item shared by every class states the filter of the one class it was written for: Dreadnaught's
// 8pc 28845 names family 5, the warlock's, and on a warrior that mask matches nothing and silences
// the listener rather than narrowing it. A class the game files state no family for - and the empty
// mask every non-class spell carries - is no evidence of anything, so both are left alone.
//
// Only another class's family counts. The game files put the potions under family 13 and a good deal
// of generic content under 0 or 1, and a mask in one of those names spells every class can use, so
// dropping it would silence a listener the wearer really does hear. With no character there is no
// class to read the mask against, which ParseEffects answers the same way: the rows that need one
// are left out rather than guessed at.
func othersFamily(character *core.Character, flags core.ClassFlags) bool {
	if character == nil {
		return false
	}

	family, stated := core.ClassSpellFamilies[character.Class]
	return stated && !flags.IsZero() && flags.Family != family && isClassFamily(flags.Family)
}

// Whether the family is the one some class files its own spells under.
func isClassFamily(family int32) bool {
	for _, classFamily := range core.ClassSpellFamilies {
		if classFamily == family {
			return true
		}
	}
	return false
}

// The roll the row states, by the source that says where it is stated. The ProcChance column is the
// roll only under ProcChanceColumn; under the others it means nothing, which is why reading it
// directly is the bug ProcChanceSource exists to prevent. 0 where the row states no roll, which under
// ProcChancePPM is a rate the caller supplies through PPM().
func (s *Spell) StatedChance() float64 {
	switch s.ProcChanceSource {
	case ProcChanceColumn:
		// 101 is the game files' "fires on its own condition" sentinel next to a real 100.
		return min(float64(s.ProcChance)/100, 1)
	case ProcChanceEffectN:
		// The position the roll is stated at can be past the effects the row carries.
		return s.EffectN(int(s.ProcChanceEffect)).Percent()
	case ProcChanceAlways:
		return 1
	}
	return 0
}

// Whether an item or enchant proc rolls the ProcChance column rather than the chance its tooltip
// states on an effect: the column holds a real roll, 1 to 99, that the effect's contradicts. Uther's
// Light 8397 states 2% on effect 1 and 4 in the column. A class spell keeps the effect's, which is
// where a talent's ranks state theirs.
func (s *Spell) ItemProcRollsTheColumn() bool {
	return s.ProcChanceSource == ProcChanceEffectN && s.RPPM == 0 && s.ProcChance > 0 && s.ProcChance < 100 &&
		float64(s.ProcChance)/100 != s.StatedChance()
}

// The column's chance on an item or enchant proc ItemProcRollsTheColumn picks it for. A rate option
// that replaces the chance has to come after this one.
func ItemProcChance(s *Spell) ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		if s.ItemProcRollsTheColumn() {
			trigger.ProcChance = float64(s.ProcChance) / 100
		}
	}
}

// A row whose rate is procs per minute states no roll: that rate is a manager, and settleProcRate
// builds it once the options have had their say about the mask it measures hits on.
func fillProcChance(s *Spell, trigger *core.ProcTrigger) {
	if s.RPPM > 0 {
		return
	}
	trigger.ProcChance = s.StatedChance()
}

// The rate the trigger ends up with. An override-supplied procs-per-minute rate wins over every
// column - the rows that carry one are the rows whose stated chance the tooltip contradicts - but it
// is measured against the mask the trigger *ends* on rather than the one the row decoded to: an
// option that narrows or blanks the mask, as a weapon proc's shape does, would otherwise be handed a
// manager counting hits the listener no longer hears.
//
// A trigger left with neither a chance nor a manager fires on every qualifying hit, which is never
// what a row with no stated rate means.
func settleProcRate(character *core.Character, s *Spell, trigger *core.ProcTrigger) {
	if trigger.ProcChance != 0 || trigger.DPM != nil {
		return
	}

	if s.RPPM > 0 {
		// A mask of nothing is the weapon-proc shape, where which hits count is the weapon's to say
		// and the manager has to be bound to it by whoever knows which weapon that is.
		if trigger.ProcMask == core.ProcMaskUnknown {
			panic(fmt.Sprintf("spelldata: spell %d (%s) states %g procs per minute but no proc mask to measure them on; pass a manager bound to what carries it",
				s.ID, s.Name, s.RPPM))
		}

		trigger.DPM = character.NewLegacyPPMManager(float64(s.RPPM), trigger.ProcMask)
		return
	}

	panic(fmt.Sprintf("spelldata: spell %d (%s) states no proc chance; pass PPM()", s.ID, s.Name))
}
