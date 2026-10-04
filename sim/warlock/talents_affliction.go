package warlock

func (warlock *Warlock) registerAfflictionTalents() {
	// Tier 1
	// Improved Life Tap implemented in lifetap.go
	warlock.applySuppression()
	warlock.applyImprovedCorruption()
	// Empowered Corruption rides on Improved Corruption's guard; see the TODO below.
	warlock.applyEmpoweredCorruption()

	// Tier 2
	// Malediction implemented in curse_of_elements.go
	warlock.applySoulHarvest()
	warlock.applyImprovedDrains()

	// Tier 3
	warlock.applyImprovedBaneOfAgony()
	warlock.applyFelConcentration()
	warlock.registerAmplifyCurse()
	warlock.applyPandemic()

	// Tier 4
	warlock.applyMalevolence()
	warlock.applyNightfall()
	warlock.applyCurseOfExhaustion()

	// Tier 5
	warlock.applySiphonLife()
	// Soul Siphon implemented in drain_life.go

	// Tier 6
	warlock.applyShadowMastery()

	// Tier 7
	warlock.applyWrack()
}

// TODO: To be implemented. Port the TBC Suppression implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) applySuppression() {
	if warlock.Talents.Suppression == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.Suppression == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_BonusHit_Percent,
	// 	FloatValue: spellData.Suppression.Effect(dbcenums.A_MOD_SPELL_HIT_CHANCE, 0).ValueAt(warlock.Talents.Suppression),
	// 	ClassMask:  WarlockAfflictionSpells,
	// })
}

// TODO: To be implemented. Port the TBC Improved Corruption implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) applyImprovedCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ImprovedCorruption == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// 	TimeValue: time.Millisecond * (-400 * time.Duration(warlock.Talents.ImprovedCorruption)),
	// 	ClassMask: WarlockSpellCorruption,
	// })
}

// TODO: To be implemented. Port the TBC Empowered Corruption implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) applyEmpoweredCorruption() {
	if warlock.Talents.ImprovedCorruption == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ImprovedCorruption == 0 {
	// 	return
	// }
	//
	// // TODO: this read Talents.EmpoweredCorruption, a talent Forever drops entirely, while
	// // the registrar's guard is ImprovedCorruption (kept) -- looks like a long-standing
	// // copy-paste bug. EmpoweredCorruption's spellData ladder is unrelated to
	// // ImprovedCorruption's rank count, so the bonus coefficient is unknown and pinned to 0
	// // until confirmed against the Forever tooltip.
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DotBonusCoeffecient_Flat,
	// 	FloatValue: 0,
	// 	ClassMask:  WarlockSpellCorruption,
	// })
}

// TODO: To be implemented. Port the TBC Nightfall implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) applyNightfall() {
	if warlock.Talents.Nightfall == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.Nightfall == 0 {
	// 	return
	// }
	//
	// warlock.NightfallProcAura = warlock.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:            "Shadow Trance",
	// 	MetricsActionID: core.ActionID{SpellID: 17941},
	// 	ClassSpellMask:  WarlockSpellShadowBolt,
	// 	Callback:        core.CallbackOnCastComplete,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if spell.CurCast.CastTime != 0 {
	// 			return
	// 		}
	// 		warlock.NightfallProcAura.Deactivate(sim)
	// 	},
	// }).AttachSpellMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_CastTime_Pct,
	// 	FloatValue: -1.0,
	// 	ClassMask:  WarlockSpellShadowBolt,
	// })
	//
	// warlock.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:           "Nightfall",
	// 	ClassSpellMask: WarlockSpellCorruption | WarlockSpellDrainLife,
	// 	// Forever puts the real per-rank chance on the effect; Rank(r).ProcChance reads a flat 100.
	// 	ProcChance: spellData.Nightfall.FractionAt(warlock.Talents.Nightfall),
	// 	Callback:   core.CallbackOnPeriodicDamageDealt,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		warlock.NightfallProcAura.Activate(sim)
	// 	},
	// })
}

// TODO: To be implemented. Port the TBC Shadow Mastery implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) applyShadowMastery() {
	if warlock.Talents.ShadowMastery == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ShadowMastery == 0 {
	// 	return
	// }
	//
	// warlock.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.ShadowMastery.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(warlock.Talents.ShadowMastery),
	// 	ClassMask:  WarlockShadowDamage,
	// })
}

// TODO: To be implemented. Port the TBC Amplify Curse implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerAmplifyCurse() {
	if !warlock.Talents.AmplifyCurse {
		return
	}

	// The TBC implementation, kept for the port:
	// if !warlock.Talents.AmplifyCurse {
	// 	return
	// }
	//
	// actionID := core.ActionID{SpellID: 18288}
	//
	// warlock.AmplifyCurseAura = warlock.GetOrRegisterAura(core.Aura{
	// 	Label:    "Amplify Curse",
	// 	Tag:      "Affliction",
	// 	ActionID: actionID,
	// 	Duration: time.Second * 30,
	// 	OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
	// 		if spell.Matches(WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom) {
	// 			warlock.AmplifyCurseAura.Deactivate(sim)
	// 		}
	// 	},
	// })
	//
	// warlock.AmplifyCurse = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID: actionID,
	// 	Flags:    core.SpellFlagAPL,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: time.Second,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    warlock.NewTimer(),
	// 			Duration: time.Minute * 3,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		warlock.AmplifyCurseAura.Activate(sim)
	// 	},
	// })
	//
	// warlock.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: warlock.AmplifyCurse,
	// 	Type:  core.CooldownTypeDPS,
	// })
}

// applySoulHarvest implements Soul Harvest, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applySoulHarvest() {
	if warlock.Talents.SoulHarvest == 0 {
		return
	}
}

// applyImprovedDrains implements Improved Drains, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyImprovedDrains() {
	if warlock.Talents.ImprovedDrains == 0 {
		return
	}
}

// applyImprovedBaneOfAgony implements Improved Bane of Agony, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyImprovedBaneOfAgony() {
	if warlock.Talents.ImprovedBaneOfAgony == 0 {
		return
	}
}

// applyFelConcentration implements Fel Concentration, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyFelConcentration() {
	if warlock.Talents.FelConcentration == 0 {
		return
	}
}

// applyPandemic implements Pandemic, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyPandemic() {
	if warlock.Talents.Pandemic == 0 {
		return
	}
}

// applyMalevolence implements Malevolence, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyMalevolence() {
	if warlock.Talents.Malevolence == 0 {
		return
	}
}

// applyCurseOfExhaustion implements Curse of Exhaustion, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyCurseOfExhaustion() {
	if !warlock.Talents.CurseOfExhaustion {
		return
	}
}

// applySiphonLife implements Siphon Life, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applySiphonLife() {
	if !warlock.Talents.SiphonLife {
		return
	}
}

// applyWrack implements Wrack, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyWrack() {
	if !warlock.Talents.Wrack {
		return
	}
}
