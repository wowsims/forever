package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerLacerateSpell() {
	if !druid.Talents.Lacerate {
		return
	}
	rank := spellData.Lacerate.Highest()
	tick := rank.PeriodicDamageEffect()
	perStack := rank.EffectN(2).Percent()
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.Dot = spelldata.DotConfig(rank, tick)
	// TODO: Ingame research needed -- the tooltip's "Causes a high amount of threat"; the row states no
	// threat effect, so the cast threatens only for its damage.
	// Each application stacks the bleed, so a tick is the row's amount per stack.
	// The bleed gains no attack power: in game its ticks stayed at the rank's amount a stack with 43 more AP.
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, tick.AverageAt(druid.level())*float64(dot.GetStacks())*druid.rendAndTearMultiplier(target), rank.TickOutcome(dot))
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		dot := spell.Dot(target)
		stacks := int32(1)
		if dot.IsActive() {
			stacks = min(dot.GetStacks()+1, dot.MaxStacks)
		}
		// The tooltip's "$s2% weapon damage per existing application" counts the application this cast adds:
		// in game the hits grow 1:2:3:4:5 with the stacks after the cast, so a first Lacerate hits for $s2%.
		weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		damage := weapon * perStack * float64(stacks) * druid.rendAndTearMultiplier(target)
		result := spell.CalcAndDealDamage(sim, target, damage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		if result.Landed() {
			dot.Apply(sim)
			dot.SetStacks(sim, stacks)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Lacerate = druid.RegisterSpell(config)
}
