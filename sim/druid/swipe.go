package druid

var swipeRank = spellData.Swipe.Highest()

// TODO: To be implemented.
// TODO: Swipe gains 3% of the druid's attack power in game; its rows do not state it.
func (druid *Druid) registerSwipeBearSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.Swipe = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: swipeRank.ID},
	// 	CastRequirement: swipeRank.CastRequirement(),
	// 	SpellSchool:     swipeRank.SpellSchool(),
	// 	DefenseType:     swipeRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask:  DruidSpellSwipe,
	// 	Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	RageCost: core.RageCostOptions{
	// 		Cost:   int32(swipeRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: swipeRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		numHits := min(3, len(druid.Env.Encounter.AllTargetUnits))
	// 		for i := 0; i < numHits; i++ {
	// 			aoeTarget := druid.Env.Encounter.AllTargetUnits[i]
	// 			baseDamage := swipeRank.DamageEffect().Average(core.CharacterLevel) + 0.07*spell.MeleeAttackPower(aoeTarget)
	// 			spell.CalcAndDealDamage(sim, aoeTarget, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	// 		}
	// 	},
	// })
}
