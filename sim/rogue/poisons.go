package rogue

import (
	"github.com/wowsims/forever/sim/core"
)

func (rogue *Rogue) applyPoisons() {
	if rogue.Consumables.MhWeaponImbueId == 0 && rogue.Consumables.OhWeaponImbueId == 0 {
		return
	}
	rogue.applyDeadlyPoison()
	rogue.applyWoundPoison()
	rogue.applyInstantPoison()
}

// TODO: To be implemented. Deadly Poison pins spell 27187 directly in the TBC body; the implementation
// needs review before it's uncommented.
func (rogue *Rogue) registerDeadlyPoisonSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(deadlyImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// rogue.DeadlyPoison = rogue.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 27187},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamageProc,
	// 	ClassSpellMask: RogueSpellDeadlyPoison,
	// 	ClassFlags:     spellData.DeadlyPoison.Highest().ClassFlags,
	// 	Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell | core.SpellFlagProc,
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label:     "Deadly Poison",
	// 			MaxStacks: 5,
	// 			Duration:  time.Second * 12,
	// 		},
	// 		NumberOfTicks: 4,
	// 		TickLength:    time.Second * 3,
	//
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			base := 45.0 * float64(dot.GetStacks())
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, base, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if !result.Landed() {
	// 			return
	// 		}
	//
	// 		dot := spell.Dot(target)
	// 		if dot.IsActive() {
	// 			dot.Refresh(sim)
	// 			dot.AddStack(sim)
	// 		} else {
	// 			dot.Apply(sim)
	// 			dot.SetStacks(sim, 1)
	// 		}
	// 	},
	// })
}

// TODO: To be implemented. Wound Poison pins spell 27189 directly in the TBC body; the implementation
// needs review before it's uncommented.
func (rogue *Rogue) registerWoundPoisonSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(woundImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// woundPoisonDebuffAura := core.Aura{
	// 	Label:     "Wound Poison",
	// 	ActionID:  core.ActionID{SpellID: 27189},
	// 	Duration:  time.Second * 15,
	// 	MaxStacks: 5,
	// 	// Wound Healing Debuff NYI
	// }
	//
	// rogue.WoundPoisonDebuffAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	return target.RegisterAura(woundPoisonDebuffAura)
	// })
	//
	// wpBaseDamage := 65.0
	//
	// wpConfig := core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 27189},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamageProc,
	// 	ClassSpellMask: RogueSpellWoundPoison,
	// 	ClassFlags:     spellData.WoundPoison.Highest().ClassFlags,
	// 	Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell | core.SpellFlagProc,
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealDamage(sim, target, wpBaseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		if result.Landed() {
	// 			rogue.WoundPoisonDebuffAuras.Get(target).Activate(sim)
	// 		}
	// 	},
	// }
	//
	// rogue.WoundPoison = rogue.RegisterSpell(wpConfig)
}

// TODO: To be implemented. Instant Poison pins spell 26890 directly in the TBC body; the implementation
// needs review before it's uncommented.
func (rogue *Rogue) registerInstantPoisonSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(instantImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// ipBaseDamage := 146.0
	// ipRange := 48
	//
	// ipConfig := core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 26890},
	// 	SpellSchool:    core.SpellSchoolNature,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamageProc,
	// 	ClassSpellMask: RogueSpellInstantPoison,
	// 	ClassFlags:     spellData.InstantPoison.Highest().ClassFlags,
	// 	Flags:          core.SpellFlagPoison | core.SpellFlagPassiveSpell | core.SpellFlagProc,
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	ThreatMultiplier:         1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.CalcAndDealDamage(sim, target, ipBaseDamage+sim.RandomFloat("Instant Poison")*float64(ipRange), spell.OutcomeMagicHitAndCrit)
	// 	},
	// }
	//
	// rogue.InstantPoison = rogue.RegisterSpell(ipConfig)
}

func (rogue *Rogue) getPoisonProcMask(poisonId int32) core.ProcMask {
	var mask core.ProcMask
	if rogue.Consumables.MhWeaponImbueId == poisonId {
		mask |= core.ProcMaskMeleeMH
	}
	if rogue.Consumables.OhWeaponImbueId == poisonId {
		mask |= core.ProcMaskMeleeOH
	}
	return mask
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) applyDeadlyPoison() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(deadlyImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// pph := 0.3 + spellData.ImprovedPoisons.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CHANCE_OF_SUCCESS)).FractionAt(rogue.Talents.ImprovedPoisons)
	// rogue.deadlyPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)
	//
	// rogue.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Deadly Poison",
	// 	Outcome:            core.OutcomeLanded,
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	TriggerImmediately: true,
	// 	ProcMask:           procMask,
	// 	IsWeaponProc:       true,
	// 	DPM:                rogue.deadlyPoisonPPHM,
	//
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		rogue.DeadlyPoison.Cast(sim, result.Target)
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) applyWoundPoison() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(woundImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// pph := 0.3 + spellData.ImprovedPoisons.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CHANCE_OF_SUCCESS)).FractionAt(rogue.Talents.ImprovedPoisons)
	// rogue.woundPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)
	//
	// rogue.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Wound Poison",
	// 	Outcome:            core.OutcomeLanded,
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	TriggerImmediately: true,
	// 	ProcMask:           procMask,
	// 	IsWeaponProc:       true,
	// 	DPM:                rogue.woundPoisonPPHM,
	//
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		rogue.WoundPoison.Cast(sim, result.Target)
	// 	},
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (rogue *Rogue) applyInstantPoison() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// procMask := rogue.getPoisonProcMask(instantImbueID)
	// if procMask == core.ProcMaskUnknown {
	// 	return
	// }
	// pph := 0.2 + spellData.ImprovedPoisons.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CHANCE_OF_SUCCESS)).FractionAt(rogue.Talents.ImprovedPoisons)
	// rogue.instantPoisonPPHM = rogue.NewFixedProcChanceManager(pph, procMask)
	//
	// rogue.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Instant Poison",
	// 	Outcome:            core.OutcomeLanded,
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	TriggerImmediately: true,
	// 	ProcMask:           procMask,
	// 	IsWeaponProc:       true,
	// 	DPM:                rogue.instantPoisonPPHM,
	//
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		rogue.InstantPoison.Cast(sim, result.Target)
	// 	},
	// })
}
