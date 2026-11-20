package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/stats"
)

// The game's proc eligibility rules, as AttachProcTriggerCallback applies them before any mask
// or outcome check. Rows are the hitting spell, columns the listener.
func TestProcTriggerEligibility(t *testing.T) {
	triggers := []struct {
		name        string
		procMask    ProcMask
		flags       SpellFlag
		defenseType DefenseType
		shield      bool
	}{
		{name: "auto attack", procMask: ProcMaskMeleeMHAuto, defenseType: DefenseTypeMelee},
		{name: "ability", procMask: ProcMaskMeleeMHSpecial, defenseType: DefenseTypeMelee},
		{name: "proc", procMask: ProcMaskMeleeMHSpecial, flags: SpellFlagProc, defenseType: DefenseTypeMelee},
		{name: "proc suppressing weapon procs", procMask: ProcMaskMeleeMHSpecial, flags: SpellFlagProc | SpellFlagSuppressWeaponProcs, defenseType: DefenseTypeMelee},
		{name: "ability suppressing weapon procs", procMask: ProcMaskMeleeMHSpecial, flags: SpellFlagSuppressWeaponProcs, defenseType: DefenseTypeMelee},
		{name: "magic item proc's damage", procMask: ProcMaskEmpty, flags: SpellFlagProc, defenseType: DefenseTypeMagic},
		{name: "spell", procMask: ProcMaskSpellDamage, defenseType: DefenseTypeMagic},
		{name: "ability requiring a shield", procMask: ProcMaskMeleeMHSpecial, defenseType: DefenseTypeMelee, shield: true},
		{name: "melee ability with magic DefenseType", procMask: ProcMaskMeleeMHSpecial, defenseType: DefenseTypeMagic},
		{name: "ranged ability with magic DefenseType", procMask: ProcMaskRangedSpecial, defenseType: DefenseTypeMagic},
		{name: "wand shot", procMask: ProcMaskRangedAuto, defenseType: DefenseTypeMagic},
		{name: "melee ability with no DefenseType", procMask: ProcMaskMeleeMHSpecial},
		{name: "auto attack with no DefenseType", procMask: ProcMaskMeleeMHAuto},
	}
	listeners := []struct {
		name   string
		config ProcTrigger
		// One expectation per trigger above.
		want []bool
	}{
		{"aura", ProcTrigger{}, []bool{true, true, false, false, true, false, true, true, true, true, true, true, true}},
		{"aura that can proc from procs", ProcTrigger{CanProcFromProcs: true}, []bool{true, true, true, true, true, true, true, true, true, true, true, true, true}},
		{"weapon proc", ProcTrigger{IsWeaponProc: true}, []bool{true, true, true, false, false, false, false, false, false, false, true, false, true}},
		{"aura marked as weapon proc", ProcTrigger{SpellFlagsExclude: SpellFlagSuppressWeaponProcs}, []bool{true, true, false, false, false, false, true, true, true, true, true, true, true}},
	}

	for _, listener := range listeners {
		if len(listener.want) != len(triggers) {
			t.Fatalf("%s states %d expectations for %d triggers", listener.name, len(listener.want), len(triggers))
		}
		for i, trigger := range triggers {
			unit := &Unit{}
			aura := &Aura{Unit: unit}
			spell := &Spell{ProcMask: trigger.procMask, Flags: trigger.flags, DefenseType: trigger.defenseType,
				CastRequirement: CastRequirement{Shield: trigger.shield}}
			result := &SpellResult{Outcome: OutcomeHit}
			fired := false

			config := listener.config
			config.Callback = CallbackOnSpellHitDealt
			config.Outcome = OutcomeLanded
			config.TriggerImmediately = true
			config.Handler = func(_ *Simulation, _ *Spell, _ *SpellResult) { fired = true }
			aura.AttachProcTriggerCallback(unit, config)
			aura.OnSpellHitDealt(aura, nil, spell, result)

			if fired != listener.want[i] {
				t.Errorf("%s hit by %s: fired=%v, want %v", listener.name, trigger.name, fired, listener.want[i])
			}
		}
	}
}

