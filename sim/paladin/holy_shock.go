package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyShockRankMap = spellData.HolyShock

// Holy Shock (talent)
// https://www.wowhead.com/forever/spell=20930
//
// Blasts the target with Holy energy, causing 334 to 362 Holy damage to an enemy, or 307 to 333
// healing to an ally.
func (paladin *Paladin) registerHolyShock() {
	HolyShockRankMap.Each(paladin.registerHolyShockRank)
}

// One of the rank's two spells off the cast row, with the effect row's id, defense type, class
// flags, spell power coefficient and roll.
func (paladin *Paladin) holyShockConfig(rank *spelldata.Spell, id int32, row *spelldata.Spell, effect *spelldata.Effect, mask core.ProcMask) core.SpellConfig {
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(mask))
	config.ActionID = core.ActionID{SpellID: id}
	config.DefenseType = row.DefenseTypeCore()
	config.ClassFlags = row.ClassFlags
	// Magic reads the coefficient and the roll off the cast row, which states neither.
	config.BonusCoefficient = effect.Coeff()
	config.DirectVariance = effect.DirectVariance()
	return config
}

func (paladin *Paladin) registerHolyShockRank(_ int32, rank *spelldata.Spell) {
	refs := rank.Refs()
	damageRow, healRow := refs[0], refs[1]
	damage, heal := damageRow.DamageEffect(), healRow.HealEffect()

	damageConfig := paladin.holyShockConfig(rank, damageRow.ID, damageRow, damage, core.ProcMaskSpellDamage)
	damageConfig.ClassSpellMask = SpellMaskHolyShock
	damageConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}
	paladin.RegisterSpell(damageConfig)

	healConfig := paladin.holyShockConfig(rank, healRow.ID, healRow, heal, core.ProcMaskSpellHealing)
	healConfig.Flags |= core.SpellFlagHelpful
	healConfig.ClassSpellMask = SpellMaskHolyShockHeal
	// In game the heal makes 0.25 threat a point healed, as Holy Light does. The damage makes full threat.
	healConfig.ThreatMultiplier = 0.5
	// Boundless Light: more crit on an ally without the paladin's Oath of Light.
	boundlessLight := 0.0
	if paladin.Talents.BoundlessLight > 0 {
		boundlessLight = spellData.BoundlessLight.ValueAt(paladin.Talents.BoundlessLight)
	}
	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if paladin.consumeLightsVigil(sim, target) {
			spell.CD.Reset()
		}
		amount := heal.Average()
		crit := boundlessLight
		if paladin.hasOathOfLight(target) {
			amount += paladin.oathOfLight.holyShock
			crit = 0
		}
		spell.BonusCritPercent += crit
		spell.CalcAndDealHealing(sim, target, amount, spell.OutcomeHealingCrit)
		spell.BonusCritPercent -= crit
	}
	paladin.RegisterSpell(healConfig)
}
