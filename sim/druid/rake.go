package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerRakeSpell() {
	rank := spellData.Rake.Highest()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// Initial hit armor in EV-0038. The ticks are periodic physical damage,
	// which core keeps out of armor anyway.
	config.Flags |= core.SpellFlagIgnoreResists
	hit := rank.DamageEffect()
	tick := rank.PeriodicDamageEffect()
	config.Dot = spelldata.DotConfig(rank, tick)
	// Each tick reads the attack power when it ticks, not the cast's.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := (tick.AverageAt(druid.level()) + tick.APCoeff()*dot.Spell.MeleeAttackPower(target)) * druid.rendAndTearMultiplier(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		initial := (hit.AverageAt(druid.level()) + hit.APCoeff()*spell.MeleeAttackPower(target)) * druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, initial, spell.OutcomeMeleeSpecialHitAndCrit)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, target, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Rake = druid.RegisterSpell(config)
}
