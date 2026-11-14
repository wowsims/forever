package core

// Adds an extra attack to a proc's handler: a copy of the main-hand swing tagged with the granting
// spell, cast count times on the hit's target, or on the attacker when the hit was one the unit
// took. The proc rolls off the triggering hit, but the extra attack swings once that hit has been
// dealt, in the same instant, and is lost if its target died before it swings. The returned OnInit
// registers the copy, so it goes on the proc's aura: the main-hand swing is configured after procs
// register, and a unit that never swings has none to copy, in which case the proc casts nothing. The
// swing it grants is not a hit the proc hears, or one proc would swing forever. Windfury Totem's
// strike is one it hears: per clean strike against per white swing, Hack and Slash 60 of 2,208
// (2.7%, white 1.5%), Thrash Blade 30 of 1,003 (3.0%, white 1.3%, or 2.5% with no off-hand weapon),
// Sword Specialization 8 of 198 (4.0%, white 4.8%) and Hand of Justice 4 of 272 (1.5%, white 1.8%).
// TODO: Ingame test needed: why Hack and Slash and Thrash Blade come off the strike about twice as often as off a white swing.
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
		sim.RunAfterCurrentEvent(func(sim *Simulation) {
			// The swing's cast checks no target, so the check is here.
			if !target.IsEnabled() {
				return
			}
			mainHand := unit.AutoAttacks.MHAuto()
			extraAttack.SetSpellSchool(mainHand.SpellSchool)
			extraAttack.BonusCoefficient = mainHand.BonusCoefficient
			for range count {
				spell := unit.AutoAttacks.MaybeReplaceMHSwing(sim, extraAttack)
				// Make sure to check range, since Judgement can proc WF
				if spell.InRange() {
					spell.Cast(sim, target)
				}
			}
		})
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
