package druid

var ferociousBiteRank = spellData.FerociousBite.Highest()
var ferociousBiteBase = ferociousBiteRank.DamageEffect().Average()

// TODO: To be implemented.
func (druid *Druid) registerFerociousBiteSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// var energyMetrics *core.ResourceMetrics
	//
	// druid.FerociousBite = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: ferociousBiteRank.ID},
	// 	CastRequirement: ferociousBiteRank.CastRequirement(),
	// 	SpellSchool:     ferociousBiteRank.SpellSchool(),
	// 	DefenseType:     ferociousBiteRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask:  DruidSpellFerociousBite,
	// 	Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: int32(ferociousBiteRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: ferociousBiteRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	// 	ExtraCastCondition: func(_ *core.Simulation, _ *core.Unit) bool {
	// 		return druid.ComboPoints() > 0
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		cp := float64(druid.ComboPoints())
	// 		ap := spell.MeleeAttackPower(target)
	// 		excessEnergy := druid.CurrentEnergy()
	// 		if excessEnergy > 0 {
	// 			druid.SpendEnergy(sim, excessEnergy, energyMetrics)
	// 			energyMetrics.Events--
	// 		}
	//
	// 		dmgPerCP := 169.0
	// 		baseDamage := ferociousBiteBase + dmgPerCP*cp + 4.1*excessEnergy + 0.05*cp*ap
	//
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			druid.SpendComboPoints(sim, spell.ComboPointMetrics())
	// 		}
	// 	},
	//
	// 	ExpectedInitialDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
	// 		cp := float64(druid.ComboPoints())
	// 		ap := spell.MeleeAttackPower(target)
	// 		dmgPerCP := 169.0
	// 		baseDamage := ferociousBiteBase + dmgPerCP*cp + 0.05*cp*ap
	// 		return spell.CalcDamage(sim, target, baseDamage, spell.OutcomeExpectedMeleeWeaponSpecialHitAndCrit)
	// 	},
	// })
	//
	// energyMetrics = druid.FerociousBite.Cost.ResourceCostImpl.(*core.EnergyCost).ResourceMetrics
}

func (druid *Druid) CurrentFerociousBiteCost() float64 {
	return druid.FerociousBite.Cost.GetCurrentCost()
}
