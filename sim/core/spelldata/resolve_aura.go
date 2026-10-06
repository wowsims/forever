package spelldata

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
)

// An addition to the resolved aura for what the game files do not state: the label the sim keys the
// aura by and the callbacks it acts through.
type AuraOpt func(*core.Aura)

// What the game files state about a spell's aura. The caller adds the callbacks and anything the
// game files do not carry to the returned value before registering it.
//
// A row that states no duration resolves to one of 0, which core refuses to activate: such an aura
// needs Permanent() or a duration of the caller's.
func AuraConfig(s *Spell, opts ...AuraOpt) core.Aura {
	aura := core.Aura{
		Label:     s.Name,
		ActionID:  core.ActionID{SpellID: s.ID},
		Duration:  s.Duration(),
		MaxStacks: maxStacks(s),
	}

	for _, opt := range opts {
		opt(&aura)
	}
	return aura
}

// The sim keeps charges and stacks in one field, and so do the game files on all but a handful of
// spells: CumulativeAura counts the stacks an aura builds up, ProcCharges counts the times it acts
// before it drops, and a spell that states both is read as a stacking one.
func maxStacks(s *Spell) int32 {
	if s.MaxStack > 0 {
		return int32(s.MaxStack)
	}
	return int32(s.ProcCharges)
}

// Two auras of the same spell on one unit need labels of their own.
func Label(label string) AuraOpt {
	return func(aura *core.Aura) {
		aura.Label = label
	}
}

// Ranks of one spell registered side by side need an aura each: ranks that share the spell's name share
// one aura on the unit, and a cast of one rank then runs every rank's callbacks. The owner's label keeps
// two casters' debuffs on one target apart.
func RankLabel(s *Spell, owner *core.Unit) AuraOpt {
	return Label(s.RankedLabel(owner))
}

// An aura that is up for the whole iteration, whatever duration the row states.
func Permanent() AuraOpt {
	return func(aura *core.Aura) {
		core.MakePermanent(aura)
	}
}

// What the game files state about a periodic effect of a spell, as the dot core registers. The
// ticks are the row's duration over the effect's period. The default OnTick deals the effect's own
// Average at the caster's level, on the stats and multipliers in force at the tick; a caller
// replaces OnTick for a heal or a value the game files do not state. OnSnapshot is nil.
//
// The aura's duration is the row's, which core recomputes from the ticks on every application.
func DotConfig(s *Spell, e *Effect, opts ...AuraOpt) core.DotConfig {
	if e.PeriodMs <= 0 {
		panic(fmt.Sprintf("spelldata: effect %d of spell %d (%s) states no tick period",
			e.Index, s.ID, s.Name))
	}
	// DurationMs, not Duration(): the game files' -1 is a permanent aura, which reads as a duration
	// longer than any encounter and would resolve to an absurd number of ticks.
	if s.DurationMs <= 0 {
		panic(fmt.Sprintf("spelldata: spell %d (%s) states no duration to tick over", s.ID, s.Name))
	}

	return core.DotConfig{
		Aura:             AuraConfig(s, opts...),
		TickLength:       e.Period(),
		NumberOfTicks:    int32(s.Duration() / e.Period()),
		BonusCoefficient: e.Coeff(),
		OnTick:           PeriodicDamageTick(e, s.TickOutcome),
	}
}

func PeriodicDamageTick(e *Effect, outcome func(*core.Dot) core.OutcomeApplier) core.OnTick {
	return func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, e.Average(), outcome(dot))
	}
}
