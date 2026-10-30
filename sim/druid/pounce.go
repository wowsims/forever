package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPounceSpell() {
	rank := spellData.Pounce.Highest()
	bleedRank := spellData.PounceBleed.Highest()

	// TODO: Ingame research needed -- Genesis names Pounce 9827 by its class flags, and the bleed 9826
	// states none, so the bleed here takes no class-flag modifier.
	bleedConfig := spelldata.SpellConfig(&druid.Unit, bleedRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	bleedTick := bleedRank.PeriodicDamageEffect()
	bleedConfig.Dot = spelldata.DotConfig(bleedRank, bleedTick)
	bleedConfig.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, bleedTick.Average()*druid.rendAndTearMultiplier(target), bleedRank.TickOutcome(dot))
	}
	druid.PounceBleed = druid.RegisterSpell(bleedConfig)

	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" Stun")))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// CastRequirement does not enforce ATTR_REQ_STEALTH. These attacks require Prowl;
	// Shadowmeld does not meet that requirement.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ProwlAura.IsActive() && druid.BehindTarget()
	}
	// No Active Defense (Attributes 0x200000): it misses or lands, never dodged, parried or blocked.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			stuns.Get(target).Activate(sim)
			druid.PounceBleed.Dot(target).Apply(sim)
			druid.awardComboPoints(sim, spell, rank)
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Pounce = druid.RegisterSpell(config)
}
