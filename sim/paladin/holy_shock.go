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
// flags and spell power coefficient.
func (paladin *Paladin) holyShockConfig(rank *spelldata.Spell, id int32, row *spelldata.Spell, effect *spelldata.Effect, mask core.ProcMask) core.SpellConfig {
	config := spelldata.SpellConfig(&paladin.Unit, rank, spelldata.Magic(mask))
	config.ActionID = core.ActionID{SpellID: id}
	config.DefenseType = row.DefenseTypeCore()
	config.ClassFlags = row.ClassFlags
	// Magic reads the coefficient off the cast row, which states none.
	config.BonusCoefficient = effect.Coeff()
	return config
}

func (paladin *Paladin) registerHolyShockRank(_ int32, rank *spelldata.Spell) {
	refs := rank.Refs()
	damageRow, healRow := refs[0], refs[1]
	damage, heal := damageRow.DamageEffect(), healRow.HealEffect()

	damageConfig := paladin.holyShockConfig(rank, rank.ID, damageRow, damage, core.ProcMaskSpellDamage)
	damageConfig.ClassSpellMask = SpellMaskHolyShock
	damageConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if vigil := paladin.activeLightsVigil(target); vigil != nil {
			vigil.consume(sim, target)
			spell.CD.Reset()
		}
		spell.CalcAndDealDamage(sim, target, damage.Roll(sim), spell.OutcomeMagicHitAndCrit)
	}
	paladin.RegisterSpell(damageConfig)

	healConfig := paladin.holyShockConfig(rank, healRow.ID, healRow, heal, core.ProcMaskSpellHealing)
	healConfig.Flags |= core.SpellFlagHelpful
	healConfig.ClassSpellMask = SpellMaskHolyShockHeal
	healConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealHealing(sim, target, heal.Roll(sim), spell.OutcomeHealingCrit)
	}
	paladin.RegisterSpell(healConfig)
}