// Suppress Caster Procs keeps a spell's hits, ticks and heals from every proc trigger the caster hears
// them with, and Suppress Target Procs keeps its hits and ticks from every one the target hears them
// with (EV-0165). A cast completing is no hit, so neither keeps it from a trigger on the cast.
func TestSuppressedProcsSkipTheirSide(t *testing.T) {
	hooks := []struct {
		name     string
		callback AuraCallback
		side     SpellFlag
		hear     func(aura *Aura, spell *Spell, result *SpellResult)
	}{
		{"a hit dealt", CallbackOnSpellHitDealt, SpellFlagSuppressCasterProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnSpellHitDealt(a, nil, s, r) }},
		{"a tick dealt", CallbackOnPeriodicDamageDealt, SpellFlagSuppressCasterProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnPeriodicDamageDealt(a, nil, s, r) }},
		{"a heal dealt", CallbackOnHealDealt, SpellFlagSuppressCasterProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnHealDealt(a, nil, s, r) }},
		{"a heal tick dealt", CallbackOnPeriodicHealDealt, SpellFlagSuppressCasterProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnPeriodicHealDealt(a, nil, s, r) }},
		{"a cast result", CallbackOnCastResult, SpellFlagSuppressCasterProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnCastResult(a, nil, s, r) }},
		{"a hit taken", CallbackOnSpellHitTaken, SpellFlagSuppressTargetProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnSpellHitTaken(a, nil, s, r) }},
		{"a tick taken", CallbackOnPeriodicDamageTaken, SpellFlagSuppressTargetProcs, func(a *Aura, s *Spell, r *SpellResult) { a.OnPeriodicDamageTaken(a, nil, s, r) }},
		{"a cast completing", CallbackOnCastComplete, SpellFlagNone, func(a *Aura, s *Spell, _ *SpellResult) { a.OnCastComplete(a, nil, s) }},
	}

	for _, hook := range hooks {
		for _, flag := range []SpellFlag{SpellFlagNone, SpellFlagSuppressCasterProcs, SpellFlagSuppressTargetProcs} {
			unit := &Unit{}
			aura := &Aura{Unit: unit}
			fired := false
			aura.AttachProcTriggerCallback(unit, ProcTrigger{
				Callback:           hook.callback,
				TriggerImmediately: true,
				Handler:            func(_ *Simulation, _ *Spell, _ *SpellResult) { fired = true },
			})
			hook.hear(aura, &Spell{ProcMask: ProcMaskSpellDamage, Flags: flag}, &SpellResult{Outcome: OutcomeHit})

			if want := !flag.Matches(hook.side); fired != want {
				t.Errorf("a trigger on %s from a spell flagged %v: fired=%v, want %v", hook.name, flag, fired, want)
			}
		}
	}
}

// A suppressed hit still lands in full: the defender's rage from the damage it takes comes as it does
// from any hit, and only the proc triggers on the suppressed side miss it (EV-0165).
func TestSuppressedProcsStillLandTheHit(t *testing.T) {
	sim := setupFakeRageSimFrom(fakeRageSimRequest())
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	mob := sim.Encounter.ActiveTargetUnits[0]

	heard := map[*Unit]map[AuraCallback]int{&fw.Unit: {}, mob: {}}
	sim.Environment.State = Constructed
	for unit, counts := range heard {
		for _, callback := range []AuraCallback{CallbackOnSpellHitDealt, CallbackOnSpellHitTaken} {
			unit.MakeProcTriggerAura(ProcTrigger{
				Name:               "Listener " + callback.String(),
				Callback:           callback,
				TriggerImmediately: true,
				Handler:            func(_ *Simulation, _ *Spell, _ *SpellResult) { counts[callback]++ },
			}).Activate(sim)
		}
	}
	sim.Environment.State = Finalized

	fw.stats[stats.Armor] = 0
	for _, flag := range []SpellFlag{SpellFlagSuppressCasterProcs, SpellFlagSuppressTargetProcs} {
		for _, attacker := range []*Unit{&fw.Unit, mob} {
			defender := mob
			if attacker == mob {
				defender = &fw.Unit
			}
			for _, counts := range heard {
				clear(counts)
			}
			strike := physicalSpecial(attacker, ProcMaskMeleeMHSpecial)
			strike.Flags |= flag

			rageBefore := fw.CurrentRage()
			result := strike.calcDamageInternal(sim, defender, 1000, 1, false, strike.OutcomeAlwaysHitNoHitCounter)
			strike.DealDamage(sim, result)

			if result.Damage <= 0 {
				t.Errorf("a hit flagged %v dealt %v damage, want some", flag, result.Damage)
			}
			if got, want := heard[attacker][CallbackOnSpellHitDealt] == 1, flag != SpellFlagSuppressCasterProcs; got != want {
				t.Errorf("the attacker's trigger heard a hit flagged %v: %v, want %v", flag, got, want)
			}
			if got, want := heard[defender][CallbackOnSpellHitTaken] == 1, flag != SpellFlagSuppressTargetProcs; got != want {
				t.Errorf("the defender's trigger heard a hit flagged %v: %v, want %v", flag, got, want)
			}
			if defender == &fw.Unit && fw.CurrentRage() <= rageBefore {
				t.Errorf("a hit flagged %v gave the defender no rage", flag)
			}
		}
	}
}
