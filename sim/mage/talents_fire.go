package mage

func (mage *Mage) registerFireTalents() {
	// Tier 1
	mage.registerWakeOfFire()
	mage.registerIncineration()
	mage.registerImprovedFireball()

	// Tier 2
	mage.registerIgnite()
	mage.registerFlameThrowing()
	mage.registerImpact()

	// Tier 3
	mage.registerBurningSoul()
	mage.registerImprovedFlamestrike()
	mage.registerPyroblastTalent()

	// Tier 4
	// Improved Scorch implemented in scorch.go
	mage.registerImprovedFireWard()
	mage.registerHeatingUp()
	mage.registerMasterOfElements()

	// Tier 5
	mage.registerCriticalMass()
	// Blast Wave implemented in blast_wave.go; registered unconditionally
	// from registerSpells (its own Talents.BlastWave guard is inside that file).

	// Tier 6
	mage.registerFirePower()

	// Tier 7
	// Combustion implemented in combustion.go; registered unconditionally
	// from registerSpells (its own Talents.Combustion guard is inside that file).
}

// registerWakeOfFire implements Wake of Fire, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerWakeOfFire() {
	if mage.Talents.WakeOfFire == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerIncineration() {
	if mage.Talents.Incineration == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.Incineration == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellFireBlast | MageSpellScorch,
	// 	FloatValue: spellData.Incineration.ValueAt(mage.Talents.Incineration),
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerImprovedFireball() {
	if mage.Talents.ImprovedFireball == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ImprovedFireball == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask: MageSpellFireball,
	// 	TimeValue: time.Millisecond * time.Duration(spellData.ImprovedFireball.Effect(dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_CASTING_TIME)).ValueAt(mage.Talents.ImprovedFireball)),
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerIgnite() {
	if mage.Talents.Ignite == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.Ignite == 0 {
	// 	return
	// }
	// igniteDamageMultiplier := float64(mage.Talents.Ignite) * .08
	//
	// igniteSpell := mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:         core.ActionID{SpellID: 12846},
	// 	SpellSchool:      core.SpellSchoolFire,
	// 	ProcMask:         core.ProcMaskSpellDamage,
	// 	ClassSpellMask:   MageSpellIgnite,
	// 	ClassFlags:       spellData.IgniteTriggered.Highest().ClassFlags,
	// 	Flags:            core.SpellFlagIgnoreModifiers | core.SpellFlagNoSpellMods | core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreResists | core.SpellFlagProc,
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label:     "Ignite",
	// 			Tag:       "IgniteDot",
	// 			MaxStacks: math.MaxInt32,
	// 		},
	// 		NumberOfTicks: 2,
	// 		TickLength:    2 * time.Second,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			dot.Spell.CalcAndDealPeriodicDamage(sim, target, dot.SnapshotBaseDamage, dot.OutcomeTick)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		spell.Dot(target).Apply(sim)
	// 	},
	// })
	//
	// refreshIgnite := func(sim *core.Simulation, target *core.Unit, damagePerTick float64) {
	// 	dot := igniteSpell.Dot(target)
	// 	igniteSpell.Cast(sim, target)
	// 	dot.SnapshotBaseDamage = damagePerTick
	// 	dot.Aura.SetStacks(sim, int32(dot.SnapshotBaseDamage))
	// }
	//
	// procTrigger := core.ProcTrigger{
	// 	Name:               "Ignite Talent",
	// 	CanProcFromProcs:   true, // 11119, 11120, 12846-12848 carry the bit.
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	ProcMask:           core.ProcMaskSpellDamage,
	// 	ClassSpellMask:     FireSpellIgnitable,
	// 	Outcome:            core.OutcomeCrit,
	// 	TriggerImmediately: true,
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
	// 		target := result.Target
	// 		dot := igniteSpell.Dot(target)
	// 		outstandingDamage := dot.OutstandingDmg()
	// 		if dot.RemainingTicks() <= 0 {
	// 			outstandingDamage = 0
	// 		}
	//
	// 		newDamage := result.Damage * igniteDamageMultiplier
	// 		totalDamage := outstandingDamage + newDamage
	// 		damagePerTick := totalDamage / float64(dot.BaseTickCount)
	//
	// 		refreshIgnite(sim, target, damagePerTick)
	// 	},
	// }
	// igniteSpell.Unit.MakeProcTriggerAura(procTrigger)
	// mage.Ignite = igniteSpell
	//
	// // This is needed because we want to listen for the spell "cast" event that refreshes the Dot
	// mage.Ignite.Flags ^= core.SpellFlagNoOnCastComplete
}

// registerFlameThrowing implements Flame Throwing, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerFlameThrowing() {
	if mage.Talents.FlameThrowing == 0 {
		return
	}
}

// registerImpact implements Impact, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImpact() {
	if mage.Talents.Impact == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerBurningSoul() {
	if mage.Talents.BurningSoul == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.BurningSoul == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolFire,
	// 	FloatValue: -.05 * float64(mage.Talents.BurningSoul),
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerImprovedFlamestrike() {
	if mage.Talents.ImprovedFlamestrike == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ImprovedFlamestrike == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellFlamestrike,
	// 	FloatValue: spellData.ImprovedFlamestrike.ValueAt(mage.Talents.ImprovedFlamestrike),
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// })
}

// registerPyroblastTalent implements the Forever talent gate for Pyroblast.
//
// TODO: Forever makes Pyroblast a talent; the baseline registration in
// pyroblast.go (registerPyroblastSpell, called unconditionally from
// registerSpells) is untouched and should be gated on this field.
func (mage *Mage) registerPyroblastTalent() {
	if !mage.Talents.Pyroblast {
		return
	}
}

// registerImprovedFireWard implements Improved Fire Ward, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImprovedFireWard() {
	if mage.Talents.ImprovedFireWard == 0 {
		return
	}
}

// registerHeatingUp implements Heating Up, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerHeatingUp() {
	if !mage.Talents.HeatingUp {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerMasterOfElements() {
	if mage.Talents.MasterOfElements == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.MasterOfElements == 0 {
	// 	return
	// }
	//
	// refundCoeff := spellData.MasterOfElements.FractionAt(mage.Talents.MasterOfElements)
	// manaMetrics := mage.NewManaMetrics(core.ActionID{SpellID: 29076})
	//
	// mage.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Master of Elements",
	// 	Duration:       core.NeverExpires,
	// 	ClassSpellMask: MageSpellFire | MageSpellFrost,
	// 	Outcome:        core.OutcomeCrit,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.CurCast.Cost == 0 {
	// 			return
	// 		}
	// 		mage.AddMana(sim, spell.DefaultCast.Cost*refundCoeff, manaMetrics)
	// 	},
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerCriticalMass() {
	if mage.Talents.CriticalMass == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.CriticalMass == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	SpellFlag:  core.SpellFlag(core.SpellSchoolFire),
	// 	FloatValue: spellData.CriticalMass.ValueAt(mage.Talents.CriticalMass),
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// })
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerFirePower() {
	if mage.Talents.FirePower == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.FirePower == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolFire,
	// 	FloatValue: spellData.FirePower.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(mage.Talents.FirePower),
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// })
}

// ------ FROST TALENTS ------
