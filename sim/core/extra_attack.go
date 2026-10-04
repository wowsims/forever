package core

// Adds an extra attack to a proc's handler: a copy of the main-hand swing tagged with the granting
// spell, cast count times after whatever the handler already does, on the hit's target - or on the
// attacker, when the hit was one the unit took. The returned OnInit registers the copy, so it goes on
// the proc's aura: the main-hand swing is configured after procs register, and a unit that never
// swings has none to copy, in which case the proc casts nothing. The swing it grants is not a hit the
// proc hears, or one proc would swing forever.
func (unit *Unit) ExtraAttackProc(trigger *ProcTrigger, tagSpellID int32, count int) OnInit {
	var extraAttack *Spell
	condition := trigger.ExtraCondition
	trigger.ExtraCondition = func(sim *Simulation, spell *Spell, result *SpellResult) bool {
		return spell != extraAttack && (condition == nil || condition(sim, spell, result))
	}
	handler := trigger.Handler
	trigger.Handler = func(sim *Simulation, spell *Spell, result *SpellResult) {
		if handler != nil {
			handler(sim, spell, result)
		}
		if extraAttack == nil {
			return
		}
		target := result.Target
		if target == unit {
			target = spell.Unit
		}
		for range count {
			unit.AutoAttacks.MaybeReplaceMHSwing(sim, extraAttack).Cast(sim, target)
		}
	}
	trigger.TriggerImmediately = true

	return func(_ *Aura, _ *Simulation) {
		if unit.AutoAttacks.MHAuto() == nil {
			return
		}
		swing := *unit.AutoAttacks.MHConfig()
		swing.ActionID = swing.ActionID.WithTag(tagSpellID)
		extraAttack = unit.GetOrRegisterSpell(swing)
	}
}
