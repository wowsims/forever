package core

import "testing"

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
		{"auto attack", ProcMaskMeleeMHAuto, SpellFlagNone, DefenseTypeMelee, false},
		{"ability", ProcMaskMeleeMHSpecial, SpellFlagNone, DefenseTypeMelee, false},
		{"proc", ProcMaskMeleeMHSpecial, SpellFlagProc, DefenseTypeMelee, false},
		{"proc suppressing weapon procs", ProcMaskMeleeMHSpecial, SpellFlagProc | SpellFlagSuppressWeaponProcs, DefenseTypeMelee, false},
		{"ability suppressing weapon procs", ProcMaskMeleeMHSpecial, SpellFlagSuppressWeaponProcs, DefenseTypeMelee, false},
		{"magic item proc's damage", ProcMaskEmpty, SpellFlagProc, DefenseTypeMagic, false},
		{"spell", ProcMaskSpellDamage, SpellFlagNone, DefenseTypeMagic, false},
		{"ability requiring a shield", ProcMaskMeleeMHSpecial, SpellFlagNone, DefenseTypeMelee, true},
		{"melee ability with magic DefenseType", ProcMaskMeleeMHSpecial, SpellFlagNone, DefenseTypeMagic, false},
		{"ranged ability with magic DefenseType", ProcMaskRangedSpecial, SpellFlagNone, DefenseTypeMagic, false},
		{"wand shot", ProcMaskRangedAuto, SpellFlagNone, DefenseTypeMagic, false},
	}
	listeners := []struct {
		name   string
		config ProcTrigger
		// One expectation per trigger above.
		want [11]bool
	}{
		{"aura", ProcTrigger{}, [11]bool{true, true, false, false, true, false, true, true, true, true, true}},
		{"aura that can proc from procs", ProcTrigger{CanProcFromProcs: true}, [11]bool{true, true, true, true, true, true, true, true, true, true, true}},
		{"weapon proc", ProcTrigger{IsWeaponProc: true}, [11]bool{true, true, true, false, false, false, false, false, false, false, true}},
		{"aura marked as weapon proc", ProcTrigger{SpellFlagsExclude: SpellFlagSuppressWeaponProcs}, [11]bool{true, true, false, false, false, false, true, true, true, true, true}},
	}

	for _, listener := range listeners {
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
