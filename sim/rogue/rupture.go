package rogue

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ruptureRank = spellData.Rupture.Highest()
var ruptureApScaling = [6]float64{0, 0.01, 0.02, 0.03, 0.03, 0.03}

func (rogue *Rogue) registerRupture() {
	tick := ruptureRank.PeriodicEffect()
	tickBase := tick.Average()
	tickPerPoint := float64(tick.PointsPerResource)
	baseTicks := int32(ruptureRank.Duration() / tick.Period())
	var comboPoints int32

	config := spelldata.SpellConfig(&rogue.Unit, ruptureRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.ClassSpellMask = RogueSpellRupture

	config.Dot = spelldata.DotConfig(ruptureRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := tickBase + tickPerPoint*float64(comboPoints) + dot.Spell.MeleeAttackPower(target)*ruptureApScaling[comboPoints]
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, ruptureRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			comboPoints = rogue.ComboPoints()
			dot := spell.Dot(target)
			dot.BaseTickCount = baseTicks + comboPoints
			dot.Apply(sim)
			rogue.ApplyFinisher(sim, spell)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		spell.SetMetricsSplit(rogue.ComboPoints())
	}

	rogue.Rupture = rogue.RegisterSpell(config)
}
