package rogue

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var exposeArmorRank = spellData.ExposeArmor.Highest()

func (rogue *Rogue) registerExposeArmorSpell() {
	rogue.ExposeArmorAuras = rogue.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// The generated aura is the five-point finisher, which is what the raid
		// config applies; a cast that spends fewer combo points needs a driver
		// that prices the aura from rogue.ComboPoints, and there is none yet.
		return buffs.NewSpellDataExposeArmorAura(target, true, 0)
	})

	config := spelldata.SpellConfig(&rogue.Unit, exposeArmorRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if rogue.CanApplyExposeArmorAura(target) {
			rogue.BreakStealth(sim)
			result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
			if result.Landed() {
				rogue.ExposeArmorAuras.Get(target).Activate(sim)

				if rogue.Talents.ImprovedExposeArmor > 0 && rogue.ComboPoints() >= 5 {
					pointsRefunded := int32(spellData.ImprovedExposeArmor.Rank(rogue.Talents.ImprovedExposeArmor).EffectN(2).BaseValue())
					rogue.ApplyFinisher(sim, spell)
					rogue.AddComboPoints(sim, pointsRefunded, spell.ComboPointMetrics())
				} else {
					rogue.ApplyFinisher(sim, spell)
				}
			} else {
				spell.IssueRefund(sim)
			}
			spell.DealOutcome(sim, result)
		}
	}

	config.MetricSplits = 6
	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		spell.SetMetricsSplit(rogue.ComboPoints())
	}

	rogue.ExposeArmor = rogue.RegisterSpell(config)
}

func (rogue *Rogue) GetExposeArmorValue() float64 {
	return math.Abs((float64(exposeArmorRank.EffectN(1).PointsPerResource) * float64(rogue.ComboPoints())))
}

func (rogue *Rogue) CanApplyExposeArmorAura(target *core.Unit) bool {
	return !rogue.ExposeArmorAuras.Get(target).IsActive() || rogue.ExposeArmorAuras.Get(target).ExclusiveEffects[0].Priority <= rogue.GetExposeArmorValue()
}
