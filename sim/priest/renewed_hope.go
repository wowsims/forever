package priest

import "github.com/wowsims/forever/sim/core"

const priestRenewedHopeSpells = PriestSpellFlashHeal |
	PriestSpellLesserHeal |
	PriestSpellHeal |
	PriestSpellGreaterHeal |
	PriestSpellBindingHeal |
	PriestSpellPenance

func (priest *Priest) withRenewedHope(sim *core.Simulation, target *core.Unit, spell *core.Spell, apply func()) {
	weakenedSoul := priest.renewedHopeTargetAura(target, spell)
	priest.withRenewedHopeBolt(target, spell, apply)
	if weakenedSoul != nil {
		priest.reduceWeakenedSoulDuration(sim, weakenedSoul)
	}
}

// For Penance's heal bolts: each takes the crit bonus while Weakened Soul is up, and only the cast shortens it.
func (priest *Priest) withRenewedHopeBolt(target *core.Unit, spell *core.Spell, apply func()) {
	if priest.renewedHopeTargetAura(target, spell) == nil {
		apply()
		return
	}

	bonusCrit := priest.renewedHopeCritBonus(spell)
	spell.BonusCritPercent += bonusCrit
	defer func() { spell.BonusCritPercent -= bonusCrit }()
	apply()
}

func (priest *Priest) renewedHopeCritBonus(spell *core.Spell) float64 {
	if priest.Talents.RenewedHope == 0 || spell.ClassSpellMask&priestRenewedHopeSpells == 0 {
		return 0
	}
	return spellData.RenewedHope.Rank(priest.Talents.RenewedHope).EffectN(1).Percent()
}

func (priest *Priest) renewedHopeTargetAura(target *core.Unit, spell *core.Spell) *core.Aura {
	if priest.Talents.RenewedHope == 0 || target == nil || target.Type == core.EnemyUnit ||
		spell.ClassSpellMask&priestRenewedHopeSpells == 0 {
		return nil
	}

	aura := priest.weakenedSoulAuras.Get(target)
	if !aura.IsActive() {
		return nil
	}
	return aura
}

func (priest *Priest) reduceWeakenedSoulDuration(sim *core.Simulation, aura *core.Aura) {
	reduction := -spellData.RenewedHope.Rank(priest.Talents.RenewedHope).EffectN(2).TimeValue()
	remaining := aura.RemainingDuration(sim)
	if remaining <= reduction {
		aura.Deactivate(sim)
		return
	}

	baseDuration := aura.Duration
	aura.Duration = remaining - reduction
	aura.Refresh(sim)
	aura.Duration = baseDuration
}
