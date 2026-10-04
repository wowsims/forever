package druid

// Package-level state the commented-out implementation used:
// var mangleBearRank = spellData.PrimalBite.ByID(1238073)

// TODO: To be implemented.
func (druid *Druid) registerMangleAuras() {
	if druid.MangleAuras != nil {
		return
	}
	panic("To be implemented")

	// The TBC implementation, kept for the port. buffs.MangleAura named below does not
	// exist: the game files' Mangle ranks deal damage and apply no aura.
	// if druid.MangleAuras != nil {
	// 	return
	// }
	// druid.MangleAuras = druid.NewEnemyAuraArray(buffs.MangleAura)
}

// TODO: To be implemented. Forever ships ONE Primal Bite -- spells 407995 and 1238069/1238070/1238073
// on the Feral Combat line, all with ShapeshiftMask [144,0], which is Bear and Dire Bear only. It
// applies no debuff.
// TODO: Primal Bite generates roughly twice the threat its rows state; measure the threat multiplier in game.
func (druid *Druid) registerPrimalBiteSpell() {
	if !druid.Talents.PrimalBite {
		return
	}
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// if !druid.Talents.Mangle {
	// 	return
	// }
	//
	// druid.registerMangleAuras()
	//
	// druid.MangleBear = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        core.ActionID{SpellID: mangleBearRank.ID},
	// 	CastRequirement: mangleBearRank.CastRequirement(),
	// 	SpellSchool:     mangleBearRank.SpellSchool(),
	// 	DefenseType:     mangleBearRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskMeleeMHSpecial,
	// 	ClassSpellMask:  DruidSpellMangleBear,
	// 	Flags:           core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
	//
	// 	RageCost: core.RageCostOptions{
	// 		Cost:   int32(mangleBearRank.Cost()),
	// 		Refund: 0.8,
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: mangleBearRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: max(mangleBearRank.Cooldown(), mangleBearRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1.15,
	// 	ThreatMultiplier: 1.5 / 1.15,
	// 	MaxRange:         core.MaxMeleeRange,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := (mangleBearRank.DamageEffect().Average(core.CharacterLevel)*1.15+druid.IdolMangleBearBonus)/1.15 + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
	//
	// 		if result.Landed() {
	// 			druid.MangleAuras.Get(target).Activate(sim)
	// 		} else {
	// 			spell.IssueRefund(sim)
	// 		}
	// 	},
	// })
}
