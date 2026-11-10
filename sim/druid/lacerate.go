package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The weapon hit every Lacerate rank casts by script; the store carries it as an extra spell.
const lacerateHitSpellID = 414647

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
	// The hit is its own row. Natural Instinct's mask covers its class flags as well as the rank's, so
	// the hit's crits and the ticks' crits both take the talent's crit damage. Its weapon percent is not
	// what lands; the rank's $s2 is.
	hitConfig := spelldata.SpellConfig(&druid.Unit, spelldata.MustFind(lacerateHitSpellID),
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	hit := druid.RegisterSpell(hitConfig)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The rank's cast is what misses, is dodged, parried or immune: every SPELL_MISSED in the log
		// names the rank, never 414647. The hit follows a landed cast and rolls block and crit, and its
		// row's No Active Defense takes block out, so it only rolls crit.
		result := spell.CalcOutcome(sim, target, spell.OutcomeMeleeSpecialHitNoHitCounter)
		if !result.Landed() {
			spell.DealOutcome(sim, result)
			druid.refundOnMiss(sim, spell, result)
			return
		}
		dot := spell.Dot(target)
		stacks := int32(1)
		if dot.IsActive() {
			stacks = min(dot.GetStacks()+1, dot.MaxStacks)
		}
		// The tooltip's "$s2% weapon damage per existing application" counts the application this cast adds:
		// in game the hits grow 1:2:3:4:5 with the stacks after the cast, so a first Lacerate hits for $s2%.
		weapon := druid.MHWeaponDamage(sim, hit.MeleeAttackPower(target))
		damage := weapon * perStack * float64(stacks) * druid.rendAndTearMultiplier(target)
		hit.CalcAndDealDamage(sim, target, damage, hit.OutcomeMeleeSpecialBlockAndCrit)
		dot.Apply(sim)
		dot.SetStacks(sim, stacks)
	}
	druid.Lacerate = druid.RegisterSpell(config)
}
