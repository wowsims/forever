package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var eviscerateRank = spellData.Eviscerate.Rank(8)

func (rogue *Rogue) registerEviscerate() {
	eviscerateBaseDamage := eviscerateRank.DamageEffect().Average()
	eviscerateDamagePerPoint := float64(eviscerateRank.DamageEffect().PointsPerResource)
	eviscerateApScalingPerPoint := 0.04 // Estimated from in-game tests.
	config := spelldata.SpellConfig(&rogue.Unit, eviscerateRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.Flags |= SpellFlagFinisher | SpellFlagColdBlooded

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		comboPoints := float64(rogue.ComboPoints())
		flatBaseDamage := eviscerateBaseDamage + eviscerateDamagePerPoint*comboPoints

		baseDamage := flatBaseDamage + eviscerateApScalingPerPoint*comboPoints*spell.MeleeAttackPower(target)

		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if result.Landed() {
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}

		spell.DealDamage(sim, result)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		spell.SetMetricsSplit(rogue.ComboPoints())
	}

	rogue.Eviscerate = rogue.RegisterSpell(config)
}
