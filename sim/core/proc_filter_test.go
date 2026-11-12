package core

import "testing"

// The game's proc eligibility rules, as AttachProcTriggerCallback applies them before any mask
// or outcome check. Rows are the hitting spell, columns the listener.
func TestProcTriggerEligibility(t *testing.T) {
	triggers := []struct {
		name     string
		procMask ProcMask
		flags    SpellFlag
	}{
		{"auto attack", ProcMaskMeleeMHAuto, SpellFlagNone},
		{"ability", ProcMaskMeleeMHSpecial, SpellFlagNone},
		{"proc", ProcMaskMeleeMHSpecial, SpellFlagProc},
		{"proc suppressing weapon procs", ProcMaskMeleeMHSpecial, SpellFlagProc | SpellFlagSuppressWeaponProcs},
		{"ability suppressing weapon procs", ProcMaskMeleeMHSpecial, SpellFlagSuppressWeaponProcs},
		{"magic item proc's damage", ProcMaskEmpty, SpellFlagProc},
		{"spell", ProcMaskSpellDamage, SpellFlagNone},
		{"ability no chance-on-hit proc hears", ProcMaskMeleeMHSpecial, SpellFlagNoChanceOnHit},
	}
	listeners := []struct {
		name   string
		config ProcTrigger
		// One expectation per trigger above.
		want [8]bool
	}{
		{"aura", ProcTrigger{}, [8]bool{true, true, false, false, true, false, true, true}},
		{"aura that can proc from procs", ProcTrigger{CanProcFromProcs: true}, [8]bool{true, true, true, true, true, true, true, true}},
		{"weapon proc", ProcTrigger{IsWeaponProc: true}, [8]bool{true, true, true, false, false, false, false, false}},
		{"aura marked as weapon proc", ProcTrigger{SpellFlagsExclude: SpellFlagSuppressWeaponProcs}, [8]bool{true, true, false, false, false, false, true, true}},
	}

	for _, listener := range listeners {
		for i, trigger := range triggers {
			unit := &Unit{}
			aura := &Aura{Unit: unit}
			spell := &Spell{ProcMask: trigger.procMask, Flags: trigger.flags}
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
