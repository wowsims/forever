package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRipSpell() {
	rank := spellData.Rip.Highest()
	tick := rank.PeriodicDamageEffect()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ExtraCastCondition = druid.requiresComboPoints(rank)
	config.Dot = spelldata.DotConfig(rank, tick)
	// Combo points are captured when the dot is applied. Attack power is read on the tick, and scales
	// per combo point with the fifth counted as the fourth.
	// TODO: Ingame research needed -- whether Rip snapshots attack power.
	comboPoints := map[int32]float64{}
	config.Dot.OnSnapshot = func(_ *core.Simulation, target *core.Unit, dot *core.Dot) {
		cp := float64(druid.ComboPoints())
		dot.SnapshotBaseDamage = tick.AverageAt(druid.level()) + float64(tick.PointsPerResource)*cp
		comboPoints[target.UnitIndex] = cp
	}
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		apScale := min(comboPoints[target.UnitIndex], 4)
		damage := (dot.SnapshotBaseDamage + tick.APCoeff()*apScale*dot.Spell.MeleeAttackPower(target)) * druid.rendAndTearMultiplier(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			druid.spendComboPoints(sim, spell)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Rip = druid.RegisterSpell(config)
}
