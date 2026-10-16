package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerLacerateSpell() {
	rank := spellData.Lacerate.Highest()
	tick := rank.PeriodicDamageEffect()
	perStack := rank.EffectN(2).Percent()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Dot = spelldata.DotConfig(rank, tick)
	// TODO: Ingame research needed -- the tooltip's "Causes a high amount of threat"; the row states no
	// threat effect, so the cast threatens only for its damage.
	// Each application stacks the bleed, so a tick is the row's amount per stack.
	// TODO: Ingame research needed -- whether the hit or the bleed gains attack power; the row states none.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.AverageAt(druid.level())*float64(dot.GetStacks())*druid.rendAndTearMultiplier(target), rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dot := spell.Dot(target)
		existing := int32(0)
		if dot.IsActive() {
			existing = dot.GetStacks()
		}
		stacks := existing
		if maxStacks := dot.MaxStacks; maxStacks == 0 || stacks < maxStacks {
			stacks++
		}
		// The tooltip's "$s2% weapon damage per existing application": a first Lacerate hits for nothing.
		weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		damage := weapon * perStack * float64(existing) * druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			dot.Apply(sim)
			dot.SetStacks(sim, stacks)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Lacerate = druid.RegisterSpell(config)
}
