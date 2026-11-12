package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerPounceSpell() {
	rank := spellData.Pounce.Highest()
	bleedRank := spellData.PounceBleed.Highest()

	// Natural Instinct's mask covers the bleed 9826's class flags, so its crit ticks take the talent's crit
	// damage.
	// Genesis names Pounce 9827 by its class flags, and the bleed 9826's flags lie outside Genesis' mask, so
	// the bleed takes no Genesis.
	bleedConfig := spelldata.SpellConfig(&druid.Unit, bleedRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
	bleedTick := bleedRank.PeriodicDamageEffect()
	bleedConfig.Dot = spelldata.DotConfig(bleedRank, bleedTick)
	// Attack power is read at each tick, and Rend and Tear raises the ticks; the bleed itself counts as the
	// bleed on the target.
	bleedConfig.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := (bleedTick.AverageAt(druid.level()) + bleedTick.APCoeff()*dot.Spell.MeleeAttackPower(target)) * druid.rendAndTearMultiplier(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, bleedRank.TickOutcome(dot))
	}
	druid.PounceBleed = druid.RegisterSpell(bleedConfig)

	stuns := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" Stun - "+druid.Label)))
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// CastRequirement does not enforce ATTR_REQ_STEALTH. These attacks require Prowl;
	// Shadowmeld does not meet that requirement.
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ProwlAura.IsActive() && druid.BehindTarget()
	}
	// No Active Defense (Attributes 0x200000): the cast is never dodged, parried or blocked, but it can
	// miss. A missed cast stuns nothing, awards no combo point, applies no bleed and refunds the row's
	// share of its cost. A landed cast always stuns and awards the combo point; the bleed then rolls its
	// own miss and dodge, and an avoided bleed refunds no energy.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if !result.Landed() {
			druid.refundOnMiss(sim, spell, result)
			return
		}
		stuns.Get(target).Activate(sim)
		druid.awardComboPoints(sim, target, spell, rank)
		bleed := druid.PounceBleed.CalcOutcome(sim, target, druid.PounceBleed.OutcomeMeleeSpecialHitNoHitCounter)
		if bleed.Landed() {
			druid.PounceBleed.Dot(target).Apply(sim)
		}
	}
	druid.Pounce = druid.RegisterSpell(config)
}
