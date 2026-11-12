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
	}
	listeners := []struct {
		name   string
		config ProcTrigger
		// One expectation per trigger above.
		want []bool
	}{
		{"aura", ProcTrigger{}, []bool{true, true, false, false, true, false, true, true, true, true, true}},
		{"aura that can proc from procs", ProcTrigger{CanProcFromProcs: true}, []bool{true, true, true, true, true, true, true, true, true, true, true}},
		{"weapon proc", ProcTrigger{IsWeaponProc: true}, []bool{true, true, true, false, false, false, false, false, false, false, true}},
		{"aura marked as weapon proc", ProcTrigger{SpellFlagsExclude: SpellFlagSuppressWeaponProcs}, []bool{true, true, false, false, false, false, true, true, true, true, true}},
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
