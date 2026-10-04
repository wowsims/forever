package druid

func (druid *Druid) registerFeralCombatTalents() {
	// Tier 1
	druid.applyFerocity()
	druid.applyHeartOfTheWild()

	// Tier 2
	druid.applyFeralSwiftness()
	druid.applyFeralInstincts()
	druid.applyBrutalImpact()
	druid.applyThickHide()

	// Tier 3
	druid.applySavageFury()
	druid.applyFeralCharge()
	druid.applySharpenedClaws()

	druid.applyShreddingAttacks()

	// Tier 4
	druid.applyShiftingPower()
	// Primal Bite implemented in mangle.go
	druid.applyPredatoryStrikes()
	druid.applyBloodFrenzy()

	// Tier 5
	druid.applyImprovedShiftingPower()
	// Leader of the Pack implemented in druid.go
	druid.applyPredatoryInstincts()

	// Tier 6
	druid.applyNaturalReaction()
	druid.applyRendAndTear()

	// Tier 7
	druid.applyBerserk()
}

// TODO: To be implemented.
// applyThickHide increases armor contribution from items by 4/7/10% (ranks 1/2/3).
// Applies to both Armor and BonusArmor equip stats
func (druid *Druid) applyThickHide() {
	if druid.Talents.ThickHide == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.ThickHide == 0 {
	// 	return
	// }
	//
	// bonusByRank := [3]float64{0.04, 0.07, 0.10}
	// bonus := bonusByRank[druid.Talents.ThickHide-1]
	// druid.ApplyEquipScaling(stats.Armor, 1.0+bonus)
	// druid.ApplyEquipScaling(stats.BonusArmor, 1.0+bonus)
}

// Predatory Instincts: +2% melee critical strike damage per rank while in Cat or Bear form.
// The aura in the game files (33859 and its ranks) is school-masked to Physical, so it covers every
// physical attack in form (abilities and auto attacks, Ravage included) and nothing else.
//
// TODO: To be implemented. Forever's regenerated aura enum no longer carries the crit damage
// aura this talent's rank data used (A_MOD_CRIT_DAMAGE_BONUS is gone from the auto-generated
// table), and this talent has no other effect to fall back on, so it is fully disabled.
func (druid *Druid) applyPredatoryInstincts() {
	if druid.Talents.PredatoryInstincts == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.PredatoryInstincts == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CritMultiplier_Pct,
	// 	School:     core.SpellSchoolPhysical,
	// 	FloatValue: spellData.PredatoryInstincts.Effect(dbcenums.A_MOD_CRIT_DAMAGE_BONUS, 1).FractionAt(druid.Talents.PredatoryInstincts),
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.HeartOfTheWild == 0 {
	// 	return
	// }
	//
	// // +2% Intellect per rank (all forms, always active). The stat aura carries misc 0 in the
	// // Forever data, as every stat-percent effect does, so the stat is the call site's choice.
	// // The Cat/Bear form-specific bonuses are handled dynamically in RegisterCatFormAura /
	// // RegisterBearFormAura.
	// druid.MultiplyStat(stats.Intellect, spellData.HeartOfTheWild.Effect(dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE, 0).MultiplierAt(druid.Talents.HeartOfTheWild))
}

// TODO: To be implemented.
func (druid *Druid) applySharpenedClaws() {
	if druid.Talents.SharpenedClaws == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.SharpenedClaws == 0 {
	// 	return
	// }
	// bonus := stats.Stats{
	// 	stats.PhysicalCritPercent: float64(druid.Talents.SharpenedClaws) * 2.0,
	// }
	// druid.CatFormAura.AttachStatsBuff(bonus)
	// druid.BearFormAura.AttachStatsBuff(bonus)
}

// TODO: To be implemented.
func (druid *Druid) applyFeralSwiftness() {
	if druid.Talents.FeralSwiftness == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.FeralSwiftness == 0 {
	// 	return
	// }
	// bonus := stats.Stats{
	// 	stats.DodgeRating: core.DodgeRatingPerDodgePercent * 2.0 * float64(druid.Talents.FeralSwiftness),
	// }
	// druid.CatFormAura.AttachStatsBuff(bonus)
	// druid.BearFormAura.AttachStatsBuff(bonus)
}

// TODO: To be implemented.
func (druid *Druid) applyPredatoryStrikes() {
	if druid.Talents.PredatoryStrikes == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.PredatoryStrikes == 0 {
	// 	return
	// }
	// bonus := stats.Stats{
	// 	stats.AttackPower: float64(druid.Talents.PredatoryStrikes) * 0.5 * core.CharacterLevel,
	// }
	// druid.CatFormAura.AttachStatsBuff(bonus)
	// druid.BearFormAura.AttachStatsBuff(bonus)
}

// TODO: To be implemented.
// applyFuror gives a 20% chance per rank to gain 40 energy when shifting into
// Cat Form, or 10 rage when shifting into Bear Form.
func (druid *Druid) applyFerocity() {
	if druid.Talents.Ferocity == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.Ferocity == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: DruidSpellRake,
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	IntValue:  -druid.Talents.Ferocity,
	// })
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: DruidSpellMangleBear | DruidSpellMaul | DruidSpellSwipe,
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	IntValue:  -druid.Talents.Ferocity,
	// })
}

// TODO: To be implemented.
func (druid *Druid) applySavageFury() {
	if druid.Talents.SavageFury == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.SavageFury == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  DruidSpellRake,
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.SavageFury.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(druid.Talents.SavageFury),
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyShreddingAttacks() {
	if druid.Talents.ShreddingAttacks == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.ShreddingAttacks == 0 {
	// 	return
	// }
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: DruidSpellShred,
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	IntValue:  -9 * druid.Talents.ShreddingAttacks,
	// })
	//
	// druid.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: DruidSpellLacerate,
	// 	Kind:      core.SpellMod_PowerCost_Flat,
	// 	IntValue:  -druid.Talents.ShreddingAttacks,
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyBloodFrenzy() {
	if druid.Talents.BloodFrenzy == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.PrimalFury == 0 {
	// 	return
	// }
	//
	// procChance := 0.5 * float64(druid.Talents.PrimalFury)
	// actionID := core.ActionID{SpellID: 37117}
	// rageMetrics := druid.NewRageMetrics(actionID)
	// cpMetrics := druid.NewComboPointMetrics(actionID)
	//
	// // Cat form: +1 combo point on builder crits.
	// druid.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Primal Fury (Cat)",
	// 	ActionID:       actionID,
	// 	Callback:       core.CallbackOnSpellHitDealt,
	// 	ClassSpellMask: DruidSpellBuilder,
	// 	Outcome:        core.OutcomeCrit,
	// 	ProcChance:     procChance,
	// 	ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
	// 		return druid.InForm(Cat)
	// 	},
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		druid.AddComboPoints(sim, 1, cpMetrics)
	// 	},
	// })
	//
	// // Bear form: +5 rage on auto-attack crits.
	// druid.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:       "Primal Fury (Bear)",
	// 	ActionID:   actionID,
	// 	Callback:   core.CallbackOnSpellHitDealt,
	// 	ProcMask:   core.ProcMaskMelee,
	// 	Outcome:    core.OutcomeCrit,
	// 	ProcChance: procChance,
	// 	ExtraCondition: func(_ *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
	// 		return druid.InForm(Bear)
	// 	},
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		druid.AddRage(sim, 5, rageMetrics)
	// 	},
	// })
}

// TODO: To be implemented.
func (druid *Druid) applyFeralInstincts() {
	if druid.Talents.FeralInstinct == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if druid.Talents.FeralInstinct == 0 {
	// 	return
	// }
	//
	// // TODO: Forever repurposes Feral Instinct: the Dire Bear threat bonus is gone and the
	// // spell now carries SPELLMOD_EFFECT1 +5/10/15 and SPELLMOD_DAMAGE +10/20/30% against an
	// // unknown spell, so the threat multiplier is pinned to the untalented 1.0.
	// threatMultiplier := 1.0
	// druid.BearFormAura.AttachMultiplicativePseudoStatBuff(&druid.PseudoStats.ThreatMultiplier, threatMultiplier)
}

// applyBrutalImpact implements Brutal Impact, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyBrutalImpact() {
	if druid.Talents.BrutalImpact == 0 {
		return
	}
}

// applyFeralCharge implements Feral Charge, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyFeralCharge() {
	if !druid.Talents.FeralCharge {
		return
	}
}

// applyShiftingPower implements Shifting Power, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyShiftingPower() {
	if !druid.Talents.ShiftingPower {
		return
	}
}

// applyImprovedShiftingPower implements Improved Shifting Power, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedShiftingPower() {
	if druid.Talents.ImprovedShiftingPower == 0 {
		return
	}
}

// applyNaturalReaction implements Natural Reaction, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyNaturalReaction() {
	if druid.Talents.NaturalReaction == 0 {
		return
	}
}

// applyRendAndTear implements Rend and Tear, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyRendAndTear() {
	if druid.Talents.RendAndTear == 0 {
		return
	}
}

// applyBerserk implements Berserk, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyBerserk() {
	if !druid.Talents.Berserk {
		return
	}
}
