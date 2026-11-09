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
	// Attack power is read at each tick. Rend and Tear does not raise the ticks, though the bleed counts
	// toward it for the other attacks.
	bleedConfig.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		damage := bleedTick.AverageAt(druid.level()) + bleedTick.APCoeff()*dot.Spell.MeleeAttackPower(target)
		dot.Spell.CalcAndDealPeriodicDamage(sim, target, damage, bleedRank.TickOutcome(dot))
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
	// No Active Defense (Attributes 0x200000): the cast misses or lands, never dodged, parried or
	// blocked. A landed cast awards the combo point and stuns, and then the bleed rolls its own miss and
	// dodge. An avoided bleed refunds no energy.
	// TODO: Ingame research needed -- whether the cast itself can miss.
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			druid.awardComboPoints(sim, spell, rank)
			stuns.Get(target).Activate(sim)
			bleed := druid.PounceBleed.CalcAndDealOutcome(sim, target, druid.PounceBleed.OutcomeMeleeSpecialHit)
			if bleed.Landed() {
				druid.PounceBleed.Dot(target).Apply(sim)
			}
		}
		druid.refundOnMiss(sim, spell, result)
	}
	druid.Pounce = druid.RegisterSpell(config)
}
