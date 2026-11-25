package core

import (
	"testing"
	"time"
)

func TestPushBackNeverLeavesMoreThanTheCastTime(t *testing.T) {
	hc := Hardcast{Expires: time.Second, CastTime: time.Second}

	steps := []struct {
		at, moved, expires time.Duration
	}{
		{at: 10 * time.Millisecond, moved: 10 * time.Millisecond, expires: 1010 * time.Millisecond},
		{at: 20 * time.Millisecond, moved: 10 * time.Millisecond, expires: 1020 * time.Millisecond},
		{at: 20 * time.Millisecond, moved: 0, expires: 1020 * time.Millisecond},
		{at: 800 * time.Millisecond, moved: SpellPushbackDuration, expires: 1520 * time.Millisecond},
		{at: 900 * time.Millisecond, moved: 380 * time.Millisecond, expires: 1900 * time.Millisecond},
	}
	for _, step := range steps {
		if moved := hc.pushBack(step.at); moved != step.moved || hc.Expires != step.expires {
			t.Fatalf("hit at %v: moved %v to %v, want %v to %v", step.at, moved, hc.Expires, step.moved, step.expires)
		}
	}
}

// A caster who tanks nothing, one second into the fight, with a hardcast that can be pushed back and an
// enemy spell that hits the caster. The count is how often the hardcast has landed.
func setupPushbackSim(t *testing.T) (*Simulation, *FakeSwingCaster, *Spell, *Spell, *int) {
	t.Helper()
	sim, fc := setupSwingResetSim(t)
	landed := 0
	cast := fc.RegisterSpell(SpellConfig{
		ActionID:       ActionID{SpellID: 910500},
		SpellSchool:    SpellSchoolHoly,
		ProcMask:       ProcMaskEmpty,
		Flags:          SpellFlagPushback | SpellFlagNoMetrics,
		ClassSpellMask: 1,
		Cast:           CastConfig{DefaultCast: Cast{CastTime: swingResetCastTime}},
		ApplyEffects:   func(*Simulation, *Unit, *Spell) { landed++ },
	})
	cast.finalize()
	hit := sim.Encounter.AllTargetUnits[0].RegisterSpell(SpellConfig{
		ActionID:         ActionID{SpellID: 910501},
		SpellSchool:      SpellSchoolFire,
		DefenseType:      DefenseTypeMagic,
		ProcMask:         ProcMaskSpellDamage,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	})
	hit.finalize()
	return sim, fc, cast, hit, &landed
}

// A hit pushes back the hardcast of a caster who tanks nothing, unless a mod protects the spell in full.
func TestHitsPushBackACastersHardcast(t *testing.T) {
	for _, tc := range []struct {
		name       string
		protection float64
		pushedBack time.Duration
	}{
		{name: "without protection the cast end moves", protection: 0, pushedBack: SpellPushbackDuration},
		{name: "full protection keeps the cast end", protection: 1, pushedBack: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, fc, cast, hit, _ := setupPushbackSim(t)
			if tc.protection > 0 {
				fc.AddStaticMod(SpellModConfig{Kind: SpellMod_PushbackProtection_Pct, ClassMask: 1, FloatValue: tc.protection})
			}
			if cast.PushbackProtection != tc.protection {
				t.Fatalf("test setup: protection %v, want %v", cast.PushbackProtection, tc.protection)
			}

			cast.Cast(sim, fc.CurrentTarget)
			castEnd := fc.Hardcast.Expires
			stepSwingSim(t, sim, sim.CurrentTime+swingResetCastTime/2)
			hit.Cast(sim, &fc.Unit)
			stepSwingSim(t, sim, sim.CurrentTime+2*SpellBatchWindow)

			if got, want := fc.Hardcast.Expires, castEnd+tc.pushedBack; got != want {
				t.Errorf("cast end %v, want %v", got, want)
			}
		})
	}
}

// A hit inside the last spell batch window of a hardcast is handled once the cast is done, so the cast
// lands once.
func TestAHitAtTheCastEndLeavesTheCastToLandOnce(t *testing.T) {
	sim, fc, cast, hit, landed := setupPushbackSim(t)
	cast.Cast(sim, fc.CurrentTarget)
	castEnd := fc.Hardcast.Expires
	stepSwingSim(t, sim, castEnd-SpellBatchWindow/2)
	hit.Cast(sim, &fc.Unit)
	stepSwingSim(t, sim, castEnd+time.Second)

	if *landed != 1 {
		t.Errorf("the cast landed %d times, want once", *landed)
	}
}

// A hit in the last spell batch window of a hardcast is handled after the next hardcast has begun, and
// leaves that cast's end where it is.
func TestAHitAtTheCastEndLeavesTheNextCastAlone(t *testing.T) {
	sim, fc, cast, hit, _ := setupPushbackSim(t)
	cast.Cast(sim, fc.CurrentTarget)
	castEnd := fc.Hardcast.Expires
	stepSwingSim(t, sim, castEnd-SpellBatchWindow/2)
	hit.Cast(sim, &fc.Unit)
	stepSwingSim(t, sim, castEnd+SpellBatchWindow/4)
	if !cast.Cast(sim, fc.CurrentTarget) {
		t.Fatalf("test setup: the next cast did not begin at %v", sim.CurrentTime)
	}
	nextEnd := fc.Hardcast.Expires
	stepSwingSim(t, sim, castEnd+2*SpellBatchWindow)

	if got := fc.Hardcast.Expires; got != nextEnd {
		t.Errorf("next cast end %v, want %v", got, nextEnd)
	}
}
