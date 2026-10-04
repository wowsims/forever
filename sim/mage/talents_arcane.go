package mage

func (mage *Mage) registerArcaneTalents() {
	// Tier 1
	mage.registerWandSpecialization()
	mage.registerArcaneFocus()
	mage.registerImprovedChanneling()

	// Tier 2
	mage.registerArcaneSubtlety()
	mage.registerMagicAbsorption()
	mage.registerArcaneConcentration()
	mage.registerArcaneResilience()

	// Tier 3
	mage.registerArcaneGeometry()
	mage.registerArcaneImpact()
	mage.registerArcaneBlastTalent()

	// Tier 4
	mage.registerArcaneShielding()
	mage.registerImprovedCounterspell()
	mage.registerArcaneMeditation()
	mage.registerMissileBarrage()

	// Tier 5
	// Presence of Mind implemented in presence_of_mind.go; registered unconditionally
	// from registerSpells (its own Talents.PresenceOfMind guard is inside that file).
	mage.registerArcaneMind()

	// Tier 6
	mage.registerArcaneInstability()

	// Tier 7
	// Arcane Power implemented in arcane_power.go; registered unconditionally
	// from registerSpells (its own Talents.ArcanePower guard is inside that file).
}

// registerWandSpecialization implements Wand Specialization, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerWandSpecialization() {
	if mage.Talents.WandSpecialization == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneFocus() {
	if mage.Talents.ArcaneFocus == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneFocus == 0 {
	// 	return
	// }
	//
	// mage.PseudoStats.SchoolBonusHitChance[stats.SchoolIndexArcane] += spellData.ArcaneFocus.ValueAt(mage.Talents.ArcaneFocus)
}

// registerImprovedChanneling implements Improved Channeling, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImprovedChanneling() {
	if mage.Talents.ImprovedChanneling == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneSubtlety() {
	if mage.Talents.ArcaneSubtlety == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneSubtlety == 0 {
	// 	return
	// }
	//
	// //all spells resist 5 & arcance spells threat 20% per rank
	// mage.AddStaticMod(core.SpellModConfig{
	// 	School:     core.SpellSchoolArcane,
	// 	FloatValue: -.20 * float64(mage.Talents.ArcaneSubtlety),
	// 	Kind:       core.SpellMod_ThreatMultiplier_Pct,
	// })
}

// registerMagicAbsorption implements Magic Absorption, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerMagicAbsorption() {
	if mage.Talents.MagicAbsorption == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below already accounts for Forever dropping Arcane Potency (bonusCrit pinned to 0, per the TODO inside); kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneConcentration() {
	if mage.Talents.ArcaneConcentration == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneConcentration == 0 {
	// 	return
	// }
	//
	// // TODO: Forever drops Arcane Potency; the Clearcasting crit bonus it fed is pinned to
	// // 0 until we know whether the effect moved onto another talent.
	// bonusCrit := 0.0
	// var proccedAt time.Duration
	// var proccedSpell *core.Spell
	//
	// mage.ClearCasting = mage.RegisterAura(core.Aura{
	// 	Label:    "Clearcasting",
	// 	ActionID: core.ActionID{SpellID: 12536},
	// 	Duration: time.Second * 15,
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		mage.AddStatDynamic(sim, stats.SpellCritPercent, bonusCrit)
	// 		aura.Unit.PseudoStats.SpellCostPercentModifier -= 100
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		mage.AddStatDynamic(sim, stats.SpellCritPercent, -bonusCrit)
	// 		aura.Unit.PseudoStats.SpellCostPercentModifier += 100
	// 	},
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.ClassSpellMask&MageSpellsAllDamaging == 0 {
	// 			return
	// 		}
	//
	// 		if spell.DefaultCast.Cost == 0 {
	// 			return
	// 		}
	//
	// 		if proccedAt == sim.CurrentTime && proccedSpell == spell {
	// 			// Means this is another hit from the same cast that procced CC.
	// 			return
	// 		}
	//
	// 		aura.Deactivate(sim)
	// 	},
	// })
	//
	// mage.RegisterAura(core.Aura{
	// 	Label:    "Arcane Concentration",
	// 	Duration: core.NeverExpires,
	// 	OnReset: func(aura *core.Aura, sim *core.Simulation) {
	// 		aura.Activate(sim)
	// 	},
	// 	OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.ClassSpellMask&MageSpellsAllDamaging == 0 {
	// 			return
	// 		}
	//
	// 		if !result.Landed() {
	// 			return
	// 		}
	//
	// 		// Forever states a flat SpellAuraOptions.ProcChance of 100 on the talent spell and puts the
	// 		// real per-rank chance on the effect, so ProcChanceAt would read 100% at every rank.
	// 		procChance := spellData.ArcaneConcentration.FractionAt(mage.Talents.ArcaneConcentration)
	// 		if sim.Proc(procChance, "Arcane Concentration") {
	// 			proccedAt = sim.CurrentTime
	// 			proccedSpell = spell
	// 			mage.ClearCasting.Activate(sim)
	// 		}
	// 	},
	// })
}

// registerArcaneResilience implements Arcane Resilience, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerArcaneResilience() {
	if mage.Talents.ArcaneResilience == 0 {
		return
	}
}

// registerArcaneGeometry implements Arcane Geometry, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerArcaneGeometry() {
	if mage.Talents.ArcaneGeometry == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneImpact() {
	if mage.Talents.ArcaneImpact == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneImpact == 0 {
	// 	return
	// }
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellArcaneBlast | MageSpellArcaneExplosion,
	// 	FloatValue: spellData.ArcaneImpact.ValueAt(mage.Talents.ArcaneImpact),
	// 	Kind:       core.SpellMod_BonusCrit_Percent,
	// })
}

// registerArcaneBlastTalent implements the Forever talent gate for Arcane Blast.
//
// TODO: Forever makes Arcane Blast a talent; the baseline registration in
// arcane_blast.go (registerArcaneBlastSpell, called unconditionally from
// registerSpells) is untouched and should be gated on this field.
func (mage *Mage) registerArcaneBlastTalent() {
	if !mage.Talents.ArcaneBlast {
		return
	}
}

// registerArcaneShielding implements Arcane Shielding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerArcaneShielding() {
	if mage.Talents.ArcaneShielding == 0 {
		return
	}
}

// registerImprovedCounterspell implements Improved Counterspell, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerImprovedCounterspell() {
	if mage.Talents.ImprovedCounterspell == 0 {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneMeditation() {
	if mage.Talents.ArcaneMeditation == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneMeditation == 0 {
	// 	return
	// }
	//
	// mage.PseudoStats.SpiritRegenRateCasting += float64(mage.Talents.ArcaneMeditation) * 0.1
	// mage.UpdateManaRegenRates()
}

// registerMissileBarrage implements Missile Barrage, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (mage *Mage) registerMissileBarrage() {
	if !mage.Talents.MissileBarrage {
		return
	}
}

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneMind() {
	if mage.Talents.ArcaneMind == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneMind == 0 {
	// 	return
	// }
	//
	// mage.MultiplyStat(stats.Intellect, 1+(float64(mage.Talents.ArcaneMind)*.03))
}

// TODO: To be implemented. TBC body below only applies the damage-done half, per the TODO inside about Forever's restated crit aura; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneInstability() {
	if mage.Talents.ArcaneInstability == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if mage.Talents.ArcaneInstability == 0 {
	// 	return
	// }
	//
	// // TODO: Forever restates the crit half as A_MOD_CRIT_PCT (+1% per rank) rather than the
	// // school-crit aura, and whether that covers spells or every attack is unknown, so only
	// // the damage-done portion below is applied.
	//
	// mage.AddStaticMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellsAll,
	// 	FloatValue: spellData.ArcaneInstability.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(mage.Talents.ArcaneInstability),
	// 	Kind:       core.SpellMod_DamageDone_Pct,
	// })
	//
}

// ------ FIRE TALENTS ------
