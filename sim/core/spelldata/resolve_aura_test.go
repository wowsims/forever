package spelldata

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
)

func TestAuraConfigFromARow(t *testing.T) {
	withRows(t, resolverRows())
	aura := AuraConfig(Find(800))

	if aura.Label != "Bleed" {
		t.Errorf("label = %q, want the spell's name", aura.Label)
	}
	if aura.ActionID != (core.ActionID{SpellID: 800}) {
		t.Errorf("ActionID = %v, want spell 800", aura.ActionID)
	}
	if aura.Duration != time.Second*21 {
		t.Errorf("duration = %v, want 21s", aura.Duration)
	}
	if aura.MaxStacks != 5 {
		t.Errorf("max stacks = %d, want the row's 5", aura.MaxStacks)
	}
}

// The game files state a permanent aura as -1, and its charges in the field the stacks share.
func TestAuraConfigPermanentRowAndCharges(t *testing.T) {
	withRows(t, resolverRows())
	aura := AuraConfig(Find(900))

	if aura.Duration != core.NeverExpires {
		t.Errorf("duration = %v, want NeverExpires", aura.Duration)
	}
	if aura.MaxStacks != 3 {
		t.Errorf("max stacks = %d, want the row's 3 charges", aura.MaxStacks)
	}
}

func TestAuraConfigOptions(t *testing.T) {
	withRows(t, resolverRows())
	aura := AuraConfig(Find(1000), Label("Tickless Aura - Debuff"), Permanent())

	if aura.Label != "Tickless Aura - Debuff" {
		t.Errorf("label = %q, want the caller's", aura.Label)
	}
	if aura.Duration != core.NeverExpires {
		t.Errorf("duration = %v, want NeverExpires", aura.Duration)
	}
	if aura.OnReset == nil {
		t.Error("a permanent aura got no OnReset to activate it")
	}
}

func TestDotConfigFromARow(t *testing.T) {
	withRows(t, resolverRows())
	bleed := Find(800)
	dot := DotConfig(bleed, bleed.PeriodicEffect())

	if dot.TickLength != time.Second*3 {
		t.Errorf("tick length = %v, want 3s", dot.TickLength)
	}
	if dot.NumberOfTicks != 7 {
		t.Errorf("ticks = %d, want 21s over 3s ticks", dot.NumberOfTicks)
	}
	if dot.BonusCoefficient != 0.1 {
		t.Errorf("bonus coefficient = %v, want the effect's 0.1", dot.BonusCoefficient)
	}
	if dot.Aura.Duration != time.Second*21 {
		t.Errorf("aura duration = %v, want 21s", dot.Aura.Duration)
	}
	if dot.Aura.Label != "Bleed" {
		t.Errorf("aura label = %q, want the spell's name", dot.Aura.Label)
	}
	if dot.OnSnapshot != nil {
		t.Error("OnSnapshot = non-nil, want nil")
	}
	if dot.OnTick == nil {
		t.Error("OnTick = nil, want the default tick")
	}
}

// The callback is a field, so a caller that needs its own replaces it.
func TestDotConfigCallerReplacesOnTick(t *testing.T) {
	withRows(t, resolverRows())
	bleed := Find(800)
	config := DotConfig(bleed, bleed.PeriodicEffect())

	ticks := 0
	config.OnTick = func(*core.Simulation, *core.Unit, *core.Dot) { ticks++ }
	config.OnTick(nil, nil, nil)

	if ticks != 1 {
		t.Errorf("the caller's OnTick ran %d times, want once", ticks)
	}
}

func TestDotConfigTakesAuraOptions(t *testing.T) {
	withRows(t, resolverRows())
	bleed := Find(800)

	if got := DotConfig(bleed, bleed.PeriodicEffect(), Label("Bleed - Off-hand")).Aura.Label; got != "Bleed - Off-hand" {
		t.Errorf("aura label = %q, want the caller's", got)
	}
}

// An effect that does not tick and a spell with no duration cannot answer a tick count.
func TestDotConfigPanicsWithoutAPeriodOrADuration(t *testing.T) {
	withRows(t, resolverRows())

	tickless := Find(1000)
	requirePanic(t, "states no tick period", func() {
		DotConfig(tickless, tickless.EffectN(1))
	})

	permanent := Find(900)
	requirePanic(t, "states no duration to tick over", func() {
		DotConfig(permanent, Find(800).PeriodicEffect())
	})
}
