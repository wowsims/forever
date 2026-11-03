package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Rank 5 is the last rank the trainer teaches; rank 6 comes from the Book of Ferocious Bite V.
var ferociousBiteTrainerRank = spellData.FerociousBite.ByID(22829)
var ferociousBiteBookRank = spellData.FerociousBite.ByID(31018)

func (druid *Druid) registerFerociousBiteSpell() {
	druid.FerociousBite = druid.registerFerociousBiteRank(ferociousBiteTrainerRank)
	druid.registerFerociousBiteRank(ferociousBiteBookRank)
}

func (druid *Druid) registerFerociousBiteRank(rank *spelldata.Spell) *DruidSpell {
	damageEffect := rank.DamageEffect()
	perEnergy := rank.EffectN(2).BaseValue() / 100
	// Predatory Discipline keeps the extra energy and still adds its damage.
	// TODO: Ingame research needed -- what the talent's dummy (base points 20) does; the tooltip names no value.
	keepsEnergy := druid.Talents.PredatoryDiscipline
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = druid.requiresComboPoints(rank)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		cp := float64(druid.ComboPoints())
		extra := druid.CurrentEnergy()
		damage := damageEffect.RollAt(sim, druid.level()) + float64(damageEffect.PointsPerResource)*cp
		damage += extra * perEnergy
		damage += damageEffect.APCoeff() * cp * spell.MeleeAttackPower(target)
		damage *= druid.rendAndTearMultiplier(target)
		// The extra energy and the combo points are spent on a landed hit only, so a failed bite costs its base energy.
		// TODO: Ingame research needed -- the row flags Discount Power On Miss, which would give a failed bite back 80%
		// of its base cost; nothing refunds it until that is checked.
		result := spell.CalcDamage(sim, target, damage, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			if extra > 0 && !keepsEnergy {
				druid.SpendEnergy(sim, extra, spell.EnergyMetrics())
			}
			druid.spendComboPoints(sim, spell)
		}
		spell.DealDamage(sim, result)
	}
	return druid.RegisterSpell(config)
}
