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
	comboPoints := map[int32]int32{}

	config := spelldata.SpellConfig(&rogue.Unit, ruptureRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.ClassSpellMask = RogueSpellRupture

	config.Dot = spelldata.DotConfig(ruptureRank, tick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		points := comboPoints[target.UnitIndex]
		damage := tickBase + tickPerPoint*float64(points) + dot.Spell.MeleeAttackPower(target)*ruptureApScaling[points]
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, ruptureRank.TickOutcome(dot))
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim)
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if result.Landed() {
			comboPoints[target.UnitIndex] = rogue.ComboPoints(target)
			dot := spell.Dot(target)
			dot.BaseTickCount = baseTicks + comboPoints[target.UnitIndex]
			dot.Apply(sim)
			rogue.ApplyFinisher(sim, target, spell)
		} else {
			spell.IssueRefund(sim)
		}
		spell.DealOutcome(sim, result)
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = rogue.setFinisherMetricsSplit

	rogue.Rupture = rogue.RegisterSpell(config)
}
