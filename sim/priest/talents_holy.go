package priest

import (
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerHolyTalents() {
	// Tier 1
	priest.applyTwilightFocus()
	priest.applyImprovedRenew()
	priest.applyHolySpecialization()

	// Tier 2
	priest.applySpellWarding()
	priest.applyDivineFury()

	// Tier 3
	priest.applyHolyNova()
	priest.applyBlessedRecovery()
	priest.applyInspiration()

	// Tier 4
	priest.applyHolyReach()
	priest.applyImprovedHealing()
	priest.applySearingLight()
	priest.applyBindingHeal()

	// Tier 5
	priest.applyLitanyOfLight()
	priest.applySpiritOfRedemption()
	priest.applySpiritualGuidance()

	// Tier 6
	priest.applySpiritualHealing()

	// Tier 7
	priest.applyPrayerOfMending()
}

// applyTwilightFocus implements Twilight Focus, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyTwilightFocus() {
	if priest.Talents.TwilightFocus == 0 {
		return
	}
}

// applyImprovedRenew implements Improved Renew, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
// TODO: Renew rank 1289450 states 35 per tick with a 0 coefficient and no level requirement; verify in game before porting Renew.
func (priest *Priest) applyImprovedRenew() {
	if priest.Talents.ImprovedRenew == 0 {
		return
	}
}

// applyHolySpecialization implements Holy Specialization, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyHolySpecialization() {
	if priest.Talents.HolySpecialization == 0 {
		return
	}
}

// applySpellWarding implements Spell Warding, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applySpellWarding() {
	if priest.Talents.SpellWarding == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyDivineFury() {
	if priest.Talents.DivineFury == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.DivineFury == 0 {
	// 	return
	// }
	// // -0.1s per rank
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:      core.SpellMod_CastTime_Flat,
	// 	TimeValue: time.Millisecond * time.Duration(-100*priest.Talents.DivineFury),
	// 	ClassMask: PriestSpellSmite | PriestSpellHolyFire,
	// })
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applyHolyNova() {
	if !priest.Talents.HolyNova {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.HolyNova {
	// 	return
	// }
	// HolyNovaRankMap.Each(func(_ int32, rank *spelldata.Spell) { priest.registerHolyNovaSpell(rank) })
}

var HolyNovaRankMap = spellData.HolyNova

// TODO: To be implemented. Holy Nova already has a full Forever rank ladder (spellData.HolyNova); the
// TBC body needs review before it's uncommented.
func (priest *Priest) registerHolyNovaSpell(rank *spelldata.Spell) {
	// The TBC implementation, kept for the port:
	// priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: rank.ID},
	// 	SpellSchool:    core.SpellSchoolHoly,
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellHolyNova,
	// 	ClassFlags:     rank.ClassFlags,
	// 	Rank:           rank.RankNumber(),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(rank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: rank.GCD(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier:         1,
	// 	DamageMultiplierAdditive: 1,
	// 	BonusCoefficient:         rank.DamageEffect().Coeff(),
	// 	ThreatMultiplier:         0,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := rank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	//
	// 		baseHeal := rank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealHealing(sim, spell.Unit, baseHeal, spell.OutcomeHealing)
	// 	},
	// })
}

// applyBlessedRecovery implements Blessed Recovery, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyBlessedRecovery() {
	if priest.Talents.BlessedRecovery == 0 {
		return
	}
}

// applyInspiration implements Inspiration, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyInspiration() {
	if priest.Talents.Inspiration == 0 {
		return
	}
}

// applyHolyReach implements Holy Reach, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyHolyReach() {
	if priest.Talents.HolyReach == 0 {
		return
	}
}

// applyImprovedHealing implements Improved Healing, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyImprovedHealing() {
	if priest.Talents.ImprovedHealing == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applySearingLight() {
	if priest.Talents.SearingLight == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.SearingLight == 0 {
	// 	return
	// }
	// // +5% damage per rank
	// priest.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.SearingLight.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 0).FractionAt(priest.Talents.SearingLight),
	// 	ClassMask:  PriestSpellSmite | PriestSpellHolyFire,
	// })
}

// applyBindingHeal implements Binding Heal, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyBindingHeal() {
	if !priest.Talents.BindingHeal {
		return
	}
}

// applyLitanyOfLight implements Litany of Light, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyLitanyOfLight() {
	if priest.Talents.LitanyOfLight == 0 {
		return
	}
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
// Spirit of Redemption's passive half: +5% total Spirit. The on-death form is not modelled.
func (priest *Priest) applySpiritOfRedemption() {
	if !priest.Talents.SpiritOfRedemption {
		return
	}

	// The TBC implementation, kept for the port:
	// if !priest.Talents.SpiritOfRedemption {
	// 	return
	// }
	// priest.MultiplyStat(stats.Spirit, 1.05)
}

// TODO: To be implemented. The TBC body needs review against Forever's tooltip/values before it's
// brought back.
func (priest *Priest) applySpiritualGuidance() {
	if priest.Talents.SpiritualGuidance == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if priest.Talents.SpiritualGuidance == 0 {
	// 	return
	// }
	// // 5% of Spirit added to spell damage per rank
	// coeff := spellData.SpiritualGuidance.Effect(dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, 126).FractionAt(priest.Talents.SpiritualGuidance)
	// priest.AddStatDependency(stats.Spirit, stats.SpellDamage, coeff) // Only scaling damage for now since no healing sim....yet!
}

// applySpiritualHealing implements Spiritual Healing, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applySpiritualHealing() {
	if priest.Talents.SpiritualHealing == 0 {
		return
	}
}

// applyPrayerOfMending implements Prayer of Mending, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyPrayerOfMending() {
	if !priest.Talents.PrayerOfMending {
		return
	}
}
